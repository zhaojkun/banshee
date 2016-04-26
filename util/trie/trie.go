// Copyright 2016 Eleme Inc. All rights reserved.
// Source from github.com/hit9/trie.

// Package trie implements a in-memory trie tree.
// Reference: Trie - Wikipedia, the free encyclopedia
package trie

import (
	"strings"
	"sync"
)

// tree is the internal tree.
type tree struct {
	value    interface{}
	children map[string]*tree
	lock     sync.RWMutex // protects value, children
}

// Trie is the trie tree.
type Trie struct {
	root   *tree // root tree, won't be rewritten
	delim  string
	length int
	lock   sync.RWMutex // protects length
}

// newTree creates a new tree.
func newTree() *tree {
	return &tree{
		children: make(map[string]*tree, 0),
	}
}

// New creates a new Trie.
func New(delim string) *Trie {
	return &Trie{
		root:  newTree(),
		delim: delim,
	}
}

// Len returns the trie length.
func (tr *Trie) Len() int {
	tr.lock.RLock()
	defer tr.lock.RUnlock()
	return tr.length
}

// Put an item to the trie.
func (tr *Trie) Put(key string, value interface{}) {
	parts := strings.Split(key, tr.delim)
	t := tr.root
	if len(parts) > 0 {
		t.lock.Lock() // touch root
	}
	for i, part := range parts {
		child, ok := t.children[part]
		if !ok {
			child = newTree()
			t.children[part] = child
		}
		t.lock.Unlock()   // leave parent
		child.lock.Lock() // touch child
		if i == len(parts)-1 {
			if child.value == nil {
				tr.lock.Lock()
				tr.length++
				tr.lock.Unlock()
			}
			child.value = value
			child.lock.Unlock() // leave child
			return
		}
		t = child // child as next parent
	}
	return
}

// Get an item from the trie.
func (tr *Trie) Get(key string) interface{} {
	parts := strings.Split(key, tr.delim)
	t := tr.root
	if len(parts) > 0 {
		t.lock.RLock() // touch root
	}
	for i, part := range parts {
		child, ok := t.children[part]
		if !ok {
			t.lock.RUnlock() // leave parent.
			return nil
		}
		t.lock.RUnlock()   // leave parent
		child.lock.RLock() // touch child
		if i == len(parts)-1 {
			child.lock.RUnlock() // leave child
			return child.value
		}
		t = child // child as next parent
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
	parts := strings.Split(key, tr.delim)
	t := tr.root
	if len(parts) > 0 {
		t.lock.Lock() // touch root
	}
	for i, part := range parts {
		child, ok := t.children[part]
		if !ok {
			t.lock.Unlock() // leave parent
			return nil
		}
		t.lock.Unlock()   // leave parent
		child.lock.Lock() // touch child
		if i == len(parts)-1 {
			if len(child.children) == 0 {
				delete(t.children, part)
			}
			value := child.value
			child.value = nil
			if value != nil {
				tr.lock.Lock()
				tr.length--
				tr.lock.Unlock()
			}
			child.lock.Unlock() // leave child
			return value
		}
		t = child
	}
	return nil
}

// Clear the trie.
func (tr *Trie) Clear() {
	// Why not just tr.lock = newTree(): we don't want to rewrite the tr.root,
	// otherwise, accessing the tr.root requires a RWMutex.
	tr.root.lock.Lock()
	defer tr.root.lock.Unlock()
	tr.root.children = make(map[string]*tree, 0)
	tr.lock.Lock()
	defer tr.lock.Unlock()
	tr.length = 0
}

// Match a wildcard like pattern in the trie, the pattern is not a traditional
// wildcard, only "*" is supported.
func (tr *Trie) Match(pattern string) map[string]interface{} {
	return tr.root.match(tr.delim, nil, strings.Split(pattern, tr.delim))
}

// match keys in the tree recursively.
func (t *tree) match(delim string, keys []string, parts []string) map[string]interface{} {
	m := make(map[string]interface{}, 0)
	t.lock.RLock() // touch root.
	if len(parts) == 0 {
		if t.value != nil {
			// Generally, strings.Split() won't give us empty results. And the
			// empty parts means unprocessed parts are empty, indicates that we
			// should pick up all processed keys and return.
			m[strings.Join(keys, delim)] = t.value
		}
		t.lock.RUnlock() // leave root
		return m
	}
	for i, part := range parts {
		if part == "*" {
			for segment, child := range t.children {
				v := child.match(delim, append(keys, segment), parts[i+1:])
				for key, value := range v {
					m[key] = value
				}
			}
			t.lock.RUnlock() // leave parent
			return m
		}
		child, ok := t.children[part]
		if !ok {
			t.lock.RUnlock() // leave parent
			return m
		}
		t.lock.RUnlock()   // leave parent
		child.lock.RLock() // touch child
		keys = append(keys, part)
		if i == len(parts)-1 { // last part
			if child.value != nil {
				m[strings.Join(keys, delim)] = child.value
			}
			child.lock.RUnlock() // leave child
			return m
		}
		t = child // child as parent
	}
	return m
}

// Map returns the full trie as a map.
func (tr *Trie) Map() map[string]interface{} {
	return tr.root._map(tr.delim, nil)
}

// map returns the full tree as a map.
func (t *tree) _map(delim string, keys []string) map[string]interface{} {
	m := make(map[string]interface{}, 0)
	t.lock.RLock() // touch root
	// Check current tree.
	if t.value != nil {
		m[strings.Join(keys, delim)] = t.value
	}
	// Check children.
	for segment, child := range t.children {
		d := child._map(delim, append(keys, segment))
		for key, value := range d {
			m[key] = value
		}
	}
	t.lock.RUnlock()
	return m
}
