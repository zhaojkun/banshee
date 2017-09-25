// Copyright 2015 Eleme Inc. All rights reserved.

package admindb

import (
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util/log"
	"github.com/eleme/banshee/util/safemap"
	"github.com/jinzhu/gorm"
)

type rulesCache struct {
	// Cache
	rules *safemap.SafeMap
	// Listeners
	lns []chan *models.Message
}

// newRulesCache creates a rulesCache.
func newRulesCache() *rulesCache {
	c := new(rulesCache)
	c.rules = safemap.New()
	c.lns = make([]chan *models.Message, 0)
	return c
}

// Init cache from db.
func (c *rulesCache) Init(db *gorm.DB) error {
	log.Debugf("init rules from admindb..")
	// Query
	var rules []models.Rule
	err := db.Find(&rules).Error
	if err != nil {
		return err
	}
	// Load
	for i := 0; i < len(rules); i++ {
		rule := &rules[i]
		rule.Share()
		c.rules.Set(rule.ID, rule)
	}
	return nil
}

// Len returns the number of rules in cache.
func (c *rulesCache) Len() int {
	return c.rules.Len()
}

// Get returns rule.
func (c *rulesCache) Get(id int) (*models.Rule, bool) {
	r, ok := c.rules.Get(id)
	if !ok {
		return nil, false
	}
	rule := r.(*models.Rule)
	return rule.Copy(), true
}

// Put a rule into cache.
func (c *rulesCache) Put(rule *models.Rule) bool {
	if c.rules.Has(rule.ID) {
		return false
	}
	r := rule.Copy()
	r.Share()
	c.rules.Set(rule.ID, r)
	c.push(&models.Message{Type: models.RULEADD, Rule: rule.Copy()})
	return true
}

// All returns all rules.
func (c *rulesCache) All() (rules []*models.Rule) {
	for _, v := range c.rules.Items() {
		rule := v.(*models.Rule)
		rules = append(rules, rule.Copy())
	}
	return rules
}

// Delete a rule from cache.
func (c *rulesCache) Delete(id int) bool {
	r, ok := c.rules.Pop(id)
	if ok {
		rule := r.(*models.Rule)
		c.push(&models.Message{Type: models.RULEDELETE, Rule: rule.Copy()})
		return true
	}
	return false
}

// OnChange listens rules changes.
func (c *rulesCache) OnChange(ch chan *models.Message) {
	c.lns = append(c.lns, ch)
}

// Pushes changed rule to listeners.
func (c *rulesCache) push(msg *models.Message) {
	for _, ch := range c.lns {
		select {
		case ch <- msg:
		default:
			log.Errorf("buffered added rules chan is full, skipping..")
		}
	}
}
