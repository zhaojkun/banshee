// Copyright 2016 Eleme Inc. All rights reserved.

package models

import (
	"github.com/eleme/banshee/util/assert"
	"testing"
)

func TestGenerateID(t *testing.T) {
	// Metric with the same name but different stamps.
	ev1 := NewEvent(&Metric{Name: "foo", Stamp: 1456815973}, nil)
	ev2 := NewEvent(&Metric{Name: "foo", Stamp: 1456815974}, nil)
	assert.Ok(t, ev1.ID != ev2.ID)
	// Metric with the same stamp but different names.
	ev1 = NewEvent(&Metric{Name: "foo", Stamp: 1456815973}, nil)
	ev2 = NewEvent(&Metric{Name: "bar", Stamp: 1456815973}, nil)
	assert.Ok(t, ev1.ID != ev2.ID)
}

func TestTranslateRuleComment(t *testing.T) {
	m := &Metric{Name: "timer.count_ps.foo.bar"}
	r := &Rule{Pattern: "timer.count_ps.*.*", Comment: "$1 and $2 timing"}
	ev := &Event{Metric: m, Rule: r}
	ev.TranslateRuleComment()
	excepted := "foo and bar timing"
	assert.Ok(t, ev.RuleTranslatedComment == excepted)
}

func TestTranslateRuleCommentNoVariables(t *testing.T) {
	m := &Metric{Name: "foo.bar"}
	r := &Rule{Pattern: "foo.*", Comment: "no variables"}
	ev := &Event{Metric: m, Rule: r}
	ev.TranslateRuleComment()
	excepted := "no variables"
	assert.Ok(t, ev.RuleTranslatedComment == excepted)
}
