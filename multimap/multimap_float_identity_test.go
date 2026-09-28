package multimap

import (
	"math"
	"testing"
)

// The generic Multimap identifies float keys by bit pattern (spec
// algorithms.md "NaN must hash and compare by bit pattern"); other key types
// keep ==.

type celsiusKey float64
type singleKey float32

func genericEdges64() []float64 {
	return []float64{
		math.Float64frombits(0x7ff8000000000000),
		math.Float64frombits(0x7ff8000000000001),
		math.Float64frombits(0xfff8000000000000),
		math.Float64frombits(0xfff8000000000001),
		0, math.Copysign(0, -1), math.Inf(1), math.Inf(-1), 1.5,
	}
}

func genericEdges32() []float32 {
	return []float32{
		math.Float32frombits(0x7fc00000),
		math.Float32frombits(0x7fc00001),
		math.Float32frombits(0xffc00000),
		math.Float32frombits(0xffc00001),
		0, math.Float32frombits(0x80000000),
		float32(math.Inf(1)), float32(math.Inf(-1)), 1.5,
	}
}

func convertGenericKeys[F, K ~float32 | ~float64](in []F) []K {
	out := make([]K, len(in))
	for i, v := range in {
		out[i] = K(v)
	}
	return out
}

// checkMultimapIdentity takes keys that are pairwise distinct under the
// expected identity. It runs on both a constructed and a zero-value Multimap.
func checkMultimapIdentity[K comparable](t *testing.T, keys []K, fresh func() *Multimap[K, int]) {
	t.Helper()
	m := fresh()
	for i, k := range keys {
		m.Put(k, i)
		m.PutAll(k, 100+i, 200+i)
	}
	if m.SizeDistinct() != len(keys) || m.Len() != 3*len(keys) {
		t.Fatalf("SizeDistinct=%d Len=%d, want %d/%d", m.SizeDistinct(), m.Len(), len(keys), 3*len(keys))
	}
	for i, k := range keys {
		got := m.Get(k)
		if !m.ContainsKey(k) || len(got) != 3 || got[0] != i || got[2] != 200+i {
			t.Errorf("Get(%v) = %v, want [%d %d %d]", k, got, i, 100+i, 200+i)
		}
	}
	seen := 0
	m.ForEachKey(func(k K, vs []int) {
		seen++
		if want := m.Get(k); len(vs) != 3 || vs[0] != want[0] {
			t.Errorf("ForEachKey(%v) = %v, Get = %v", k, vs, want)
		}
	})
	if seen != len(keys) {
		t.Errorf("ForEachKey visited %d keys, want %d", seen, len(keys))
	}
	pairs := 0
	for range m.All() {
		pairs++
	}
	if pairs != 3*len(keys) {
		t.Errorf("All yielded %d pairs, want %d", pairs, 3*len(keys))
	}
	for i, k := range keys {
		if removed := m.RemoveAll(k); len(removed) != 3 || removed[0] != i {
			t.Fatalf("RemoveAll(%v) = %v", k, removed)
		}
		if m.ContainsKey(k) {
			t.Fatalf("ContainsKey(%v) after RemoveAll", k)
		}
		for _, rest := range keys[i+1:] {
			if !m.ContainsKey(rest) {
				t.Fatalf("RemoveAll(%v) lost %v", k, rest)
			}
		}
	}
	if m.Len() != 0 || m.SizeDistinct() != 0 {
		t.Fatalf("after RemoveAll: Len=%d SizeDistinct=%d", m.Len(), m.SizeDistinct())
	}
	m.Put(keys[0], 1)
	m.Clear()
	if m.Len() != 0 || m.SizeDistinct() != 0 || m.ContainsKey(keys[0]) {
		t.Fatal("Clear left entries")
	}
	m.Put(keys[0], 7)
	if got := m.Get(keys[0]); len(got) != 1 || got[0] != 7 {
		t.Fatalf("Put after Clear: Get = %v", got)
	}
}

func runMultimapIdentity[K comparable](t *testing.T, keys []K) {
	t.Run("constructed", func(t *testing.T) { checkMultimapIdentity(t, keys, NewMultimap[K, int]) })
	t.Run("zero value", func(t *testing.T) {
		checkMultimapIdentity(t, keys, func() *Multimap[K, int] { return &Multimap[K, int]{} })
	})
}

func TestMultimapFloatKeyIdentity(t *testing.T) {
	t.Run("float64", func(t *testing.T) { runMultimapIdentity(t, genericEdges64()) })
	t.Run("float32", func(t *testing.T) { runMultimapIdentity(t, genericEdges32()) })
	t.Run("named float64", func(t *testing.T) {
		runMultimapIdentity(t, convertGenericKeys[float64, celsiusKey](genericEdges64()))
	})
	t.Run("named float32", func(t *testing.T) {
		runMultimapIdentity(t, convertGenericKeys[float32, singleKey](genericEdges32()))
	})
	t.Run("int unchanged", func(t *testing.T) { runMultimapIdentity(t, []int{0, -1, 1, math.MaxInt}) })
	t.Run("string unchanged", func(t *testing.T) { runMultimapIdentity(t, []string{"", "NaN", "-0"}) })
}

// Iteration yields the original float key (not its bits), so a NaN key read
// back from Keys() finds its values.
func TestMultimapFloatKeyRoundTrip(t *testing.T) {
	m := NewMultimap[float64, string]()
	nan := math.Float64frombits(0x7ff8000000000001)
	m.Put(nan, "a")
	for k := range m.Keys() {
		if math.Float64bits(k) != 0x7ff8000000000001 {
			t.Fatalf("Keys yielded bits %#x", math.Float64bits(k))
		}
		if got := m.Get(k); len(got) != 1 || got[0] != "a" {
			t.Fatalf("Get(Keys()) = %v", got)
		}
	}
}
