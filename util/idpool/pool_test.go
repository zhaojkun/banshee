// Copyright 2016 Eleme Inc. All rights reserved.

package idpool

import (
	"github.com/eleme/banshee/util"
	"math"
	"math/rand"
	"testing"
)

func TestAllocate(t *testing.T) {
	p := New(1, 5)
	util.Must(t, p.Allocate() == 1)
	util.Must(t, p.Allocate() == 2)
	util.Must(t, p.Allocate() == 3)
	util.Must(t, p.Allocate() == 4)
	util.Must(t, p.Allocate() == 5)
	util.Must(t, p.Allocate() == 5)
}

func TestReserve(t *testing.T) {
	p := New(1, 5)
	util.Must(t, p.Allocate() == 1)
	p.Reserve(2)
	util.Must(t, p.Allocate() == 3)
}

func TestRelease(t *testing.T) {
	p := New(1, 5)
	util.Must(t, p.Allocate() == 1)
	util.Must(t, p.Allocate() == 2)
	p.Release(2)
	util.Must(t, p.Allocate() == 2)
	util.Must(t, p.Allocate() == 3)
}

func TestClear(t *testing.T) {
	p := New(1, 5)
	util.Must(t, p.Allocate() == 1)
	p.Clear()
	util.Must(t, p.Allocate() == 1)
}

func TestLen(t *testing.T) {
	p := New(1, 8)
	util.Must(t, p.Allocate() == 1)
	util.Must(t, p.Len() == 1)
	util.Must(t, p.Allocate() == 2)
	util.Must(t, p.Len() == 2)
	p.Reserve(2)
	util.Must(t, p.Len() == 2)
	p.Reserve(3)
	util.Must(t, p.Len() == 3)
	p.Release(3)
	util.Must(t, p.Len() == 2)
	p.Release(1)
	util.Must(t, p.Len() == 1)
}

func TestLargeHigh(t *testing.T) {
	N := math.MaxUint32
	n := 1024
	p := New(1, N)
	for i := 0; i < n; i++ {
		util.Must(t, p.Allocate() == i+1)
	}
	id := rand.Intn(n-1) + 1
	p.Release(id)
	util.Must(t, p.Allocate() == id)
	util.Must(t, p.Allocate() == n+1)
}

func BenchmarkAllocate(b *testing.B) {
	p := New(0, b.N)
	for i := 0; i < b.N; i++ {
		p.Allocate()
	}
}

func BenchmarkRelease(b *testing.B) {
	p := New(0, b.N)
	for i := 0; i < b.N; i++ {
		p.Allocate()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Release(i)
	}
}
