// Copyright 2016 Eleme Inc. All rights reserved.

package idpool

import (
	"github.com/eleme/banshee/util"
	"math"
	"math/rand"
	"testing"
)

func TestReserve(t *testing.T) {
	p := New(1, 5)
	util.Must(t, p.Reserve() == 1)
	util.Must(t, p.Reserve() == 2)
	util.Must(t, p.Reserve() == 3)
	util.Must(t, p.Reserve() == 4)
	util.Must(t, p.Reserve() == 5)
	util.Must(t, p.Reserve() == 5)
}

func TestRelease(t *testing.T) {
	p := New(1, 5)
	util.Must(t, p.Reserve() == 1)
	util.Must(t, p.Reserve() == 2)
	p.Release(2)
	util.Must(t, p.Reserve() == 2)
	util.Must(t, p.Reserve() == 3)
}

func TestLargeHigh(t *testing.T) {
	N := math.MaxUint32
	n := 1024
	p := New(1, N)
	for i := 0; i < n; i++ {
		util.Must(t, p.Reserve() == i+1)
	}
	id := rand.Intn(n-1) + 1
	p.Release(id)
	util.Must(t, p.Reserve() == id)
	util.Must(t, p.Reserve() == n+1)
}

func BenchmarkReserve(b *testing.B) {
	p := New(0, b.N)
	for i := 0; i < b.N; i++ {
		p.Reserve()
	}
}

func BenchmarkRelease(b *testing.B) {
	p := New(0, b.N)
	for i := 0; i < b.N; i++ {
		p.Reserve()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Release(i)
	}
}
