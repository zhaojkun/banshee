package algo

import (
	"fmt"
	"sort"

	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util/mathutil"
)

func HistorySigmaIterMean(m *models.Metric, bms []models.BulkMetric) {
	sort.Sort(models.ByStamp(bms))
	var vals, avgs, stds, todayVals []float64
	size := len(bms)
	scale := 1.0
	for i := 0; i < size; i++ {
		var localVals []float64
		for j := 0; j < len(bms[i].Ms); j++ {
			ms := bms[i].Ms[j]
			score := scale * ms.Score
			if score >= 1 || score <= -1 {
				continue
			}
			localVals = append(localVals, ms.Value)
		}
		if len(localVals) == 0 {
			continue
		}
		avg := mathutil.Average(localVals)
		std := mathutil.StdDev(localVals, avg)
		vals = append(vals, localVals...)
		avgs = append(avgs, avg)
		stds = append(stds, std)
		scale *= 0.5
	}
	if len(vals) == 0 {
		m.Score = 0
		m.Average = m.Value
		return
	}
	if len(vals) <= int(cfg.Detector.LeastCount) { // Number of values not enough
		m.Average = mathutil.Average(vals)
		m.Score = 0
		return
	}
	for i := 0; i < len(bms[size-1].Ms); i++ {
		todayVals = append(todayVals, bms[size-1].Ms[i].Value)
	}
	stdAvg := mathutil.AverageTrim(stds, 1)
	avg := mathutil.Average(avgs)
	std := mathutil.StdDev(vals, avg)
	iteraions := 4
	for i := 0; i < iteraions; i++ {
		low := avg - 3*std
		high := avg + 3*std
		var validVals []float64
		for j := 0; j < len(todayVals); j++ {
			if low <= todayVals[j] && todayVals[j] <= high {
				validVals = append(validVals, todayVals[j])
			}
		}
		if len(validVals) == 0 {
			fmt.Println("It seems all today data bad")
			break
		}
		avg = mathutil.Average(validVals)
		std = stdAvg
	}
	m.Average = avg
	m.Score = mathutil.Score(m.Value, avg, stdAvg)
}
