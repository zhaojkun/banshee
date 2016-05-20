// Copyright 2015 Eleme Inc. All rights reserved.

package models

// Metric is a data container for time series datapoint.
type Metric struct {
	// Name
	Name string `json:"name"`
	// Metric unix time stamp
	Stamp uint32 `json:"stamp"`
	// Metric value
	Value float64 `json:"value"`
	// Anomaly score
	Score float64 `json:"score"`
	// Average old
	Average float64 `json:"average"`
	// Link between index and metric.
	Link uint32 `json:"link"`
}

// LinkTo links the metric to an index.
func (m *Metric) LinkTo(idx *Index) {
	m.Link = idx.Link
}

// Validate the metric.
func (m *Metric) Validate() (err error) {
	if err = ValidateMetricName(m.Name); err != nil {
		return
	}
	if err = ValidateMetricStamp(m.Stamp); err != nil {
		return
	}
	return
}
