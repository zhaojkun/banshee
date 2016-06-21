// Copyright 2016 Eleme Inc. All rights reserved.

package alerter

import (
	"testing"

	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util"
	"github.com/eleme/banshee/util/safemap"
)

func TestHourInRange(t *testing.T) {
	util.Must(t, hourInRange(3, 0, 6))
	util.Must(t, !hourInRange(7, 0, 6))
	util.Must(t, !hourInRange(6, 0, 6))
	util.Must(t, hourInRange(23, 19, 10))
	util.Must(t, hourInRange(6, 19, 10))
	util.Must(t, !hourInRange(13, 19, 10))
}

func TestAlertRecord(t *testing.T) {
	a := &Alerter{alertRecords: safemap.New()}
	metrics := &models.Metric{Name: "test", Stamp: 80, Value: 80}
	util.Must(t, hourInRange(23, 19, 10))
	util.Must(t, !a.checkAlertCount(metrics))
	metrics.Stamp = 81
	a.setAlertRecord(metrics)
	metrics.Stamp = 82
	a.setAlertRecord(metrics)
	metrics.Stamp = 83
	a.setAlertRecord(metrics)
	metrics.Stamp = 84
	a.setAlertRecord(metrics)
	util.Must(t, a.checkAlertCount(metrics))
}
