// Copyright 2015 Eleme Inc. All rights reserved.

package storage

import (
	"github.com/eleme/banshee/storage/admindb"
	"github.com/eleme/banshee/storage/indexdb"
	"github.com/eleme/banshee/storage/metricdb"
	"github.com/eleme/banshee/util/log"
	"os"
	"path"
)

// DB file mode.
const filemode = 0755

// Child db filename.
const (
	admindbFileName  = "admin"
	indexdbFileName  = "index"
	metricdbFileName = "metric"
)

// Options is to open DB.
type Options struct {
	Period     uint32
	Expiration uint32
}

// DB handles the storage on leveldb.
type DB struct {
	// Child db
	Admin  *admindb.DB
	Index  *indexdb.DB
	Metric *metricdb.DB
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
	// Admindb.
	db := new(DB)
	db.Admin, err = admindb.Open(path.Join(fileName, admindbFileName))
	if err != nil {
		return nil, err
	}
	// Indexdb.
	db.Index, err = indexdb.Open(path.Join(fileName, indexdbFileName))
	if err != nil {
		return nil, err
	}
	// Metricdb.
	var options *metricdb.Options
	if opts != nil {
		options = &metricdb.Options{
			Period:     opts.Period,
			Expiration: opts.Expiration,
		}
	}
	db.Metric, err = metricdb.Open(path.Join(fileName, metricdbFileName), options)
	if err != nil {
		return nil, err
	}
	log.Debugf("storage is opened successfully")
	return db, nil
}

// Close a DB.
func (db *DB) Close() error {
	// Admindb.
	if err := db.Admin.Close(); err != nil {
		return err
	}
	// Indexdb.
	if err := db.Index.Close(); err != nil {
		return err
	}
	// Metricdb.
	if err := db.Metric.Close(); err != nil {
		return err
	}
	return nil
}
