// Copyright 2016 Eleme Inc. All rights reserved.

// Package mathutil provides math util functions.
package mathutil

import (
	"math"
	"sort"
)

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

func AverageTrim(src []float64, n int) float64 {
	vals := make([]float64, len(src))
	copy(vals, src)
	sort.Float64s(vals)
	var sum float64
	for i := 0; i < len(vals); i++ {
		sum += vals[i]
	}
	for i := 0; i < n; i++ {
		size := len(vals)
		if size == 0 {
			return 0
		}
		avg := sum / float64(size)
		lowSpan := avg - vals[0]
		highSpan := avg - vals[size-1]
		if lowSpan >= highSpan {
			sum -= vals[0]
			vals = vals[1:]
		} else {
			sum -= vals[size-1]
			vals = vals[:size-1]
		}
	}
	if len(vals) == 0 {
		return 0
	}
	return sum / float64(len(vals))
}

// div3Sigma sets given metric score and average via 3-sigma.
//	states that nearly all values (99.7%) lie within the 3 standard deviations
//	of the mean in a normal distribution.
func Score(last float64, vals []float64, avg float64, std float64) float64 {
	if len(vals) == 0 || len(vals) <= int(cfg.Detector.LeastCount) {
		return 0
	}
	if std == 0 { // Eadger
		var score float64
		switch {
		case last == avg:
			score = 0
		case last > avg:
			score = 1
		case last < avg:
			score = -1
		}
		return score
	}
	return (last - avg) / (3 * std) // 3-sigma
}
