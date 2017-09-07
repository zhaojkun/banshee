// Copyright 2016 Eleme Inc. All rights reserved.

package models

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
)

// Event is the alert event.
type Event struct {
	ID     string  `json:"id"`
	Index  *Index  `json:"index"`
	Metric *Metric `json:"metric"`
	Rule   *Rule   `json:"rule"`
}

// NewEvent returns a new event from metric and index.
func NewEvent(m *Metric, idx *Index, rule *Rule) *Event {
	ev := &Event{Metric: m, Index: idx, Rule: rule}
	ev.generateID()
	return ev
}

// generateID generates a sha1 string id for the event.
func (ev *Event) generateID() {
	slug := fmt.Sprintf("%s:%d:%d", ev.Metric.Name, ev.Metric.Stamp, ev.Rule.ID)
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
//	ev.TranslateRuleComment()  // => "foo timing"
//
func (ev *Event) TranslateRuleComment() string {
	patternParts := strings.Split(ev.Rule.Pattern, ".")
	metricParts := strings.Split(ev.Metric.Name, ".")
	if len(patternParts) != len(metricParts) { // Unexcepted input metric and pattern.
		return ev.Rule.Comment // Use original comment
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
	return s
}

// EventWrapper is a wrapper of Event for tmp usage.
type EventWrapper struct {
	*Event
	Team                  *Team    `json:"team"`
	Project               *Project `json:"project"`
	User                  *User    `json:"user,omitempty"`
	AlarmUsers            []*User  `json:"alarmUsers,omitempty"`
	RuleTranslatedComment string   `json:"ruleTranslatedComment"`
}

// NewWrapperOfEvent creates an event wrapper from given event.
func NewWrapperOfEvent(ev *Event) *EventWrapper {
	return &EventWrapper{Event: ev, RuleTranslatedComment: ev.TranslateRuleComment()}
}
