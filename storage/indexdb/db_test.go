// Copyright 2015 Eleme Inc. All rights reserved.

package indexdb

import (
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"strconv"
	"testing"
)

func TestOpen(t *testing.T) {
	fileName := "db-testing"
	db, err := Open(fileName, nil)
	util.Must(t, err == nil)
	util.Must(t, util.IsFileExist(fileName))
	db.Close()
	os.RemoveAll(fileName)
}

func TestLoad(t *testing.T) {
	fileName := "db-testing"
	db, _ := Open(fileName, nil)
	defer os.RemoveAll(fileName)
	defer db.Close()
	idx := &models.Index{Name: "foo", Stamp: 1450430839, Score: 0.7, Average: 78.5}
	// Add one
	db.Put(idx)
	util.Must(t, db.tr.Has(idx.Name))
	util.Must(t, idx.Link == 1) // the first link
	// Clear cache
	db.tr.Clear()
	db.idp.Clear()
	util.Must(t, db.tr.Len() == 0)
	util.Must(t, !db.tr.Has(idx.Name))
	// Reload
	db.load()
	// Must not empty and idx in cache
	util.Must(t, db.tr.Len() == 1)
	util.Must(t, db.tr.Has(idx.Name))
	t.Logf("idp len: %v", db.idp.Len())
	util.Must(t, db.idp.Len() == 1)
	// Get again.
	i, err := db.Get(idx.Name)
	util.Must(t, err == nil && i.Equal(idx))
	util.Must(t, i.Link == 1) // link should be correct
}

func TestLoadExpired(t *testing.T) {
	fileName := "db-testing"
	opts := &Options{86400 * 7}
	db, _ := Open(fileName, opts)
	defer os.RemoveAll(fileName)
	defer db.Close()
	idx := &models.Index{Name: "foo", Stamp: 1450430839, Score: 0.7, Average: 78.5}
	db.Put(idx)
	// Clear cache
	db.tr.Clear()
	db.idp.Clear()
	db.load()
	util.Must(t, db.tr.Len() == 0)
	util.Must(t, !db.tr.Has(idx.Name))
}

func TestPut(t *testing.T) {
	// Open db.
	fileName := "db-testing"
	db, _ := Open(fileName, nil)
	defer os.RemoveAll(fileName)
	defer db.Close()
	// Test.
	idx := &models.Index{Name: "foo", Stamp: 1450430837, Score: 1.2, Average: 109.5}
	err := db.Put(idx)
	util.Must(t, err == nil)
	// Must in cache
	util.Must(t, db.tr.Has(idx.Name))
	// Must have link
	util.Must(t, db.idp.Len() == 1)
	util.Must(t, idx.Link == 1)
	// Must in db file
	v, err := db.db.Get([]byte(idx.Name), nil)
	util.Must(t, err == nil)
	idx1 := &models.Index{}
	decode(v, idx1)
	util.Must(t, idx1.Stamp == idx.Stamp)
	util.Must(t, idx1.Score == idx.Score)
	util.Must(t, idx1.Average == idx.Average)
	// RePut with another value.
	idx.Score = 1.3
	util.Must(t, db.Put(idx) == nil)
	util.Must(t, db.idp.Len() == 1) // pool shouldn't change
	util.Must(t, idx.Link == 1)
}

func TestGet(t *testing.T) {
	// Open db.
	fileName := "db-testing"
	db, _ := Open(fileName, nil)
	defer os.RemoveAll(fileName)
	defer db.Close()
	// Not found.
	_, err := db.Get("Not-exist")
	util.Must(t, ErrNotFound == err)
	// Put one.
	idx := &models.Index{Name: "foo", Stamp: 1450430837, Score: 0.3, Average: 100}
	db.Put(idx)
	// Get it from cache.
	i, err := db.Get(idx.Name)
	util.Must(t, nil == err)
	util.Must(t, i.Equal(idx))
}

func TestDelete(t *testing.T) {
	// Open db.
	fileName := "db-testing"
	db, _ := Open(fileName, nil)
	defer os.RemoveAll(fileName)
	defer db.Close()
	// Add one.
	idx := &models.Index{Name: "foo", Stamp: 1450430837, Score: 0.3, Average: 100}
	db.Put(idx)
	// Must in cache.
	util.Must(t, db.tr.Has(idx.Name))
	util.Must(t, db.idp.Len() == 1)
	util.Must(t, idx.Link == 1)
	// Delete it.
	err := db.Delete(idx.Name)
	util.Must(t, err == nil)
	// Must not exist in cache
	util.Must(t, !db.tr.Has(idx.Name))
	// Must not in db.
	_, err = db.db.Get([]byte(idx.Name), nil)
	util.Must(t, err == leveldb.ErrNotFound)
	// Must not in link pool.
	util.Must(t, db.idp.Len() == 0)
	// Cant get again.
	_, err = db.Get(idx.Name)
	util.Must(t, ErrNotFound == err)
}

func TestFilter(t *testing.T) {
	// Open db.
	fileName := "db-testing"
	db, _ := Open(fileName, nil)
	defer os.RemoveAll(fileName)
	defer db.Close()
	// Add indexes.
	excludeName := "a.b.c.d.e.f"
	db.Put(&models.Index{Name: "a.b.c.e.f.g"})
	db.Put(&models.Index{Name: "a.b.c.d.f.g"})
	db.Put(&models.Index{Name: excludeName})
	// Filter
	l := db.Filter("a.b.c.*.f.*")
	util.Must(t, len(l) == 2)
	util.Must(t, l[0].Name != excludeName && l[1].Name != excludeName)
}

func TestLen(t *testing.T) {
	// Open db.
	fileName := "db-testing"
	db, _ := Open(fileName, nil)
	defer os.RemoveAll(fileName)
	defer db.Close()
	// Add indexes.
	db.Put(&models.Index{Name: "abc"})
	db.Put(&models.Index{Name: "efg"})
	db.Put(&models.Index{Name: "fgh"})
	// Len
	util.Must(t, db.Len() == 3)
}

func BenchmarkGet10K(b *testing.B) {
	// Open db.
	fileName := "db-testing"
	db, _ := Open(fileName, nil)
	defer os.RemoveAll(fileName)
	defer db.Close()
	// Suite
	for i := 0; i < 10*1024; i++ {
		db.Put(&models.Index{Name: strconv.FormatInt(int64(i), 2)})
	}
	// Benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.Get(strconv.FormatInt(int64(i), 2))
	}
}

func BenchmarkPut(b *testing.B) {
	// Open db.
	fileName := "db-testing"
	db, _ := Open(fileName, nil)
	defer os.RemoveAll(fileName)
	defer db.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.Put(&models.Index{Name: strconv.FormatInt(int64(i), 2)})
	}
}
