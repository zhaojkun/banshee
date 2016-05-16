// Copyright 2015 Eleme Inc. All rights reserved.

/*

Package indexdb handles the indexes storage.

Design

The DB is a leveldb instance, and the key-value format is:

	|--- Key --|------------------ Value (24) -------------------|
	+----------+-----------+-----------+-----------+-------------+
	| Name (X) |  Link (4) | Stamp (4) | Score (8) | Average (8) |
	+----------+-----------+-----------+-----------+-------------+

Cache

To access indexes faster, indexes are cached in memory, in a trie with
goroutine safety.

Read operations are in cache.

Write operations are to persistence and cache.

*/
package indexdb

import (
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util/idpool"
	"github.com/eleme/banshee/util/log"
	"github.com/eleme/banshee/util/trie"
	"github.com/syndtr/goleveldb/leveldb"
)

// MaxNumIndex is the max value of the number indexes this db can handle.
// Note that this number has no other meanings, just a limitation of the
// indexes capacity, it can be larger, theoretically can be MaxUint32.
const MaxNumIndex = 16 * 1024 * 1024

// DB handles indexes storage.
type DB struct {
	// LevelDB.
	db *leveldb.DB
	// Cache.
	tr  *trie.Trie
	idp *idpool.Pool
}

// Open a DB by fileName.
func Open(fileName string) (*DB, error) {
	ldb, err := leveldb.OpenFile(fileName, nil)
	if err != nil {
		return nil, err
	}
	db := new(DB)
	db.db = ldb
	db.tr = trie.New()
	db.idp = idpool.New(1, MaxNumIndex) // low is 1 to distinct default 0
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
		db.idp.Reserve(int(idx.Link))
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
	if !db.tr.Has(idx.Name) { // It's new
		idx.Link = uint32(db.idp.Allocate()) // allocate link
	}
	if idx.Link == 0 {
		return ErrNoLink
	}
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
	v := db.tr.Pop(name)
	if v == nil {
		return ErrNotFound
	}
	idx := v.(*models.Index)
	// Delete from db.
	key := []byte(name)
	if err := db.db.Delete(key, nil); err != nil {
		return err
	}
	// Release link.
	db.idp.Release(int(idx.Link))
	return nil
}

// Has checks if an index is in db.
func (db *DB) Has(name string) bool {
	return db.tr.Has(name)
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
