// Copyright 2015 Eleme Inc. All rights reserved.

// Package config handles configuration parsing.
package config

import (
	"io/ioutil"

	"github.com/eleme/banshee/util/log"
	"gopkg.in/yaml.v2"
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
	DefaultIntervalHitLimit uint32 = 512
	// Default period for all metrics in seconds.
	DefaultPeriod uint32 = 1 * Day
	// Default metric expiration.
	DefaultExpiration uint32 = 7 * Day
	// Default filter offset to query history metrics.
	DefaultFilterOffset float64 = 0.01
	// Default filter times to query history metrics.
	DefaultFilterTimes int = 4
	// Default value of alerting interval.
	DefaultAlerterInterval uint32 = 20 * Minute
	// Default value of alerting check interval.
	DefaultAlerterCheckInterval uint32 = Minute
	// Default value of number of alerts after which we should send notifications.
	DefaultNotifyAfter = 1
	// Default value of alert times limit in one day for the same metric
	DefaultAlerterOneDayLimit uint32 = 10
	// Default value of least count.
	DefaultLeastCount uint32 = 5 * Minute / DefaultInterval
	// Default alerting silent time range.
	DefaultSilentTimeStart int = 0
	DefaultSilentTimeEnd   int = 6
	// Default language for webapp.
	DefaultWebappLanguage string = "en"
	// Default detection warning timeout, in ms.
	DefaultDetectionWarningTimeout = 300
	// Default alert command execution timeout, in seconds.
	DefaultAlertExecCommandTimeout = 5
	// Default trending factor for low level rules.
	DefaultTrendingFactorLowLevel float64 = 0.1
	// Default trending factor for middle level rules.
	DefaultTrendingFactorMiddleLevel float64 = 0.2
	// Default trending factor for high level rules.
	DefaultTrendingFactorHighLevel float64 = 0.3
	// Default idle metric check interval, in seconds.
	DefaultIdleMetricCheckInterval = 60
	// Default idle metric track limit.
	DefaultIdleMetricTrackLimit = 60
	// Default bool if detector should use recent data.
	DefaultUsingRecentDataPoints = true
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
	// Min value for the period.
	MinPeriod uint32 = 1 * Hour // 1h
)

// WebappSupportedLanguages lists webapp supported languages.
var WebappSupportedLanguages = []string{"en", "zh"}

// Config is the configuration container.
type Config struct {
	Interval   uint32         `json:"interval" yaml:"interval"`
	Period     uint32         `json:"period" yaml:"period"`
	Expiration uint32         `json:"expiration" yaml:"expiration"`
	Storage    configStorage  `json:"storage" yaml:"storage"`
	Detector   configDetector `json:"detector" yaml:"detector"`
	Webapp     configWebapp   `json:"webapp" yaml:"webapp"`
	Alerter    configAlerter  `json:"alerter" yaml:"alerter"`
	Notifier   configNotifier `json:"notifier" yaml:"notifier"`
	Cluster    configCluster  `json:"cluster" yaml:"cluster"`
}

type configStorage struct {
	Path  string      `json:"path" yaml:"path"`
	Admin configAdmin `json:"admin" yaml:"admin"`
}

type configAdmin struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	DBName   string `json:"dbName" yaml:"dbname"`
}
type configDetector struct {
	Port                      int                `json:"port" yaml:"port"`
	TrendingFactorLowLevel    float64            `json:"trendingFactorLowLevel" yaml:"trending_factor_low_level"`
	TrendingFactorMiddleLevel float64            `json:"trendingFactorMiddleLevel" yaml:"trending_factor_middle_level"`
	TrendingFactorHighLevel   float64            `json:"trendingFactorHighLevel" yaml:"trending_factor_high_level"`
	FilterOffset              float64            `json:"filterOffset" yaml:"filter_offset"`
	FilterTimes               int                `json:"filterTimes" yaml:"filter_times"`
	LeastCount                uint32             `json:"leastCount" yaml:"least_count"`
	BlackList                 []string           `json:"blackList" yaml:"blacklist"`
	EnableIntervalHitLimit    bool               `json:"enableIntervalHitLimit" yaml:"enable_interval_hit_limit"`
	IntervalHitLimit          uint32             `json:"intervalHitLimit" yaml:"interval_hit_limit"`
	IntervalLimitIgnoreList   []string           `json:"intervalLimitIgnoreList" yaml:"interval_limit_ignore_list"`
	DefaultThresholdMaxs      map[string]float64 `json:"defaultThresholdMaxs" yaml:"default_threshold_maxs"`
	DefaultThresholdMins      map[string]float64 `json:"defaultThresholdMins" yaml:"default_threshold_mins"`
	FillBlankZeros            []string           `json:"fillBlankZeros" yaml:"fill_blank_zeros"`
	WarningTimeout            int                `json:"warningTimeout" yaml:"warning_timeout"`
	IdleMetricCheckList       []string           `json:"idleMetricCheckList" yaml:"idle_metric_check_list"`
	IdleMetricCheckInterval   int                `json:"idleMetricCheckInterval" yaml:"idle_metric_check_interval"`
	IdleMetricTrackLimit      int                `json:"idleMetricTrackLimit" yaml:"idle_metric_track_limit"`
	UsingRecentDataPoints     bool               `json:"using_recent_data_points" yaml:"using_recent_data_points"`
}

type configWebapp struct {
	Port          int      `json:"port" yaml:"port"`
	Auth          []string `json:"auth" yaml:"auth"`
	Static        string   `json:"static" yaml:"static"`
	Language      string   `json:"language" yaml:"language"`
	URLPrefix     string   `json:"urlPrefix" yaml:"url_prefix"`
	PrivateDocURL string   `json:"privateDocUrl" yaml:"private_doc_url"`
	GraphiteURL   string   `json:"graphiteUrl" yaml:"graphite_url"`
}

type configAlerter struct {
	Command                string   `json:"command" yaml:"command"`
	ExecCommandTimeout     int      `json:"execCommandTimeOut" yaml:"exec_command_time_out"`
	Workers                int      `json:"workers" yaml:"workers"`
	Interval               uint32   `json:"interval" yaml:"interval"`
	AlertCheckInterval     uint32   `json:"alert_check_interval" yaml:"alert_check_interval"`
	NotifyAfter            int      `json:"notify_after" yaml:"notify_after"`
	OneDayLimit            uint32   `json:"oneDayLimit" yaml:"one_day_limit"`
	DefaultSilentTimeRange []int    `json:"defaultSilentTimeRange" yaml:"default_silent_time_range"`
	BlackList              []string `json:"blackList" yaml:"blacklist"`
}

type configNotifier struct {
	SlackURL string `json:"slackURL" yaml:"slack_url"`
}

type configCluster struct {
	Master       bool   `json:"master" yaml:"master"`
	QueueDSN     string `json:"queueDSN" yaml:"queue_dsn"`
	VHost        string `json:"vHost" yaml:"v_host"`
	ExchangeName string `json:"exchangeName" yaml:"exchange_name"`
	QueueName    string `json:"queueName" yaml:"queue_name"`
}

// New creates a Config with default values.
func New() *Config {
	c := new(Config)
	c.Interval = DefaultInterval
	c.Period = DefaultPeriod
	c.Expiration = DefaultExpiration
	c.Storage.Path = "./data"
	c.Storage.Admin.Host = "127.0.0.1"
	c.Storage.Admin.Port = 3306
	c.Storage.Admin.User = "banshee"
	c.Storage.Admin.Password = ""
	c.Storage.Admin.DBName = "banshee"
	c.Detector.Port = 2015
	c.Detector.TrendingFactorLowLevel = DefaultTrendingFactorLowLevel
	c.Detector.TrendingFactorMiddleLevel = DefaultTrendingFactorMiddleLevel
	c.Detector.TrendingFactorHighLevel = DefaultTrendingFactorHighLevel
	c.Detector.FilterOffset = DefaultFilterOffset
	c.Detector.FilterTimes = DefaultFilterTimes
	c.Detector.LeastCount = DefaultLeastCount
	c.Detector.BlackList = []string{}
	c.Detector.EnableIntervalHitLimit = true
	c.Detector.IntervalHitLimit = DefaultIntervalHitLimit
	c.Detector.DefaultThresholdMaxs = make(map[string]float64, 0)
	c.Detector.DefaultThresholdMins = make(map[string]float64, 0)
	c.Detector.FillBlankZeros = []string{}
	c.Detector.WarningTimeout = DefaultDetectionWarningTimeout
	c.Detector.IdleMetricCheckList = []string{}
	c.Detector.IdleMetricCheckInterval = DefaultIdleMetricCheckInterval
	c.Detector.IdleMetricTrackLimit = DefaultIdleMetricTrackLimit
	c.Detector.UsingRecentDataPoints = DefaultUsingRecentDataPoints
	c.Webapp.Port = 2016
	c.Webapp.Auth = []string{"admin", "admin"}
	c.Webapp.Static = "static/dist"
	c.Webapp.Language = DefaultWebappLanguage
	c.Webapp.PrivateDocURL = ""
	c.Webapp.GraphiteURL = ""
	c.Alerter.Command = ""
	c.Alerter.ExecCommandTimeout = DefaultAlertExecCommandTimeout
	c.Alerter.Workers = 4
	c.Alerter.Interval = DefaultAlerterInterval
	c.Alerter.AlertCheckInterval = DefaultAlerterCheckInterval
	c.Alerter.NotifyAfter = DefaultNotifyAfter
	c.Alerter.OneDayLimit = DefaultAlerterOneDayLimit
	c.Alerter.DefaultSilentTimeRange = []int{DefaultSilentTimeStart, DefaultSilentTimeEnd}
	c.Alerter.BlackList = []string{}
	c.Cluster.Master = true
	return c
}

// UpdateWithYamlFile updates the config from a yaml file.
func (c *Config) UpdateWithYamlFile(fileName string) error {
	log.Debugf("read config from %s..", fileName)
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(b, c)
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
	cfg.Storage.Admin = c.Storage.Admin
	cfg.Detector.Port = c.Detector.Port
	cfg.Detector.TrendingFactorLowLevel = c.Detector.TrendingFactorLowLevel
	cfg.Detector.TrendingFactorMiddleLevel = c.Detector.TrendingFactorMiddleLevel
	cfg.Detector.TrendingFactorHighLevel = c.Detector.TrendingFactorHighLevel
	cfg.Detector.FilterOffset = c.Detector.FilterOffset
	cfg.Detector.FilterTimes = c.Detector.FilterTimes
	cfg.Detector.LeastCount = c.Detector.LeastCount
	cfg.Detector.BlackList = c.Detector.BlackList
	cfg.Detector.DefaultThresholdMaxs = c.Detector.DefaultThresholdMaxs
	cfg.Detector.DefaultThresholdMins = c.Detector.DefaultThresholdMins
	cfg.Detector.FillBlankZeros = c.Detector.FillBlankZeros
	cfg.Detector.EnableIntervalHitLimit = c.Detector.EnableIntervalHitLimit
	cfg.Detector.IntervalHitLimit = c.Detector.IntervalHitLimit
	cfg.Detector.WarningTimeout = c.Detector.WarningTimeout
	cfg.Detector.IdleMetricCheckList = c.Detector.IdleMetricCheckList
	cfg.Detector.IdleMetricCheckInterval = c.Detector.IdleMetricCheckInterval
	cfg.Detector.IdleMetricTrackLimit = c.Detector.IdleMetricTrackLimit
	cfg.Detector.UsingRecentDataPoints = c.Detector.UsingRecentDataPoints
	cfg.Webapp.Port = c.Webapp.Port
	cfg.Webapp.Auth = c.Webapp.Auth
	cfg.Webapp.Static = c.Webapp.Static
	cfg.Webapp.Language = c.Webapp.Language
	cfg.Webapp.URLPrefix = c.Webapp.URLPrefix
	cfg.Webapp.PrivateDocURL = c.Webapp.PrivateDocURL
	cfg.Webapp.GraphiteURL = c.Webapp.GraphiteURL
	cfg.Alerter.Command = c.Alerter.Command
	cfg.Alerter.ExecCommandTimeout = c.Alerter.ExecCommandTimeout
	cfg.Alerter.Workers = c.Alerter.Workers
	cfg.Alerter.Interval = c.Alerter.Interval
	cfg.Alerter.OneDayLimit = c.Alerter.OneDayLimit
	cfg.Alerter.DefaultSilentTimeRange = c.Alerter.DefaultSilentTimeRange
	cfg.Alerter.NotifyAfter = c.Alerter.NotifyAfter
	cfg.Alerter.AlertCheckInterval = c.Alerter.AlertCheckInterval
	cfg.Alerter.BlackList = c.Alerter.BlackList
	cfg.Notifier.SlackURL = c.Notifier.SlackURL
	cfg.Cluster = c.Cluster
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
	return c.Alerter.validateAlerter()
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
	// Should: Period >= MinPeriod
	if c.Period < MinPeriod {
		return ErrPeriodTooSmall
	}
	// Should: Expiration/Period = integer.
	if c.Expiration/c.Period*c.Period != c.Expiration {
		return ErrExpirationDivPeriodClean
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
	if c.TrendingFactorLowLevel <= 0 || c.TrendingFactorLowLevel >= 1 {
		return ErrDetectorTrendingFactor
	}
	if c.TrendingFactorMiddleLevel <= 0 || c.TrendingFactorMiddleLevel >= 1 {
		return ErrDetectorTrendingFactor
	}
	if c.TrendingFactorHighLevel <= 0 || c.TrendingFactorHighLevel >= 1 {
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
