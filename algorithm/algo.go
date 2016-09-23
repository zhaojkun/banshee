// Copyright 2016 Eleme Inc. All rights reserved.

// Package algo provides several anomalous detection algorithms
package algo

import "github.com/eleme/banshee/config"

// Globals
var (
	// Config
	cfg *config.Config
)

// Init algorithm
func Init(config *config.Config) {
	cfg = config
}
