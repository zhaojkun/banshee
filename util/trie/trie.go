// Copyright 2016 Eleme Inc. All rights reserved.
// Source from github.com/hit9/trie.

// Package trie implements a in-memory trie tree.
// Reference: Trie - Wikipedia, the free encyclopedia
package trie

import (
	"strings"
	"sync"
)

// delim is the metric name delimeter, in banshee is a single dot.
const delim = "."

// tree is the internal tree.
type tree struct {
	value    interface{}
	children map[string]*tree
}

// Trie is the trie tree.
type Trie struct {
	root   *tree // root tree, won't be rewritten
	length int
	lock   sync.RWMutex // protects the whole trie
}

// newTree creates a new tree.
func newTree() *tree {
	return &tree{
		children: make(map[string]*tree, 0),
	}
}

// New creates a new Trie.
func New() *Trie {
	return &Trie{
		root: newTree(),
	}
}

// Len returns the trie length.
func (tr *Trie) Len() int {
	tr.lock.RLock()
	defer tr.lock.RUnlock()
	return tr.length
}

// Put an item to the trie.
// Replace if the key conflicts.
func (tr *Trie) Put(key string, value interface{}) {
	tr.lock.Lock()
	defer tr.lock.Unlock()
	parts := strings.Split(key, delim)
	t := tr.root
	for i, part := range parts {
		child, ok := t.children[part]
		if !ok {
			child = newTree()
			t.children[part] = child
		}
		if i == len(parts)-1 {
			if child.value == nil {
				tr.length++
			}
			child.value = value
			return
		}
		t = child
	}
	return
}

// Get an item from the trie.
// Returns nil if not found.
func (tr *Trie) Get(key string) interface{} {
	tr.lock.RLock()
	defer tr.lock.RUnlock()
	parts := strings.Split(key, delim)
	t := tr.root
	for i, part := range parts {
		child, ok := t.children[part]
		if !ok {
			return nil
		}
		if i == len(parts)-1 {
			return child.value
		}
		t = child
	}
	return nil
}

// Has checks if an item is in trie.
// Returns true if the given key is in the trie.
func (tr *Trie) Has(key string) bool {
	return tr.Get(key) != nil
}

// Pop an item from the trie.
// Returns nil if the given key is not in the trie.
func (tr *Trie) Pop(key string) interface{} {
	tr.lock.Lock()
	defer tr.lock.Unlock()
	parts := strings.Split(key, delim)
	t := tr.root
	for i, part := range parts {
		child, ok := t.children[part]
		if !ok {
			return nil
		}
		if i == len(parts)-1 {
			if len(child.children) == 0 {
				delete(t.children, part)
			}
			value := child.value
			child.value = nil
			if value != nil {
				tr.length--
			}
			return value
		}
		t = child
	}
	return nil
}

// Clear the trie.
func (tr *Trie) Clear() {
	tr.lock.Lock()
	defer tr.lock.Unlock()
	tr.root = newTree() // gc
	tr.length = 0
}

// Match trie items by a wildcard like pattern, the pattern is not a
// traditional wildcard, only "*" is supported, a sinle "*" represents a single
// word.
// Returns an empty map if the given pattern matches no items.
func (tr *Trie) Match(pattern string) map[string]interface{} {
	tr.lock.RLock()
	defer tr.lock.RUnlock()
	return tr.root.match(nil, strings.Split(pattern, delim))
}

// match keys in the tree recursively.
func (t *tree) match(keys []string, parts []string) map[string]interface{} {
	m := make(map[string]interface{})
	if len(parts) == 0 {
		if t.value != nil {
			// Generally, strings.Split() won't give us empty results. And the
			// empty parts means unprocessed parts are empty, indicates that we
			// should pick up all processed keys and return.
			m[strings.Join(keys, delim)] = t.value
		}
		return m
	}
	for i, part := range parts {
		if part == "*" {
			for segment, child := range t.children {
				v := child.match(append(keys, segment), parts[i+1:])
				for key, value := range v {
					m[key] = value
				}
			}
			return m
		}
		child, ok := t.children[part]
		if !ok {
			return m
		}
		keys = append(keys, part)
		if i == len(parts)-1 { // last part
			if child.value != nil {
				m[strings.Join(keys, delim)] = child.value
			}
			return m
		}
		t = child // child as parent
	}
	return m
}

// Map returns the full trie as a map.
func (tr *Trie) Map() map[string]interface{} {
	tr.lock.RLock()
	defer tr.lock.RUnlock()
	return tr.root._map(nil)
}

// map returns the full tree as a map.
func (t *tree) _map(keys []string) map[string]interface{} {
	m := make(map[string]interface{})
	// Check current tree.
	if t.value != nil {
		m[strings.Join(keys, delim)] = t.value
	}
	// Check children.
	for segment, child := range t.children {
		d := child._map(append(keys, segment))
		for key, value := range d {
			m[key] = value
		}
	}
	return m
}

// Matched uses the trie items as the wildcard like patterns, filters out the
// items matches the given string.
// Returns an empty map if the given strings matches no patterns.
func (tr *Trie) Matched(s string) map[string]interface{} {
	tr.lock.RLock()
	defer tr.lock.RUnlock()
	return tr.root.matched(nil, strings.Split(s, delim))
}

// matched returns the patterns matched the given string.
func (t *tree) matched(keys, parts []string) map[string]interface{} {
	m := make(map[string]interface{})
	if len(parts) == 0 && t.value != nil {
		m[strings.Join(keys, delim)] = t.value
		return m
	}
	if len(parts) > 0 {
		if child, ok := t.children["*"]; ok {
			for k, v := range child.matched(append(keys, "*"), parts[1:]) {
				m[k] = v
			}
		}
		if child, ok := t.children[parts[0]]; ok {
			for k, v := range child.matched(append(keys, parts[0]), parts[1:]) {
				m[k] = v
			}
		}
	}
	return m
}
