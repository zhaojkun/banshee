// Copyright 2016 Eleme Inc. All rights reserved.

package metricdb

import (
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/util"
	"reflect"
	"testing"
	"time"
)

func TestMemStorageNew(t *testing.T) {
	mp := newMemStoragePool(nil)
	util.Must(t, len(mp.pool) == 0)
	util.Must(t, mp.initOK == 0)
	util.Must(t, mp.initErr == 0)
}

func TestMemStorageInit(t *testing.T) {
}

func TestMemStoragePut(t *testing.T) {
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	mp := newMemStoragePool(opts)
	m := &models.Metric{
		Name:    "foo",
		Link:    1,
		Stamp:   1452758773,
		Value:   3.14,
		Score:   0.1892,
		Average: 3.133,
	}
	util.Must(t, nil == mp.put(m))
	// Must in pool.
	util.Must(t, mp.has(m.Link))
	util.Must(t, len(mp.pool) == 1)
	util.Must(t, len(mp.pool[0].get(m.Link, 0, m.Stamp+1)) == 1)
	// Must in the skiplist.
	sl := mp.pool[0].htree.Get(&node{link: m.Link}).(*node).sl
	util.Must(t, sl.Has(&metricWrapper{m}))
	// Must the same value
	m1 := sl.Get(&metricWrapper{m}).(*metricWrapper).m
	util.Must(t, reflect.DeepEqual(m, m1))
}

func TestMemStorageGet(t *testing.T) {
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	mp := newMemStoragePool(opts)
	// Nothing.
	util.Must(t, 0 == len(mp.get(23333, 1234, 12345)))
	// Put some
	mp.put(&models.Metric{Name: "foo", Link: 1, Stamp: 1452758723})
	mp.put(&models.Metric{Name: "foo", Link: 1, Stamp: 1452758733, Value: 1.89, Score: 1.12, Average: 1.72})
	mp.put(&models.Metric{Name: "foo", Link: 1, Stamp: 1452758743})
	mp.put(&models.Metric{Name: "foo", Link: 1, Stamp: 1452758753})
	// Get again.
	ms := mp.get(1, 1452758733, 1452758753)
	util.Must(t, len(ms) == 2)
	// Test the value.
	m := ms[0]
	util.Must(t, m.Value == 1.89 && m.Score == 1.12 && m.Link == 1)
}

func TestMemStorageGetAcrossStorages(t *testing.T) {
	opts := &Options{Period: 86400, Expiration: 86400 * 4}
	mp := newMemStoragePool(opts)
	// Force creating 4+1 storages.
	base := uint32(time.Now().Unix())
	mp.put(&models.Metric{Link: 1, Stamp: base})                    // 0
	mp.put(&models.Metric{Link: 1, Stamp: base + mp.opts.Period*1}) // 1
	mp.put(&models.Metric{Link: 1, Stamp: base + mp.opts.Period*2}) // 2
	mp.put(&models.Metric{Link: 1, Stamp: base + mp.opts.Period*3}) // 3
	mp.put(&models.Metric{Link: 1, Stamp: base + mp.opts.Period*4}) // 4
	// Get
	ms := mp.get(1, base, base+mp.opts.Period*3)
	util.Must(t, len(ms) == 3)
	util.Must(t, ms[0].Stamp == base)
	util.Must(t, ms[1].Stamp == base+mp.opts.Period*1)
	util.Must(t, ms[2].Stamp == base+mp.opts.Period*2)
}

func TestMemStorageExpire(t *testing.T) {
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	mp := newMemStoragePool(opts)
	// Force creating 7+1 storages.
	base := uint32(time.Now().Unix())
	mp.put(&models.Metric{Link: 1, Stamp: base})                    // 0
	htree := mp.pool[0].htree                                       // record the tree to be deleted
	mp.put(&models.Metric{Link: 1, Stamp: base + mp.opts.Period*1}) // 1
	mp.put(&models.Metric{Link: 1, Stamp: base + mp.opts.Period*2}) // 2
	mp.put(&models.Metric{Link: 1, Stamp: base + mp.opts.Period*3}) // 3
	mp.put(&models.Metric{Link: 1, Stamp: base + mp.opts.Period*4}) // 4
	mp.put(&models.Metric{Link: 1, Stamp: base + mp.opts.Period*5}) // 5
	mp.put(&models.Metric{Link: 1, Stamp: base + mp.opts.Period*6}) // 6
	util.Must(t, len(mp.pool) == 7)
	mp.put(&models.Metric{Link: 1, Stamp: base + mp.opts.Period*7}) // 7
	util.Must(t, len(mp.pool) == 8)
	mp.put(&models.Metric{Link: 1, Stamp: base + mp.opts.Period*8}) // 8
	util.Must(t, len(mp.pool) == 8)                                 // Full storages: 1,2,3,4,5,6,7
	util.Must(t, htree.Len() == 0)
}

func BenchmarkMemStoragePut(b *testing.B) {
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	mp := newMemStoragePool(opts)
	base := uint32(time.Now().Unix())
	// Bench
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mp.put(&models.Metric{Link: 1, Stamp: base + uint32(i)*7, Value: float64(i)})
	}
}

func BenchmarkMemStoragePutX10(b *testing.B) {
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	mp := newMemStoragePool(opts)
	base := uint32(time.Now().Unix())
	// Bench
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 10; j++ {
			mp.put(&models.Metric{Link: uint32(j), Stamp: base + uint32(i)*7, Value: float64(i)})
		}
	}
}

func BenchmarkMemStorageGet100K(b *testing.B) {
	opts := &Options{Period: 86400, Expiration: 86400 * 7}
	mp := newMemStoragePool(opts)
	// Put suite
	base := uint32(time.Now().Unix())
	for i := 0; i < 1024*100; i++ {
		for j := 0; j < 10; j++ {
			mp.put(&models.Metric{Link: uint32(j), Stamp: base + uint32(i)*7, Value: float64(i)})
		}
	}
	// Benchmark.
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mp.get(uint32(i%10), base, base+100*10)
	}
}
