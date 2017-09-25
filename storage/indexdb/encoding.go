// Copyright 2015 Eleme Inc. All rights reserved.

package indexdb

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/eleme/banshee/models"
)

// Format
//
//	|--- Key --|------------------ Value (24) -------------------|
//	+----------+-----------+-----------+-----------+-------------+
//	| Name (X) |  Link (4) | Stamp (4) | Score (8) | Average (8) |
//	+----------+-----------+-----------+-----------+-------------+
//

// encode encodes db value from index.
func encode(idx *models.Index) []byte {
	b := make([]byte, 4+4+8+8)
	binary.BigEndian.PutUint32(b[:4], idx.Link)                           // 4
	binary.BigEndian.PutUint32(b[4:4+4], idx.Stamp)                       // 4
	binary.BigEndian.PutUint64(b[4+4:4+4+8], math.Float64bits(idx.Score)) // 8
	binary.BigEndian.PutUint64(b[4+4+8:], math.Float64bits(idx.Average))  // 8
	return b
}

// decode decodes db value into index.
func decode(value []byte, idx *models.Index) error {
	if len(value) != 4+4+8+8 {
		return ErrCorrupted
	}
	r := bytes.NewReader(value)
	if err := binary.Read(r, binary.BigEndian, &idx.Link); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &idx.Stamp); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &idx.Score); err != nil {
		return err
	}
	return binary.Read(r, binary.BigEndian, &idx.Average)
}
