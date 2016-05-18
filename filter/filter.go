// Copyright 2015 Eleme Inc. All rights reserved.

// Package filter implements fast wildcard like filtering based on trie.
package filter

import (
	"sync/atomic"
	"time"

	"github.com/eleme/banshee/config"
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/storage"
	"github.com/eleme/banshee/util/log"
	"github.com/eleme/banshee/util/trie"
)

// Filter is to filter metrics by rules.
type Filter struct {
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
	rule *models.Rule
	hits uint32
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
			f.addRule(rule)
		}
	}()
}

// initDelRuleListener starts a goroutine to listen on rule deletes.
func (f *Filter) initDelRuleListener() {
	go func() {
		for {
			rule := <-f.delRuleCh
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

// initRuleHitReseter starts a goroutine to reset all hit counters.
func (f *Filter) initRuleHitReseter() {
	ticker := time.NewTicker(time.Second * time.Duration(f.cfg.Interval))
	go func() {
		for _ = range ticker.C {
			for _, v := range f.trie.Map() {
				n := v.(*node)
				atomic.StoreUint32(&n.hits, 0)
			}
		}
	}()
}

// Init filter.
func (f *Filter) Init(db *storage.DB) {
	f.initFromDB(db)
	f.initAddRuleListener()
	f.initDelRuleListener()
	f.initRuleHitReseter()
}

// addRule adds a rule to the filter.
func (f *Filter) addRule(rule *models.Rule) {
	n := &node{rule: rule, hits: 0}
	f.trie.Put(rule.Pattern, n)
}

// delRule deletes a rule from the filter.
func (f *Filter) delRule(rule *models.Rule) {
	f.trie.Pop(rule.Pattern)
}

// MatchedRules returns the matched rules by metric name.
func (f *Filter) MatchedRules(m *models.Metric) (rules []*models.Rule) {
	d := f.trie.Matched(m.Name)
	for _, v := range d {
		n := v.(*node)
		if f.cfg.Detector.EnableIntervalHitLimit {
			hits := atomic.AddUint32(&n.hits, 1)
			if hits > f.cfg.Detector.IntervalHitLimit {
				log.Warnf("%s hits over interval hit limit", n.rule.Pattern)
				return []*models.Rule{}
			}
		}
		rules = append(rules, n.rule)
	}
	return
}
