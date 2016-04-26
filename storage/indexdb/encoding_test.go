// Copyright 2015 Eleme Inc. All rights reserved.

package indexdb

import (
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util"
	"testing"
)

func TestEncoding(t *testing.T) {
	idx := &models.Index{Stamp: 1450426828, Score: 0.678888, Average: 877.234}
	value := encode(idx)
	idx1 := &models.Index{}
	err := decode(value, idx1)
	util.Must(t, err == nil)
	util.Must(t, idx1.Stamp == idx.Stamp)
	util.Must(t, idx1.Score == 0.678888)
	util.Must(t, idx1.Average == 877.234)
}
