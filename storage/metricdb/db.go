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

// Options is to open DB.
type Options struct {
	Period     uint32
	Expiration uint32
}

// DB is the top level metric storage handler.
type DB struct {
	opts *Options
	smgr *storageManager
}
