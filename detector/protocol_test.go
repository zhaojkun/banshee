// Copyright 2015 Eleme Inc. All rights reserved.

package detector

import (
	"github.com/eleme/banshee/util"
	"testing"
)

func TestParseMetric(t *testing.T) {
	line := "foo 1449655769 3.14"
	m, err := parseMetric(line)
	util.Must(t, err == nil)
	util.Must(t, m.Name == "foo")
	util.Must(t, m.Stamp == uint32(1449655769))
	util.Must(t, m.Value == 3.14)
}

func TestParseMetricBadLine(t *testing.T) {
	line := "foo 1.3 1.234"
	m, err := parseMetric(line)
	util.Must(t, err != nil)
	util.Must(t, m == nil)
}
