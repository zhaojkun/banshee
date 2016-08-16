package models

// BulkMetric is a data container for time series data points among a few days.
type BulkMetric struct {
	Err   error
	Ms    []*Metric
	Start uint32
	Stop  uint32
}
