// Copyright 2016 Eleme Inc. All rights reserved.

package alerter

import (
	"sync"
	"testing"

	"github.com/eleme/banshee/config"
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

func TestAlertRecordAlertNotifyAfterConfigDisabled(t *testing.T) {
	cfg := config.New()
	cfg.Alerter.NotifyAfter = 0
	a := &Alerter{cfg: cfg, alertRecords: safemap.New(), lock: &sync.RWMutex{}}
	ew := models.NewWrapperOfEvent(&models.Event{
		Rule:   &models.Rule{},
		Metric: &models.Metric{Name: "test", Stamp: 0, Value: 80},
	})
	for i := 0; i <= 100; i++ {
		ew.Metric.Stamp = uint32(i)
		util.Must(t, !a.checkAlertCount(ew))
		a.setAlertRecord(ew)
	}
}

func TestAlertRecordAlertNotifyAfterConfigSetNotifyAfterToTwo(t *testing.T) {
	cfg := config.New()
	cfg.Alerter.NotifyAfter = 2
	a := &Alerter{cfg: cfg, alertRecords: safemap.New(), lock: &sync.RWMutex{}}
	ew := models.NewWrapperOfEvent(&models.Event{
		Rule:   &models.Rule{},
		Metric: &models.Metric{Name: "test", Stamp: 80, Value: 80},
	})

	util.Must(t, a.checkAlertCount(ew))
	a.setAlertRecord(ew)
	ew.Metric.Stamp = 81
	util.Must(t, a.checkAlertCount(ew))
	a.setAlertRecord(ew)

	ew.Metric.Stamp = 82
	util.Must(t, !a.checkAlertCount(ew))
	a.setAlertRecord(ew)

}

func TestAlertRecordAlertNotifyAfterConfigSetNotifyAfterToOne(t *testing.T) {
	cfg := config.New()
	cfg.Alerter.NotifyAfter = 1
	a := &Alerter{cfg: cfg, alertRecords: safemap.New(), lock: &sync.RWMutex{}}
	ew := models.NewWrapperOfEvent(&models.Event{
		Rule:   &models.Rule{},
		Metric: &models.Metric{Name: "test", Stamp: 80, Value: 80},
	})
	util.Must(t, a.checkAlertCount(ew))
	a.setAlertRecord(ew)
	ew.Metric.Stamp = 81
	util.Must(t, !a.checkAlertCount(ew))
	a.setAlertRecord(ew)
}

func TestAlertRecordAlertNotifyWithDifferentRule(t *testing.T) {
	cfg := config.New()
	cfg.Alerter.NotifyAfter = 1
	a := &Alerter{cfg: cfg, alertRecords: safemap.New(), lock: &sync.RWMutex{}}
	ew := models.NewWrapperOfEvent(&models.Event{
		Rule:   &models.Rule{},
		Metric: &models.Metric{Name: "test", Stamp: 80, Value: 80},
	})
	util.Must(t, a.checkAlertCount(ew))
	a.setAlertRecord(ew)
	util.Must(t, !a.checkAlertCount(ew))
	ew.Rule.ID = 2
	util.Must(t, a.checkAlertCount(ew))
}
