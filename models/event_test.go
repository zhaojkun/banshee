// Copyright 2016 Eleme Inc. All rights reserved.

package models

import (
	"github.com/eleme/banshee/util"
	"testing"
)

func TestEventGenerateID(t *testing.T) {
	rule := &Rule{ID: 1}
	// Metric with the same name but different stamps.
	ev1 := NewEvent(&Metric{Name: "foo", Stamp: 1456815973}, nil, rule)
	ev2 := NewEvent(&Metric{Name: "foo", Stamp: 1456815974}, nil, rule)
	util.Must(t, ev1.ID != ev2.ID)
	// Metric with the same stamp but different names.
	ev1 = NewEvent(&Metric{Name: "foo", Stamp: 1456815973}, nil, rule)
	ev2 = NewEvent(&Metric{Name: "bar", Stamp: 1456815973}, nil, rule)
	util.Must(t, ev1.ID != ev2.ID)
}

func TestEventTranslateRuleComment(t *testing.T) {
	m := &Metric{Name: "timer.count_ps.foo.bar"}
	r := &Rule{Pattern: "timer.count_ps.*.*", Comment: "$1 and $2 timing"}
	ev := &Event{Metric: m, Rule: r}
	excepted := "foo and bar timing"
	util.Must(t, ev.TranslateRuleComment() == excepted)
}

func TestEventTranslateRuleCommentNoVariables(t *testing.T) {
	m := &Metric{Name: "foo.bar"}
	r := &Rule{Pattern: "foo.*", Comment: "no variables"}
	ev := &Event{Metric: m, Rule: r}
	excepted := "no variables"
	util.Must(t, ev.TranslateRuleComment() == excepted)
}
