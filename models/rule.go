// Copyright 2016 Eleme Inc. All rights reserved.

package models

import (
	"path/filepath"
	"time"

	"github.com/eleme/banshee/config"
)

// Rule Levels
const (
	RuleLevelLow = iota
	RuleLevelMiddle
	RuleLevelHigh
)

// Rule types
const (
	RULEADD    = "add"
	RULEDELETE = "delete"
)

// Message is a add or delete event.
type Message struct {
	Type string `json:"type"`
	Rule *Rule  `json:"rule"`
}

// Rule is a type to describe alerter rule.
type Rule struct {
	// Rule may be cached.
	cache `sql:"-" json:"-"`
	// ID in db.
	ID int `gorm:"primary_key" json:"id"`
	// Project belongs to
	ProjectID int `sql:"index;not null" json:"projectID"`
	// Pattern is a wildcard string
	Pattern string `sql:"size:400;not null" json:"pattern"`
	// Trend
	TrendUp   bool `json:"trendUp"`
	TrendDown bool `json:"trendDown"`
	// Optional thresholds data, only used if the rule condition is about
	// threshold. Although we really don't need any thresholds for trending
	// analyzation and alertings, but we still provide a way to alert by
	// thresholds.
	ThresholdMax float64 `json:"thresholdMax"`
	ThresholdMin float64 `json:"thresholdMin"`
	// Number of metrics matched.
	NumMetrics int `sql:"-" json:"numMetrics"`
	// Comment
	Comment string `sql:"type:varchar(256)" json:"comment"`
	// Level
	Level int `json:"level"`
	// Disabled
	Disabled bool `sql:"default:false" json:"disabled"`
	// DisabledFor
	DisabledFor int `json:"disabledFor"`
	// DisabledAt
	DisabledAt time.Time `json:"disabledAt"`
	// TrackIdle
	TrackIdle bool `json:"trackIdle"`
	// FillZero
	NeverFillZero bool `sql:"default:false" json:"neverFillZero"`
}

// Copy the rule.
func (rule *Rule) Copy() *Rule {
	dst := &Rule{}
	rule.CopyTo(dst)
	return dst
}

// CopyTo copy the rule to another.
func (rule *Rule) CopyTo(r *Rule) {
	rule.RLock()
	defer rule.RUnlock()
	r.Lock()
	defer r.Unlock()
	r.ID = rule.ID
	r.ProjectID = rule.ProjectID
	r.Pattern = rule.Pattern
	r.TrendUp = rule.TrendUp
	r.TrendDown = rule.TrendDown
	r.ThresholdMax = rule.ThresholdMax
	r.ThresholdMin = rule.ThresholdMin
	r.Comment = rule.Comment
	r.Level = rule.Level
	r.Disabled = rule.Disabled
	r.DisabledFor = rule.DisabledFor
	r.DisabledAt = rule.DisabledAt
	r.TrackIdle = rule.TrackIdle
}

// Equal tests rule equality
func (rule *Rule) Equal(r *Rule) bool {
	rule.RLock()
	defer rule.RUnlock()
	r.RLock()
	defer rule.RUnlock()
	return (r.ID == rule.ID &&
		r.ProjectID == rule.ProjectID &&
		r.Pattern == rule.Pattern &&
		r.TrendUp == rule.TrendUp &&
		r.TrendDown == rule.TrendDown &&
		r.ThresholdMax == rule.ThresholdMax &&
		r.ThresholdMin == rule.ThresholdMin &&
		r.Comment == rule.Comment &&
		r.Level == rule.Level &&
		r.Disabled == rule.Disabled &&
		r.DisabledFor == rule.DisabledFor &&
		r.DisabledAt.Equal(rule.DisabledAt) &&
		r.TrackIdle == rule.TrackIdle)
}

// IsTrendRelated returns true if any trend options is set.
func (rule *Rule) IsTrendRelated() bool {
	rule.RLock()
	defer rule.RUnlock()
	return rule.TrendUp || rule.TrendDown
}

// Test if a metric hits this rule.
//
//	1. For trend related conditions, index.Score will be used.
//	2. For value related conditions, metric.Value will be used.
//
func (rule *Rule) Test(m *Metric, idx *Index, cfg *config.Config) bool {
	// RLock if shared.
	rule.RLock()
	defer rule.RUnlock()
	if rule.Disabled { // Disable for a while
		if rule.DisabledFor <= 0 { // Disable forever
			return false
		}
		disabledBefore := rule.DisabledAt.Add(time.Duration(rule.DisabledFor) * time.Minute)
		if time.Now().Before(disabledBefore) { // Disabled for a while
			return false
		}
	}
	if rule.TrackIdle {
		if m.Value == 0 && m.Average == 0 && m.Score == 0 {
			return true
		}
	}
	// Default thresholds.
	var defaultThresholdMax float64
	var defaultThresholdMin float64
	if cfg != nil {
		// Check defaults
		for p, v := range cfg.Detector.DefaultThresholdMaxs {
			if ok, _ := filepath.Match(p, m.Name); ok && v != 0 {
				defaultThresholdMax = v
				break
			}
		}
	}
	if cfg != nil {
		// Check defaults
		for p, v := range cfg.Detector.DefaultThresholdMins {
			if ok, _ := filepath.Match(p, m.Name); ok && v != 0 {
				defaultThresholdMin = v
				break
			}
		}
	}
	// Conditions
	ok := false
	if !ok && rule.TrendUp {
		// TrendUp
		ok = idx.Score > 1
		if rule.ThresholdMax != 0 {
			// TrendUp And Value >= ThresholdMax
			ok = ok && m.Value >= rule.ThresholdMax
		} else if defaultThresholdMax != 0 {
			// TrendUp And Value >= DefaultThresholdMax
			ok = ok && m.Value >= defaultThresholdMax
		}
	}
	if !ok && !rule.TrendUp && rule.ThresholdMax != 0 {
		// Value >= ThresholdMax
		ok = m.Value >= rule.ThresholdMax
	}
	if !ok && rule.TrendDown {
		// TrendDown
		ok = idx.Score < -1
		if rule.ThresholdMin != 0 {
			// TrendDown And Value <= ThresholdMin
			ok = ok && m.Value <= rule.ThresholdMin
		} else if defaultThresholdMin != 0 {
			// TrendUp And Value >= DefaultThresholdMin
			ok = ok && m.Value <= defaultThresholdMin
		}
	}
	if !ok && !rule.TrendDown && rule.ThresholdMin != 0 {
		// Value <= ThresholdMin
		ok = m.Value <= rule.ThresholdMin
	}
	return ok
}

// SetNumMetrics sets the rule's number of metrics matched.
func (rule *Rule) SetNumMetrics(n int) {
	// Lock if shared.
	rule.Lock()
	defer rule.Unlock()
	rule.NumMetrics = n
}

// AnyTrendRelated checks if all rules is trend related.
func AnyTrendRelated(rules []*Rule) bool {
	for _, rule := range rules {
		if rule.IsTrendRelated() {
			return true
		}
	}
	return false
}
