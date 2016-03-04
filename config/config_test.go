// Copyright 2015 Eleme Inc. All rights reserved.

package config

import (
	"github.com/eleme/banshee/util/assert"
	"reflect"
	"testing"
)

func TestExampleConfigParsing(t *testing.T) {
	c := New()
	err := c.UpdateWithJSONFile("./exampleConfig.json")
	assert.Ok(t, err == nil)
	defaults := New()
	assert.Ok(t, reflect.DeepEqual(c, defaults))
}

func TestExampleConfigValidate(t *testing.T) {
	c := New()
	c.UpdateWithJSONFile("./exampleConfig.json")
	assert.Ok(t, c.Validate() == nil)
}
