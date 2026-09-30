package hashmap

import (
	"math"
	"testing"
)

// Copy-order regressions (fable72 finding 1). Select, Reject and user copy
// loops insert a table's entries, in the table's own slot order, into a fresh
// table that starts small and grows. When the home bucket is a prefix of a
// hash that every table size shares (top-k bits of the product, mapdb-golang
// b757951), slot order is hash order, the destination's low buckets fill first
// and every insert probes to the end of one growing cluster: O(n^2) in total.
//
// The check is deterministic: it replays that copy into a fresh table and
// counts, after every insert, how far the key sits from its home bucket
// (the insert's probe length). With the bit-reversed index (home i in the
// source is home i mod c in a c-slot destination) this copy averages about
// 5-15 probes at this size, growing slowly with n; the top-k-bit index
// averages about 1,300 here and doubles with every doubling of n. The bound
// sits between the two with wide margins on both sides; it counts slots, not
// time, so it cannot flake on a loaded host.

const (
	copyProbeN       = 1 << 15
	copyProbeMaxMean = 64.0
)

func checkCopyProbes(t *testing.T, name string, total, n int) {
	t.Helper()
	mean := float64(total) / float64(n)
	t.Logf("%s: n=%d mean probe length %.2f", name, n, mean)
	if mean > copyProbeMaxMean {
		t.Fatalf("%s: copying %d entries in slot order into a fresh table averaged %.1f probes per insert (limit %.0f): copy is quadratic",
			name, n, mean, copyProbeMaxMean)
	}
}

func TestCopyProbes_Int32Int32(t *testing.T) {
	src := NewInt32Int32()
	for i := 0; i < copyProbeN; i++ {
		src.Put(int32(i), int32(i))
	}
	dst := NewInt32Int32()
	total := 0
	for k, v := range src.All() {
		dst.Put(k, v)
		mask := len(dst.entries) - 1
		idx, d := dst.home(k), 0
		for !dst.entries[idx].occupied || dst.entries[idx].key != k {
			idx, d = (idx+1)&mask, d+1
		}
		total += d
	}
	checkCopyProbes(t, "Int32Int32", total, copyProbeN)
	if got := src.Select(func(int32, int32) bool { return true }); got.Len() != copyProbeN || !got.Equals(src) {
		t.Fatalf("Select(all) lost entries: len %d", got.Len())
	}
}

func TestCopyProbes_Float64Int32(t *testing.T) {
	src := NewFloat64Int32()
	for i := 0; i < copyProbeN; i++ {
		src.Put(float64(i)/4, int32(i))
	}
	dst := NewFloat64Int32()
	total := 0
	for k, v := range src.All() {
		dst.Put(k, v)
		mask := len(dst.entries) - 1
		idx, d := dst.home(k), 0
		for !dst.entries[idx].occupied || math.Float64bits(dst.entries[idx].key) != math.Float64bits(k) {
			idx, d = (idx+1)&mask, d+1
		}
		total += d
	}
	checkCopyProbes(t, "Float64Int32", total, copyProbeN)
}

func TestCopyProbes_Int64Object(t *testing.T) {
	src := NewInt64Object[int]()
	for i := 0; i < copyProbeN; i++ {
		src.Put(int64(i)<<32, i)
	}
	dst := NewInt64Object[int]()
	total := 0
	for k, v := range src.All() {
		dst.Put(k, v)
		mask := len(dst.keys) - 1
		idx, d := dst.home(k), 0
		for !dst.occupied[idx] || dst.keys[idx] != k {
			idx, d = (idx+1)&mask, d+1
		}
		total += d
	}
	checkCopyProbes(t, "Int64Object", total, copyProbeN)
	if got := src.Reject(func(int64, int) bool { return false }); got.Len() != copyProbeN {
		t.Fatalf("Reject(none) lost entries: len %d", got.Len())
	}
}
