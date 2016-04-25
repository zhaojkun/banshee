// Copyright 2016 Eleme Inc. All rights reserved.
// Source from github.com/hit9/idpool.

// Package idpool implements a reusable integer id pool.
//
// Example
//
//	pool := idpool.New(5,1024)
//	pool.Allocate() // 5
//	pool.Allocate() // 6
//	pool.Allocate(5)
//	pool.Allocate() // 5
//
package idpool

import (
	"math"
	"math/big"
	"sync"
)

// Pool is the id pool.
type Pool struct {
	lock   sync.RWMutex
	table  *big.Int
	high   int
	low    int
	length int
}

// New returns a new Pool for given range.
// Range [low,high) is left open and right closed.
// Setting high to 0 means high is MaxInt32.
func New(low, high int) *Pool {
	if high == 0 {
		high = math.MaxInt32
	}
	return &Pool{
		high:  high,
		low:   low,
		table: big.NewInt(0),
	}
}

// Allocate an id from the pool.
// Returns high if no id is available.
func (p *Pool) Allocate() int {
	p.lock.Lock()
	defer p.lock.Unlock()
	for i := p.low; i < p.high; i++ {
		if p.table.Bit(i) == 0 {
			p.table.SetBit(p.table, i, 1)
			p.length++
			return i
		}
	}
	return p.high
}

// Reserve an id from the pool.
func (p *Pool) Reserve(id int) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.table.Bit(id) == 0 {
		p.table.SetBit(p.table, id, 1)
		p.length++
	}
}

// Release an id back to the pool.
// Do nothing if the id is outside of the range.
func (p *Pool) Release(id int) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if id >= p.low && id < p.high {
		if p.table.Bit(id) == 1 {
			p.table.SetBit(p.table, id, 0)
			p.length--
		}
	}
}

// Clear the pool.
func (p *Pool) Clear() {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.table = big.NewInt(0)
}

// High returns the high.
func (p *Pool) High() int {
	return p.high
}

// Low returns the low.
func (p *Pool) Low() int {
	return p.low
}

// Len returns the number of id reserved or allocated.
func (p *Pool) Len() int {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.length
}
