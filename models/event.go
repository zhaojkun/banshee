// Copyright 2016 Eleme Inc. All rights reserved.

package models

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
)

// Event is the alerting event.
type Event struct {
	ID                    string   `json:"id"`
	Project               *Project `json:"project"`
	User                  *User    `json:"user"`
	Rule                  *Rule    `json:"rule"`
	Index                 *Index   `json:"index"`
	Metric                *Metric  `json:"metric"`
	RuleTranslatedComment string   `json:"ruleTranslatedComment"`
}

// NewEvent returns a new event from metric and index.
func NewEvent(m *Metric, idx *Index) *Event {
	ev := &Event{Metric: m, Index: idx}
	ev.generateID()
	return ev
}

// generateID generates a sha1 string id for the event.
func (ev *Event) generateID() {
	slug := fmt.Sprintf("%s:%d", ev.Metric.Name, ev.Metric.Stamp)
	hash := sha1.New()
	hash.Write([]byte(slug))
	ev.ID = hex.EncodeToString(hash.Sum(nil))
}

// TranslateRuleComment translates rule comment variables with metric name and
// rule pattern.
//
//	m := &Metric{Name: "timer.count_ps.foo"}
//	r := &Rule{Pattern: "timer.count_ps.*", Comment: "$1 timing"}
//	ev := &Event{Metric:m, Rule:r}
//	ev.TranslateRuleComment()  // ev.RuleTranslatedComment => "foo timing"
//
func (ev *Event) TranslateRuleComment() {
	patternParts := strings.Split(ev.Rule.Pattern, ".")
	metricParts := strings.Split(ev.Metric.Name, ".")
	if len(patternParts) != len(metricParts) { // Unexcepted input metric and pattern.
		ev.RuleTranslatedComment = ev.Rule.Comment // Use original comment
		return
	}
	i := 0
	s := ev.Rule.Comment
	for j, patternPart := range patternParts {
		if patternPart == "*" {
			i++
			repl := fmt.Sprintf("$%d", i)
			s = strings.Replace(s, repl, metricParts[j], 1)
		}
	}
	ev.RuleTranslatedComment = s
}
