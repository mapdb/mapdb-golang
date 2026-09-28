// Copyright (c) 2026 Jan Kotek.
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package object

import (
	"math"
	"math/rand"
	"testing"
)

// Float keys in object/ collections must use bit-pattern identity and the
// IEEE 754 totalOrder (spec algorithms.md "NaN must hash and compare by bit
// pattern", "Float ordering").

// floatSpecials holds the edge values of one float width, built from bits so
// NaN payloads and signs are exact.
type floatSpecials[F float32 | float64] struct {
	nan, nanPayload, negNaN, negNaNPayload F
	posZero, negZero, posInf, negInf       F
	bits                                   func(F) uint64
}

func specials32() floatSpecials[float32] {
	return floatSpecials[float32]{
		nan:           math.Float32frombits(0x7fc00000),
		nanPayload:    math.Float32frombits(0x7fc00001),
		negNaN:        math.Float32frombits(0xffc00000),
		negNaNPayload: math.Float32frombits(0xffc00001),
		posZero:       0,
		negZero:       math.Float32frombits(0x80000000),
		posInf:        float32(math.Inf(1)),
		negInf:        float32(math.Inf(-1)),
		bits:          func(f float32) uint64 { return uint64(math.Float32bits(f)) },
	}
}

func specials64() floatSpecials[float64] {
	return floatSpecials[float64]{
		nan:           math.Float64frombits(0x7ff8000000000000),
		nanPayload:    math.Float64frombits(0x7ff8000000000001),
		negNaN:        math.Float64frombits(0xfff8000000000000),
		negNaNPayload: math.Float64frombits(0xfff8000000000001),
		posZero:       0,
		negZero:       math.Copysign(0, -1),
		posInf:        math.Inf(1),
		negInf:        math.Inf(-1),
		bits:          math.Float64bits,
	}
}

// ladder returns the totalOrder ladder in ascending order.
func (s floatSpecials[F]) ladder() []F {
	var one F = 1
	return []F{
		s.negNaNPayload, s.negNaN, s.negInf, -one, s.negZero,
		s.posZero, one, s.posInf, s.nan, s.nanPayload,
	}
}

type hashMapFloatCase[F float32 | float64] struct {
	name string
	run  func(t *testing.T, s floatSpecials[F], m *HashMap[F, int])
}

func hashMapFloatCases[F float32 | float64]() []hashMapFloatCase[F] {
	return []hashMapFloatCase[F]{
		{"NaNKey_Findable", func(t *testing.T, s floatSpecials[F], m *HashMap[F, int]) {
			m.Put(s.nan, 7)
			if v, ok := m.Get(s.nan); !ok || v != 7 {
				t.Fatalf("Get(NaN) = %d,%v; want 7,true", v, ok)
			}
			if !m.ContainsKey(s.nan) {
				t.Fatal("ContainsKey(NaN) = false")
			}
			if got := m.GetOrDefault(s.nan, -1); got != 7 {
				t.Fatalf("GetOrDefault(NaN) = %d", got)
			}
		}},
		{"NaNKey_Replaces", func(t *testing.T, s floatSpecials[F], m *HashMap[F, int]) {
			if _, existed := m.Put(s.nan, 1); existed {
				t.Fatal("first Put(NaN) reported existing")
			}
			old, existed := m.Put(s.nan, 2)
			if !existed || old != 1 {
				t.Fatalf("second Put(NaN) = %d,%v; want 1,true", old, existed)
			}
			if m.Len() != 1 {
				t.Fatalf("Len = %d; want 1", m.Len())
			}
		}},
		{"NaNKey_Remove", func(t *testing.T, s floatSpecials[F], m *HashMap[F, int]) {
			m.Put(s.nan, 1)
			m.Put(1, 10)
			if v, ok := m.Remove(s.nan); !ok || v != 1 {
				t.Fatalf("Remove(NaN) = %d,%v; want 1,true", v, ok)
			}
			if m.ContainsKey(s.nan) || m.Len() != 1 {
				t.Fatalf("after Remove(NaN): contains=%v len=%d", m.ContainsKey(s.nan), m.Len())
			}
			if _, ok := m.Remove(s.nan); ok {
				t.Fatal("second Remove(NaN) reported present")
			}
		}},
		{"NegativeZeroDistinct", func(t *testing.T, s floatSpecials[F], m *HashMap[F, int]) {
			m.Put(s.posZero, 1)
			m.Put(s.negZero, 2)
			if m.Len() != 2 {
				t.Fatalf("Len = %d; want 2", m.Len())
			}
			if v, _ := m.Get(s.posZero); v != 1 {
				t.Fatalf("Get(+0) = %d; want 1", v)
			}
			if v, _ := m.Get(s.negZero); v != 2 {
				t.Fatalf("Get(-0) = %d; want 2", v)
			}
			m.Remove(s.negZero)
			if !m.ContainsKey(s.posZero) || m.ContainsKey(s.negZero) {
				t.Fatal("Remove(-0) must leave +0 and drop -0")
			}
			for k := range m.Keys() {
				if s.bits(k) != s.bits(s.posZero) {
					t.Fatalf("remaining key has bits %x; want +0", s.bits(k))
				}
			}
		}},
		{"InfinityKeys", func(t *testing.T, s floatSpecials[F], m *HashMap[F, int]) {
			m.Put(s.posInf, 1)
			m.Put(s.negInf, 2)
			m.Put(s.posInf, 3)
			if m.Len() != 2 {
				t.Fatalf("Len = %d; want 2", m.Len())
			}
			if v, _ := m.Get(s.posInf); v != 3 {
				t.Fatalf("Get(+Inf) = %d; want 3", v)
			}
			if v, _ := m.Get(s.negInf); v != 2 {
				t.Fatalf("Get(-Inf) = %d; want 2", v)
			}
		}},
		{"NaNPayloadsDistinct", func(t *testing.T, s floatSpecials[F], m *HashMap[F, int]) {
			nans := []F{s.nan, s.nanPayload, s.negNaN, s.negNaNPayload}
			for i, k := range nans {
				m.Put(k, i)
			}
			if m.Len() != len(nans) {
				t.Fatalf("Len = %d; want %d", m.Len(), len(nans))
			}
			for i, k := range nans {
				if v, ok := m.Get(k); !ok || v != i {
					t.Fatalf("Get(nan #%d) = %d,%v", i, v, ok)
				}
			}
			m.Remove(s.nanPayload)
			if m.ContainsKey(s.nanPayload) || !m.ContainsKey(s.nan) {
				t.Fatal("removing one NaN payload affected another")
			}
		}},
		{"SelectKeepsIdentity", func(t *testing.T, s floatSpecials[F], m *HashMap[F, int]) {
			m.Put(s.nan, 1)
			m.Put(s.posZero, 2)
			m.Put(s.negZero, 3)
			sel := m.Select(func(F, int) bool { return true })
			if sel.Len() != 3 || !sel.ContainsKey(s.nan) || !sel.ContainsKey(s.negZero) {
				t.Fatalf("Select lost float identity: %v", sel)
			}
		}},
	}
}

func runHashMapFloat[F float32 | float64](t *testing.T, s floatSpecials[F]) {
	ctors := []struct {
		name string
		mk   func() *HashMap[F, int]
	}{
		{"New", NewHashMap[F, int]},
		{"WithCapacity", func() *HashMap[F, int] { return NewHashMapWithCapacity[F, int](64) }},
		{"ZeroValue", func() *HashMap[F, int] { return &HashMap[F, int]{} }},
	}
	for _, c := range hashMapFloatCases[F]() {
		for _, ctor := range ctors {
			t.Run(c.name+"/"+ctor.name, func(t *testing.T) { c.run(t, s, ctor.mk()) })
		}
	}
}

func TestHashMapFloat32KeyBitIdentity(t *testing.T) { runHashMapFloat(t, specials32()) }
func TestHashMapFloat64KeyBitIdentity(t *testing.T) { runHashMapFloat(t, specials64()) }

type hashSetFloatCase[F float32 | float64] struct {
	name string
	run  func(t *testing.T, s floatSpecials[F], set *HashSet[F])
}

func hashSetFloatCases[F float32 | float64]() []hashSetFloatCase[F] {
	return []hashSetFloatCase[F]{
		{"NaNKey_Findable", func(t *testing.T, s floatSpecials[F], set *HashSet[F]) {
			set.Add(s.nan)
			if !set.Contains(s.nan) {
				t.Fatal("Contains(NaN) = false")
			}
		}},
		{"NaNKey_Replaces", func(t *testing.T, s floatSpecials[F], set *HashSet[F]) {
			if !set.Add(s.nan) {
				t.Fatal("first Add(NaN) = false")
			}
			if set.Add(s.nan) {
				t.Fatal("second Add(NaN) = true; NaN duplicated")
			}
			if set.Len() != 1 {
				t.Fatalf("Len = %d; want 1", set.Len())
			}
		}},
		{"NaNKey_Remove", func(t *testing.T, s floatSpecials[F], set *HashSet[F]) {
			set.Add(s.nan)
			set.Add(1)
			if !set.Remove(s.nan) {
				t.Fatal("Remove(NaN) = false")
			}
			if set.Contains(s.nan) || set.Len() != 1 {
				t.Fatal("NaN still present after Remove")
			}
			if set.Remove(s.nan) {
				t.Fatal("second Remove(NaN) = true")
			}
			if !set.Add(s.nan) {
				t.Fatal("re-Add(NaN) after Remove = false")
			}
		}},
		{"NegativeZeroDistinct", func(t *testing.T, s floatSpecials[F], set *HashSet[F]) {
			set.Add(s.posZero)
			if !set.Add(s.negZero) {
				t.Fatal("Add(-0) after Add(+0) = false; zeros collapsed")
			}
			if set.Len() != 2 || !set.Contains(s.posZero) || !set.Contains(s.negZero) {
				t.Fatalf("Len = %d; want 2 with both zeros", set.Len())
			}
			set.Remove(s.posZero)
			if set.Contains(s.posZero) || !set.Contains(s.negZero) {
				t.Fatal("Remove(+0) must leave -0")
			}
		}},
		{"InfinityKeys", func(t *testing.T, s floatSpecials[F], set *HashSet[F]) {
			set.Add(s.posInf)
			set.Add(s.negInf)
			set.Add(s.posInf)
			if set.Len() != 2 || !set.Contains(s.posInf) || !set.Contains(s.negInf) {
				t.Fatalf("Len = %d; want 2 infinities", set.Len())
			}
		}},
		{"NaNPayloadsDistinct", func(t *testing.T, s floatSpecials[F], set *HashSet[F]) {
			for _, k := range []F{s.nan, s.nanPayload, s.negNaN, s.negNaNPayload, s.nan} {
				set.Add(k)
			}
			if set.Len() != 4 {
				t.Fatalf("Len = %d; want 4 distinct NaNs", set.Len())
			}
		}},
		{"SetOpsKeepIdentity", func(t *testing.T, s floatSpecials[F], set *HashSet[F]) {
			set.Add(s.nan)
			set.Add(s.negZero)
			other := NewHashSetFrom(s.nan, s.posZero)
			if got := set.Intersect(other); got.Len() != 1 || !got.Contains(s.nan) {
				t.Fatalf("Intersect = %v; want {NaN}", got)
			}
			if got := set.Union(other); got.Len() != 3 {
				t.Fatalf("Union = %v; want 3 elements", got)
			}
			if got := set.Difference(other); got.Len() != 1 || !got.Contains(s.negZero) {
				t.Fatalf("Difference = %v; want {-0}", got)
			}
		}},
	}
}

func runHashSetFloat[F float32 | float64](t *testing.T, s floatSpecials[F]) {
	ctors := []struct {
		name string
		mk   func() *HashSet[F]
	}{
		{"New", NewHashSet[F]},
		{"From", func() *HashSet[F] { return NewHashSetFrom[F]() }},
		{"ZeroValue", func() *HashSet[F] { return &HashSet[F]{} }},
	}
	for _, c := range hashSetFloatCases[F]() {
		for _, ctor := range ctors {
			t.Run(c.name+"/"+ctor.name, func(t *testing.T) { c.run(t, s, ctor.mk()) })
		}
	}
}

func TestHashSetFloat32KeyBitIdentity(t *testing.T) { runHashSetFloat(t, specials32()) }
func TestHashSetFloat64KeyBitIdentity(t *testing.T) { runHashSetFloat(t, specials64()) }

// The other builtin-map-backed object/ hash collections share keyIndex.
func TestOtherHashCollectionsFloatKeyBitIdentity(t *testing.T) {
	s := specials64()
	keys := []float64{s.nan, s.nan, s.nanPayload, s.posZero, s.negZero, s.posInf, s.negInf}
	const distinct = 6

	t.Run("HashBag", func(t *testing.T) {
		b := NewHashBagFrom(keys...)
		if b.SizeDistinct() != distinct || b.OccurrencesOf(s.nan) != 2 || b.OccurrencesOf(s.negZero) != 1 {
			t.Fatalf("bag = %v", b)
		}
		if !b.Remove(s.nan) || b.OccurrencesOf(s.nan) != 1 {
			t.Fatal("Remove(NaN) did not decrement")
		}
	})
	t.Run("LinkedHashMap", func(t *testing.T) {
		m := NewLinkedHashMap[float64, int]()
		for i, k := range keys {
			m.Put(k, i)
		}
		if v, ok := m.Get(s.nan); m.Len() != distinct || !ok || v != 1 {
			t.Fatalf("len=%d Get(NaN)=%d,%v", m.Len(), v, ok)
		}
		if _, ok := m.Remove(s.negZero); !ok || !m.ContainsKey(s.posZero) {
			t.Fatal("Remove(-0) failed or removed +0")
		}
	})
	t.Run("LinkedHashSet", func(t *testing.T) {
		set := NewLinkedHashSetFrom(keys...)
		if set.Len() != distinct || !set.Contains(s.nan) || !set.Remove(s.nan) || set.Contains(s.nan) {
			t.Fatalf("set = %v", set)
		}
	})
	t.Run("HashBiMap", func(t *testing.T) {
		b := NewHashBiMap[float64, float64]()
		b.Put(s.nan, s.negZero)
		b.Put(s.posZero, s.nan)
		if b.Len() != 2 {
			t.Fatalf("Len = %d; want 2", b.Len())
		}
		if k, ok := b.GetInverse(s.negZero); !ok || !isNaN(k) {
			t.Fatalf("GetInverse(-0) = %v,%v; want NaN", k, ok)
		}
		b.Put(s.nan, s.nan) // value NaN moves from key +0 to key NaN
		if b.Len() != 1 || b.ContainsKey(s.posZero) || b.ContainsValue(s.negZero) {
			t.Fatalf("bijection broken: %v", b)
		}
		if _, ok := b.RemoveInverse(s.nan); !ok || b.Len() != 0 {
			t.Fatal("RemoveInverse(NaN) failed")
		}
	})
	t.Run("HashMultimap", func(t *testing.T) {
		mm := NewHashMultimap[float64, int]()
		for i, k := range keys {
			mm.Put(k, i)
		}
		if mm.SizeDistinct() != distinct || len(mm.Get(s.nan)) != 2 {
			t.Fatalf("multimap = %v", mm)
		}
		if vs := mm.RemoveKey(s.negZero); len(vs) != 1 || !mm.ContainsKey(s.posZero) {
			t.Fatal("RemoveKey(-0) failed or removed +0")
		}
	})
}

func isNaN(f float64) bool { return f != f }

type celsius float32

// Named types with a float underlying type get the same identity.
func TestHashMapNamedFloatKeyBitIdentity(t *testing.T) {
	m := NewHashMap[celsius, int]()
	nan := celsius(math.Float32frombits(0x7fc00000))
	m.Put(nan, 1)
	m.Put(nan, 2)
	m.Put(0, 3)
	m.Put(celsius(math.Float32frombits(0x80000000)), 4)
	if v, ok := m.Get(nan); m.Len() != 3 || !ok || v != 2 {
		t.Fatalf("len=%d Get(NaN)=%d,%v", m.Len(), v, ok)
	}
}

// ── Comparator total order ────────────────────────────────────────────

func checkTotalOrder[F float32 | float64](t *testing.T, s floatSpecials[F]) {
	ladder := s.ladder()
	nat := NaturalComparator[F]()
	rev := ReverseComparator[F]()
	byField := ComparatorByField(func(v struct{ f F }) F { return v.f })
	sign := func(x int) int {
		switch {
		case x < 0:
			return -1
		case x > 0:
			return 1
		}
		return 0
	}
	for i, a := range ladder {
		for j, b := range ladder {
			want := sign(i - j)
			if got := sign(nat(a, b)); got != want {
				t.Errorf("Natural(#%d %x, #%d %x) = %d; want %d", i, s.bits(a), j, s.bits(b), got, want)
			}
			if got := sign(rev(a, b)); got != -want {
				t.Errorf("Reverse(#%d, #%d) = %d; want %d", i, j, got, -want)
			}
			if got := sign(byField(struct{ f F }{a}, struct{ f F }{b})); got != want {
				t.Errorf("ComparatorByField(#%d, #%d) = %d; want %d", i, j, got, want)
			}
		}
	}
}

func TestNaturalComparatorFloat32TotalOrder(t *testing.T) { checkTotalOrder(t, specials32()) }
func TestNaturalComparatorFloat64TotalOrder(t *testing.T) { checkTotalOrder(t, specials64()) }

func TestNaturalComparatorNonFloatUnchanged(t *testing.T) {
	if NaturalComparator[int]()(1, 2) >= 0 || NaturalComparator[string]()("b", "a") <= 0 ||
		ReverseComparator[int]()(1, 2) <= 0 || NaturalComparator[int]()(3, 3) != 0 {
		t.Fatal("non-float natural/reverse ordering changed")
	}
}

func checkTreeSetTotalOrder[F float32 | float64](t *testing.T, s floatSpecials[F]) {
	ladder := s.ladder()
	shuffled := append([]F(nil), ladder...)
	rand.New(rand.NewSource(42)).Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	set := NewTreeSet[F](NaturalComparator[F]())
	for _, v := range shuffled {
		set.Add(v)
	}
	if set.Add(s.nan) || set.Add(s.negZero) {
		t.Fatal("re-adding NaN / -0 reported new element")
	}
	if set.Len() != len(ladder) {
		t.Fatalf("Len = %d; want %d", set.Len(), len(ladder))
	}
	if !set.Contains(s.negZero) || !set.Contains(s.posZero) {
		t.Fatal("-0 and +0 must both be members")
	}
	got := set.ToSlice()
	for i := range ladder {
		if s.bits(got[i]) != s.bits(ladder[i]) {
			t.Fatalf("ToSlice[%d] bits %x; want %x", i, s.bits(got[i]), s.bits(ladder[i]))
		}
	}
	if hi, _ := set.Max(); s.bits(hi) != s.bits(s.nanPayload) {
		t.Fatalf("Max bits %x; want +NaN payload", s.bits(hi))
	}
	if lo, _ := set.Min(); s.bits(lo) != s.bits(s.negNaNPayload) {
		t.Fatalf("Min bits %x; want -NaN payload", s.bits(lo))
	}
	if h, ok := set.Higher(s.negZero); !ok || s.bits(h) != s.bits(s.posZero) {
		t.Fatal("Higher(-0) must be +0")
	}
	if h, ok := set.Higher(s.posInf); !ok || s.bits(h) != s.bits(s.nan) {
		t.Fatal("Higher(+Inf) must be +NaN")
	}
	if !set.Remove(s.negZero) || !set.Contains(s.posZero) {
		t.Fatal("Remove(-0) failed or removed +0")
	}
}

func TestTreeSetFloat32NaturalTotalOrder(t *testing.T) { checkTreeSetTotalOrder(t, specials32()) }
func TestTreeSetFloat64NaturalTotalOrder(t *testing.T) { checkTreeSetTotalOrder(t, specials64()) }

func TestTreeMapFloat64NaturalKeepsZerosAndNaNs(t *testing.T) {
	s := specials64()
	m := NewTreeMap[float64, int](NaturalComparator[float64]())
	for i, k := range []float64{s.posZero, s.negZero, s.nan, s.nanPayload, s.nan} {
		m.Put(k, i)
	}
	if m.Len() != 4 {
		t.Fatalf("Len = %d; want 4", m.Len())
	}
	if v, ok := m.Get(s.nan); !ok || v != 4 {
		t.Fatalf("Get(NaN) = %d,%v; want 4,true", v, ok)
	}
}

// TestFloatKeyHashSpread pins the bucket spread of the float-key strategy
// hash. Round float values have long runs of zero low bits; a hash whose low
// bits (the table index) come from the Fibonacci product's low bits would put
// all of them in one bucket.
func TestFloatKeyHashSpread(t *testing.T) {
	const mask = 2047 // a 2048-slot table
	s32, ok32 := floatBitsStrategy[float32]()
	s64, ok64 := floatBitsStrategy[float64]()
	if !ok32 || !ok64 {
		t.Fatal("floatBitsStrategy: float kinds not recognised")
	}
	cases := []struct {
		name string
		hash func(i int) uint64
	}{
		{"float32 1..1000", func(i int) uint64 { return s32.HashCode(float32(i)) }},
		{"float64 1..1000", func(i int) uint64 { return s64.HashCode(float64(i)) }},
		{"float64 i/2", func(i int) uint64 { return s64.HashCode(float64(i) / 2) }},
		{"float64 2^(i-500)", func(i int) uint64 { return s64.HashCode(math.Ldexp(1, i-500)) }},
	}
	for _, c := range cases {
		buckets := map[uint64]bool{}
		for i := 1; i <= 1000; i++ {
			buckets[c.hash(i)&mask] = true
		}
		if len(buckets) < 700 { // a random hash fills ~790
			t.Errorf("%s: 1000 keys hit only %d of 2048 buckets", c.name, len(buckets))
		}
	}
}
