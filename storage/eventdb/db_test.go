// Copyright 2016 Eleme Inc. All rights reserved.

package eventdb

import (
	"github.com/eleme/banshee/util"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestOpenOptionsNil(t *testing.T) {
	fileName := "db-testing"
	db, err := Open(fileName, nil)
	util.Must(t, err == nil)
	util.Must(t, util.IsFileExist(fileName))
	defer os.RemoveAll(fileName)
	defer db.Close()
	util.Must(t, len(db.pool) == 0) // should have nothing in pool..

}

func TestOpenInit(t *testing.T) {
	fileName := "db-testing"
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	db, err := Open(fileName, opts)
	util.Must(t, err == nil)
	defer os.RemoveAll(fileName)
	stamp := uint32(time.Now().Unix())
	db.Put(&EventWrapper{Stamp: stamp})
	db.Close()
	// Reopen.
	db, err = Open(fileName, opts)
	util.Must(t, err == nil)
	util.Must(t, len(db.pool) == 1) // should have one storage
	util.Must(t, db.pool[0].id*db.opts.Period <= stamp)
}

func TestPut(t *testing.T) {
	// Open db.
	fileName := "db-testing"
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	db, _ := Open(fileName, opts)
	defer os.RemoveAll(fileName)
	defer db.Close()
	// Put.
	ew := &EventWrapper{
		ID:        "20160525152356",
		RuleID:    1,
		ProjectID: 1,
		Level:     2,
		Name:      "foo",
		Stamp:     1452758773,
	}
	util.Must(t, db.Put(ew) == nil)
	// Must in db.
	gdb := db.pool[0].db
	ew1 := &EventWrapper{}
	util.Must(t, gdb.Where("id = ?", ew.ID).First(ew1).Error == nil)
	t.Logf("%v \n %v", ew, ew1)
	util.Must(t, reflect.DeepEqual(ew, ew1))
}
