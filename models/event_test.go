// Copyright 2016 Eleme Inc. All rights reserved.

package models

import (
	"github.com/eleme/banshee/util/assert"
	"testing"
)

func TestGenerateID(t *testing.T) {
	// Metric with the same name but different stamps.
	ev1 := NewEvent(&Metric{Name: "foo", Stamp: 1456815973}, nil)
	ev2 := NewEvent(&Metric{Name: "foo", Stamp: 1456815974}, nil)
	assert.Ok(t, ev1.ID != ev2.ID)
	// Metric with the same stamp but different names.
	ev1 = NewEvent(&Metric{Name: "foo", Stamp: 1456815973}, nil)
	ev2 = NewEvent(&Metric{Name: "bar", Stamp: 1456815973}, nil)
	assert.Ok(t, ev1.ID != ev2.ID)
}
