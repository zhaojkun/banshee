// Copyright 2015 Eleme Inc. All rights reserved.

package metricdb

import (
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util"
	"os"
	"path"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestOpenOptionsNil(t *testing.T) {
	fileName := "db-testing"
	db, err := Open(fileName, nil)
	util.Must(t, err == nil)
	util.Must(t, util.IsFileExist(fileName))
	defer os.RemoveAll(fileName)
	defer db.Close()
	util.Must(t, len(db.pool) == 0) // should have nothing in pool..
}

func TestOpenInit(t *testing.T) {
	fileName := "db-testing"
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	db, err := Open(fileName, opts)
	util.Must(t, err == nil)
	defer os.RemoveAll(fileName)
	stamp := uint32(time.Now().Unix())
	db.Put(&models.Metric{Stamp: stamp, Link: 1})
	db.Close()
	// Reopen.
	db, err = Open(fileName, opts)
	util.Must(t, err == nil)
	util.Must(t, len(db.pool) == 1) // should have one storage in pool
	util.Must(t, db.pool[0].id*db.opts.Period <= stamp)
}

func TestPut(t *testing.T) {
	// Open db.
	fileName := "db-testing"
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	db, _ := Open(fileName, opts)
	defer os.RemoveAll(fileName)
	defer db.Close()
	// Put.
	m := &models.Metric{
		Name:    "foo",
		Link:    1,
		Stamp:   1452758773,
		Value:   3.14,
		Score:   0.1892,
		Average: 3.133,
	}
	err := db.Put(m)
	util.Must(t, err == nil)
	// Must in db
	key := encodeKey(m)
	value, err := db.pool[0].db.Get(key, nil)
	util.Must(t, err == nil)
	m1 := &models.Metric{
		Name:  m.Name,
		Stamp: m.Stamp,
		Link:  1,
	}
	err = decodeValue(value, m1)
	util.Must(t, err == nil)
	util.Must(t, reflect.DeepEqual(m, m1))
}

func TestGet(t *testing.T) {
	// Open db.
	fileName := "db-testing"
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	db, _ := Open(fileName, opts)
	defer os.RemoveAll(fileName)
	defer db.Close()
	// Nothing.
	ms, err := db.Get("not-exist", 1234, 0, 1452758773)
	util.Must(t, err == nil)
	util.Must(t, len(ms) == 0)
	// Put some.
	db.Put(&models.Metric{Name: "foo", Link: 1, Stamp: 1452758723})
	db.Put(&models.Metric{Name: "foo", Link: 1, Stamp: 1452758733, Value: 1.89, Score: 1.12, Average: 1.72})
	db.Put(&models.Metric{Name: "foo", Link: 1, Stamp: 1452758743})
	db.Put(&models.Metric{Name: "foo", Link: 1, Stamp: 1452758753})
	// Get again.
	ms, err = db.Get("foo", 1, 1452758733, 1452758753)
	util.Must(t, err == nil)
	util.Must(t, len(ms) == 2)
	// Test the value.
	m := ms[0]
	util.Must(t, m.Value == 1.89 && m.Score == 1.12 && m.Link == 1)
}

func TestGetAcrossStorages(t *testing.T) {
	// Open db.
	fileName := "db-testing"
	opts := &Options{Period: 86400, Expiration: 86400 * 4}
	db, _ := Open(fileName, opts)
	defer os.RemoveAll(fileName)
	defer db.Close()
	// Force creating 4+1 storages.
	base := uint32(time.Now().Unix())
	db.Put(&models.Metric{Link: 1, Stamp: base})                    // 0
	db.Put(&models.Metric{Link: 1, Stamp: base + db.opts.Period*1}) // 1
	db.Put(&models.Metric{Link: 1, Stamp: base + db.opts.Period*2}) // 2
	db.Put(&models.Metric{Link: 1, Stamp: base + db.opts.Period*3}) // 3
	db.Put(&models.Metric{Link: 1, Stamp: base + db.opts.Period*4}) // 4
	// Get
	ms, err := db.Get("whatever", 1, base, base+db.opts.Period*3)
	util.Must(t, err == nil)
	util.Must(t, len(ms) == 3)
	util.Must(t, ms[0].Stamp == base)
	util.Must(t, ms[1].Stamp == base+db.opts.Period*1)
	util.Must(t, ms[2].Stamp == base+db.opts.Period*2)
}

func TestStorageExpire(t *testing.T) {
	// Open db.
	fileName := "db-testing"
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	db, _ := Open(fileName, opts)
	defer os.RemoveAll(fileName)
	defer db.Close()
	// Force creating 7+1 storages.
	base := uint32(time.Now().Unix())
	db.Put(&models.Metric{Link: 1, Stamp: base})                    // 0
	id := db.pool[0].id                                             // record the id to be deleted
	db.Put(&models.Metric{Link: 1, Stamp: base + db.opts.Period*1}) // 1
	db.Put(&models.Metric{Link: 1, Stamp: base + db.opts.Period*2}) // 2
	db.Put(&models.Metric{Link: 1, Stamp: base + db.opts.Period*3}) // 3
	db.Put(&models.Metric{Link: 1, Stamp: base + db.opts.Period*4}) // 4
	db.Put(&models.Metric{Link: 1, Stamp: base + db.opts.Period*5}) // 5
	db.Put(&models.Metric{Link: 1, Stamp: base + db.opts.Period*6}) // 6
	util.Must(t, len(db.pool) == 7)
	db.Put(&models.Metric{Link: 1, Stamp: base + db.opts.Period*7}) // 7
	util.Must(t, len(db.pool) == 8)
	db.Put(&models.Metric{Link: 1, Stamp: base + db.opts.Period*8}) // 8
	util.Must(t, len(db.pool) == 8)                                 // Full storages: 1,2,3,4,5,6,7
	// Files must be deleted.
	deleteFileName := path.Join(fileName, strconv.FormatUint(uint64(id), 10))
	util.Must(t, !util.IsFileExist(deleteFileName))
}

func BenchmarkPut(b *testing.B) {
	// Open db.
	fileName := "db-testing"
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	db, _ := Open(fileName, opts)
	defer os.RemoveAll(fileName)
	defer db.Close()
	base := uint32(time.Now().Unix())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.Put(&models.Metric{Link: 1, Stamp: base + uint32(i)*7, Value: float64(i)})
	}
}

func BenchmarkPutX10(b *testing.B) {
	// Open db.
	fileName := "db-testing"
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	db, _ := Open(fileName, opts)
	defer os.RemoveAll(fileName)
	defer db.Close()
	base := uint32(time.Now().Unix())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 10; j++ {
			db.Put(&models.Metric{Link: uint32(j), Stamp: base + uint32(i)*7, Value: float64(i)})
		}
	}
}

func BenchmarkGet100K(b *testing.B) {
	// Open db.
	fileName := "db-testing"
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	db, _ := Open(fileName, opts)
	defer os.RemoveAll(fileName)
	defer db.Close()
	// Put suite
	base := uint32(time.Now().Unix())
	for i := 0; i < 1024*100; i++ {
		for j := 0; j < 10; j++ {
			db.Put(&models.Metric{Link: uint32(j), Stamp: base + uint32(i)*7, Value: float64(i)})
		}
	}
	// Benchmark.
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.Get("whatever", uint32(i%10), base, base+100)
	}
}
