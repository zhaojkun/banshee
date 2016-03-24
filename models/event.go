// Copyright 2016 Eleme Inc. All rights reserved.

package models

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

// Event is the alerting event.
type Event struct {
	ID      string   `json:"id"`
	Project *Project `json:"project"`
	User    *User    `json:"user"`
	Rule    *Rule    `json:"rule"`
	Index   *Index   `json:"index"`
	Metric  *Metric  `json:"metric"`
}

// NewEvent returns a new event from metric and index.
func NewEvent(m *Metric, idx *Index) *Event {
	ev := &Event{Metric: m, Index: idx}
	ev.generateID()
	return ev
}

// generateID generates a sha1 string id for the event.
func (ev *Event) generateID() {
	slug := fmt.Sprintf("%s:%s", ev.Metric.Name, ev.Metric.Stamp)
	hash := sha1.New()
	hash.Write([]byte(slug))
	ev.ID = hex.EncodeToString(hash.Sum(nil))
}
