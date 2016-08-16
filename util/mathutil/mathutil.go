// Copyright 2016 Eleme Inc. All rights reserved.

// Package mathutil provides math util functions.
package mathutil

import (
	"math"

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

// Average returns the mean value of float64 values.
// Returns zero if the vals length is 0.
func Average(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var sum float64
	for i := 0; i < len(vals); i++ {
		sum += vals[i]
	}
	return sum / float64(len(vals))
}

// StdDev returns the standard deviation of float64 values, with an input
// average.
// Returns zero if the vals length is 0.
func StdDev(vals []float64, avg float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var sum float64
	for i := 0; i < len(vals); i++ {
		dis := vals[i] - avg
		sum += dis * dis
	}
	return math.Sqrt(sum / float64(len(vals)))
}

func Div3Sigma(m *models.Metric, bms []models.BulkMetric) {
	var vals []float64
	for i := 0; i < len(bms); i++ {
		for j := 0; j < len(bms[i].Ms); j++ {
			vals = append(vals, bms[i].Ms[j].Value)
		}
	}
	m.Average, m.Score = div3Sigma(m.Value, vals)
}

// div3Sigma sets given metric score and average via 3-sigma.
//	states that nearly all values (99.7%) lie within the 3 standard deviations
//	of the mean in a normal distribution.
func div3Sigma(last float64, vals []float64) (avg, score float64) {
	if len(vals) == 0 {
		return last, 0
	}
	// Values average and standard deviation.
	avg = Average(vals)
	std := StdDev(vals, avg)
	// Set metric score
	if len(vals) <= int(cfg.Detector.LeastCount) { // Number of values not enough
		score = 0
		return
	}
	if std == 0 { // Eadger
		switch {
		case last == avg:
			score = 0
		case last > avg:
			score = 1
		case last < avg:
			score = -1
		}
		return
	}
	score = (last - avg) / (3 * std) // 3-sigma
	return
}
