// Copyright 2016 Eleme Inc. All rights reserved.

package detector

import (
	"github.com/eleme/banshee/config"
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util"
	"testing"
)

func TestFill0Issue470(t *testing.T) {
	// Case https://github.com/eleme/banshee/issues/470
	cfg := config.New()
	d := &Detector{cfg: cfg}
	ms := []*models.Metric{
		&models.Metric{Stamp: 80, Value: 80},
		&models.Metric{Stamp: 90, Value: 90},
		&models.Metric{Stamp: 120, Value: 120},
	}
	start, stop := uint32(60), uint32(150)
	excepted := []float64{80, 90, 0, 0, 120, 0, 0}
	actually := d.fill0(ms, start, stop)
	util.Must(t, len(actually) == len(excepted))
	for i := 0; i < len(excepted); i++ {
		util.Must(t, excepted[i] == actually[i])
	}
}
