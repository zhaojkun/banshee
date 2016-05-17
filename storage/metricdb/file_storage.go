// Copyright 2016 Eleme Inc. All rights reserved.

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

// fileStorage is the file based storage.
type fileStorage struct {
	id  uint32
	ldb *leveldb.DB
}

// close the file storage.
func (s *fileStorage) close() error { return s.ldb.Close() }

// put a metric into the file storage.
func (s *fileStorage) put(m *models.Metric) error {
	if m.Link == 0 {
		return ErrNoLink
	}
	return s.ldb.Put(encodeKey(m), encodeValue(m), nil)
}

// get metrics in a stamp range.
func (s *fileStorage) get(name string, link, start, end uint32) (ms []*models.Metric, err error) {
	startKey := encodeKey(&models.Metric{Link: link, Stamp: start})
	endKey := encodeKey(&models.Metric{Link: link, Stamp: end})
	iter := s.ldb.NewIterator(&util.Range{Start: startKey, Limit: endKey}, nil)
	for iter.Next() {
		m := &models.Metric{Name: name}
		if err = decodeKey(iter.Key(), m); err != nil {
			return
		}
		if err = decodeValue(iter.Value(), m); err != nil {
			return
		}
		ms = append(ms, m)
	}
	return
}

// openFileStorage opens a fileStorage by filename.
func openFileStorage(fileName string) (*fileStorage, error) {
	n, err := strconv.ParseUint(path.Base(fileName), 10, 32)
	if err != nil {
		return nil, err
	}
	ldb, err := leveldb.OpenFile(fileName, nil)
	if err != nil {
		return nil, err
	}
	return &fileStorage{uint32(n), ldb}, nil
}

// fileStoragesByID implements sort.Interface for a slice of storages.
type fileStoragesByID []*fileStorage

func (b fileStoragesByID) Len() int           { return len(b) }
func (b fileStoragesByID) Less(i, j int) bool { return b[i].id < b[j].id }
func (b fileStoragesByID) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

// fileStoragePool is the file storage pool.
type fileStoragePool struct {
	opts *Options
	name string
	pool []*fileStorage
	lock sync.RWMutex // protects the pool
}

// openFileStoragePool opens a fileStoragePool for given filename.
func openFileStoragePool(fileName string, opts *Options) (p *fileStoragePool, err error) {
	_, err = os.Stat(fileName)
	if os.IsNotExist(err) { // Create if not exist
		if err = os.Mkdir(fileName, filemode); err != nil {
			return
		}
		log.Debugf("dir %s created", fileName)
	}
	p = &fileStoragePool{opts: opts, name: fileName}
	if err = p.init(); err != nil {
		return
	}
	return
}

// init all file storages.
func (p *fileStoragePool) init() error {
	infs, err := ioutil.ReadDir(p.name)
	if err != nil {
		return err
	}
	for _, info := range infs {
		fileName := path.Join(p.name, info.Name())
		s, err := openFileStorage(fileName)
		if err != nil {
			return err
		}
		p.pool = append(p.pool, s)
		log.Debugf("file storage %d opened", s.id)
	}
	sort.Sort(fileStoragesByID(p.pool))
	return nil
}

// close the pool.
func (p *fileStoragePool) close() (err error) {
	for _, s := range p.pool {
		if err = s.close(); err != nil {
			return
		}
	}
	return
}

// create a file storage for given stamp.
func (p *fileStoragePool) create(stamp uint32) error {
	id := stamp / p.opts.Period
	if len(p.pool) > 0 && id <= p.pool[len(p.pool)-1].id {
		return nil // Not large enough
	}
	fileName := path.Join(p.name, strconv.FormatUint(uint64(id), 10))
	ldb, err := leveldb.OpenFile(fileName, nil)
	if err != nil {
		return err
	}
	p.pool = append(p.pool, &fileStorage{id, ldb})
	log.Infof("file storage %d created", id)
	return nil
}

// expire oudated file storages.
func (p *fileStoragePool) expire() (err error) {
	if len(p.pool) == 0 {
		return nil
	}
	id := p.pool[len(p.pool)-1].id - p.opts.Expiration/p.opts.Period
	for i, s := range p.pool {
		if s.id < id {
			if err = s.close(); err != nil {
				return
			}
			fileName := path.Join(p.name, strconv.FormatUint(uint64(s.id), 10))
			if err = os.RemoveAll(fileName); err != nil {
				return
			}
			p.pool = p.pool[i+1:]
			log.Infof("file storage %d expired", s.id)
		}
	}
	return
}

// adjust the pool.
func (p *fileStoragePool) adjust(stamp uint32) (err error) {
	if err = p.create(stamp); err != nil {
		return
	}
	if err = p.expire(); err != nil {
		return
	}
	return
}

// put a metric into pool.
// Returns ErrNoFileStorage if no storage is available.
func (p *fileStoragePool) put(m *models.Metric) (err error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if err = p.adjust(m.Stamp); err != nil {
		return
	}
	if len(p.pool) == 0 {
		return ErrNoFileStorage
	}
	for i := len(p.pool) - 1; i >= 0; i-- {
		s := p.pool[i]
		if s.id*p.opts.Period <= m.Stamp && m.Stamp < (s.id+1)*p.opts.Period {
			return s.put(m)
		}
	}
	return
}

// get metrics in a stamp range, the range is left open and right closed.
func (p *fileStoragePool) get(name string, link, start, end uint32) (ms []*models.Metric, err error) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if len(p.pool) == 0 {
		return
	}
	for _, s := range p.pool {
		min := s.id * p.opts.Period
		max := (s.id + 1) * p.opts.Period
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
	return
}
