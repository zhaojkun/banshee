// Copyright 2015 Eleme Inc. All rights reserved.

package storage

import (
	"errors"
	"io/ioutil"
	"os"
	"path"

	yaml "gopkg.in/yaml.v2"

	"github.com/eleme/banshee/storage/admindb"
	"github.com/eleme/banshee/storage/eventdb"
	"github.com/eleme/banshee/storage/indexdb"
	"github.com/eleme/banshee/storage/metricdb"
	"github.com/eleme/banshee/util/log"
)

// DB file mode.
const filemode = 0755

// Child db filename.
const (
	optionlockFileName = "option.lock"
	admindbFileName    = "admin"
	indexdbFileName    = "index"
	metricdbFileName   = "metric"
	eventdbFileName    = "event"
)

var (
	errPeriodNotMatched = errors.New("period has been changed,you should migrate data firstly")
)

// Options is to open DB.
type Options struct {
	Period       uint32
	Expiration   uint32
	FilterOffset float64
}

func (p *Options) validateWithYamlFile(fileName string) error {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	var opts Options
	err = yaml.Unmarshal(b, &opts)
	if err != nil {
		return err
	}
	if opts.Period != p.Period {
		return errPeriodNotMatched
	}
	return nil
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
	// Check with lock file
	lockFilePath := path.Join(fileName, optionlockFileName)
	err = opts.validateWithYamlFile(lockFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			b, _ := yaml.Marshal(opts)
			ioutil.WriteFile(lockFilePath, b, 0644)
		} else {
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
	// Eventdb.
	if err := db.Event.Close(); err != nil {
		return err
	}
	return nil
}
