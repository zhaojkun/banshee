// Copyright 2016 Eleme Inc. All rights reserved.

package detector

import (
	"testing"

	"github.com/eleme/banshee/config"
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util"
)

func TestFill0Issue470(t *testing.T) {
	// Case https://github.com/eleme/banshee/issues/470
	cfg := config.New()
	d := &Detector{cfg: cfg}
	ms := []*models.Metric{
		&models.Metric{Stamp: 80, Value: 80},
		&models.Metric{Stamp: 90, Value: 90},
		&models.Metric{Stamp: 120, Value: 120},
	}
	start, stop := uint32(60), uint32(150)
	excepted := []*models.Metric{
		&models.Metric{Stamp: 80, Value: 80},
		&models.Metric{Stamp: 90, Value: 90},
		&models.Metric{Value: 0},
		&models.Metric{Value: 0},
		&models.Metric{Stamp: 120, Value: 120},
		&models.Metric{Value: 0},
		&models.Metric{Value: 0},
	}
	actually := d.fill0(ms, start, stop)
	util.Must(t, len(actually) == len(excepted))
	for i := 0; i < len(excepted); i++ {
		util.Must(t, excepted[i].Value == actually[i].Value)
	}
}

func TestPickTrendingFactor(t *testing.T) {
	cfg := config.New()
	d := &Detector{cfg: cfg}

	rules := []*models.Rule{
		{Level: models.RuleLevelLow},
	}
	util.Must(t, d.pickTrendingFactor(rules) == cfg.Detector.TrendingFactorLowLevel)

	rules = []*models.Rule{
		{Level: models.RuleLevelMiddle},
	}
	util.Must(t, d.pickTrendingFactor(rules) == cfg.Detector.TrendingFactorMiddleLevel)

	rules = []*models.Rule{
		{Level: models.RuleLevelHigh},
	}
	util.Must(t, d.pickTrendingFactor(rules) == cfg.Detector.TrendingFactorHighLevel)

	rules = []*models.Rule{
		{Level: models.RuleLevelLow},
		{Level: models.RuleLevelHigh},
		{Level: models.RuleLevelMiddle},
	}
	util.Must(t, d.pickTrendingFactor(rules) == cfg.Detector.TrendingFactorHighLevel)
}
