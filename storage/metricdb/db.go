// Copyright 2015 Eleme Inc. All rights reserved.

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

*/
package metricdb

import (
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util/log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"sync"
)

// storage is the actual data store handler.
type storage struct {
	id uint32
	db *leveldb.DB
}

const filemode = 0755

// byID implements sort.Interface.
type byID []*storage

func (b byID) Len() int           { return len(b) }
func (b byID) Less(i, j int) bool { return b[i].id < b[j].id }
func (b byID) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

// Options is to open DB.
type Options struct {
	Period     uint32
	Expiration uint32
}

// DB is the top level metric storage handler.
type DB struct {
	name string // dirname
	opts *Options
	pool []*storage   // sorted byID
	lock sync.RWMutex // protects runtime pool
}

// openStorage opens a storage by filename.
func openStorage(fileName string) (*storage, error) {
	base := path.Base(fileName)
	n, err := strconv.ParseUint(base, 10, 32)
	if err != nil {
		return nil, err
	}
	id := uint32(n)
	db, err := leveldb.OpenFile(fileName, nil)
	if err != nil {
		return nil, err
	}
	return &storage{id: id, db: db}, nil
}

// close the storage.
func (s *storage) close() error { return s.db.Close() }

// Open a DB by filename.
func Open(fileName string, opts *Options) (*DB, error) {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		// Create if not exist
		if err := os.Mkdir(fileName, filemode); err != nil {
			return nil, err
		}
		log.Debugf("dir %s created", fileName)
	}
	db := &DB{opts: opts, name: fileName}
	if err = db.init(); err != nil {
		return nil, err
	}
	return db, nil
}

// init opens all storages on DB open.
func (db *DB) init() error {
	infos, err := ioutil.ReadDir(db.name)
	if err != nil {
		return err
	}
	for _, info := range infos {
		fileName := path.Join(db.name, info.Name())
		s, err := openStorage(fileName)
		if err != nil {
			return err
		}
		db.pool = append(db.pool, s)
		log.Debugf("storage %d opened", s.id)
	}
	sort.Sort(byID(db.pool))
	return nil
}

// Close the DB.
func (db *DB) Close() (err error) {
	for _, s := range db.pool {
		if err = s.close(); err != nil {
			return
		}
	}
	return nil
}

// createStorage creates a storage for given stamp.
// Dose nothing if the stamp is not large enough.
func (db *DB) createStorage(stamp uint32) error {
	id := stamp / db.opts.Period
	if len(db.pool) > 0 && id <= db.pool[len(db.pool)-1].id {
		// stamp is not large enough.
		return nil
	}
	baseName := strconv.FormatUint(uint64(id), 10)
	fileName := path.Join(db.name, baseName)
	ldb, err := leveldb.OpenFile(fileName, nil)
	if err != nil {
		return err
	}
	s := &storage{db: ldb, id: id}
	db.pool = append(db.pool, s)
	log.Infof("storage %d created", id)
	return nil
}

// expireStorage expire storages.
// Dose nothing if the pool needs no expiration.
func (db *DB) expireStorages() error {
	if len(db.pool) == 0 {
		return nil
	}
	id := db.pool[len(db.pool)-1].id - db.opts.Expiration/db.opts.Period
	pool := db.pool
	for i, s := range db.pool {
		if s.id < id {
			if err := s.close(); err != nil {
				return err
			}
			baseName := strconv.FormatUint(uint64(s.id), 10)
			fileName := path.Join(db.name, baseName)
			if err := os.RemoveAll(fileName); err != nil {
				return err
			}
			pool = db.pool[i+1:]
			log.Infof("storage %d expired", s.id)
		}
	}
	db.pool = pool
	return nil
}

// Put a metric into db.
// Returns ErrNoStorage if no storage is available.
func (db *DB) Put(m *models.Metric) (err error) {
	db.lock.Lock()
	defer db.lock.Unlock()
	// Adjust storage pool.
	if err = db.createStorage(m.Stamp); err != nil {
		return
	}
	if err = db.expireStorages(); err != nil {
		return
	}
	// Select a storage.
	if len(db.pool) == 0 {
		return ErrNoStorage
	}
	for i := len(db.pool) - 1; i >= 0; i-- {
		s := db.pool[i]
		if s.id*db.opts.Period <= m.Stamp && m.Stamp < (s.id+1)*db.opts.Period {
			return s.put(m)
		}
	}
	return ErrNoStorage
}

// put a metric into storage.
// Returns ErrNoLink if given metric has no link.
func (s *storage) put(m *models.Metric) error {
	if m.Link == 0 {
		return ErrNoLink
	}
	key := encodeKey(m)
	value := encodeValue(m)
	return s.db.Put(key, value, nil)
}

// get metrics in a timestamp range.
func (s *storage) get(name string, link, start, end uint32) ([]*models.Metric, error) {
	startKey := encodeKey(&models.Metric{Link: link, Stamp: start})
	endKey := encodeKey(&models.Metric{Link: link, Stamp: end})
	iter := s.db.NewIterator(&util.Range{Start: startKey, Limit: endKey}, nil)
	var ms []*models.Metric
	for iter.Next() {
		m := &models.Metric{Name: name}
		key := iter.Key()
		value := iter.Value()
		if err := decodeKey(key, m); err != nil {
			return nil, err
		}
		if err := decodeValue(value, m); err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	return ms, nil
}

// Get metrics in a timestamp range, the range is left open and right closed.
func (db *DB) Get(name string, link, start, end uint32) ([]*models.Metric, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()
	var ms []*models.Metric
	if len(db.pool) == 0 {
		return ms, nil
	}
	for _, s := range db.pool {
		min := s.id * db.opts.Period
		max := (s.id + 1) * db.opts.Period
		if start >= max || end < min {
			continue
		}
		st, ed := start, end
		if start < min {
			st = min
		}
		if end > max {
			ed = max
		}
		l, err := s.get(name, link, st, ed)
		if err != nil {
			return nil, err
		}
		ms = append(ms, l...)
	}
	return ms, nil
}
