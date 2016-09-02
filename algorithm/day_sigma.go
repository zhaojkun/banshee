package algo

import (
	"sort"

	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util/mathutil"
)

// DivDaySigma sets given metric score and average with the asumption that every day metrics belong to
//   a normal distribution and everyday distribution has same sigma but different mean
func DivDaySigma(m *models.Metric, bms []models.BulkMetric) {
	sort.Sort(models.ByStamp(bms))
	period := len(bms)
	daysVals, n := scoreFilter(bms)
	if n == 0 {
		m.Score = 0
		m.Average = m.Value
		return
	}
	var vals, avgs, stds, todayVals []float64
	var nums []int
	for i := 0; i < period; i++ {
		if len(daysVals[i]) == 0 {
			continue
		}
		avg := mathutil.Average(daysVals[i])
		std := mathutil.StdDev(daysVals[i], avg)
		avgs = append(avgs, avg)
		stds = append(stds, std)
		nums = append(nums, len(daysVals[i]))
		vals = append(vals, daysVals[i]...)
	}
	if n <= int(cfg.Detector.LeastCount) {
		m.Average = mathutil.Average(vals)
		m.Score = 0
		return
	}
	for i := 0; i < len(bms[period-1].Ms); i++ {
		todayVals = append(todayVals, bms[period-1].Ms[i].Value)
	}
	avg := mathutil.Average(avgs)
	stdAvg := mathutil.StdAverage(stds, nums)
	iteraions := 2
	for i := 0; i < iteraions; i++ {
		var std float64
		if i == 0 {
			std = mathutil.StdDev(vals, avg)
		} else {
			std = stdAvg
		}
		low := avg - 3*std
		high := avg + 3*std
		var validVals []float64
		for j := 0; j < len(todayVals); j++ {
			if low <= todayVals[j] && todayVals[j] <= high {
				validVals = append(validVals, todayVals[j])
			}
		}
		if len(validVals) == 0 { //for today values too large or small than other days
			todayAvg := mathutil.Average(todayVals)
			m.Average = todayAvg
			m.Score = averageScore(todayAvg, avgs[:len(avgs)-1])
			return
		}
		avg = mathutil.Average(validVals)
	}
	m.Average = avg
	m.Score = mathutil.Score(m.Value, avg, stdAvg)
}
func scoreFilter(bms []models.BulkMetric) (vals [][]float64, n int) {
	threshold := 1.0
	for i := 0; i < len(bms); i++ {
		var localVals []float64
		for j := 0; j < len(bms[i].Ms); j++ {
			ms := bms[i].Ms[j]
			if ms.Score > threshold || ms.Score <= -1.0*threshold {
				continue
			}
			localVals = append(localVals, ms.Value)
		}
		n += len(localVals)
		vals = append(vals, localVals)
		threshold *= 1.4
	}
	return
}
func averageScore(last float64, vals []float64) float64 {
	if len(vals) < 2 {
		return 2.0
	}
	min := mathutil.Min(vals)
	max := mathutil.Max(vals)
	if max == min {
		return 2.0
	}
	var score float64
	if last > max {
		score = (last - max) / (max - min)
	} else if last < min {
		score = (last - min) / (max - min)
	}
	return mathutil.Saturation(score, 20, -20)
}
