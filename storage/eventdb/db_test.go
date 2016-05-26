// Copyright 2016 Eleme Inc. All rights reserved.

package eventdb

import (
	"github.com/eleme/banshee/util"
	"os"
	"path"
	"reflect"
	"strconv"
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
	util.Must(t, reflect.DeepEqual(ew, ew1))
}

func TestGetByProjectID(t *testing.T) {
	// Open db.
	fileName := "db-testing"
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	db, _ := Open(fileName, opts)
	defer os.RemoveAll(fileName)
	defer db.Close()
	// Nothing.
	ews, err := db.GetByProjectID(1, 0, 0, 1464162569)
	util.Must(t, err == nil)
	util.Must(t, len(ews) == 0)
	// Put some.
	db.Put(&EventWrapper{ID: "20160525155330.1", RuleID: 1, ProjectID: 1, Level: 2, Name: "foo", Stamp: 1464162569})
	db.Put(&EventWrapper{ID: "20160525155330.2", RuleID: 1, ProjectID: 1, Level: 2, Name: "foo", Stamp: 1464162579})
	db.Put(&EventWrapper{ID: "20160525155330.3", RuleID: 1, ProjectID: 1, Level: 2, Name: "foo", Stamp: 1464162589})
	db.Put(&EventWrapper{ID: "20160525155330.4", RuleID: 1, ProjectID: 1, Level: 2, Name: "foo", Stamp: 1464162599})
	// Get again.
	ews, err = db.GetByProjectID(1, 0, 0, 1464162599)
	util.Must(t, err == nil)
	util.Must(t, len(ews) == 3) // right is closed
	// Test the value
	ew := ews[0]
	util.Must(t, ew.ID == "20160525155330.1")
}

func TestGetAcrossStorages(t *testing.T) {
	// Open db.
	fileName := "db-testing"
	opts := &Options{Period: 86400, Expiration: 86400 * 4}
	db, _ := Open(fileName, opts)
	defer os.RemoveAll(fileName)
	defer db.Close()
	// Force creating 4+1 storages.
	base := uint32(time.Now().Unix())
	db.Put(&EventWrapper{ID: "20160525155730.1", RuleID: 1, ProjectID: 1, Level: 2, Name: "foo", Stamp: base})                    // 0
	db.Put(&EventWrapper{ID: "20160525155730.2", RuleID: 1, ProjectID: 1, Level: 2, Name: "foo", Stamp: base + db.opts.Period*1}) // 1
	db.Put(&EventWrapper{ID: "20160525155730.3", RuleID: 1, ProjectID: 1, Level: 2, Name: "foo", Stamp: base + db.opts.Period*2}) // 2
	db.Put(&EventWrapper{ID: "20160525155730.4", RuleID: 1, ProjectID: 1, Level: 2, Name: "foo", Stamp: base + db.opts.Period*3}) // 3
	db.Put(&EventWrapper{ID: "20160525155730.5", RuleID: 1, ProjectID: 1, Level: 2, Name: "foo", Stamp: base + db.opts.Period*4}) // 4
	// Get by project id.
	ews, err := db.GetByProjectID(1, 0, base, base+db.opts.Period*3)
	util.Must(t, err == nil)
	util.Must(t, len(ews) == 3)
	util.Must(t, ews[0].Stamp == base)
	util.Must(t, ews[1].Stamp == base+db.opts.Period*1)
	util.Must(t, ews[2].Stamp == base+db.opts.Period*2)
}

func TestStorageExpire(t *testing.T) {
	// Open db.
	fileName := "db-testing"
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	db, _ := Open(fileName, opts)
	defer os.RemoveAll(fileName)
	defer db.Close()
	// Force creating 7+1 storages.
	base := uint32(time.Now().Unix())
	db.Put(&EventWrapper{ID: "20160525155730.1", RuleID: 1, ProjectID: 1, Level: 2, Name: "foo", Stamp: base})                    // 0
	id := db.pool[0].id                                                                                                           // record the id to be deleted
	db.Put(&EventWrapper{ID: "20160525155730.2", RuleID: 1, ProjectID: 1, Level: 2, Name: "foo", Stamp: base + db.opts.Period*1}) // 1
	db.Put(&EventWrapper{ID: "20160525155730.3", RuleID: 1, ProjectID: 1, Level: 2, Name: "foo", Stamp: base + db.opts.Period*2}) // 2
	db.Put(&EventWrapper{ID: "20160525155730.4", RuleID: 1, ProjectID: 1, Level: 2, Name: "foo", Stamp: base + db.opts.Period*3}) // 3
	db.Put(&EventWrapper{ID: "20160525155730.5", RuleID: 1, ProjectID: 1, Level: 2, Name: "foo", Stamp: base + db.opts.Period*4}) // 4
	db.Put(&EventWrapper{ID: "20160525155730.6", RuleID: 1, ProjectID: 1, Level: 2, Name: "foo", Stamp: base + db.opts.Period*5}) // 5
	db.Put(&EventWrapper{ID: "20160525155730.7", RuleID: 1, ProjectID: 1, Level: 2, Name: "foo", Stamp: base + db.opts.Period*6}) // 6
	util.Must(t, len(db.pool) == 7)
	db.Put(&EventWrapper{ID: "20160525155730.8", RuleID: 1, ProjectID: 1, Level: 2, Name: "foo", Stamp: base + db.opts.Period*7}) // 7
	util.Must(t, len(db.pool) == 8)
	db.Put(&EventWrapper{ID: "20160525155730.9", RuleID: 1, ProjectID: 1, Level: 2, Name: "foo", Stamp: base + db.opts.Period*8}) // 8
	util.Must(t, len(db.pool) == 8)                                                                                               // Full storages: 1,2,3,4,5,6,7
	// Files must be deleted.
	deleteFileName := path.Join(fileName, strconv.FormatUint(uint64(id), 10))
	util.Must(t, !util.IsFileExist(deleteFileName))
}
