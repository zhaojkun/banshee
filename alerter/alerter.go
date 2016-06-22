// Copyright 2015 Eleme Inc. All rights reserved.

package alerter

import (
	"encoding/json"
	"errors"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/eleme/banshee/config"
	"github.com/eleme/banshee/health"
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/storage"
	"github.com/eleme/banshee/storage/eventdb"
	"github.com/eleme/banshee/util/log"
	"github.com/eleme/banshee/util/safemap"
)

// bufferedEventsLimit is the limitation of the number of events waiting to be
// processed by alerter.
const bufferedEventsLimit = 10 * 1024 // 10k

// Alerter is the alert service abstraction.
type Alerter struct {
	cfg          *config.Config
	db           *storage.DB
	In           chan *models.Event
	alertAts     *safemap.SafeMap
	alertNums    *safemap.SafeMap
	alertRecords *safemap.SafeMap
	lock         *sync.RWMutex
}

// New creates a new Alerter.
func New(cfg *config.Config, db *storage.DB) *Alerter {
	return &Alerter{
		cfg:          cfg,
		db:           db,
		In:           make(chan *models.Event, bufferedEventsLimit),
		alertAts:     safemap.New(),
		alertNums:    safemap.New(),
		alertRecords: safemap.New(),
		lock:         &sync.RWMutex{},
	}
}

// initAlerterNumsReseter starts a ticker to reset alert numbers by one day.
func (al *Alerter) initAlerterNumsReseter() {
	go func() {
		ticker := time.NewTicker(time.Hour * 24)
		for _ = range ticker.C {
			al.alertNums.Clear()
		}
	}()
}

// Start workers to wait for events.
func (al *Alerter) Start() {
	log.Infof("start %d alerter workers..", al.cfg.Alerter.Workers)
	for i := 0; i < al.cfg.Alerter.Workers; i++ {
		go al.work()
	}
	al.initAlerterNumsReseter()
}

// hourInRange returns true if given hour is in range [start, end).
// Examples:
//	hourInRange(10, 7, 19) // true
//	hourInRange(10, 20, 19) // false
func hourInRange(hour, start, end int) bool {
	switch {
	case start < end:
		return start <= hour && hour < end
	case start > end:
		return start <= hour || hour < end
	case start == end:
		return start == hour
	}
	return false
}

// shouldProjBeSilent returns true if given project should be silent at this
// time.
func (al *Alerter) shoudProjBeSilent(proj *models.Project) bool {
	var start, end int
	if proj.EnableSilent {
		start = proj.SilentTimeStart
		end = proj.SilentTimeEnd
	} else {
		start = al.cfg.Alerter.DefaultSilentTimeRange[0]
		end = al.cfg.Alerter.DefaultSilentTimeRange[1]
	}
	now := time.Now().Hour()
	return hourInRange(now, start, end)
}

// execCommand executes configured command with configured timeout.
func (al *Alerter) execCommand(ew *models.EventWrapper) (err error) {
	b, _ := json.Marshal(ew) // No error would occur
	done := make(chan error)
	cmd := exec.Command(al.cfg.Alerter.Command, string(b))
	go func() {
		done <- cmd.Run()
	}()
	timeout := time.After(time.Second * time.Duration(al.cfg.Alerter.ExecCommandTimeout))
	select {
	case <-timeout:
		defer func() {
			go func() {
				<-done // Avoid goroutine links
			}()
		}()
		if cmd.Process == nil {
			return // May exit
		}
		if err = cmd.Process.Kill(); err != nil {
			return
		}
		return errors.New("command timed out, killed")
	case err = <-done:
		return
	}
}

// getUniversalUsers returns universal users.
func (al *Alerter) getUniversalUsers() (univs []models.User, err error) {
	if err = al.db.Admin.DB().Where("universal = ?", true).Find(&univs).Error; err != nil {
		return
	}
	return
}

// checkOneDayAlerts returns true if given metric exceeds the one day
// limit.
func (al *Alerter) checkOneDayAlerts(m *models.Metric) bool {
	v, ok := al.alertNums.Get(m.Name)
	if ok && atomic.LoadUint32(v.(*uint32)) > al.cfg.Alerter.OneDayLimit {
		return true
	}
	return false
}

// incrAlertNum increases alert number by 1 for given metric.
func (al *Alerter) incrAlertNum(m *models.Metric) {
	v, ok := al.alertNums.Get(m.Name)
	if !ok {
		n := uint32(1)
		al.alertNums.Set(m.Name, &n)
		return
	}
	atomic.AddUint32(v.(*uint32), 1)
}

// checkAlertCount returns true if given metric has issued an alert
// with in a minimal given period.
func (al *Alerter) checkAlertCount(m *models.Metric) bool {
	if al.cfg.Alerter.NotifyAfter <= 0 {
		return false
	}
	al.lock.RLock()
	defer al.lock.RUnlock()
	v, ok := al.alertRecords.Get(m.Name)
	if !ok {
		return true
	}
	alerted := 0
	for _, timeStamp := range v.([]uint32) {
		if timeStamp > 0 && m.Stamp-timeStamp < al.cfg.Alerter.AlertCheckInterval {
			alerted++
		}
	}
	return alerted < al.cfg.Alerter.NotifyAfter
}

// checkAlertAt returns true if given metric still not reaches the minimal
// alert interval.
func (al *Alerter) checkAlertAt(m *models.Metric) bool {
	v, ok := al.alertAts.Get(m.Name)
	return ok && m.Stamp < v.(uint32)+al.cfg.Alerter.Interval
}

// setAlertRecord sets the alert record for given metric.
func (al *Alerter) setAlertRecord(m *models.Metric) {
	var records []uint32
	al.lock.Lock()
	defer al.lock.Unlock()
	v, ok := al.alertRecords.Get(m.Name)
	if ok {
		records = v.([]uint32)
	} else {
		records = make([]uint32, 10)
	}
	if len(records) >= 10 {
		records = append(records[1:], m.Stamp)
	}
	al.alertRecords.Set(m.Name, records)
}

// setAlertAt sets the alert timestamp for given metric.
func (al *Alerter) setAlertAt(m *models.Metric) {
	al.alertAts.Set(m.Name, m.Stamp)
}

// getProjByRule returns the project for given rule.
func (al *Alerter) getProjByRule(rule *models.Rule) (proj *models.Project, err error) {
	proj = &models.Project{}
	if err = al.db.Admin.DB().Model(rule).Related(proj).Error; err != nil {
		return
	}
	return
}

// getUsersByProj returns the users for given project.
func (al *Alerter) getUsersByProj(proj *models.Project) (users []models.User, err error) {
	var univs []models.User
	if univs, err = al.getUniversalUsers(); err != nil {
		return
	}
	if err = al.db.Admin.DB().Model(proj).Related(&users, "Users").Error; err != nil {
		return
	}
	users = append(users, univs...)
	return
}

// storeEvent stores an event into db.
func (al *Alerter) storeEvent(ev *models.Event) (err error) {
	if err = al.db.Event.Put(eventdb.NewEventWrapper(ev)); err != nil {
		return
	}
	return
}

// work waits for events to alert.
func (al *Alerter) work() {
	for {
		ev := <-al.In
		ew := models.NewWrapperOfEvent(ev) // Avoid locks
		if al.checkAlertAt(ew.Metric) {    // Check alert interval
			continue
		}
		if al.checkOneDayAlerts(ew.Metric) { // Check one day limit
			continue
		}
		al.incrAlertNum(ew.Metric)
		// Store event
		if err := al.storeEvent(ev); err != nil {
			log.Warnf("failed to store event:%v, skipping..", err)
			continue
		}
		// Avoid noises by issuing alerts only when same alert has occurred
		// predefined times.
		if al.checkAlertCount(ew.Metric) {
			al.setAlertRecord(ew.Metric)
			continue
		}
		al.setAlertRecord(ew.Metric)
		// Do alert.
		var err error
		if ew.Project, err = al.getProjByRule(ew.Rule); err != nil {
			continue
		}
		if al.shoudProjBeSilent(ew.Project) {
			continue
		}
		var users []models.User
		if users, err = al.getUsersByProj(ew.Project); err != nil {
			continue
		}
		for _, user := range users {
			ew.User = &user
			if ew.Rule.Level < user.RuleLevel {
				continue
			}
			if len(al.cfg.Alerter.Command) == 0 {
				log.Warnf("alert command not configured")
				continue
			}
			if err = al.execCommand(ew); err != nil { // Execute command
				log.Errorf("exec %s: %v", al.cfg.Alerter.Command, err)
				continue
			}
			log.Infof("send to %s with %s ok", user.Name, ew.Metric.Name)
		}
		if len(users) != 0 {
			al.setAlertAt(ew.Metric)
			health.IncrNumAlertingEvents(1)
		}
	}
}
