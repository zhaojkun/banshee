// Copyright 2015 Eleme Inc. All rights reserved.

package metricdb

import (
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util"
	"reflect"
	"testing"
)

func TestEncodingKey(t *testing.T) {
	m := &models.Metric{
		Link:  1,
		Stamp: 1452758773,
	}
	key := encodeKey(m)
	m1 := &models.Metric{}
	util.Must(t, nil == decodeKey(key, m1))
	util.Must(t, reflect.DeepEqual(m, m1))
}

func TestEncodingValue(t *testing.T) {
	m := &models.Metric{
		Value:   3.1415926,
		Score:   0.1892,
		Average: 3.1333333,
	}
	value := encodeValue(m)
	m1 := &models.Metric{}
	util.Must(t, nil == decodeValue(value, m1))
	util.Must(t, reflect.DeepEqual(m, m1))
}
