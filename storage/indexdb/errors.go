// Copyright 2015 Eleme Inc. All rights reserved.

package indexdb

import "errors"

var (
	// ErrNotFound is returned when requested data not found.
	ErrNotFound = errors.New("indexdb: not found")
	// ErrCorrupted is returned when corrupted data found.
	ErrCorrupted = errors.New("indexdb: corrupted data found")
	// ErrNoLink is returned when the index to put has no link.
	ErrNoLink = errors.New("indexdb: no link")
)
