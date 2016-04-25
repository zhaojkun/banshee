// Copyright 2015 Eleme Inc. All rights reserved.

package admindb

import (
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util"
	"os"
	"testing"
)

func TestOpen(t *testing.T) {
	fileName := "db-testing"
	db, err := Open(fileName)
	util.Must(t, nil == err)
	util.Must(t, db != nil)
	util.Must(t, util.IsFileExist(fileName))
	defer os.RemoveAll(fileName)
	defer db.Close()
	util.Must(t, db.DB().HasTable(&models.User{}))
	util.Must(t, db.DB().HasTable(&models.Rule{}))
	util.Must(t, db.DB().HasTable(&models.Project{}))
}
