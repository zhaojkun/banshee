// Copyright 2015 Eleme Inc. All rights reserved.

package detector

import (
	"bufio"
	"fmt"
	"net"
	"path/filepath"
	"sync"
	"time"

	"github.com/eleme/banshee/algorithm"
	"github.com/eleme/banshee/config"
	"github.com/eleme/banshee/filter"
	"github.com/eleme/banshee/health"
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/storage"
	"github.com/eleme/banshee/storage/indexdb"
	"github.com/eleme/banshee/util"
	"github.com/eleme/banshee/util/log"
)

// Detector is to detect anomalies.
type Detector struct {
	cfg  *config.Config
	db   *storage.DB
	flt  *filter.Filter
	outs []chan *models.Event
	// Idle tracking
	idleMA map[string]int64 // indexName:trackTimes
	idleMB map[string]int64 // indexName:trackTimes
	lock   sync.Mutex       // protects idleM*
}

// New creates a new Detector.
func New(cfg *config.Config, db *storage.DB, flt *filter.Filter) *Detector {
	return &Detector{
		cfg:    cfg,
		db:     db,
		flt:    flt,
		outs:   make([]chan *models.Event, 0),
		idleMA: make(map[string]int64, 0),
		idleMB: make(map[string]int64, 0),
	}
}

// Out adds a chan to receive detection results.
func (d *Detector) Out(ch chan *models.Event) {
	d.outs = append(d.outs, ch)
}

// output detected metrics to all chans in outs.
// Skip if the target chan is full.
func (d *Detector) output(ev *models.Event) {
	for _, ch := range d.outs {
		select {
		case ch <- ev:
		default:
			log.Errorf("output channel is full, skipping..")
			continue
		}
	}
}

// Start the tcp server.
func (d *Detector) Start() {
	addr := fmt.Sprintf("0.0.0.0:%d", d.cfg.Detector.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	log.Infof("detector is listening on %s", addr)
	go d.startIdleTracking()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Errorf("cannot accept conn: %v, skipping..", err)
			continue
		}
		go d.handle(conn)
	}
}

// handle a new connection:
// Steps:
//	1. Read input from connection line by line.
//	2. Parse each line into a metric.
//	3. Validate the metric
//	4. Process the metric.
func (d *Detector) handle(conn net.Conn) {
	addr := conn.RemoteAddr()
	health.IncrNumClients(1)
	defer health.DecrNumClients(1)
	log.Infof("conn %s established", addr)
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() { // Read line by line.
		if err := scanner.Err(); err != nil { // Close on read error.
			log.Errorf("read error: %v, closing conn..", err)
			break
		}
		line := scanner.Text()
		m, err := parseMetric(line) // Parse
		if err != nil {
			log.Errorf("parse error %v : %v, skipping..", line, err)
			continue
		}
		if err = m.Validate(); err != nil {
			log.Debugf("invalid metric %v : %v, skipping..", m.Name, err)
			continue
		}
		d.process(m)
	}
	conn.Close()
	log.Infof("conn %s disconnected", addr)
}

// process the input metric.
// Steps:
//	1. Match metric with all rules.
//	2. Detect the metric with matched rules.
//	3. Output detection results to receivers.
func (d *Detector) process(m *models.Metric) {
	health.IncrNumMetricIncomed(1)
	timer := util.NewTimer() // Detection cost timer
	// Match
	ok, rules := d.match(m)
	if !ok {
		return
	}
	d.adjustIdleM(m, rules)
	// Detect
	evs, err := d.detect(m, rules)
	if err != nil {
		log.Errorf("detect: %v, skipping..", err)
		return
	}
	health.IncrNumMetricDetected(1)
	// Output
	for _, ev := range evs {
		d.output(ev)
	}
	// Time end.
	elapsed := timer.Elapsed()
	if elapsed > float64(d.cfg.Detector.WarningTimeout) {
		log.Warnf("detection is slow: %.2fms", elapsed)
	}
	health.AddDetectionCost(elapsed)
}

func (d *Detector) processIdle(m *models.Metric) {
	ok, rules := d.match(m)
	if !ok {
		return
	}
	idx, err := d.db.Index.Get(m.Name)
	if err != nil {
		return
	}
	m.LinkTo(idx)
	var evs []*models.Event
	for _, rule := range d.test(m, idx, rules) {
		evs = append(evs, models.NewEvent(m, idx, rule))
	}
	for _, ev := range evs {
		d.output(ev)
	}
	return

}

// match a metric with rules, and return matched rules.
// Details:
//	1. If no rules matched, return false.
//	2. If any black patterns matched, return false.
//	3. Else, return true and matched rules.
func (d *Detector) match(m *models.Metric) (bool, []*models.Rule) {
	// Check rules.
	timer := util.NewTimer() // Filter timer
	rules := d.flt.MatchedRules(m, true)
	elapsed := timer.Elapsed()
	health.AddFilterCost(elapsed)
	if len(rules) == 0 { // Hit no rules.
		return false, rules
	}
	// Check blacklist.
	for _, p := range d.cfg.Detector.BlackList {
		ok, err := filepath.Match(p, m.Name)
		if err != nil {
			// Invalid black pattern.
			log.Errorf("invalid black pattern: %s, %v", p, err)
			continue
		}
		if ok {
			// Hit black pattern.
			log.Debugf("%s hit black pattern %s", m.Name, p)
			return false, rules
		}
	}
	return true, rules // OK
}

// adjustIdleM puts a metric name to idle track maps if it should be tracked
// and pops from the maps if it shouldn't.
func (d *Detector) adjustIdleM(m *models.Metric, rules []*models.Rule) {
	b := d.shoudTrackIdle(m, rules)
	d.lock.Lock()
	defer d.lock.Unlock()
	if b { // Should be tracked.
		delete(d.idleMA, m.Name)
		d.idleMB[m.Name] = int64(m.Stamp)
	} else { // Shouldn't be tracked.
		delete(d.idleMA, m.Name)
		delete(d.idleMB, m.Name)
	}
}

// shouldTrackIdle returns true if given metric should be tracked for idle.
// A metric should be tracked for idle states if it matches configured check
// pattern list or its matched rules have an option TrackIdle set.
func (d *Detector) shoudTrackIdle(m *models.Metric, rules []*models.Rule) bool {
	for _, rule := range rules {
		if rule.TrackIdle {
			return true
		}
	}
	isHighLevel := false
	for _, rule := range rules { // IdleMetricCheckList only works for high level rules.
		if rule.Level == models.RuleLevelHigh {
			isHighLevel = true
		}
	}
	for _, p := range d.cfg.Detector.IdleMetricCheckList {
		ok, err := filepath.Match(p, m.Name)
		if err != nil {
			log.Errorf("invalid idleMetricCheck pattern: %s, %v", p, err)
			continue
		}
		if ok && isHighLevel {
			return true
		}
	}
	return false
}

// startIdleTracking starts a goroutine to track idle metrics by interval.
func (d *Detector) startIdleTracking() {
	go func() {
		log.Debugf("starting idle metrics tracking..")
		interval := time.Duration(d.cfg.Detector.IdleMetricCheckInterval) * time.Second
		ticker := time.NewTicker(interval)
		for _ = range ticker.C {
			d.checkIdles()
		}
	}()
}

// checkIdles checks all idle metrics.
func (d *Detector) checkIdles() {
	d.lock.Lock()
	defer d.lock.Unlock()
	now := time.Now().Unix()
	offset := int64(d.cfg.Detector.IdleMetricTrackLimit * d.cfg.Detector.IdleMetricCheckInterval)
	for name, n := range d.idleMA {
		mockedMetric := &models.Metric{
			Name:  name,
			Stamp: uint32(time.Now().Unix()),
			Value: 0,
			Score: 0,
		}
		d.processIdle(mockedMetric)
		// Copy mapA to mapB.
		// Reason: mapA should be much smaller than mapB.
		if now-n < offset {
			d.idleMB[name] = n
		}
	}
	// Rename mapB to mapA and reset mapB.
	d.idleMA = d.idleMB
	d.idleMB = make(map[string]int64, 0)
}

// detect input metric with its matched rules.
// Details:
//	1. Get its index from db.
//	2. If the metric need to be analyzed, analyze it.
//	3. Save its index and the metricinto db.
//	4. Test the metric with matched rules.
//	5. Make events with its tested rules.
func (d *Detector) detect(m *models.Metric, rules []*models.Rule) (evs []*models.Event, err error) {
	var idx *models.Index
	if idx, err = d.db.Index.Get(m.Name); err != nil {
		if err == indexdb.ErrNotFound {
			idx = nil
		} else {
			return nil, err
		}
	}
	idx = d.analyze(idx, m, rules)
	if err = d.save(m, idx); err != nil {
		return
	}
	for _, rule := range d.test(m, idx, rules) {
		evs = append(evs, models.NewEvent(m, idx, rule))
	}
	return
}

// analyze given metric with 3sigma, returns the new index.
// Steps:
//	1. Get index.
//	2. Get history values.
//	3. Do 3sigma calculation.
//	4. Move the index next.
func (d *Detector) analyze(idx *models.Index, m *models.Metric, rules []*models.Rule) *models.Index {
	fz := idx != nil && d.shouldFill0(m, rules)
	if idx != nil {
		m.LinkTo(idx)
	}
	if models.AnyTrendRelated(rules) {
		bms, err := d.values(m, fz)
		if err != nil {
			return nil
		}
		algo.DivDaySigma(m, bms)
	} else {
		m.Average = m.Value
	}
	return d.nextIdx(idx, m, d.pickTrendingFactor(rules))
}

// Test metric and index with rules.
func (d *Detector) test(m *models.Metric, idx *models.Index, rules []*models.Rule) (testedRules []*models.Rule) {
	for _, rule := range rules {
		if rule.Test(m, idx, d.cfg) {
			testedRules = append(testedRules, rule)
		}
	}
	return
}

// Save metric and index into db.
func (d *Detector) save(m *models.Metric, idx *models.Index) error {
	// Save index.
	if err := d.db.Index.Put(idx); err != nil {
		return err
	}
	// Save metric.
	m.LinkTo(idx) // Important
	return d.db.Metric.Put(m)
}

// shouldFill0 returns true if given metric needs to fill blanks with zeros to
// its hidtory values.
// A metric should fill0 if it matches configured fill blank zero patterns and
// the matching rules have no option NeverFillZero set.
func (d *Detector) shouldFill0(m *models.Metric, rules []*models.Rule) bool {
	for _, p := range d.cfg.Detector.FillBlankZeros {
		ok, err := filepath.Match(p, m.Name)
		if err != nil {
			// Invalid pattern.
			log.Errorf("invalid fillBlankZeros pattern: %s, %v", p, err)
			continue
		}
		if ok {
			// Matched the fill zeros patterns, then check its rules.
			for _, rule := range rules {
				if rule.NeverFillZero {
					return false
				}
			}
			return true // OK
		}
	}
	return false
}

// Fill blank with zeros into history values, mainly for dispersed
// metrics such as counters. The start and stop is for periodicity
// reasons.
func (d *Detector) fill0(ms []*models.Metric, start, stop uint32) []*models.Metric {
	i := 0 // record real-metric.
	step := d.cfg.Interval
	var rms []*models.Metric
	for start < stop {
		if i < len(ms) {
			m := ms[i]
			// start is smaller than current stamp.
			for ; start < m.Stamp; start += step {
				if len(rms) >= 1 && rms[0].Value != 0 { // issue#470
					rms = append(rms, &models.Metric{
						Value: 0,
					})
				}
			}
			rms = append(rms, m) //Append a real metric
			i++
		} else { // No more real metric.
			if len(rms) >= 1 && rms[0].Value != 0 { // issue#470
				rms = append(rms, &models.Metric{
					Value: 0,
				})
			}
		}
		start += step
	}
	return rms
}

// Get history values for the input metric, will only fetch the history
// values with the same phase around this timestamp, within an filter
// offset.
func (d *Detector) values(m *models.Metric, fz bool) ([]models.BulkMetric, error) {
	timer := util.NewTimer()
	defer func() {
		elapsed := timer.Elapsed()
		health.AddQueryCost(elapsed)
	}()
	offset := uint32(d.cfg.Detector.FilterOffset * float64(d.cfg.Period))
	expiration := d.cfg.Expiration
	period := d.cfg.Period
	ftimes := d.cfg.Detector.FilterTimes
	// Get values with the same phase.
	n := 0 // number of goroutines to luanch
	ch := make(chan models.BulkMetric)
	var stamp uint32
	if d.cfg.Detector.UsingRecentDataPoints {
		stamp = m.Stamp
	} else {
		stamp = m.Stamp - period
	}
	for ; stamp+expiration > m.Stamp; stamp -= period {
		start := stamp - offset
		stop := stamp + offset
		// Range (m.Stamp,m.Stamp+offset) has no data as it is the future
		if stamp == m.Stamp {
			stop = m.Stamp
		}
		go func() {
			ms, err := d.db.Metric.Get(m.Name, m.Link, start, stop)
			ch <- models.BulkMetric{Err: err, Ms: ms, Start: start, Stop: stop}
		}()
		n++
		if n >= ftimes {
			break
		}
	}
	// Concat chunks.
	var bms []models.BulkMetric
	var err error
	for i := 0; i < n; i++ {
		r := <-ch
		if r.Err != nil {
			// Record error but DONOT return directly.
			// Must receive n times from ch, otherwise the goroutine will
			// be hanged and the ch won't be gc, yet memory leaks.
			err = r.Err
			continue
		}
		if err != nil {
			continue
		}
		// Append to values.
		if fz {
			r.Ms = d.fill0(r.Ms, r.Start, r.Stop)
		}
		bms = append(bms, r)
	}
	return bms, err
}

// nextIdx creates the next index via the weighted exponentia moving average.
//	t[0] = x[1], f: 0~1
//	t[n] = t[n-1] * (1 - f) + f * x[n]
// Index score is the trending description of metric score.
func (d *Detector) nextIdx(idx *models.Index, m *models.Metric, f float64) *models.Index {
	n := &models.Index{Name: m.Name, Stamp: m.Stamp}
	if idx == nil {
		// As first
		n.Score = m.Score
		n.Average = m.Value
		return n
	}
	// Move next
	n.Score = idx.Score*(1-f) + f*m.Score
	n.Average = m.Average
	n.Link = idx.Link
	return n
}

// pickTrendingFactor picks a trending factor by rules levels.
func (d *Detector) pickTrendingFactor(rules []*models.Rule) float64 {
	maxLevel := models.RuleLevelLow
	for _, rule := range rules {
		if rule.Level > maxLevel {
			maxLevel = rule.Level
		}
	}
	if maxLevel == models.RuleLevelLow {
		return d.cfg.Detector.TrendingFactorLowLevel
	} else if maxLevel == models.RuleLevelMiddle {
		return d.cfg.Detector.TrendingFactorMiddleLevel
	} else {
		return d.cfg.Detector.TrendingFactorHighLevel
	}
}
