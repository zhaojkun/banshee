// Copyright 2015 Eleme Inc. All rights reserved.

// Package filter implements fast wildcard like filtering based on trie.
package filter

import (
	"sync"

	"github.com/eleme/banshee/config"
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/storage"
	"github.com/eleme/banshee/util/log"
	"github.com/eleme/banshee/util/trie"
)

// Filter is to filter metrics by rules.
type Filter struct {
	sync.Mutex
	// Config
	cfg *config.Config
	// Rule changes
	addRuleCh chan *models.Rule
	delRuleCh chan *models.Rule
	// Trie
	trie *trie.Trie
}

// node is the trie node.
type node struct {
	sync.Mutex
	// Pattern
	pattern string
	// Rule
	rules []*models.Rule
	// Hit number about the rule in 'interval' seconds.
	hits uint32
	// resetStamp will be reset when income metric is after the time resetStamp+interval.
	resetStamp uint32
	interval   uint32
}

func (n *node) incrHits(m *models.Metric) uint32 {
	n.Lock()
	defer n.Unlock()
	if m.Stamp >= n.resetStamp+n.interval {
		n.resetStamp = m.Stamp / n.interval * n.interval
		n.hits = 0
	}
	n.hits++
	return n.hits
}

func (n *node) pushRule(r *models.Rule) {
	n.Lock()
	defer n.Unlock()
	for _, rule := range n.rules {
		if rule.ID == r.ID {
			return
		}
	}
	n.rules = append(n.rules, r)
}

func (n *node) popRule(r *models.Rule) {
	n.Lock()
	defer n.Unlock()
	idx := -1
	for i, rule := range n.rules {
		if rule.ID == r.ID {
			idx = i
			break
		}
	}
	if idx != -1 {
		n.rules = append(n.rules[:idx], n.rules[idx+1:]...)
	}
	return
}

func (n *node) currentRules() []*models.Rule {
	var rules []*models.Rule
	n.Lock()
	defer n.Unlock()
	for _, rule := range n.rules {
		rules = append(rules, rule)
	}
	return rules
}

// Limit for buffered changed rules
const bufferedChangedRulesLimit = 128

// New creates a new Filter.
func New(cfg *config.Config) *Filter {
	return &Filter{
		cfg:       cfg,
		addRuleCh: make(chan *models.Rule, bufferedChangedRulesLimit),
		delRuleCh: make(chan *models.Rule, bufferedChangedRulesLimit),
		trie:      trie.New(),
	}
}

// initAddRuleListener starts a goroutine to listen on new rules.
func (f *Filter) initAddRuleListener() {
	go func() {
		for {
			rule := <-f.addRuleCh
			log.Infof("filter add rule %s", rule.Pattern)
			f.addRule(rule)
		}
	}()
}

// initDelRuleListener starts a goroutine to listen on rule deletes.
func (f *Filter) initDelRuleListener() {
	go func() {
		for {
			rule := <-f.delRuleCh
			log.Infof("filter delete rule %s", rule.Pattern)
			f.delRule(rule)
		}
	}()
}

// initFromDB inits rules from db.
func (f *Filter) initFromDB(db *storage.DB) {
	log.Debugf("init filter's rules from cache..")
	// Listen rules changes.
	db.Admin.RulesCache.OnAdd(f.addRuleCh)
	db.Admin.RulesCache.OnDel(f.delRuleCh)
	// Add rules from cache
	rules := db.Admin.RulesCache.All()
	for _, rule := range rules {
		f.addRule(rule)
	}
}

// Init filter.
func (f *Filter) Init(db *storage.DB) {
	f.initFromDB(db)
	f.initAddRuleListener()
	f.initDelRuleListener()
}

// addRule adds a rule to the filter.
func (f *Filter) addRule(rule *models.Rule) {
	f.Lock()
	defer f.Unlock()
	var n *node
	v := f.trie.Get(rule.Pattern)
	if v == nil {
		n = &node{pattern: rule.Pattern, hits: 0, interval: f.cfg.Interval}
		f.trie.Put(rule.Pattern, n)
	} else {
		n = v.(*node)
	}
	n.pushRule(rule)
}

// delRule deletes a rule from the filter.
func (f *Filter) delRule(rule *models.Rule) {
	f.Lock()
	defer f.Unlock()
	v := f.trie.Get(rule.Pattern)
	if v != nil {
		n := v.(*node)
		n.popRule(rule)
		if len(n.currentRules()) == 0 {
			f.trie.Pop(rule.Pattern)
		}
	}
}

// MatchedRules returns the matched rules by metric name.
func (f *Filter) MatchedRules(m *models.Metric, shouldLimitHit bool) (rules []*models.Rule) {
	d := f.trie.Matched(m.Name)
	for _, v := range d {
		n := v.(*node)
		if shouldLimitHit && f.cfg.Detector.EnableIntervalHitLimit {
			hits := n.incrHits(m)
			if hits > f.cfg.Detector.IntervalHitLimit {
				log.Debugf("%s hits over interval hit limit", n.pattern)
				return []*models.Rule{}
			}
		}
		rules = append(rules, n.currentRules()...)
	}
	return
}
