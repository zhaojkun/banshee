package models

// BulkMetric is a data container for time series data points among a few days.
type BulkMetric struct {
	Err   error
	Ms    []*Metric
	Start uint32
	Stop  uint32
}

// ByStamp sort BulkMetric by start stamp
type ByStamp []BulkMetric

func (s ByStamp) Len() int { return len(s) }

func (s ByStamp) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s ByStamp) Less(i, j int) bool { return s[i].Start < s[j].Start }
