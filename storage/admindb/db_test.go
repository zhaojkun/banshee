// Copyright 2015 Eleme Inc. All rights reserved.

package admindb

import (
	"os"
	"testing"

	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util"
	"github.com/jinzhu/gorm"
)

func TestOpen(t *testing.T) {
	fileName := "db-testing"
	gdb, _ := gorm.Open("sqlite3", fileName)
	db, err := Open(gdb)
	util.Must(t, nil == err)
	util.Must(t, db != nil)
	util.Must(t, util.IsFileExist(fileName))
	defer os.RemoveAll(fileName)
	defer db.Close()
	util.Must(t, db.DB().HasTable(&models.User{}))
	util.Must(t, db.DB().HasTable(&models.Rule{}))
	util.Must(t, db.DB().HasTable(&models.Project{}))
}
