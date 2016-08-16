package models

// Result struct help to receive multiple return values.
type BulkMetric struct {
	Err   error
	Ms    []*Metric
	Start uint32
	Stop  uint32
}
