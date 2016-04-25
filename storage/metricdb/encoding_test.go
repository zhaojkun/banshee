// Copyright 2015 Eleme Inc. All rights reserved.

package metricdb

import (
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util"
	"testing"
)

func TestEncodeKey(t *testing.T) {
	m := &models.Metric{Name: "foo", Stamp: horizon + 0xf}
	key := encodeKey(m)
	s := "foo000000f"
	util.Must(t, s == string(key))
}

func TestDecodeKey(t *testing.T) {
	key := []byte("foo000001f")
	m := &models.Metric{}
	err := decodeKey(key, m)
	util.Must(t, err == nil)
	util.Must(t, m.Name == "foo")
	util.Must(t, m.Stamp == 36+0xf+horizon)
}

func TestStampLenEnoughToUse(t *testing.T) {
	stamp := uint32(90*365*24*60*60) + horizon
	m := &models.Metric{Name: "foo", Stamp: stamp}
	key := encodeKey(m)
	n := &models.Metric{}
	err := decodeKey(key, n)
	util.Must(t, err == nil)
	util.Must(t, n.Name == m.Name)
	util.Must(t, n.Stamp == m.Stamp)
}

func TestEncodeValue(t *testing.T) {
	m := &models.Metric{Value: 1.23, Score: 0.72, Average: 0.798766}
	value := encodeValue(m)
	s := "1.23:0.72:0.79877"
	util.Must(t, s == string(value))
}

func TestDecodeValue(t *testing.T) {
	m := &models.Metric{}
	value := []byte("1.23:0.72:0.79")
	err := decodeValue(value, m)
	util.Must(t, err == nil)
	util.Must(t, m.Value == 1.23)
	util.Must(t, m.Score == 0.72)
	util.Must(t, m.Average == 0.79)
}
