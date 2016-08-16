// Copyright 2015 Eleme Inc. All rights reserved.

// Banshee is a real-time anomalies or outliers detection system for periodic
// metrics.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/eleme/banshee/alerter"
	"github.com/eleme/banshee/config"
	"github.com/eleme/banshee/detector"
	"github.com/eleme/banshee/filter"
	"github.com/eleme/banshee/health"
	"github.com/eleme/banshee/storage"
	"github.com/eleme/banshee/util/log"
	"github.com/eleme/banshee/util/mathutil"
	"github.com/eleme/banshee/version"
	"github.com/eleme/banshee/webapp"
)

var (
	// Arguments
	debug       = flag.Bool("d", false, "debug mode")
	fileName    = flag.String("c", "config.yaml", "config file path")
	showVersion = flag.Bool("v", false, "show version")
	// Variables
	cfg = config.New()
	db  *storage.DB
	flt *filter.Filter
)

// usage prints command line usage to stderr.
func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  ./banshee -c filename [-d]\n")
	fmt.Fprintf(os.Stderr, "  ./banshee -v\n")
	fmt.Fprintf(os.Stderr, "%s@%s %s\n", version.Product, version.Version, version.Website)
	os.Exit(2)
}

// initLog initializes logging.
func initLog() {
	if *debug {
		log.SetLevel(log.DEBUG)
	}
}

func initConfig() {
	// Config parsing.
	if flag.NFlag() == 0 || (flag.NFlag() == 1 && *debug) { // Case ./program [-d]
		log.Warnf("no config specified, using default..")
	} else { // Case ./program -c filename
		err := cfg.UpdateWithYamlFile(*fileName)
		if err != nil {
			log.Fatalf("failed to load %s, %s", *fileName, err)
		}
	}
	// Config validation.
	err := cfg.Validate()
	if err != nil {
		log.Fatalf("config: %s", err)
	}
}

func initDB() {
	// Rely on config.
	if cfg == nil {
		panic(errors.New("db require config"))
	}
	path := cfg.Storage.Path
	opts := &storage.Options{
		Period:     cfg.Period,
		Expiration: cfg.Expiration,
	}
	var err error
	db, err = storage.Open(path, opts)
	if err != nil {
		log.Fatalf("failed to open %s: %v", path, err)
	}
}

func initFilter() {
	// Rely on db and config.
	if db == nil || cfg == nil {
		panic(errors.New("filter require db and config"))
	}
	// Init filter
	flt = filter.New(cfg)
	flt.Init(db)
}

func initMath() {
	if cfg == nil {
		panic(errors.New("filter require db and config"))
	}
	mathutil.Init(cfg)
}

func init() {
	// Arguments
	flag.Usage = usage
	flag.Parse()
	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Version)
		os.Exit(1)
	}
	// Init
	initLog()
	initConfig()
	initDB()
	initFilter()
	initMath()
}

func main() {
	health.Init(db)
	go health.Start()

	alerter := alerter.New(cfg, db)
	alerter.Start()

	go webapp.Start(cfg, db, flt)

	detector := detector.New(cfg, db, flt)
	detector.Out(alerter.In)
	detector.Start()
}
