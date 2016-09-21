// Copyright 2015 Eleme Inc. All rights reserved.

package alerter

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"sync/atomic"
	"time"

	"github.com/eleme/banshee/config"
	"github.com/eleme/banshee/health"
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/storage"
	"github.com/eleme/banshee/util/log"
	"github.com/eleme/banshee/util/safemap"
)

const (
	// Limit for buffered detected metric results, further results will be dropped
	// if this limit is reached.
	bufferedMetricResultsLimit = 10 * 1024
	// Exec command timeout in second
	execCommandTimeout = 5 * time.Second
)

// Alerter alerts on anomalies detected.
type Alerter struct {
	// Storage
	db *storage.DB
	// Config
	cfg *config.Config
	// Input
	In chan *models.Event
	// Alertings stamps
	m *safemap.SafeMap
	// Alertings counters
	c *safemap.SafeMap
}

// New creates a alerter.
func New(cfg *config.Config, db *storage.DB) *Alerter {
	al := new(Alerter)
	al.cfg = cfg
	al.db = db
	al.In = make(chan *models.Event, bufferedMetricResultsLimit)
	al.m = safemap.New()
	al.c = safemap.New()
	return al
}

// Start several goroutines to wait for detected metrics, then check each
// metric with all the rules, the configured shell command will be executed
// once a rule is hit.
func (al *Alerter) Start() {
	log.Infof("start %d alerter workers..", al.cfg.Alerter.Workers)
	for i := 0; i < al.cfg.Alerter.Workers; i++ {
		go al.work()
	}
	go func() {
		ticker := time.NewTicker(time.Hour * 24)
		for _ = range ticker.C {
			al.c.Clear()
		}
	}()
}

// Test if an hour is in [start, end)
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

// Test if alerter should silent now for a project.
func (al *Alerter) shouldSilent(proj *models.Project) bool {
	var start, end int
	if proj.EnableSilent {
		// Use project defined.
		start = proj.SilentTimeStart
		end = proj.SilentTimeEnd
	} else {
		// Default
		start = al.cfg.Alerter.DefaultSilentTimeRange[0]
		end = al.cfg.Alerter.DefaultSilentTimeRange[1]
	}
	now := time.Now().Hour()
	return hourInRange(now, start, end)
}

// execute command with event within certain timeout.
func (al *Alerter) execCommand(ev *models.Event) error {
	b, _ := json.Marshal(ev)
	arg := string(b)
	done := make(chan error)
	cmd := exec.Command(al.cfg.Alerter.Command, arg)
	go func() {
		done <- cmd.Run()
	}()
	timeout := time.After(execCommandTimeout)
	select {
	case <-timeout:
		err := cmd.Process.Kill()
		if err == nil {
			err = errors.New("command timed out, killed")
		} else {
			s := fmt.Sprintf("failed to kill command: %v", err)
			err = errors.New(s)
		}
		go func() {
			<-done // exit the prev goroutine
		}()
		return err
	case err := <-done:
		return err
	}
}

// execute webhook
func (al *Alerter) triggerWebHook(ev *models.Event) error {

}

// work waits for detected metrics, then check each metric with all the
// rules, the configured shell command will be executed once a rule is hit.
func (al *Alerter) work() {
	for {
		ev := <-al.In
		// Check interval.
		v, ok := al.m.Get(ev.Metric.Name)
		if ok && ev.Metric.Stamp-v.(uint32) < al.cfg.Alerter.Interval {
			continue
		}
		// Check alert times in one day
		v, ok = al.c.Get(ev.Metric.Name)
		if ok && atomic.LoadUint32(v.(*uint32)) > al.cfg.Alerter.OneDayLimit {
			log.Warnf("%s hit alerting one day limit, skipping..", ev.Metric.Name)
			continue
		}
		if !ok {
			var newCounter uint32
			newCounter = 1
			al.c.Set(ev.Metric.Name, &newCounter)
		} else {
			atomic.AddUint32(v.(*uint32), 1)
		}
		// Universals
		var univs []models.User
		if err := al.db.Admin.DB().Where("universal = ?", true).Find(&univs).Error; err != nil {
			log.Errorf("get universal users: %v, skiping..", err)
			continue
		}
		for _, rule := range ev.Metric.TestedRules {
			ev.Rule = rule
			ev.TranslateRuleComment()
			// Project
			proj := &models.Project{}
			if err := al.db.Admin.DB().Model(rule).Related(proj).Error; err != nil {
				log.Errorf("project, %v, skiping..", err)
				continue
			}
			ev.Project = proj
			// Silent
			if al.shouldSilent(proj) {
				continue
			}
			// Users
			var users []models.User
			if err := al.db.Admin.DB().Model(proj).Related(&users, "Users").Error; err != nil {
				log.Errorf("get users: %v, skiping..", err)
				continue
			}
			users = append(users, univs...)
			// Send
			for _, user := range users {
				ev.User = &user
				if rule.Level < user.RuleLevel {
					continue
				}
				// Exec
				if len(al.cfg.Alerter.Command) == 0 {
					log.Warnf("alert command not configured")
					continue
				}
				if err := al.execCommand(ev); err != nil {
					log.Errorf("exec %s: %v", al.cfg.Alerter.Command, err)
					continue
				}
				log.Infof("send message to %s with %s ok", user.Name, ev.Metric.Name)
			}
			if len(users) != 0 {
				al.m.Set(ev.Metric.Name, ev.Metric.Stamp)
				health.IncrNumAlertingEvents(1)
			}
		}
	}
}
