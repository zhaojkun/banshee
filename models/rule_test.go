// Copyright 2016 Eleme Inc. All rights reserved.

package models

import (
	"github.com/eleme/banshee/config"
	"github.com/eleme/banshee/util"
	"testing"
	"time"
)

func TestRuleTest(t *testing.T) {
	var rule *Rule
	// TrendUp
	rule = &Rule{TrendUp: true}
	util.Must(t, rule.Test(&Metric{}, &Index{Score: 1.2}, nil))
	util.Must(t, !rule.Test(&Metric{}, &Index{Score: 0.8}, nil))
	// TrendDown
	rule = &Rule{TrendDown: true}
	util.Must(t, rule.Test(&Metric{}, &Index{Score: -1.2}, nil))
	util.Must(t, !rule.Test(&Metric{}, &Index{Score: 1.2}, nil))
	// TrendUp And Value >= X
	rule = &Rule{TrendUp: true, ThresholdMax: 39}
	util.Must(t, rule.Test(&Metric{Value: 50}, &Index{Score: 1.3}, nil))
	util.Must(t, !rule.Test(&Metric{Value: 38}, &Index{Score: 1.5}, nil))
	util.Must(t, !rule.Test(&Metric{Value: 60}, &Index{Score: 0.9}, nil))
	// TrendDown And Value <= X
	rule = &Rule{TrendDown: true, ThresholdMin: 40}
	util.Must(t, rule.Test(&Metric{Value: 10}, &Index{Score: -1.2}, nil))
	util.Must(t, !rule.Test(&Metric{Value: 41}, &Index{Score: -1.2}, nil))
	util.Must(t, !rule.Test(&Metric{Value: 12}, &Index{Score: -0.2}, nil))
	// (TrendUp And Value >= X) Or TrendDown
	rule = &Rule{TrendUp: true, TrendDown: true, ThresholdMax: 90}
	util.Must(t, rule.Test(&Metric{Value: 100}, &Index{Score: 1.1}, nil))
	util.Must(t, rule.Test(&Metric{}, &Index{Score: -1.1}, nil))
	util.Must(t, !rule.Test(&Metric{}, &Index{Score: -0.1}, nil))
	util.Must(t, !rule.Test(&Metric{Value: 89}, &Index{Score: 1.3}, nil))
	util.Must(t, !rule.Test(&Metric{Value: 189}, &Index{Score: 0.3}, nil))
	// (TrendUp And Value >= X) Or (TrendDown And Value <= X)
	rule = &Rule{TrendUp: true, TrendDown: true, ThresholdMax: 90, ThresholdMin: 10}
	util.Must(t, rule.Test(&Metric{Value: 100}, &Index{Score: 1.2}, nil))
	util.Must(t, rule.Test(&Metric{Value: 9}, &Index{Score: -1.2}, nil))
	util.Must(t, !rule.Test(&Metric{Value: 12}, &Index{Score: 1.2}, nil))
	util.Must(t, !rule.Test(&Metric{Value: 102}, &Index{Score: 0.2}, nil))
	util.Must(t, !rule.Test(&Metric{Value: 2}, &Index{Score: 0.9}, nil))
	// Default thresholdMaxs
	cfg := config.New()
	cfg.Detector.DefaultThresholdMaxs["fo*"] = 300
	rule = &Rule{TrendUp: true}
	util.Must(t, rule.Test(&Metric{Value: 310, Name: "foo"}, &Index{Score: 1.3}, cfg))
	util.Must(t, !rule.Test(&Metric{Value: 120, Name: "foo"}, &Index{Score: 1.3}, cfg))
	// Default thresholdMins
	cfg = config.New()
	cfg.Detector.DefaultThresholdMins["fo*"] = 10
	rule = &Rule{TrendDown: true}
	util.Must(t, !rule.Test(&Metric{Value: 19, Name: "foo"}, &Index{Score: -1.2}, cfg))
	util.Must(t, rule.Test(&Metric{Value: 8, Name: "foo"}, &Index{Score: -1.2}, cfg))
	// Bug#456: DefaultThresholdMax intercepts the testing for later trendDown.
	cfg = config.New()
	cfg.Detector.DefaultThresholdMaxs["fo*"] = 10
	rule = &Rule{TrendDown: true}
	util.Must(t, !rule.Test(&Metric{Value: 19, Name: "foo"}, &Index{Score: 0.37}, cfg))
}

func TestRuleDisabled(t *testing.T) {
	var rule *Rule
	// Forever disabled
	rule = &Rule{Disabled: true, DisabledFor: 0}
	util.Must(t, !rule.Test(&Metric{}, nil, nil))
	// Tmp disabled
	rule = &Rule{Disabled: true, DisabledFor: 1}
	util.Must(t, !rule.Test(&Metric{}, nil, nil))
	// Don't disabled.
	rule = &Rule{Disabled: true, DisabledFor: 1, DisabledAt: time.Now().Add(time.Minute * -1), ThresholdMax: 1}
	util.Must(t, rule.Test(&Metric{Value: 2}, nil, nil))
	// Default
	rule = &Rule{Disabled: true, DisabledFor: 1, DisabledAt: time.Time{}, ThresholdMax: 2}
	util.Must(t, rule.Test(&Metric{Value: 3}, nil, nil))
}

func BenchmarkRuleTest(b *testing.B) {
	cfg := config.New()
	m := &Metric{Value: 102}
	idx := &Index{Score: 1.2}
	rule := &Rule{TrendUp: true, ThresholdMax: 100}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rule.Test(m, idx, cfg)
	}
}

func BenchmarkRuleTestWithDefaultThresholdMaxsNum4(b *testing.B) {
	cfg := config.New()
	cfg.Detector.DefaultThresholdMaxs = map[string]float64{
		"timer.count_ps.*": 30,
		"timer.upper_90.*": 500,
		"counter.*":        10,
		"timer.mean_90.*":  300,
	}
	m := &Metric{Name: "timer.mean_90.foo", Value: 1700}
	idx := &Index{Name: m.Name, Score: 1.2}
	rule := Rule{TrendUp: true}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rule.Test(m, idx, cfg)
	}
}

func BenchmarkRuleTestWithDefaultThresholdMaxsNum8(b *testing.B) {
	cfg := config.New()
	cfg.Detector.DefaultThresholdMaxs = map[string]float64{
		"timer.count_ps.y.*": 30,
		"timer.upper_90.y.*": 500,
		"counter.y.*":        10,
		"timer.mean_90.y.*":  300,
		"timer.count_ps.x.*": 100,
		"timer.upper_90.x.*": 1500,
		"counter.x.*":        15,
		"timer.mean_90.x.*":  1000,
	}
	m := &Metric{Name: "timer.mean_90.x.foo", Value: 1700}
	idx := &Index{Name: m.Name, Score: 1.2}
	rule := Rule{TrendUp: true}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rule.Test(m, idx, cfg)
	}
}
