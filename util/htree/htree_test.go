// Copyright 2016 Eleme Inc. All rights reserved.
// Source from github.com/hit9/htree.

package htree

import (
	"github.com/eleme/banshee/util"
	"math/rand"
	"testing"
)

func TestPrimesLargerThanUint32(t *testing.T) {
	s := uint64(1)
	for i := 0; i < len(primes); i++ {
		s *= uint64(primes[i])
	}
	util.Must(t, s > uint64(^uint32(0)))
}

func TestTreeInside(t *testing.T) {
	/*
	       root
	     /     \
	    0       1     %2
	   /|\     /|\
	  6 4 2   3 7 5   %3
	      |   |
	      8   9       %5
	*/
	tree := New()
	for i := 0; i < 10; i++ {
		tree.Put(Uint32(i))
	}
	n1_0 := tree.root.children[0]
	n1_1 := tree.root.children[1]
	n2_0_0 := n1_0.children[0]
	n2_0_1 := n1_0.children[1]
	n2_0_2 := n1_0.children[2]
	n2_1_0 := n1_1.children[0]
	n2_1_1 := n1_1.children[1]
	n2_1_2 := n1_1.children[2]
	util.Must(t, n1_0.item == Uint32(0))
	util.Must(t, n1_1.item == Uint32(1))
	util.Must(t, n2_0_0.item == Uint32(6))
	util.Must(t, n2_0_1.item == Uint32(4))
	util.Must(t, n2_0_2.item == Uint32(2))
	util.Must(t, n2_1_0.item == Uint32(3))
	util.Must(t, n2_1_1.item == Uint32(7))
	util.Must(t, n2_1_2.item == Uint32(5))
	util.Must(t, len(n2_0_2.children) == 1)
	util.Must(t, n2_0_2.children[0].item == Uint32(8))
	util.Must(t, len(n2_1_0.children) == 1)
	util.Must(t, n2_1_0.children[0].item == Uint32(9))
	util.Must(t, n1_0.remainder == 0)
	util.Must(t, n1_1.remainder == 1)
	util.Must(t, n2_0_0.remainder == 0)
	util.Must(t, n2_0_1.remainder == 1)
	util.Must(t, n2_0_2.remainder == 2)
	util.Must(t, n2_1_0.remainder == 0)
	util.Must(t, n2_1_1.remainder == 1)
	util.Must(t, n2_1_2.remainder == 2)
}

func TestPutN(t *testing.T) {
	tree := New()
	n := 1024
	for i := 0; i < n; i++ {
		item := Uint32(rand.Uint32())
		// util.Must put
		util.Must(t, tree.Put(item) != nil)
		// util.Must get
		util.Must(t, tree.Get(item) == item)
		// util.Must len++
		util.Must(t, tree.Len()+tree.Conflicts() == i+1)
	}
}

func TestPutReuse(t *testing.T) {
	/*
	       root
	     /     \
	    0       1     %2
	   /|\     /|\
	  6 4 2   3 7 5   %3
	      |   |
	      8   9       %5
	*/
	tree := New()
	for i := 0; i < 10; i++ {
		tree.Put(Uint32(i))
	}
	util.Must(t, tree.Len() == 10)
	item := Uint32(9)
	util.Must(t, tree.Put(item) == item)
	util.Must(t, tree.Conflicts() == 1)
	util.Must(t, tree.Len() == 10)
}

func TestPutNewNode(t *testing.T) {
	/*
	       root
	     /     \
	    0       1     %2
	   /|\     /|\
	  6 4 2   3 7 5   %3
	      |   |
	      8   9       %5
	*/
	tree := New()
	for i := 0; i < 10; i++ {
		tree.Put(Uint32(i))
	}
	util.Must(t, tree.Len() == 10)
	item := Uint32(10)
	util.Must(t, tree.Put(item) == item)
	util.Must(t, tree.Conflicts() == 0)
	util.Must(t, tree.Len() == 11)
}

func TestGetN(t *testing.T) {
	tree := New()
	n := 1024
	for i := 0; i < n; i++ {
		item := Uint32(rand.Uint32())
		tree.Put(item)
		// util.Must get
		util.Must(t, tree.Get(item) == item)
		// util.Must cant get
		util.Must(t, tree.Get(Uint32(n+i)) == nil)
	}
}

func TestDeleteN(t *testing.T) {
	tree := New()
	n := 1024
	for i := 0; i < n; i++ {
		item := Uint32(rand.Uint32())
		tree.Put(item)
		// util.Must delete
		util.Must(t, tree.Delete(item) == item)
		// util.Must cant delete
		util.Must(t, tree.Delete(Uint32(n+i)) == nil)
		// util.Must len--
		util.Must(t, tree.Len() == 0)
	}
}

func TestDeleteReplace(t *testing.T) {
	/*
	       root
	     /     \
	    0       1     %2
	   /|\     /|\
	  6 4 2   3 7 5   %3
	  |    |
	  42   8          %5
	*/
	tree := New()
	for i := 0; i < 9; i++ {
		tree.Put(Uint32(i))
	}
	tree.Put(Uint32(42))
	util.Must(t, tree.Len() == 10)
	item := Uint32(0)
	// util.Must delete
	util.Must(t, tree.Delete(item) == item)
	// Original child must be replaced by new node:42
	util.Must(t, tree.root.children[0].item == Uint32(42))
	util.Must(t, tree.root.children[0].remainder == 0)
	util.Must(t, tree.root.children[0].depth == 1)
	// The children shouldnt be changed
	util.Must(t, len(tree.root.children[0].children) == 3)
	// Node must be a leaf now.
	leaf := tree.root.children[0].children[0]
	util.Must(t, len(leaf.children) == 0)
	// util.Must length--
	util.Must(t, tree.Len() == 9)
}

func TestDeleteLeaf(t *testing.T) {
	/*
	       root
	     /     \
	    0       1     %2
	   /|\     /|\
	  6 4 2   3 7 5   %3
	*/
	tree := New()
	for i := 0; i < 8; i++ {
		tree.Put(Uint32(i))
	}
	util.Must(t, tree.Len() == 8)
	item := Uint32(7)
	// util.Must delete
	util.Must(t, tree.Delete(item) == item)
	// util.Must node(1) has 2 nodes now
	util.Must(t, len(tree.root.children[1].children) == 2)
	// util.Must length--
	util.Must(t, tree.Len() == 7)
}

func TestIteratorEmpty(t *testing.T) {
	tree := New()
	i := 0
	iter := tree.NewIterator()
	for iter.Next() {
		i++
	}
	// util.Must iterates 0 times
	util.Must(t, i == 0)
}

func TestIteratorOrder(t *testing.T) {
	/*
	      root
	     /    \
	    0      1     %2
	   / \    / \
	  4   2  3   5   %3
	*/
	tree := New()
	tree.Put(Uint32(0))
	tree.Put(Uint32(1))
	tree.Put(Uint32(2))
	tree.Put(Uint32(3))
	tree.Put(Uint32(4))
	tree.Put(Uint32(5))
	iter := tree.NewIterator()
	util.Must(t, iter.Next() && iter.Item() == Uint32(0))
	util.Must(t, iter.Next() && iter.Item() == Uint32(4))
	util.Must(t, iter.Next() && iter.Item() == Uint32(2))
	util.Must(t, iter.Next() && iter.Item() == Uint32(1))
	util.Must(t, iter.Next() && iter.Item() == Uint32(3))
	util.Must(t, iter.Next() && iter.Item() == Uint32(5))
}

func TestIteratorLarge(t *testing.T) {
	tree := New()
	n := 1024 * 10
	for i := 0; i < n; i++ {
		item := Uint32(rand.Uint32())
		tree.Put(item)
	}
	j := 0
	iter := tree.NewIterator()
	for iter.Next() {
		j++
	}
	util.Must(t, j == tree.Len())
}

func BenchmarkPut(b *testing.B) {
	t := New()
	for i := 0; i < b.N; i++ {
		t.Put(Uint32(i))
	}
}

func BenchmarkGet(b *testing.B) {
	t := New()
	for i := 0; i < b.N; i++ {
		t.Put(Uint32(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Get(Uint32(i))
	}
}

func BenchmarkIteratorNext(b *testing.B) {
	t := New()
	for i := 0; i < b.N; i++ {
		t.Put(Uint32(i))
	}
	iter := t.NewIterator()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Next()
	}
}

func BenchmarkGetLargeTree(b *testing.B) {
	t := New()
	n := 1000 * 1000 // Million
	for i := 0; i < n; i++ {
		t.Put(Uint32(rand.Uint32()))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Get(Uint32(i))
	}
}

func BenchmarkPutLargeTree(b *testing.B) {
	t := New()
	n := 1000 * 1000 // Million
	for i := 0; i < n; i++ {
		t.Put(Uint32(rand.Uint32()))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Put(Uint32(i))
	}
}

func BenchmarkDeleteLargeTree(b *testing.B) {
	t := New()
	n := 1000 * 1000 // Million
	for i := 0; i < n; i++ {
		t.Put(Uint32(rand.Uint32()))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Delete(Uint32(i))
	}
}
