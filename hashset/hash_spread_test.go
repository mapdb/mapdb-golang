package hashset

import (
	"math"
	"testing"
)

// Spread regressions for the primitive-key hashset tables (spec algorithms.md
// §"Hash function": 64-bit Fibonacci multiply; its entropy is in the product's
// HIGH bits, so a low-bit mask must see them). Each case inserts a key family
// into a real table and counts the distinct HOME buckets the table itself
// computes (its own hash >> its real index shift), requiring at least 40% of
// min(n, capacity) distinct buckets. Before the fix, float keys 1..1000, i/2
// and powers of two all landed in a single bucket.

const spreadMinFraction = 0.40

func float64Families() map[string][]float64 {
	seq := make([]float64, 0, 1000)
	half := make([]float64, 0, 1000)
	for i := 1; i <= 1000; i++ {
		seq = append(seq, float64(i))
		half = append(half, float64(i)/2)
	}
	pow2 := make([]float64, 0, 1000)
	for j := -500; j < 500; j++ {
		pow2 = append(pow2, math.Ldexp(1, j))
	}
	return map[string][]float64{"1..1000": seq, "i/2": half, "pow2": pow2}
}

func float32Families() map[string][]float32 {
	seq := make([]float32, 0, 1000)
	half := make([]float32, 0, 1000)
	for i := 1; i <= 1000; i++ {
		seq = append(seq, float32(i))
		half = append(half, float32(i)/2)
	}
	pow2 := make([]float32, 0, 277)
	for j := -149; j <= 127; j++ { // every float32 power of two, subnormals included
		pow2 = append(pow2, float32(math.Ldexp(1, j)))
	}
	return map[string][]float32{"1..1000": seq, "i/2": half, "pow2": pow2}
}

func int64Families() map[string][]int64 {
	seq := make([]int64, 0, 1000)
	high := make([]int64, 0, 1000)
	high1 := make([]int64, 0, 1000)
	shl40 := make([]int64, 0, 1000)
	for i := int64(1); i <= 1000; i++ {
		seq = append(seq, i)
		high = append(high, i<<32)     // high-word-only keys (spec-pinned family)
		high1 = append(high1, i<<32|1) // {1, 2^32+1, ...}
		shl40 = append(shl40, i<<40)   // >= 43 trailing zero bits, like float64 bit patterns
	}
	pow2 := make([]int64, 0, 63)
	for j := 0; j < 63; j++ {
		pow2 = append(pow2, int64(1)<<j)
	}
	return map[string][]int64{"1..1000": seq, "high-word": high, "high-word+1": high1, "i<<40": shl40, "pow2": pow2}
}

// checkSpread counts distinct home buckets; capacity is the table's real slot
// count. slot(i) reports the key stored in slot i and whether it is occupied;
// every key must be reachable by linear probing from its home bucket, which
// proves home() is the index the table really uses.
func checkSpread[K comparable](t *testing.T, name string, keys []K, capacity int, home func(K) int, slot func(int) (K, bool)) {
	t.Helper()
	seen := make(map[int]struct{}, len(keys))
	for _, k := range keys {
		b := home(k)
		if b < 0 || b >= capacity {
			t.Fatalf("%s: home bucket %d outside [0,%d)", name, b, capacity)
		}
		seen[b] = struct{}{}
		found := false
		for i, n := b, 0; n < capacity; i, n = (i+1)&(capacity-1), n+1 {
			sk, occ := slot(i)
			if !occ {
				break
			}
			if sk == k {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%s: key %v not reachable by probing from home bucket %d", name, k, b)
		}
	}
	limit := min(len(keys), capacity)
	t.Logf("%s: n=%d cap=%d distinct home buckets=%d (%.1f%% of %d)",
		name, len(keys), capacity, len(seen), 100*float64(len(seen))/float64(limit), limit)
	if float64(len(seen)) < spreadMinFraction*float64(limit) {
		t.Errorf("%s: only %d distinct home buckets for %d keys in %d slots (want >= %.0f%% of %d)",
			name, len(seen), len(keys), capacity, 100*spreadMinFraction, limit)
	}
}

func TestHashSpread_Float64Set(t *testing.T) {
	for fam, keys := range float64Families() {
		s := NewFloat64()
		for _, k := range keys {
			s.Add(k)
		}
		checkSpread(t, "Float64 set "+fam, keys, len(s.entries), func(k float64) int { return int(s.hash(k) >> s.shift) },
			func(i int) (float64, bool) { return s.entries[i].key, s.entries[i].occupied })
	}
}

func TestHashSpread_Float32Set(t *testing.T) {
	for fam, keys := range float32Families() {
		s := NewFloat32()
		for _, k := range keys {
			s.Add(k)
		}
		checkSpread(t, "Float32 set "+fam, keys, len(s.entries), func(k float32) int { return int(s.hash(k) >> s.shift) },
			func(i int) (float32, bool) { return s.entries[i].key, s.entries[i].occupied })
	}
}

func TestHashSpread_Int64Set(t *testing.T) {
	for fam, keys := range int64Families() {
		s := NewInt64()
		for _, k := range keys {
			s.Add(k)
		}
		checkSpread(t, "Int64 set "+fam, keys, len(s.entries), func(k int64) int { return int(s.hash(k) >> s.shift) },
			func(i int) (int64, bool) { return s.entries[i].key, s.entries[i].occupied })
	}
}
