// Copyright 2015 Eleme Inc. All rights reserved.

package filter

import (
	"github.com/eleme/banshee/config"
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util"
	"github.com/eleme/banshee/util/log"
	"path/filepath"
	"testing"
	"time"
)

func TestSimple(t *testing.T) {
	// New and add rules.
	cfg := config.New()
	cfg.Detector.EnableIntervalHitLimit = false
	filter := New(cfg)
	rule1 := &models.Rule{Pattern: "a.*.c.d"}
	rule2 := &models.Rule{Pattern: "a.b.c.*"}
	filter.addRule(rule1)
	filter.addRule(rule2)
	// Test
	rules1 := filter.MatchedRules(&models.Metric{Name: "nothing"})
	util.Must(t, 0 == len(rules1))

	rules2 := filter.MatchedRules(&models.Metric{Name: "a.b.c.e"})
	util.Must(t, 1 == len(rules2))
	util.Must(t, rules2[0] == rule2)

	rules3 := filter.MatchedRules(&models.Metric{Name: "a.e.c.d"})
	util.Must(t, 1 == len(rules3))
	util.Must(t, rules3[0] == rule1)

	rules4 := filter.MatchedRules(&models.Metric{Name: "a.b.c.d"})
	util.Must(t, 2 == len(rules4))
}

func TestHitLimit(t *testing.T) {
	// Currently disable logging
	log.Disable()
	defer log.Enable()
	// New and add rules.
	cfg := config.New()
	cfg.Interval = 1
	cfg.Detector.EnableIntervalHitLimit = true
	cfg.Detector.IntervalHitLimit = 100
	rule1 := &models.Rule{Pattern: "a.*.c.d"}
	filter := New(cfg)
	filter.initRuleHitReseter()
	filter.addRule(rule1)

	for i := uint32(0); i < cfg.Detector.IntervalHitLimit; i++ {
		// hit rule when counter < intervalHitLimit
		rules := filter.MatchedRules(&models.Metric{Name: "a.b.c.d"})
		util.Must(t, 1 == len(rules))

	}
	rules := filter.MatchedRules(&models.Metric{Name: "a.b.c.d"})
	time.Sleep(time.Second * 2)
	// after interval counter is cleared, matched rules = 1
	rules = filter.MatchedRules(&models.Metric{Name: "a.b.c.d"})
	util.Must(t, 1 == len(rules))
}

func BenchmarkRules1KNativeBest(b *testing.B) {
	var rules []*models.Rule
	for i := 0; i < 1024; i++ {
		rules = append(rules, &models.Rule{Pattern: "a.*.c." + string(i)})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < 1024; i++ {
			filepath.Match(rules[i].Pattern, "x")
		}
	}
}

func BenchmarkRules1kBest(b *testing.B) {
	cfg := config.New()
	cfg.Detector.EnableIntervalHitLimit = false
	filter := New(cfg)
	for i := 0; i < 1024; i++ {
		filter.addRule(&models.Rule{Pattern: "a.*.c." + string(i)})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.MatchedRules(&models.Metric{Name: "x.b.c." + string(i&1024)})
	}
}

func BenchmarkRules1kWorst(b *testing.B) {
	cfg := config.New()
	cfg.Detector.EnableIntervalHitLimit = false
	filter := New(cfg)
	for i := 0; i < 1024; i++ {
		filter.addRule(&models.Rule{Pattern: "a.*.c." + string(i)})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.MatchedRules(&models.Metric{Name: "a.b.c." + string(i&1024)})
	}
}

func BenchmarkRules2kWorst(b *testing.B) {
	cfg := config.New()
	cfg.Detector.EnableIntervalHitLimit = false
	filter := New(cfg)
	for i := 0; i < 1024*2; i++ {
		filter.addRule(&models.Rule{Pattern: "a.*.c." + string(i)})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.MatchedRules(&models.Metric{Name: "a.b.c." + string(i&65535)})
	}
}
