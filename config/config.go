// Copyright 2015 Eleme Inc. All rights reserved.

package config

import (
	"encoding/json"
	"github.com/eleme/banshee/util/log"
	"io/ioutil"
)

// Measures
const (
	// Time
	Second uint32 = 1
	Minute        = 60 * Second
	Hour          = 60 * Minute
	Day           = 24 * Hour
)

// Defaults
const (
	// Default time interval for all metrics in seconds.
	DefaultInterval uint32 = 10 * Second
	// Default hit limit to a rule in an interval
	DefaultIntervalHitLimit int = 100
	// Default period for all metrics in seconds.
	DefaultPeriod uint32 = 1 * Day
	// Default metric expiration.
	DefaultExpiration uint32 = 7 * Day
	// Default weight factor for moving average.
	DefaultTrendingFactor float64 = 0.1
	// Default filter offset to query history metrics.
	DefaultFilterOffset float64 = 0.01
	// Default cleaner interval.
	DefaultCleanerInterval uint32 = 3 * Hour
	// Default cleaner threshold.
	DefaultCleanerThreshold uint32 = 3 * Day
	// Default value of alerting interval.
	DefaultAlerterInterval uint32 = 20 * Minute
	// Default value of alert times limit in one day for the same metric
	DefaultAlerterOneDayLimit uint32 = 5
	// Default value of least count.
	DefaultLeastCount uint32 = 5 * Minute / DefaultInterval
	// Default alerting silent time range.
	DefaultSilentTimeStart int = 0
	DefaultSilentTimeEnd   int = 6
	// Default language for webapp.
	DefaultWebappLanguage string = "en"
)

// Limitations
const (
	// Max value for the number of DefaultThresholdMaxs.
	MaxNumDefaultThresholdMaxs = 8
	// Max value for the number of DefaultThresholdMins.
	MaxNumDefaultThresholdMins = 8
	// Max value for the number of FillBlankZeros.
	MaxFillBlankZerosLen = 8
	// Min value for the expiration to period.
	MinExpirationNumToPeriod uint32 = 5
	// Min value for the cleaner threshold to period.
	MinCleanerThresholdNumToPeriod uint32 = 2
)

// WebappSupportedLanguages lists webapp supported languages.
var WebappSupportedLanguages = []string{"en", "zh"}

// Config is the configuration container.
type Config struct {
	Interval   uint32         `json:"interval"`
	Period     uint32         `json:"period"`
	Expiration uint32         `json:"expiration"`
	Storage    configStorage  `json:"storage"`
	Detector   configDetector `json:"detector"`
	Webapp     configWebapp   `json:"webapp"`
	Alerter    configAlerter  `json:"alerter"`
	Cleaner    configCleaner  `json:"cleaner"`
}

type configStorage struct {
	Path string `json:"path"`
}

type configDetector struct {
	Port                 int                `json:"port"`
	TrendingFactor       float64            `json:"trendingFactor"`
	FilterOffset         float64            `json:"filterOffset"`
	LeastCount           uint32             `json:"leastCount"`
	BlackList            []string           `json:"blackList"`
	IntervalHitLimit     int                `json:"intervalHitLimit"`
	DefaultThresholdMaxs map[string]float64 `json:"defaultThresholdMaxs"`
	DefaultThresholdMins map[string]float64 `json:"defaultThresholdMins"`
	FillBlankZeros       []string           `json:"fillBlankZeros"`
}

type configWebapp struct {
	Port     int               `json:"port"`
	Auth     [2]string         `json:"auth"`
	Static   string            `json:"static"`
	Notice   map[string]string `json:"notice"`
	Language string            `json:"language"`
}

type configAlerter struct {
	Command                string `json:"command"`
	Workers                int    `json:"workers"`
	Interval               uint32 `json:"inteval"`
	OneDayLimit            uint32 `json:"oneDayLimit"`
	DefaultSilentTimeRange [2]int `json:"defaultSilentTimeRange"`
}

type configCleaner struct {
	Interval  uint32 `json:"interval"`
	Threshold uint32 `json:"threshold"`
}

// New creates a Config with default values.
func New() *Config {
	config := new(Config)
	config.Interval = DefaultInterval
	config.Period = DefaultPeriod
	config.Expiration = DefaultExpiration
	config.Storage.Path = "storage/"
	config.Detector.Port = 2015
	config.Detector.TrendingFactor = DefaultTrendingFactor
	config.Detector.FilterOffset = DefaultFilterOffset
	config.Detector.LeastCount = DefaultLeastCount
	config.Detector.BlackList = []string{}
	config.Detector.IntervalHitLimit = DefaultIntervalHitLimit
	config.Detector.DefaultThresholdMaxs = make(map[string]float64, 0)
	config.Detector.DefaultThresholdMins = make(map[string]float64, 0)
	config.Detector.FillBlankZeros = []string{}
	config.Webapp.Port = 2016
	config.Webapp.Auth = [2]string{"admin", "admin"}
	config.Webapp.Static = "static/dist"
	config.Webapp.Notice = make(map[string]string, 0)
	config.Webapp.Language = DefaultWebappLanguage
	config.Alerter.Command = ""
	config.Alerter.Workers = 4
	config.Alerter.Interval = DefaultAlerterInterval
	config.Alerter.OneDayLimit = DefaultAlerterOneDayLimit
	config.Alerter.DefaultSilentTimeRange = [2]int{DefaultSilentTimeStart, DefaultSilentTimeEnd}
	config.Cleaner.Interval = DefaultCleanerInterval
	config.Cleaner.Threshold = DefaultCleanerThreshold
	return config
}

// UpdateWithJSONFile update the config from a json file.
func (config *Config) UpdateWithJSONFile(fileName string) error {
	log.Debug("read config from %s..", fileName)
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, config)
	if err != nil {
		return err
	}
	return err
}

// Copy config.
func (config *Config) Copy() *Config {
	c := New()
	c.Interval = config.Interval
	c.Period = config.Period
	c.Expiration = config.Expiration
	c.Storage.Path = config.Storage.Path
	c.Detector.Port = config.Detector.Port
	c.Detector.TrendingFactor = config.Detector.TrendingFactor
	c.Detector.FilterOffset = config.Detector.FilterOffset
	c.Detector.LeastCount = config.Detector.LeastCount
	c.Detector.BlackList = config.Detector.BlackList
	c.Detector.DefaultThresholdMaxs = config.Detector.DefaultThresholdMaxs
	c.Detector.DefaultThresholdMins = config.Detector.DefaultThresholdMins
	c.Detector.FillBlankZeros = config.Detector.FillBlankZeros
	c.Detector.IntervalHitLimit = config.Detector.IntervalHitLimit
	c.Webapp.Port = config.Webapp.Port
	c.Webapp.Auth = config.Webapp.Auth
	c.Webapp.Static = config.Webapp.Static
	c.Webapp.Notice = config.Webapp.Notice
	c.Webapp.Language = config.Webapp.Language
	c.Alerter.Command = config.Alerter.Command
	c.Alerter.Workers = config.Alerter.Workers
	c.Alerter.Interval = config.Alerter.Interval
	c.Alerter.OneDayLimit = config.Alerter.OneDayLimit
	c.Alerter.DefaultSilentTimeRange = config.Alerter.DefaultSilentTimeRange
	c.Cleaner.Interval = config.Cleaner.Interval
	c.Cleaner.Threshold = config.Cleaner.Threshold
	return c
}

// Validate config.
func (c *Config) Validate() error {
	if err := c.validateGlobals(); err != nil {
		return err
	}
	if err := c.Detector.validateDetector(); err != nil {
		return err
	}
	if err := c.Webapp.validateWebapp(); err != nil {
		return err
	}
	if err := c.Alerter.validateAlerter(); err != nil {
		return err
	}
	if err := c.Cleaner.validateCleaner(c.Period); err != nil {
		return err
	}
	return nil
}

func (c *Config) validateGlobals() error {
	// Should: 1 Second <= Interval <= 5 Minute
	if c.Interval < 1*Second || c.Interval > 5*Minute {
		return ErrInterval
	}
	// Should: Period >= Interval
	if c.Interval > c.Period {
		return ErrPeriod
	}
	// Should: Expiration >= Period * 5
	if c.Expiration < c.Period*MinExpirationNumToPeriod {
		return ErrExpiration
	}
	return nil
}

func (c *configDetector) validateDetector() error {
	// Should: 0 < Port < 65536
	if c.Port < 1 || c.Port > 65535 {
		return ErrDetectorPort
	}
	// Should: 0 < TrendingFactor < 1
	if c.TrendingFactor <= 0 || c.TrendingFactor >= 1 {
		return ErrDetectorTrendingFactor
	}
	// Should: len(DefaultThresholdMaxs) <= 8
	if len(c.DefaultThresholdMaxs) > MaxNumDefaultThresholdMaxs {
		return ErrDetectorDefaultThresholdMaxsLen
	}
	// Should: len(DefaultThresholdMins) <= 8
	if len(c.DefaultThresholdMins) > MaxNumDefaultThresholdMins {
		return ErrDetectorDefaultThresholdMinsLen
	}
	// Should: No zero values in DefaultThresholdMaxs
	for _, v := range c.DefaultThresholdMaxs {
		if v == 0 {
			return ErrDetectorDefaultThresholdMaxZero
		}
	}
	// Should: No zero values in DefaultThresholdMins
	for _, v := range c.DefaultThresholdMins {
		if v == 0 {
			return ErrDetectorDefaultThresholdMinZero
		}
	}
	// Should: len(FillBlankZeros) <= 8
	if len(c.FillBlankZeros) > MaxFillBlankZerosLen {
		return ErrDetectorFillBlankZerosLen
	}
	return nil
}

func (c *configWebapp) validateWebapp() error {
	// Should: 0 < Port < 65536
	if c.Port < 1 || c.Port > 65535 {
		return ErrWebappPort
	}
	// Should : Language in Supported
	b := false
	for _, lg := range WebappSupportedLanguages {
		if lg == c.Language {
			b = true
			break
		}
	}
	if !b {
		return ErrWebappLanguage
	}
	return nil
}

func (c *configAlerter) validateAlerter() error {
	// Should: Interval > 0
	if c.Interval <= 0 {
		return ErrAlerterInterval
	}
	// Should: OneDayLimit > 0
	if c.OneDayLimit <= 0 {
		return ErrAlerterOneDayLimit
	}
	// Should: 0 <= SilentStart <= 23
	if c.DefaultSilentTimeRange[0] < 0 || c.DefaultSilentTimeRange[0] > 23 {
		return ErrAlerterDefaultSilentTimeRange
	}
	// Should: 0 <= SilentEnd <= 23
	if c.DefaultSilentTimeRange[1] < 0 || c.DefaultSilentTimeRange[1] > 23 {
		return ErrAlerterDefaultSilentTimeRange
	}
	return nil
}

func (c *configCleaner) validateCleaner(period uint32) error {
	// Should: Threshold >= 2 * Period
	if c.Threshold < period*MinCleanerThresholdNumToPeriod {
		return ErrCleanerThreshold
	}
	return nil
}
