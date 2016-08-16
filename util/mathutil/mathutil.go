// Copyright 2016 Eleme Inc. All rights reserved.

// Package mathutil provides math util functions.
package mathutil

import (
	"github.com/eleme/banshee/config"
	"github.com/eleme/banshee/models"
)

// Globals
var (
	// Config
	cfg *config.Config
)

func Init(config *config.Config) {
	cfg = config
}

// Div3Sigma sets given metric score and average when metric data is considered as normal distribution
func Div3Sigma(m *models.Metric, bms []models.BulkMetric) {
	var vals []float64
	for i := 0; i < len(bms); i++ {
		for j := 0; j < len(bms[i].Ms); j++ {
			vals = append(vals, bms[i].Ms[j].Value)
		}
	}
	if len(vals) == 0 {
		m.Score = 0
		m.Average = m.Value
		return
	}
	m.Average = Average(vals)
	if len(vals) <= int(cfg.Detector.LeastCount) { // Number of values not enough
		m.Score = 0
		return
	}
	std := StdDev(vals, m.Average)
	m.Score = Score(m.Value, avg, std)
}
