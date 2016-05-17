// Copyright 2015 Eleme Inc. All rights reserved.

package metricdb

import (
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util/log"
	"github.com/syndtr/goleveldb/leveldb"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"sync"
)

// File structure: period=1day, expiration=7days
//
//	storage/       (period=1day, expiration=7days)
//	  |- admindb
//	  |- indexdb/
//	  |- metricdb/
//	        |- 16912 -- Outdated
//	        |- 16913 -- -7
//	        |- 16914 -- -6
//	        |- 16915 -- -5
//	        |- 16916 -- -4
//	        |- 16917 -- -3
//	        |- 16918 -- -2
//	        |- 16919 -- -1
//	        |- 16920 -- Active
//

// filemode is the file mode to open a new storage.
const filemode = 0755

// storage is the leveldb.DB wrapper with an id.
type storage struct {
	id uint32
	db *leveldb.DB
}

// byID implements sort.Interface for a slice of storages.
type byID []*storage

func (b byID) Len() int           { return len(b) }
func (b byID) Less(i, j int) bool { return b[i].id < b[j].id }
func (b byID) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

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

// get metrics in a timestamp range from the storage.
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

// storageManager is the storage manager.
type storageManager struct {
	opts *Options     // *Options ref
	name string       // dirname
	pool []*storage   // storages
	lock sync.RWMutex // protects pool
}

// openStorageManager opens a storageManager by filename.
func openStorageManager(fileName string, opts *Options) (*storageManager, error) {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) { // Create if not exist
		if err := os.Mkdir(fileName, filemode); err != nil {
			return nil, err
		}
		log.Debugf("dir %s created", fileName)
	}
	smgr := &storageManager{opts: opts, name: name}
	if err := smgr.init(); err != nil {
		return nil, err
	}
	return smgr, nil
}

// init opens all storages on storageManager open.
func (smgr *storageManager) init() error {
	infos, err := ioutil.ReadDir(smgr.name)
	if err != nil {
		return err
	}
	for _, info := range infos {
		fileName := path.Join(smgr.name, info.Name())
		s, err := openStorage(fileName)
		if err != nil {
			return err
		}
		smgr.pool = append(smgr.pool, s)
		log.Debugf("storage %d opened", s.id)
	}
	sort.Sort(byID(smgr.pool))
	return nil
}

// close the storageManager.
func (smgr *storageManager) close() (err error) {
	for _, s := range smgr.pool {
		if err = s.close(); err != nil {
			return
		}
	}
	return nil
}

// createStorage creates a storage for given stamp.
// Dose nothing if the stamp is not large enough.
func (smgr *storageManager) createStorage(stamp uint32) error {
	id := stamp / smgr.opts.Period
	if len(smgr.pool) > 0 && id <= smgr.pool[len(smgr.pool)-1].id {
		// stamp is not large enough.
		return nil
	}
	baseName := strconv.FormatUint(uint64(id), 10)
	fileName := path.Join(smgr.name, baseName)
	ldb, err := leveldb.OpenFile(fileName, nil)
	if err != nil {
		return err
	}
	s := &storage{db: ldb, id: id}
	smgr.pool = append(smgr.pool, s)
	log.Infof("storage %d created", id)
	return nil
}

// expireStorage expires storages.
// Dose nothing if the pool needs no expiration.
func (smgr *storageManager) expireStorages() error {
	if len(smgr.pool) == 0 {
		return nil
	}
	id := smgr.pool[len(smgr.pool)-1].id - smgr.opts.Expiration/smgr.opts.Period
	pool := smgr.pool
	for i, s := range smgr.pool {
		if s.id < id {
			if err := s.close(); err != nil {
				return err
			}
			baseName := strconv.FormatUint(uint64(s.id), 10)
			fileName := path.Join(smgr.name, baseName)
			if err := os.RemoveAll(fileName); err != nil {
				return err
			}
			pool = smgr.pool[i+1:]
			log.Infof("storage %d expired", s.id)
		}
	}
	smgr.pool = pool
	return nil
}

// put a metric into storageManager.
// Returns ErrNoStorage if no storage is available.
func (smgr *storageManager) put(m *models.Metric) (err error) {
	smgr.lock.Lock()
	defer smgr.lock.Unlock()
	// Adjust storage pool.
	if err = smgr.createStorage(m.Stamp); err != nil {
		return
	}
	if err = smgr.expireStorages(); err != nil {
		return
	}
	// Select a storage.
	if len(smgr.pool) == 0 {
		return ErrNoStorage
	}
	for i := len(smgr.pool) - 1; i >= 0; i-- {
		s := smgr.pool[i]
		if s.id*smgr.opts.Period <= m.Stamp && m.Stamp < (s.id+1)*smgr.opts.Period {
			return s.put(m)
		}
	}
	return ErrNoStorage
}

// get metrics in a timestamp range, the range is left open and right closed.
func (smgr *storageManager) get(name string, link, start, end uint32) ([]*models.Metric, error) {
	smgr.lock.RLock()
	defer smgr.lock.RUnlock()
	var ms []*models.Metric
	if len(smgr.pool) == 0 {
		return ms, nil
	}
	for _, s := range smgr.pool {
		min := s.id * smgr.opts.Period
		max := (s.id + 1) * smgr.opts.Period
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
