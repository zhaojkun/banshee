// Copyright 2015 Eleme Inc. All rights reserved.

package config

import (
	"github.com/eleme/banshee/util"
	"reflect"
	"testing"
)

func TestExampleConfigParsing(t *testing.T) {
	c := New()
	err := c.UpdateWithJSONFile("./exampleConfig.json")
	util.Must(t, err == nil)
	defaults := New()
	util.Must(t, reflect.DeepEqual(c, defaults))
}

func TestExampleConfigValidate(t *testing.T) {
	c := New()
	c.UpdateWithJSONFile("./exampleConfig.json")
	util.Must(t, c.Validate() == nil)
}
