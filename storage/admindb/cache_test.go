// Copyright 2015 Eleme Inc. All rights reserved.

package admindb

import (
	"os"
	"testing"

	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util"
	"github.com/jinzhu/gorm"
)

func TestInit(t *testing.T) {
	fileName := "db-testing"
	gdb, _ := gorm.Open("sqlite3", fileName)
	db, _ := Open(gdb)
	defer db.Close()
	defer os.RemoveAll(fileName)
	rule1 := &models.Rule{Pattern: "a.b.*"}
	rule2 := &models.Rule{Pattern: "a.b.*.c"}
	rule3 := &models.Rule{Pattern: "a.*.c.d"}
	// Add to db.
	db.DB().Create(rule1)
	db.DB().Create(rule2)
	db.DB().Create(rule3)
	// Clear cache.
	db.RulesCache.rules.Clear()
	// Reload
	util.Must(t, nil == db.RulesCache.Init(db.DB()))
	// Get rule
	r1, ok := db.RulesCache.Get(rule1.ID)
	util.Must(t, ok)
	util.Must(t, r1.Pattern == rule1.Pattern)
	r2, ok := db.RulesCache.Get(rule2.ID)
	util.Must(t, ok)
	util.Must(t, r2.Pattern == rule2.Pattern)
	r3, ok := db.RulesCache.Get(rule3.ID)
	util.Must(t, ok)
	util.Must(t, r3.Pattern == rule3.Pattern)
}
