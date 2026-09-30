package sentinelhashmap

import (
	"math"
	"testing"
)

// Copy-order regressions (fable72 finding 1): Select, Reject and user copy
// loops put a map's entries, in its iteration order, into a fresh map that
// starts small and grows. With a home bucket that is a prefix of a hash shared
// by every table size (top-k bits, b757951) that copy is O(n^2). The check
// replays the copy and counts each insert's probe length deterministically;
// see hashmap/copy_probe_test.go. Keys 0 and 1 live outside the table.

const (
	copyProbeN       = 1 << 15
	copyProbeMaxMean = 64.0
)

func checkCopyProbes(t *testing.T, name string, total, n int) {
	t.Helper()
	mean := float64(total) / float64(n)
	t.Logf("%s: n=%d mean probe length %.2f", name, n, mean)
	if mean > copyProbeMaxMean {
		t.Fatalf("%s: copying %d entries in iteration order into a fresh map averaged %.1f probes per insert (limit %.0f): copy is quadratic",
			name, n, mean, copyProbeMaxMean)
	}
}

func TestCopyProbes_Int32Int32Sentinel(t *testing.T) {
	src := NewInt32Int32()
	for i := 2; i < copyProbeN+2; i++ {
		src.Put(int32(i), int32(i))
	}
	dst := NewInt32Int32()
	total := 0
	for k, v := range src.All() {
		dst.Put(k, v)
		mask := len(dst.keys) - 1
		idx, d := dst.home(k), 0
		for dst.keys[idx] != k {
			idx, d = (idx+1)&mask, d+1
		}
		total += d
	}
	checkCopyProbes(t, "sentinel Int32Int32", total, copyProbeN)
	if got := src.Select(func(int32, int32) bool { return true }); got.Len() != copyProbeN {
		t.Fatalf("Select(all) lost entries: len %d", got.Len())
	}
}

func TestCopyProbes_Float64Int32Sentinel(t *testing.T) {
	src := NewFloat64Int32()
	for i := 2; i < copyProbeN+2; i++ {
		src.Put(float64(i)/4, int32(i))
	}
	dst := NewFloat64Int32()
	total, n := 0, 0
	for k, v := range src.All() {
		dst.Put(k, v)
		if k == 0 || k == 1 { // stored outside the table
			continue
		}
		n++
		mask := len(dst.keys) - 1
		idx, d := dst.home(k), 0
		for math.Float64bits(dst.keys[idx]) != math.Float64bits(k) {
			idx, d = (idx+1)&mask, d+1
		}
		total += d
	}
	checkCopyProbes(t, "sentinel Float64Int32", total, n)
}
