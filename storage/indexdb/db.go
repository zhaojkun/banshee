// Copyright 2015 Eleme Inc. All rights reserved.

package indexdb

import (
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util/log"
	"github.com/eleme/banshee/util/trie"
	"github.com/syndtr/goleveldb/leveldb"
)

// delim is the metric name delimeter.
const delim = "."

// DB handles indexes storage.
type DB struct {
	// LevelDB.
	db *leveldb.DB
	// Cache.
	tr *trie.Trie
}

// Open a DB by fileName.
func Open(fileName string) (*DB, error) {
	ldb, err := leveldb.OpenFile(fileName, nil)
	if err != nil {
		return nil, err
	}
	tr := trie.New(delim)
	db := new(DB)
	db.db = ldb
	db.tr = tr
	db.load()
	return db, nil
}

// Close the DB.
func (db *DB) Close() error {
	return db.db.Close()
}

// load indexes from db to cache.
func (db *DB) load() {
	log.Debugf("init index from indexdb..")
	// Scan values to memory.
	iter := db.db.NewIterator(nil, nil)
	for iter.Next() {
		// Decode
		key := iter.Key()
		value := iter.Value()
		idx := &models.Index{}
		idx.Name = string(key)
		err := decode(value, idx)
		if err != nil {
			// Skip corrupted values
			log.Warn("corrupted data found, skipping..")
			continue
		}
		idx.Share()
		db.tr.Put(idx.Name, idx)
	}
}

// get index by name.
func (db *DB) get(name string) (*models.Index, bool) {
	v := db.tr.Get(name)
	if v == nil {
		return nil, false
	}
	return v.(*models.Index), true
}

// Operations.

// Put an index into db.
func (db *DB) Put(idx *models.Index) error {
	// Save to db.
	key := []byte(idx.Name)
	value := encode(idx)
	err := db.db.Put(key, value, nil)
	if err != nil {
		return err
	}
	// Use an copy.
	idx = idx.Copy()
	// Add to cache.
	idx.Share()
	db.tr.Put(idx.Name, idx)
	return nil
}

// Get an index by name.
func (db *DB) Get(name string) (*models.Index, error) {
	i, ok := db.get(name)
	if !ok {
		return nil, ErrNotFound
	}
	return i.Copy(), nil
}

// Delete an index by name.
func (db *DB) Delete(name string) error {
	// Delete in cache.
	db.tr.Pop(name)
	key := []byte(name)
	return db.db.Delete(key, nil)
}

// Filter indexes by pattern.
func (db *DB) Filter(pattern string) (l []*models.Index) {
	for _, v := range db.tr.Match(pattern) {
		idx := v.(*models.Index)
		l = append(l, idx.Copy())
	}
	return
}

// All returns all indexes.
func (db *DB) All() (l []*models.Index) {
	m := db.tr.Map()
	for _, v := range m {
		idx := v.(*models.Index)
		l = append(l, idx.Copy())
	}
	return
}

// Len returns the number of indexes.
func (db *DB) Len() int {
	return db.tr.Len()
}
