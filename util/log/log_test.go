// Copyright 2015 Eleme Inc. All rights reserved.

package log

import (
	"testing"
)

func TestLog(t *testing.T) {
	// No assertions.
	SetLevel(DEBUG)
	Debug(nil)
	Info(nil)
	Warn(nil)
	Error(nil)
	Debugf("hello %s", "world")
	Infof("hello %s", "world")
	Warnf("hello %s", "world")
	Errorf("hello %s", "world")
}
