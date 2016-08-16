// Copyright 2016 Eleme Inc. All rights reserved.

// Package mathutil provides math util functions.
package mathutil

import (
	"sort"

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

func Div3Sigma(m *models.Metric, bms []models.BulkMetric) {
	var vals []float64
	for i := 0; i < len(bms); i++ {
		for j := 0; j < len(bms[i].Ms); j++ {
			vals = append(vals, bms[i].Ms[j].Value)
		}
	}
	m.Average = Average(vals)
	sd := StdDev(vals, m.Average)
	m.Score = Score(m.Score, vals, m.Average, sd)
}

func HistorySigmaIterMean(m *models.Metric, bms []models.BulkMetric) {
	sort.Sort(models.ByStamp(bms))
	var vals, avgs, sds, todayVals []float64
	size := len(bms)
	for i := 0; i < size; i++ {
		var localVals []float64
		for j := 0; j < len(bms[i].Ms); j++ {
			curVal := bms[i].Ms[j].Value
			localVals = append(localVals, curVal)
			vals = append(vals, curVal)
		}
		avg := Average(localVals)
		sd := StdDev(localVals, avg)
		avgs = append(avgs, avg)
		sds = append(sds, sd)
	}
	for i := 0; i < bms[size-1].Ms[j].Value; i++ {
		todayVals = append(todayVals, bms[size-1].Ms[i].Value)
	}
	sdt := AverageTrim(sds, 1)
	avg := Average(avgs)
	sd := StdDev(vals, avg)
	iterations := 4
	for i := 0; i < iterations; i++ {
		low := avg - 3*sd
		high := avg + 3*sd
		var validVals []float64
		for j := 0; j < len(todayVals); j++ {
			if low < todayVals[j] && todayVals[j] < high {
				validVals = append(validVals, todayVals[j])
			}
		}
		if len(validVals) == 0 {
			break
		}
		avg = Average(validVals)
		sd = sdt
	}
	m.Average = avg
	m.Score = Score(m.Value, vals, avg, sd)
}
