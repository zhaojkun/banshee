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
	"github.com/eleme/banshee/alerter/notifier"
	"github.com/eleme/banshee/algorithm"
	"github.com/eleme/banshee/cluster"
	"github.com/eleme/banshee/config"
	"github.com/eleme/banshee/detector"
	"github.com/eleme/banshee/filter"
	"github.com/eleme/banshee/health"
	"github.com/eleme/banshee/storage"
	"github.com/eleme/banshee/util/log"
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
	msg *cluster.Hub
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
	err = db.InitAdminDB(storage.AdminOptions{
		Host:     cfg.Storage.Admin.Host,
		Port:     cfg.Storage.Admin.Port,
		User:     cfg.Storage.Admin.User,
		Password: cfg.Storage.Admin.Password,
		DBName:   cfg.Storage.Admin.DBName,
	})
	if err != nil {
		log.Fatalf("failed to init admin db : %v", err)
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

func initAlgo() {
	// Rely on config.
	if cfg == nil {
		panic(errors.New("filter require db and config"))
	}
	algo.Init(cfg)
}
func initNotifier() {
	// Rely on config.
	if cfg == nil {
		panic(errors.New("filter require db and config"))
	}
	notifier.Init(cfg)
}

func initCluster() {
	if cfg == nil {
		panic(errors.New("cluster require db and config"))
	}
	var err error
	if cfg.Cluster.QueueDSN != "" {
		opts := cluster.Options{
			Master:       cfg.Cluster.Master,
			DSN:          cfg.Cluster.QueueDSN,
			VHost:        cfg.Cluster.VHost,
			ExchangeName: cfg.Cluster.ExchangeName,
			QueueName:    cfg.Cluster.QueueName,
		}
		msg, err = cluster.New(&opts, db)
		if err != nil {
			log.Errorf("cluster message queue open error: %s", err.Error())
		}
	}
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
	initAlgo()
	initNotifier()
	initCluster()
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
