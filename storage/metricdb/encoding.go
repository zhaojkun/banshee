// Copyright 2015 Eleme Inc. All rights reserved.

package metricdb

import (
	"bytes"
	"encoding/binary"
	"github.com/eleme/banshee/models"
	"math"
)

// Format
//
//	|------- Key (8) ------|-------------- Value (24) -----------|
//	+----------+-----------+-----------+-----------+-------------+
//	| Link (4) | Stamp (4) | Value (8) | Score (8) | Average (8) |
//	+----------+-----------+-----------+-----------+-------------+
//

// encodeKey encodes db key from metric.
func encodeKey(m *models.Metric) []byte {
	b := make([]byte, 4+4)
	binary.BigEndian.PutUint32(b[:4], m.Link)
	binary.BigEndian.PutUint32(b[4:], m.Stamp)
	return b
}

// encodeValue encodes db value from metric.
func encodeValue(m *models.Metric) []byte {
	b := make([]byte, 8+8+8)
	binary.BigEndian.PutUint64(b[:8], math.Float64bits(m.Value))
	binary.BigEndian.PutUint64(b[8:8+8], math.Float64bits(m.Score))
	binary.BigEndian.PutUint64(b[8+8:], math.Float64bits(m.Average))
	return b
}

// decodeKey decodes db key into metric, this will fill metric name and metric
// stamp.
func decodeKey(key []byte, m *models.Metric) (err error) {
	if len(key) != 4+4 {
		return ErrCorrupted
	}
	r := bytes.NewReader(key)
	if err = binary.Read(r, binary.BigEndian, &m.Link); err != nil {
		return
	}
	if err = binary.Read(r, binary.BigEndian, &m.Stamp); err != nil {
		return
	}
	return nil
}

// decodeValue decodes db value into metric, this will fill metric value,
// average and stddev.
func decodeValue(value []byte, m *models.Metric) (err error) {
	if len(value) != 8+8+8 {
		return ErrCorrupted
	}
	r := bytes.NewReader(value)
	if err = binary.Read(r, binary.BigEndian, &m.Value); err != nil {
		return
	}
	if err = binary.Read(r, binary.BigEndian, &m.Score); err != nil {
		return
	}
	if err = binary.Read(r, binary.BigEndian, &m.Average); err != nil {
		return
	}
	return nil
}
