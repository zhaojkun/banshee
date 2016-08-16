package models

// Result struct help to receive multiple return values.
type BulkMetric struct {
	Err   error
	Ms    []*Metric
	Start uint32
	Stop  uint32
}

type ByStamp []BulkMetric

func (s ByStamp) Len() int {
	return len(s)
}

func (s ByStamp) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByStamp) Less(i, j int) bool {
	return s[i].Start < s[j].Start
}
