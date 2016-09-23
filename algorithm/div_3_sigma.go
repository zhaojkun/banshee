package algo

import (
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util/mathutil"
)

// Div3Sigma sets given metric score and average with the asumption that  metrics belong to normal distribution
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
	m.Average = mathutil.Average(vals)
	if len(vals) <= int(cfg.Detector.LeastCount) { // Number of values not enough
		m.Score = 0
		return
	}
	std := mathutil.StdDev(vals, m.Average)
	m.Score = mathutil.Score(m.Value, m.Average, std)
}
