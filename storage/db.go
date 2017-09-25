// Copyright 2015 Eleme Inc. All rights reserved.

package storage

import (
	"fmt"
	"os"
	"path"

	"github.com/eleme/banshee/storage/admindb"
	"github.com/eleme/banshee/storage/eventdb"
	"github.com/eleme/banshee/storage/indexdb"
	"github.com/eleme/banshee/storage/metricdb"
	"github.com/eleme/banshee/util/log"
	"github.com/jinzhu/gorm"
)

// DB file mode.
const filemode = 0755

// Child db filename.
const (
	indexdbFileName  = "index"
	metricdbFileName = "metric"
	eventdbFileName  = "event"
)

// Options is to open DB.
type Options struct {
	Period       uint32
	Expiration   uint32
	FilterOffset float64
}

// DB handles the storage on leveldb.
type DB struct {
	// Child db
	Admin  *admindb.DB
	Index  *indexdb.DB
	Metric *metricdb.DB
	Event  *eventdb.DB
}

// Open a DB by fileName and options.
func Open(fileName string, opts *Options) (*DB, error) {
	// Create if not exist
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		log.Debugf("create dir %s", fileName)
		err := os.Mkdir(fileName, filemode)
		if err != nil {
			return nil, err
		}
	}
	db := new(DB)
	// Indexdb.
	var indexdbOpts *indexdb.Options
	if opts != nil {
		indexdbOpts = &indexdb.Options{Expiration: opts.Expiration}
	}
	db.Index, err = indexdb.Open(path.Join(fileName, indexdbFileName), indexdbOpts)
	if err != nil {
		return nil, err
	}
	// Metricdb.
	var options *metricdb.Options
	if opts != nil {
		options = &metricdb.Options{
			Period:       opts.Period,
			Expiration:   opts.Expiration,
			FilterOffset: opts.FilterOffset,
		}
	}
	db.Metric, err = metricdb.Open(path.Join(fileName, metricdbFileName), options)
	if err != nil {
		return nil, err
	}
	// Eventdb.
	var eventdbOpts *eventdb.Options
	if opts != nil {
		eventdbOpts = &eventdb.Options{
			Period:     opts.Period,
			Expiration: opts.Expiration,
		}
	}
	db.Event, err = eventdb.Open(path.Join(fileName, eventdbFileName), eventdbOpts)
	if err != nil {
		return nil, err
	}
	log.Debugf("storage is opened successfully")
	return db, nil
}

// AdminOptions for Admin database.
type AdminOptions struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbName"`
}

// InitAdminDB init admin database.
func (db *DB) InitAdminDB(opts AdminOptions) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local&interpolateParams=True",
		opts.User, opts.Password, opts.Host, opts.Port, opts.DBName)
	gdb, err := gorm.Open("mysql", dsn)
	if err != nil {
		return err
	}
	db.Admin, err = admindb.Open(gdb)
	return err
}

// Close a DB.
func (db *DB) Close() error {
	// Admindb.
	if db.Admin != nil {
		if err := db.Admin.Close(); err != nil {
			return err
		}
	}
	// Indexdb.
	if err := db.Index.Close(); err != nil {
		return err
	}
	// Metricdb.
	if err := db.Metric.Close(); err != nil {
		return err
	}
	// Eventdb.
	return db.Event.Close()
}
