// Copyright 2016 Eleme Inc. All rights reserved.
// Source from github.com/hit9/skiplist.

package skiplist

import (
	"github.com/eleme/banshee/util"
	"math/rand"
	"testing"
)

func TestPut(t *testing.T) {
	sl := New(16)
	n := 1024 * 10
	for i := 0; i < n; i++ {
		item := Int(rand.Int())
		sl.Put(item)
		// util.Must get
		util.Must(t, equal(sl.Get(item), item))
		// util.Must len++
		util.Must(t, sl.Len() == i+1)
	}
}

func TestGet(t *testing.T) {
	sl := New(16)
	n := 1024 * 10
	for i := 0; i < n; i++ {
		item := Int(rand.Int() % n)
		sl.Put(item)
		// util.Must get
		util.Must(t, equal(sl.Get(item), item))
		// util.Must cant get
		util.Must(t, sl.Get(Int(n+rand.Int())) == nil)
	}
}

func TestDelete(t *testing.T) {
	sl := New(16)
	n := 1024 * 10
	for i := 0; i < n; i++ {
		item := Int(rand.Int() % n)
		sl.Put(item)
		util.Must(t, sl.Len() == 1)
		// util.Must delete
		util.Must(t, sl.Delete(item) == item)
		// util.Must cant delete
		util.Must(t, sl.Delete(Int(n+rand.Int())) == nil)
		util.Must(t, sl.Len() == 0)
	}
}

func TestIteratorNil(t *testing.T) {
	sl := New(7)
	n := 1024
	for i := n - 1; i >= 0; i-- {
		sl.Put(Int(i))
	}
	iter := sl.NewIterator(nil)
	i := 0
	for iter.Next() {
		// util.Must equal
		util.Must(t, Int(i) == iter.Item())
		i++
	}
}

func TestIteratorStart(t *testing.T) {
	sl := New(7)
	n := 1024
	for i := n - 1; i >= 0; i-- {
		sl.Put(Int(i))
	}
	start := rand.Intn(n)
	iter := sl.NewIterator(Int(start))
	i := 0
	for iter.Next() {
		// util.Must equal
		util.Must(t, Int(i+start) == iter.Item())
		i++
	}
	util.Must(t, i == n-start)
}

// The maxLevel masters the bench results.
func BenchmarkPut(b *testing.B) {
	sl := New(50)
	for i := 0; i < b.N; i++ {
		sl.Put(Int(i))
	}
}

func BenchmarkGet(b *testing.B) {
	sl := New(50)
	for i := 0; i < b.N; i++ {
		sl.Put(Int(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sl.Get(Int(i))
	}
}
