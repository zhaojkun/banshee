// Copyright 2016 Eleme Inc. All rights reserved.

// Package mathutil provides math util functions.
package algo

import "github.com/eleme/banshee/config"

// Globals
var (
	// Config
	cfg *config.Config
)

func Init(config *config.Config) {
	cfg = config
}
