// Copyright 2015 Eleme Inc. All rights reserved.

package safemap

import (
	"github.com/eleme/banshee/util"
	"testing"
)

func TestBasic(t *testing.T) {
	m := New()
	// Set
	m.Set("key1", "val1")
	m.Set("key2", "val2")
	m.Set("key3", "val3")
	util.Must(t, m.Len() == 3)
	// Get
	val1, ok := m.Get("key1")
	util.Must(t, ok)
	util.Must(t, val1 == "val1")
	// Items
	util.Must(t, m.Items()["key1"] == "val1")
	util.Must(t, m.Items()["key2"] == "val2")
	util.Must(t, m.Items()["key3"] == "val3")
	// Len
	util.Must(t, m.Len() == 3)
	// Delete
	util.Must(t, m.Delete("key1"))
	util.Must(t, !m.Delete("key-not-exist"))
	util.Must(t, m.Len() == 2)
	// Clear
	m.Clear()
	util.Must(t, m.Len() == 0)
}
