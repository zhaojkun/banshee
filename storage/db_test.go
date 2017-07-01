// Copyright 2015 Eleme Inc. All rights reserved.

package storage

import (
	"os"
	"path"
	"testing"

	"github.com/eleme/banshee/util"
)

func TestOpen(t *testing.T) {
	// Open db.
	fileName := "storage_test"
	db, err := Open(fileName, nil)
	util.Must(t, err == nil)
	util.Must(t, db != nil)
	// Defer close and remove files.
	defer db.Close()
	defer os.RemoveAll(fileName)
	// Check if child db file exist
	util.Must(t, util.IsFileExist(path.Join(fileName, indexdbFileName)))
	util.Must(t, util.IsFileExist(path.Join(fileName, metricdbFileName)))
	util.Must(t, util.IsFileExist(path.Join(fileName, eventdbFileName)))
}
