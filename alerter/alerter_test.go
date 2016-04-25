// Copyright 2016 Eleme Inc. All rights reserved.

package alerter

import (
	"github.com/eleme/banshee/util"
	"testing"
)

func TestHourInRange(t *testing.T) {
	util.Must(t, hourInRange(3, 0, 6))
	util.Must(t, !hourInRange(7, 0, 6))
	util.Must(t, !hourInRange(6, 0, 6))
	util.Must(t, hourInRange(23, 19, 10))
	util.Must(t, hourInRange(6, 19, 10))
	util.Must(t, !hourInRange(13, 19, 10))
}
