// Copyright 2015 Eleme Inc. All rights reserved.

package metricdb

import (
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util"
	"os"
	"reflect"
	"testing"
)

func TestOpen(t *testing.T) {
	fileName := "db-testing"
	db, err := Open(fileName)
	util.Must(t, err == nil)
	util.Must(t, util.IsFileExist(fileName))
	db.Close()
	os.RemoveAll(fileName)
}

func TestPut(t *testing.T) {
	// Open db.
	fileName := "db-testing"
	db, _ := Open(fileName)
	defer os.RemoveAll(fileName)
	defer db.Close()
	// Put.
	m := &models.Metric{
		Name:    "foo",
		Link:    1,
		Stamp:   1452758773,
		Value:   3.14,
		Score:   0.1892,
		Average: 3.133,
	}
	err := db.Put(m)
	util.Must(t, err == nil)
	// Must in db
	key := encodeKey(m)
	value, err := db.db.Get(key, nil)
	util.Must(t, err == nil)
	m1 := &models.Metric{
		Name:  m.Name,
		Stamp: m.Stamp,
		Link:  1,
	}
	err = decodeValue(value, m1)
	util.Must(t, err == nil)
	util.Must(t, reflect.DeepEqual(m, m1))
}

func TestGet(t *testing.T) {
	// Open db.
	fileName := "db-testing"
	db, _ := Open(fileName)
	defer os.RemoveAll(fileName)
	defer db.Close()
	// Nothing.
	ms, err := db.Get("not-exist", 1234, 0, 1452758773)
	util.Must(t, err == nil)
	util.Must(t, len(ms) == 0)
	// Put some.
	db.Put(&models.Metric{Name: "foo", Link: 1, Stamp: 1452758723})
	db.Put(&models.Metric{Name: "foo", Link: 1, Stamp: 1452758733, Value: 1.89, Score: 1.12, Average: 1.72})
	db.Put(&models.Metric{Name: "foo", Link: 1, Stamp: 1452758743})
	db.Put(&models.Metric{Name: "foo", Link: 1, Stamp: 1452758753})
	// Get again.
	ms, err = db.Get("foo", 1, 1452758733, 1452758753)
	util.Must(t, err == nil)
	util.Must(t, len(ms) == 2)
	// Test the value.
	m := ms[0]
	util.Must(t, m.Value == 1.89 && m.Score == 1.12 && m.Link == 1)
}

func TestDelete(t *testing.T) {
	// Open db.
	fileName := "db-testing"
	db, _ := Open(fileName)
	defer os.RemoveAll(fileName)
	defer db.Close()
	// Nothing.
	n, err := db.Delete(1222222, 0, 1452758773)
	util.Must(t, err == nil && n == 0)
	// Put some.
	db.Put(&models.Metric{Name: "foo", Link: 1, Stamp: 1452758723})
	db.Put(&models.Metric{Name: "foo", Link: 1, Stamp: 1452758733})
	db.Put(&models.Metric{Name: "foo", Link: 1, Stamp: 1452758743})
	db.Put(&models.Metric{Name: "foo", Link: 1, Stamp: 1452758753})
	// Delete again
	n, err = db.Delete(1, 1452758733, 1452758753)
	util.Must(t, err == nil && n == 2)
	// Get
	ms, err := db.Get("foo", 1, 1452758723, 1452758763)
	util.Must(t, len(ms) == 2)
	util.Must(t, ms[0].Stamp == 1452758723)
	util.Must(t, ms[1].Stamp == 1452758753)
}
