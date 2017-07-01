// Copyright 2015 Eleme Inc. All rights reserved.

package admindb

import (
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util/log"
	_ "github.com/go-sql-driver/mysql" // Import but no use
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3" // Import but no use
)

// SQL db dialect
const dialect = "mysql"

// Gorm logging?
const gormLogMode = false

// DB handles admin storage.
type DB struct {
	// DB
	db *gorm.DB
	// Cache
	RulesCache *rulesCache
}

// Open DB by fileName.
func Open(gdb *gorm.DB) (*DB, error) {
	db := new(DB)
	db.db = gdb
	// Migration
	if err := db.migrate(); err != nil {
		return nil, err
	}
	// Cache
	db.RulesCache = newRulesCache()
	if err := db.RulesCache.Init(db.db); err != nil {
		return nil, err
	}
	// Log Mode
	db.db.LogMode(gormLogMode)
	return db, nil
}

// Close DB.
func (db *DB) Close() error {
	return db.db.Close()
}

// DB returns db handle.
func (db *DB) DB() *gorm.DB {
	return db.db
}

// migrate db schema.
func (db *DB) migrate() error {
	log.Debugf("migrate sql schemas..")
	rule := &models.Rule{}
	user := &models.User{}
	proj := &models.Project{}
	team := &models.Team{}
	webHook := &models.WebHook{}
	return db.db.AutoMigrate(rule, user, proj, team, webHook).Error

}
