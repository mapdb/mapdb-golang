package hashset

import (
	"math"
	"testing"
)

// Copy-order regressions (fable72 finding 1): Intersect, Difference, Select
// and user copy loops add a set's elements, in its slot order, to a fresh set
// that starts small and grows. With a home bucket that is a prefix of a hash
// shared by every table size (top-k bits, b757951) that copy is O(n^2). The
// check replays the copy and counts each insert's probe length
// deterministically; see hashmap/copy_probe_test.go.

const (
	copyProbeN       = 1 << 15
	copyProbeMaxMean = 64.0
)

func checkCopyProbes(t *testing.T, name string, total, n int) {
	t.Helper()
	mean := float64(total) / float64(n)
	t.Logf("%s: n=%d mean probe length %.2f", name, n, mean)
	if mean > copyProbeMaxMean {
		t.Fatalf("%s: copying %d elements in slot order into a fresh set averaged %.1f probes per insert (limit %.0f): copy is quadratic",
			name, n, mean, copyProbeMaxMean)
	}
}

func TestCopyProbes_Int32Set(t *testing.T) {
	src := NewInt32()
	for i := 0; i < copyProbeN; i++ {
		src.Add(int32(i))
	}
	dst := NewInt32()
	total := 0
	for k := range src.All() {
		dst.Add(k)
		mask := len(dst.entries) - 1
		idx, d := dst.home(k), 0
		for !dst.entries[idx].occupied || dst.entries[idx].key != k {
			idx, d = (idx+1)&mask, d+1
		}
		total += d
	}
	checkCopyProbes(t, "Int32 set", total, copyProbeN)
	if got := src.Intersect(src); !got.Equals(src) {
		t.Fatalf("Intersect(self) lost elements: len %d", got.Len())
	}
	if got := src.Difference(NewInt32()); !got.Equals(src) {
		t.Fatalf("Difference(empty) lost elements: len %d", got.Len())
	}
}

func TestCopyProbes_Float64Set(t *testing.T) {
	src := NewFloat64()
	for i := 0; i < copyProbeN; i++ {
		src.Add(float64(i) / 4)
	}
	dst := NewFloat64()
	total := 0
	for k := range src.All() {
		dst.Add(k)
		mask := len(dst.entries) - 1
		idx, d := dst.home(k), 0
		for !dst.entries[idx].occupied || math.Float64bits(dst.entries[idx].key) != math.Float64bits(k) {
			idx, d = (idx+1)&mask, d+1
		}
		total += d
	}
	checkCopyProbes(t, "Float64 set", total, copyProbeN)
}
