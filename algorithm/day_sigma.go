package algo

import (
	"math"
	"sort"

	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util/log"
	"github.com/eleme/banshee/util/mathutil"
)

var (
	scoreFactor     = 1.4
	scoreMaxAverage = math.Pow(scoreFactor, 4)
	scoreMax        = math.Pow(scoreFactor, 8)
)

// DivDaySigma sets given metric score and average with the asumption that every day metrics belong to
//   a normal distribution and everyday distribution has same sigma but different mean
func DivDaySigma(m *models.Metric, bms []models.BulkMetric) {
	sort.Sort(models.ByStamp(bms))
	period := len(bms)
	daysVals := scoreFilter(bms)
	if len(daysVals) == 0 || len(daysVals[len(daysVals)-1]) == 0 {
		m.Average = m.Value
		m.Score = 0
		return
	}
	var vals, avgs, stds []float64
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
	if len(avgs) == 0 || len(vals) <= int(cfg.Detector.LeastCount) {
		m.Average = mathutil.Average(vals)
		m.Score = 0
		return
	}
	if used := tryAverageScore(m, bms, vals, avgs); used {
		log.Infof("AppID:%s now using average score", m.Name)
	}
	if -1 < m.Score && m.Score < 1 {
		stdAvg := mathutil.StdAverage(stds, nums)
		avg := avgs[len(avgs)-1]
		m.Average = avg
		m.Score = mathutil.Score(m.Value, avg, stdAvg)
	}
	m.Score = mathutil.Saturation(m.Score, -1.0*scoreMax, scoreMax)
}
func scoreFilter(bms []models.BulkMetric) (vals [][]float64) {
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
		vals = append(vals, localVals)
		threshold *= scoreFactor
	}
	return
}
func tryAverageScore(m *models.Metric, bms []models.BulkMetric, vals []float64, avgs []float64) bool { //for today values too large or small than other days
	if len(avgs) <= 2 {
		return false
	}
	period := len(bms)
	var todayVals []float64
	for i := 0; i < len(bms[period-1].Ms); i++ {
		todayVals = append(todayVals, bms[period-1].Ms[i].Value)
	}
	avg := mathutil.Average(avgs)
	std := mathutil.StdDev(vals, avg)
	low := avg - 3*std
	high := avg + 3*std
	var validVals []float64
	for i := 0; i < len(todayVals); i++ {
		if low <= todayVals[i] && todayVals[i] <= high {
			validVals = append(validVals, todayVals[i])
		}
	}
	if len(validVals) > len(todayVals)/2 {
		return false
	}
	var todayAvg float64
	if low <= m.Value && m.Value <= high {
		validVals = append(validVals, m.Value)
		todayAvg = mathutil.Average(validVals)
	} else {
		todayVals = append(todayVals, m.Value)
		todayAvg = mathutil.Average(todayVals)
	}
	m.Average = todayAvg
	m.Score = averageScore(todayAvg, avgs[:len(avgs)-1])
	return true
}
func averageScore(last float64, vals []float64) float64 {
	min := mathutil.Min(vals)
	max := mathutil.Max(vals)
	if max == min {
		return 0
	}
	var dis []float64
	for _, v := range vals {
		dis = append(dis, last-v)
	}
	minDis := mathutil.AbsMin(dis)
	score := minDis / (max - min)
	return mathutil.Saturation(score, -1.0*scoreMaxAverage, scoreMaxAverage)
}
