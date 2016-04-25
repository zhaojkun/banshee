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
	// Default filter times to query history metrics.
	DefaultFilterTimes int = 4
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
}

type configStorage struct {
	Path string `json:"path"`
}

type configDetector struct {
	Port                 int                `json:"port"`
	TrendingFactor       float64            `json:"trendingFactor"`
	FilterOffset         float64            `json:"filterOffset"`
	FilterTimes          int                `json:"filterTimes"`
	LeastCount           uint32             `json:"leastCount"`
	BlackList            []string           `json:"blackList"`
	IntervalHitLimit     int                `json:"intervalHitLimit"`
	DefaultThresholdMaxs map[string]float64 `json:"defaultThresholdMaxs"`
	DefaultThresholdMins map[string]float64 `json:"defaultThresholdMins"`
	FillBlankZeros       []string           `json:"fillBlankZeros"`
}

type configWebapp struct {
	Port          int       `json:"port"`
	Auth          [2]string `json:"auth"`
	Static        string    `json:"static"`
	Language      string    `json:"language"`
	PrivateDocURL string    `json:"privateDocUrl"`
}

type configAlerter struct {
	Command                string `json:"command"`
	Workers                int    `json:"workers"`
	Interval               uint32 `json:"inteval"`
	OneDayLimit            uint32 `json:"oneDayLimit"`
	DefaultSilentTimeRange [2]int `json:"defaultSilentTimeRange"`
}

// New creates a Config with default values.
func New() *Config {
	c := new(Config)
	c.Interval = DefaultInterval
	c.Period = DefaultPeriod
	c.Expiration = DefaultExpiration
	c.Storage.Path = "storage/"
	c.Detector.Port = 2015
	c.Detector.TrendingFactor = DefaultTrendingFactor
	c.Detector.FilterOffset = DefaultFilterOffset
	c.Detector.FilterTimes = DefaultFilterTimes
	c.Detector.LeastCount = DefaultLeastCount
	c.Detector.BlackList = []string{}
	c.Detector.IntervalHitLimit = DefaultIntervalHitLimit
	c.Detector.DefaultThresholdMaxs = make(map[string]float64, 0)
	c.Detector.DefaultThresholdMins = make(map[string]float64, 0)
	c.Detector.FillBlankZeros = []string{}
	c.Webapp.Port = 2016
	c.Webapp.Auth = [2]string{"admin", "admin"}
	c.Webapp.Static = "static/dist"
	c.Webapp.Language = DefaultWebappLanguage
	c.Webapp.PrivateDocURL = ""
	c.Alerter.Command = ""
	c.Alerter.Workers = 4
	c.Alerter.Interval = DefaultAlerterInterval
	c.Alerter.OneDayLimit = DefaultAlerterOneDayLimit
	c.Alerter.DefaultSilentTimeRange = [2]int{DefaultSilentTimeStart, DefaultSilentTimeEnd}
	return c
}

// UpdateWithJSONFile update the config from a json file.
func (c *Config) UpdateWithJSONFile(fileName string) error {
	log.Debugf("read config from %s..", fileName)
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, c)
	if err != nil {
		return err
	}
	return err
}

// Copy config.
func (c *Config) Copy() *Config {
	cfg := New()
	cfg.Interval = c.Interval
	cfg.Period = c.Period
	cfg.Expiration = c.Expiration
	cfg.Storage.Path = c.Storage.Path
	cfg.Detector.Port = c.Detector.Port
	cfg.Detector.TrendingFactor = c.Detector.TrendingFactor
	cfg.Detector.FilterOffset = c.Detector.FilterOffset
	cfg.Detector.FilterTimes = c.Detector.FilterTimes
	cfg.Detector.LeastCount = c.Detector.LeastCount
	cfg.Detector.BlackList = c.Detector.BlackList
	cfg.Detector.DefaultThresholdMaxs = c.Detector.DefaultThresholdMaxs
	cfg.Detector.DefaultThresholdMins = c.Detector.DefaultThresholdMins
	cfg.Detector.FillBlankZeros = c.Detector.FillBlankZeros
	cfg.Detector.IntervalHitLimit = c.Detector.IntervalHitLimit
	cfg.Webapp.Port = c.Webapp.Port
	cfg.Webapp.Auth = c.Webapp.Auth
	cfg.Webapp.Static = c.Webapp.Static
	cfg.Webapp.Language = c.Webapp.Language
	cfg.Webapp.PrivateDocURL = c.Webapp.PrivateDocURL
	cfg.Alerter.Command = c.Alerter.Command
	cfg.Alerter.Workers = c.Alerter.Workers
	cfg.Alerter.Interval = c.Alerter.Interval
	cfg.Alerter.OneDayLimit = c.Alerter.OneDayLimit
	cfg.Alerter.DefaultSilentTimeRange = c.Alerter.DefaultSilentTimeRange
	return cfg
}

// Validate config.
func (c *Config) Validate() error {
	if err := c.validateGlobals(); err != nil {
		return err
	}
	if err := c.Detector.validateDetector(c.Period, c.Expiration); err != nil {
		return err
	}
	if err := c.Webapp.validateWebapp(); err != nil {
		return err
	}
	if err := c.Alerter.validateAlerter(); err != nil {
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

func (c *configDetector) validateDetector(period uint32, expiration uint32) error {
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
	// Should: FilterTimes * Period < Expiration
	if uint32(c.FilterTimes)*period > expiration {
		return ErrDetectorFilterTimes
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
