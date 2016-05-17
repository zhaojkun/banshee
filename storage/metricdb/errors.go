// Copyright 2015 Eleme Inc. All rights reserved.

package metricdb

import "errors"

var (
	// ErrNotFound is returned when requested data not found.
	ErrNotFound = errors.New("metricdb: not found")
	// ErrCorrupted is returned when corrupted data found.
	ErrCorrupted = errors.New("metricdb: corrupted data found")
	// ErrNoLink is returned when the metric to put has no link.
	ErrNoLink = errors.New("metricdb: no link")
	// ErrNoFileStorage is returned when no file storage is able to serve,
	// which indicates that given stamp or stamp range may be invalid.
	ErrNoFileStorage = errors.New("metricdb: no storage")
	// ErrNoMemeStorage is returned when no mem storage is able to serve,
	// which indicates that given stamp or stamp range may be invalid.
	ErrNoMemStorage = errors.New("metricdb: no storage")
)
