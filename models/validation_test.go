// Copyright 2016 Eleme Inc. All rights reserved.

package models

import (
	"github.com/eleme/banshee/util"
	"math/rand"
	"testing"
)

func genLongString(length int) string {
	d := "abcdefghijk0123456789"
	s := ""
	for i := 0; i < length; i++ {
		s = s + string(d[rand.Intn(len(d))])
	}
	return s
}

func TestValidateProjectName(t *testing.T) {
	util.Must(t, ValidateProjectName("") == ErrProjectNameEmpty)
	util.Must(t, ValidateProjectName(genLongString(MaxProjectNameLen+1)) == ErrProjectNameTooLong)
	util.Must(t, ValidateProjectName("project") == nil)
}

func TestValidateProjectSilentTimeRange(t *testing.T) {
	util.Must(t, ValidateProjectSilentRange(39, 6) == ErrProjectSilentTimeStart)
	util.Must(t, ValidateProjectSilentRange(0, 29) == ErrProjectSilentTimeEnd)
	util.Must(t, ValidateProjectSilentRange(7, 4) == ErrProjectSilentTimeRange)
	util.Must(t, ValidateProjectSilentRange(1, 9) == nil)
}

func TestValidateUserName(t *testing.T) {
	util.Must(t, ValidateUserName("") == ErrUserNameEmpty)
	util.Must(t, ValidateUserName(genLongString(MaxUserNameLen+1)) == ErrUserNameTooLong)
	util.Must(t, ValidateUserName("user") == nil)
}

func TestValidateUserEmail(t *testing.T) {
	util.Must(t, ValidateUserEmail("") == ErrUserEmailEmpty)
	util.Must(t, ValidateUserEmail("abc") == ErrUserEmailFormat)
	util.Must(t, ValidateUserEmail("hit9@ele.me") == nil)
}

func TestValidateUserPhone(t *testing.T) {
	util.Must(t, ValidateUserPhone("123456789012") == ErrUserPhoneLen)
	util.Must(t, ValidateUserPhone("12345678a01") == ErrUserPhoneFormat)
	util.Must(t, ValidateUserPhone("18701616177") == nil)
}

func TestValidateRulePattern(t *testing.T) {
	util.Must(t, ValidateRulePattern("") == ErrRulePatternEmpty)
	util.Must(t, ValidateRulePattern("abc efg") == ErrRulePatternContainsSpace)
	util.Must(t, ValidateRulePattern("abc*.s") == ErrRulePatternFormat)
	util.Must(t, ValidateRulePattern("abc.*.s") == nil)
	util.Must(t, ValidateRulePattern("abc.*.*") == nil)
	util.Must(t, ValidateRulePattern("*.abc.*") == nil)
}

func TestValidateRuleLevel(t *testing.T) {
	util.Must(t, ValidateRuleLevel(RuleLevelLow) == nil)
	util.Must(t, ValidateRuleLevel(RuleLevelMiddle) == nil)
	util.Must(t, ValidateRuleLevel(RuleLevelHigh) == nil)
	util.Must(t, ValidateRuleLevel(2016) == ErrRuleLevel)
}

func TestValidateMetricName(t *testing.T) {
	util.Must(t, ValidateMetricName("") == ErrMetricNameEmpty)
	util.Must(t, ValidateMetricName(genLongString(MaxMetricNameLen+1)) == ErrMetricNameTooLong)
}

func TestValidateMetricStamp(t *testing.T) {
	util.Must(t, ValidateMetricStamp(123) == ErrMetricStampTooSmall)
}
