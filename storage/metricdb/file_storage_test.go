// Copyright 2016 Eleme Inc. All rights reserved.

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

func TestFileStoragePoolOpenNil(t *testing.T) {
	fileName := "db-testing"
	fp, err := openFileStoragePool(fileName, nil)
	util.Must(t, err == nil)
	util.Must(t, util.IsFileExist(fileName))
	defer os.RemoveAll(fileName)
	defer fp.close()
	util.Must(t, len(fp.pool) == 0) // should have nothing in pool
}

func TestFileStoragePoolOpenInit(t *testing.T) {
	fileName := "db-testing"
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	fp, err := openFileStoragePool(fileName, opts)
	util.Must(t, err == nil)
	defer os.RemoveAll(fileName)
	stamp := uint32(time.Now().Unix())
	fp.put(&models.Metric{Stamp: stamp, Link: 1})
	fp.close()
	// Reopen.
	fp, err = openFileStoragePool(fileName, opts)
	util.Must(t, err == nil)
	util.Must(t, len(fp.pool) == 1)
	util.Must(t, fp.pool[0].id*fp.opts.Period <= stamp)
}

func TestFileStoragePut(t *testing.T) {
	// Open file storage.
	fileName := "db-testing"
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	fp, _ := openFileStoragePool(fileName, opts)
	defer os.RemoveAll(fileName)
	defer fp.close()
	// Put.
	m := &models.Metric{
		Name:    "foo",
		Link:    1,
		Stamp:   1452758773,
		Value:   3.14,
		Score:   0.1892,
		Average: 3.133,
	}
	util.Must(t, nil == fp.put(m))
	// Must in pool.
	key := encodeKey(m)
	value, err := fp.pool[0].ldb.Get(key, nil)
	util.Must(t, err == nil)
	m1 := &models.Metric{Name: m.Name, Stamp: m.Stamp, Link: 1}
	err = decodeValue(value, m1)
	util.Must(t, err == nil)
	util.Must(t, reflect.DeepEqual(m, m1))
}

func TestFileStorageGet(t *testing.T) {
	// Open file storage.
	fileName := "db-testing"
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	fp, _ := openFileStoragePool(fileName, opts)
	defer os.RemoveAll(fileName)
	defer fp.close()
	// Nothing.
	ms, err := fp.get("not-exist", 1234, 0, 1452758773)
	util.Must(t, err == nil)
	util.Must(t, len(ms) == 0)
	// Put some
	fp.put(&models.Metric{Name: "foo", Link: 1, Stamp: 1452758723})
	fp.put(&models.Metric{Name: "foo", Link: 1, Stamp: 1452758733, Value: 1.89, Score: 1.12, Average: 1.72})
	fp.put(&models.Metric{Name: "foo", Link: 1, Stamp: 1452758743})
	fp.put(&models.Metric{Name: "foo", Link: 1, Stamp: 1452758753})
	// Get again.
	ms, err = fp.get("foo", 1, 1452758733, 1452758753)
	util.Must(t, err == nil)
	util.Must(t, len(ms) == 2)
	// Test the value.
	m := ms[0]
	util.Must(t, m.Value == 1.89 && m.Score == 1.12 && m.Link == 1)
}

func TestFileStorageGetAcrossStorages(t *testing.T) {
	// Open file storage.
	fileName := "db-testing"
	opts := &Options{Period: 86400, Expiration: 86400 * 4}
	fp, _ := openFileStoragePool(fileName, opts)
	defer os.RemoveAll(fileName)
	defer fp.close()
	// Force creating 4+1 storages.
	base := uint32(time.Now().Unix())
	fp.put(&models.Metric{Link: 1, Stamp: base})                    // 0
	fp.put(&models.Metric{Link: 1, Stamp: base + fp.opts.Period*1}) // 1
	fp.put(&models.Metric{Link: 1, Stamp: base + fp.opts.Period*2}) // 2
	fp.put(&models.Metric{Link: 1, Stamp: base + fp.opts.Period*3}) // 3
	fp.put(&models.Metric{Link: 1, Stamp: base + fp.opts.Period*4}) // 4
	// Get
	ms, err := fp.get("whatever", 1, base, base+fp.opts.Period*3)
	util.Must(t, err == nil)
	util.Must(t, len(ms) == 3)
	util.Must(t, ms[0].Stamp == base)
	util.Must(t, ms[1].Stamp == base+fp.opts.Period*1)
	util.Must(t, ms[2].Stamp == base+fp.opts.Period*2)
}

func TestFileStorageExpire(t *testing.T) {
	// Open file storage.
	fileName := "db-testing"
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	fp, _ := openFileStoragePool(fileName, opts)
	defer os.RemoveAll(fileName)
	defer fp.close()
	// Force creating 7+1 storages.
	base := uint32(time.Now().Unix())
	fp.put(&models.Metric{Link: 1, Stamp: base})                    // 0
	id := fp.pool[0].id                                             // record the id to be deleted
	fp.put(&models.Metric{Link: 1, Stamp: base + fp.opts.Period*1}) // 1
	fp.put(&models.Metric{Link: 1, Stamp: base + fp.opts.Period*2}) // 2
	fp.put(&models.Metric{Link: 1, Stamp: base + fp.opts.Period*3}) // 3
	fp.put(&models.Metric{Link: 1, Stamp: base + fp.opts.Period*4}) // 4
	fp.put(&models.Metric{Link: 1, Stamp: base + fp.opts.Period*5}) // 5
	fp.put(&models.Metric{Link: 1, Stamp: base + fp.opts.Period*6}) // 6
	util.Must(t, len(fp.pool) == 7)
	fp.put(&models.Metric{Link: 1, Stamp: base + fp.opts.Period*7}) // 7
	util.Must(t, len(fp.pool) == 8)
	fp.put(&models.Metric{Link: 1, Stamp: base + fp.opts.Period*8}) // 8
	util.Must(t, len(fp.pool) == 8)                                 // Full storages: 1,2,3,4,5,6,7
	// Files must be deleted.
	deleteFileName := path.Join(fileName, strconv.FormatUint(uint64(id), 10))
	util.Must(t, !util.IsFileExist(deleteFileName))
}

func BenchmarkFileStoragePut(b *testing.B) {
	// Open file storage pool.
	fileName := "db-testing"
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	fp, _ := openFileStoragePool(fileName, opts)
	defer os.RemoveAll(fileName)
	defer fp.close()
	base := uint32(time.Now().Unix())
	// Bench
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fp.put(&models.Metric{Link: 1, Stamp: base + uint32(i)*7, Value: float64(i)})
	}
}

func BenchmarkFileStoragePutX10(b *testing.B) {
	// Open file storage pool.
	fileName := "db-testing"
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	fp, _ := openFileStoragePool(fileName, opts)
	defer os.RemoveAll(fileName)
	defer fp.close()
	base := uint32(time.Now().Unix())
	// Bench
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 10; j++ {
			fp.put(&models.Metric{Link: uint32(j), Stamp: base + uint32(i)*7, Value: float64(i)})
		}
	}
}

func BenchmarkFileStorageGet100K(b *testing.B) {
	// Open file storage pool.
	fileName := "db-testing"
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	fp, _ := openFileStoragePool(fileName, opts)
	defer os.RemoveAll(fileName)
	defer fp.close()
	// Put suite
	base := uint32(time.Now().Unix())
	for i := 0; i < 1024*100; i++ {
		for j := 0; j < 10; j++ {
			fp.put(&models.Metric{Link: uint32(j), Stamp: base + uint32(i)*7, Value: float64(i)})
		}
	}
	// Benchmark.
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fp.get("whatever", uint32(i%10), base, base+100*10)
	}
}
