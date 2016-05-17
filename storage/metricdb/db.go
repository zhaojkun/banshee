// Copyright 2016 Eleme Inc. All rights reserved.

/*

Package metricdb handles the metrics storage.

The DB contains multiple leveldb instances, a new leveldb instance would be
created and also an old instance would be expired every day.

File Structure

Example file structure for period=1day, expiration=7days:

	storage/       (period=1day, expiration=7days)
	  |- admindb
	  |- indexdb/
	  |- metricdb/
	        |- 16912 -- Outdated
	        |- 16913 -- -7
	        |- 16914 -- -6
	        |- 16915 -- -5
	        |- 16916 -- -4
	        |- 16917 -- -3
	        |- 16918 -- -2
	        |- 16919 -- -1
	        |- 16920 -- Active


Entry Format

Key-Value design in leveldb:

	|------- Key (8) ------|-------------- Value (24) -----------|
	+----------+-----------+-----------+-----------+-------------+
	| Link (4) | Stamp (4) | Value (8) | Score (8) | Average (8) |
	+----------+-----------+-----------+-----------+-------------+

Memory Cache

Metrics may be mirrored to memory to improve read performance.

*/
package metricdb

import (
	"github.com/eleme/banshee/models"
	"math/rand"
)

// Options is to open DB.
type Options struct {
	Period          uint32
	Expiration      uint32
	EnableCache     bool
	CachePercentage float64
}

// DB is the top level metric storage handler.
type DB struct {
	opts *Options
	mp   *memStoragePool
	fp   *fileStoragePool
}

// Open a DB.
func Open(fileName string, idxs []*models.Index, opts *Options) (db *DB, err error) {
	db = &DB{opts: opts}
	if db.fp, err = openFileStoragePool(fileName, opts); err != nil {
		return
	}
	if opts.EnableCache {
		db.mp = newMemStoragePool(opts)
		go db.mp.init(db.fp, idxs)
	}
	return
}

// Close the DB.
func (db *DB) Close() error { return db.fp.close() }

// Put a metric into db.
func (db *DB) Put(m *models.Metric) (err error) {
	if err = db.fp.put(m); err != nil {
		return
	}
	if db.opts.EnableCache && !db.mp.isInitErr() {
		if db.mp.has(m.Link) || rand.Float64() < db.opts.CachePercentage {
			if err = db.mp.put(m); err != nil {
				return
			}
		}
	}
	return
}

// Get metrics in a stamp range.
func (db *DB) Get(name string, link, start, end uint32) (ms []*models.Metric, err error) {
	if db.opts.EnableCache && db.mp.isInitOK() && db.mp.has(link) {
		return db.mp.get(link, start, end), nil
	}
	return db.fp.get(name, link, start, end)
}
