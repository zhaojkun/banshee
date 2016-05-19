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
	// ErrNoStorage is returned when no storage is able to serve, which
	// indicates that given stamp or stamp range may be invalid.
	ErrNoStorage = errors.New("metricdb: no storage")
)
