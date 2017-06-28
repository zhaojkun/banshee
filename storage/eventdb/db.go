// Copyright 2016 Eleme Inc. All rights reserved.

/*

Package eventdb handles the events storage.

The DB contains multiple sqlite instances, a new sqlite instance would be
created and also an old instance would be expired every day.

File Structure

Example file structure for period=1day, expiration=7days:

	storage/
	   |- event/
            |- 16912 -- Outdated
	        |- 16913 -- -7
	        |- 16914 -- -6
	        |- 16915 -- -5
	        |- 16916 -- -4
	        |- 16917 -- -3
	        |- 16918 -- -2
	        |- 16919 -- -1
	        |- 16920 -- Active

*/
package eventdb

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"sync"

	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util/log"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3" // Import but no use
)

// SQL db dialect
const dialect = "sqlite3"
const filemode = 0755

// ErrNoStorage is returned when no storage is able to serve, which
// indicates that given stamp or stamp range may be invalid.
var ErrNoStorage = errors.New("eventdb: no storage")

// storage is the actual data store handler.
type storage struct {
	id uint32
	db *gorm.DB
}

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

// DB is the top level event storage handler.
type DB struct {
	name string // dirname
	opts *Options
	pool []*storage   // sorted byID
	lock sync.RWMutex // protects runtime pool
}

// openStorage opens astorage by filename.
func openStorage(fileName string) (*storage, error) {
	base := path.Base(fileName)
	n, err := strconv.ParseUint(base, 10, 32)
	if err != nil {
		return nil, err
	}
	id := uint32(n)
	db, err := gorm.Open(dialect, fileName)
	if err != nil {
		return nil, err
	}
	s := &storage{id: id, db: db}
	if err = s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

// close the storage.
func (s *storage) close() error { return s.db.Close() }

// migrate storage schmea.
func (s *storage) migrate() error {
	log.Debugf("migrate storage %d schema", s.id)
	return s.db.AutoMigrate(&EventWrapper{}).Error
}

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
	gdb, err := gorm.Open(dialect, fileName)
	if err != nil {
		return err
	}
	s := &storage{db: gdb, id: id}
	if err = s.migrate(); err != nil { // DoNot forget
		return err
	}
	db.pool = append(db.pool, s)
	log.Infof("storage %d created", id)
	return nil
}

// expireStorage expires storages.
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

// EventWrapper is an event wrapper to store into sqlite.
type EventWrapper struct {
	ID                string  `gorm:"primary_key" json:"id"`
	RuleID            int     `sql:"index;not null" json:"ruleID"`
	ProjectID         int     `sql:"index;not null" json:"projectID"`
	Level             int     `sql:"index;not null" json:"level"`
	Comment           string  `sql:"type:varchar(256)" json:"comment"` // rule.Comment
	Name              string  `sql:"size:257;not null" json:"name"`
	Stamp             uint32  `sql:"index;not null" json:"stamp"`
	Score             float64 `json:"score"`   // index.Score
	Average           float64 `json:"average"` // index.Average
	Value             float64 `json:"value"`   // metric.Value
	TranslatedComment string  `sql:"type:varchar(513)" json:"translatedComment"`
}

// NewEventWrapper creates a new EventWrapper from models.Event.
func NewEventWrapper(ev *models.Event) *EventWrapper {
	// Note: No need to rlock the index or rule, all rules/indexes are copied
	// of the shared rules/indexes.
	return &EventWrapper{
		ID:                ev.ID,
		RuleID:            ev.Rule.ID,
		ProjectID:         ev.Rule.ProjectID,
		Level:             ev.Rule.Level,
		Comment:           ev.Rule.Comment,
		Name:              ev.Index.Name,
		Stamp:             ev.Index.Stamp,
		Score:             ev.Index.Score,
		Average:           ev.Index.Average,
		Value:             ev.Metric.Value,
		TranslatedComment: ev.TranslateRuleComment(),
	}
}

// adjust db storages pool.
func (db *DB) adjust(stamp uint32) (err error) {
	db.lock.Lock()
	defer db.lock.Unlock()
	if err = db.createStorage(stamp); err != nil {
		return
	}
	if err = db.expireStorages(); err != nil {
		return
	}
	return
}

// Put an event wraper into db.
func (db *DB) Put(ew *EventWrapper) (err error) {
	// Adjust storage pool.
	if err = db.adjust(ew.Stamp); err != nil {
		return
	}
	db.lock.RLock()
	defer db.lock.RUnlock()
	// Select a storage.
	if len(db.pool) == 0 {
		return ErrNoStorage
	}
	for i := len(db.pool) - 1; i >= 0; i-- {
		s := db.pool[i]
		if s.id*db.opts.Period <= ew.Stamp && ew.Stamp < (s.id+1)*db.opts.Period {
			return s.put(ew)
		}
	}
	return ErrNoStorage
}

// put an event wrapper into storage.
func (s *storage) put(ew *EventWrapper) (err error) {
	if err = s.db.Create(ew).Error; err != nil {
		return
	}
	return
}

// rangeOpt is the timestamp range options.
type rangeOpt struct {
	s     *storage
	start uint32
	end   uint32
}

// getRangeOpt returns the storages along with timestamp range.
func (db *DB) getRangeOpts(start, end uint32) (opts []*rangeOpt) {
	if len(db.pool) == 0 {
		return
	}
	for _, s := range db.pool {
		min := s.id * db.opts.Period
		max := (s.id + 1) * db.opts.Period
		if start >= max || end < min {
			continue
		}
		opt := &rangeOpt{s, start, end}
		if opt.start < min {
			opt.start = min
		}
		if opt.end > max {
			opt.end = max
		}
		opts = append(opts, opt)
	}
	return
}

// getByProjectID returns event wrappers by project id and time range.
func (s *storage) getByProjectID(projectID, lowestLevel int, start, end uint32) (ews []EventWrapper, err error) {
	if err = s.db.Where("project_id = ? AND stamp >= ? AND stamp < ? AND level >= ?", projectID, start, end, lowestLevel).Find(&ews).Error; err != nil {
		return
	}
	return
}

// getRange returns event wrappers by lowest level and time range.
func (s *storage) getRange(lowestLevel int, start, end uint32) (ews []EventWrapper, err error) {
	if err = s.db.Where("stamp >= ? AND stamp < ? AND level >= ?", start, end, lowestLevel).Find(&ews).Error; err != nil {
		return
	}
	return
}

// GetByProjectID returns event wrappers by project id and time range.
func (db *DB) GetByProjectID(projectID, lowestLevel int, start, end uint32) (ews []EventWrapper, err error) {
	db.lock.RLock()
	defer db.lock.RUnlock()
	for _, opt := range db.getRangeOpts(start, end) {
		var chunk []EventWrapper
		if chunk, err = opt.s.getByProjectID(projectID, lowestLevel, opt.start, opt.end); err != nil {
			return
		}
		ews = append(ews, chunk...)
	}
	return
}

// GetRange returns event wrappers by lowest level and time range.
func (db *DB) GetRange(lowestLevel int, start, end uint32) (ews []EventWrapper, err error) {
	db.lock.RLock()
	defer db.lock.RUnlock()
	for _, opt := range db.getRangeOpts(start, end) {
		var chunk []EventWrapper
		if chunk, err = opt.s.getRange(lowestLevel, opt.start, opt.end); err != nil {
			return
		}
		ews = append(ews, chunk...)
	}
	return
}
