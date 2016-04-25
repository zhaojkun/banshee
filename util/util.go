// Copyright 2015 Eleme Inc. All rights reserved.

// Package util provides util functions.
package util

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Must asserts the given value is True for testing.
func Must(t *testing.T, v bool) {
	if !v {
		_, fileName, line, _ := runtime.Caller(1)
		t.Errorf("\n unexcepted: %s:%d", fileName, line)
		t.FailNow()
	}
}

// ToFixed truncates float64 type to a particular percision in string.
func ToFixed(n float64, prec int) string {
	s := strconv.FormatFloat(n, 'f', prec, 64)
	return strings.TrimRight(strings.TrimRight(s, "0"), ".")
}

// IsFileExist test whether a filepath is exist.
func IsFileExist(fileName string) bool {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// Timer is a minimal timer util.
type Timer struct {
	startAt time.Time
}

// NewTimer creates a minimal timer.
func NewTimer() *Timer {
	return &Timer{time.Now()}
}

// Elapsed returns milliseconds elapsed.
func (t *Timer) Elapsed() float64 {
	return float64(time.Since(t.startAt).Nanoseconds()) / float64(1000*1000)
}
