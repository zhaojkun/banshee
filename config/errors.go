package config

import "errors"

// Errors
var (
	// Error
	ErrInterval                        = errors.New("interval should be an integer between 1s~10min")
	ErrPeriod                          = errors.New("period should be an integer greater than interval")
	ErrPeriodTooSmall                  = errors.New("period should be at least 1 hour")
	ErrExpiration                      = errors.New("expiration should be an integer greater than (or equal to) 5 * period")
	ErrExpirationDivPeriodClean        = errors.New("expiration should be a multiple of period")
	ErrDetectorPort                    = errors.New("invalid detector.port")
	ErrDetectorTrendingFactor          = errors.New("detector.trending_factor should be a float between 0 and 1")
	ErrDetectorFilterTimes             = errors.New("detector.filter_times should be smaller")
	ErrDetectorDefaultThresholdMaxsLen = errors.New("detector.default_threshold_maxs should have up to 8 items")
	ErrDetectorDefaultThresholdMinsLen = errors.New("detector.default_threshold_mins should have up to 8 items")
	ErrDetectorDefaultThresholdMaxZero = errors.New("detector.default_threshold_maxs should not contain zeros")
	ErrDetectorDefaultThresholdMinZero = errors.New("detector.default_threshold_mins should not contain zeros")
	ErrDetectorFillBlankZerosLen       = errors.New("detector.fill_blank_zeros should have up to 8 items")
	ErrWebappPort                      = errors.New("invalid webapp.port")
	ErrWebappLanguage                  = errors.New("invalid webapp language")
	ErrAlerterInterval                 = errors.New("alerter.interval should be greater than 0")
	ErrAlerterOneDayLimit              = errors.New("alerter.one_day_limit should be greater than 0")
	ErrAlerterDefaultSilentTimeRange   = errors.New("alerter.default_silent_time_range should be 2 numbers between 0~24")
	// Warn
	ErrAlerterCommandEmpty = errors.New("alerter.command is empty")
)
