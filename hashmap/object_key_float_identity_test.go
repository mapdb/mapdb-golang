package hashmap

import (
	"math"
	"math/rand"
	"testing"
)

// Object-keyed maps (Object<Value>[K]) with a float K must identify keys by
// bit pattern (spec algorithms.md "NaN must hash and compare by bit
// pattern"): NaN finds itself, NaN payloads and signs are distinct keys,
// -0.0 != +0.0.

type celsiusKey float64
type singleKey float32

// floatEdges64 returns pairwise bit-distinct float64 edge values.
func floatEdges64() []float64 {
	return []float64{
		math.Float64frombits(0x7ff8000000000000), // NaN
		math.Float64frombits(0x7ff8000000000001), // NaN, other payload
		math.Float64frombits(0xfff8000000000000), // -NaN
		math.Float64frombits(0xfff8000000000001), // -NaN, other payload
		0,
		math.Copysign(0, -1),
		math.Inf(1),
		math.Inf(-1),
		1.5,
	}
}

func floatEdges32() []float32 {
	return []float32{
		math.Float32frombits(0x7fc00000),
		math.Float32frombits(0x7fc00001),
		math.Float32frombits(0xffc00000),
		math.Float32frombits(0xffc00001),
		0,
		math.Float32frombits(0x80000000),
		float32(math.Inf(1)),
		float32(math.Inf(-1)),
		1.5,
	}
}

func convertKeys[F, K ~float32 | ~float64](in []F) []K {
	out := make([]K, len(in))
	for i, v := range in {
		out[i] = K(v)
	}
	return out
}

// checkObjectInt32Identity puts every key (pairwise distinct under the
// expected identity) twice and checks Len, Get, ContainsKey, overwrite and
// Remove, plus the immutable wrapper.
func checkObjectInt32Identity[K comparable](t *testing.T, keys []K) {
	t.Helper()
	m := NewObjectInt32[K]()
	for i, k := range keys {
		if _, existed := m.Put(k, int32(i)); existed {
			t.Fatalf("Put(%v) #%d: reported existing key on first insert", k, i)
		}
	}
	for i, k := range keys {
		old, existed := m.Put(k, int32(100+i))
		if !existed || old != int32(i) {
			t.Fatalf("Put(%v) overwrite = (%d, %v), want (%d, true)", k, old, existed, i)
		}
	}
	if m.Len() != len(keys) {
		t.Fatalf("Len = %d, want %d", m.Len(), len(keys))
	}
	imm := NewImmutableObjectInt32From(m)
	for i, k := range keys {
		if v, ok := m.Get(k); !ok || v != int32(100+i) {
			t.Errorf("Get(%v) = (%d, %v), want (%d, true)", k, v, ok, 100+i)
		}
		if !imm.ContainsKey(k) {
			t.Errorf("immutable ContainsKey(%v) = false", k)
		}
	}
	for i, k := range keys {
		if v, ok := m.Remove(k); !ok || v != int32(100+i) {
			t.Fatalf("Remove(%v) = (%d, %v), want (%d, true)", k, v, ok, 100+i)
		}
		if m.ContainsKey(k) {
			t.Fatalf("ContainsKey(%v) after Remove", k)
		}
		for _, rest := range keys[i+1:] {
			if !m.ContainsKey(rest) {
				t.Fatalf("Remove(%v) lost %v", k, rest)
			}
		}
	}
	if m.Len() != 0 {
		t.Fatalf("Len after removing all = %d", m.Len())
	}
}

func TestObjectKeyFloatIdentity(t *testing.T) {
	t.Run("float64", func(t *testing.T) { checkObjectInt32Identity(t, floatEdges64()) })
	t.Run("float32", func(t *testing.T) { checkObjectInt32Identity(t, floatEdges32()) })
	t.Run("named float64", func(t *testing.T) {
		checkObjectInt32Identity(t, convertKeys[float64, celsiusKey](floatEdges64()))
	})
	t.Run("named float32", func(t *testing.T) {
		checkObjectInt32Identity(t, convertKeys[float32, singleKey](floatEdges32()))
	})
	t.Run("int unchanged", func(t *testing.T) { checkObjectInt32Identity(t, []int{0, -1, 1, 42, math.MaxInt}) })
	t.Run("string unchanged", func(t *testing.T) { checkObjectInt32Identity(t, []string{"", "a", "NaN", "-0"}) })
}

// A struct key that contains a float keeps Go == semantics (documented
// scope): a NaN field makes the key unequal to itself.
func TestObjectKeyCompositeKeepsEquality(t *testing.T) {
	type pair struct{ F float64 }
	m := NewObjectInt32[pair]()
	m.Put(pair{0}, 1)
	if v, ok := m.Get(pair{math.Copysign(0, -1)}); !ok || v != 1 {
		t.Fatalf("composite key: -0.0 field should == +0.0 field, got (%d, %v)", v, ok)
	}
}

// Many NaN payloads drive resize and backward-shift deletion.
func TestObjectKeyManyNaNPayloads(t *testing.T) {
	const n = 2000
	m := NewObjectFloat64[float64]()
	for i := range n {
		m.Put(math.Float64frombits(0x7ff8000000000000|uint64(i)), float64(i))
	}
	if m.Len() != n {
		t.Fatalf("Len = %d, want %d", m.Len(), n)
	}
	for i := 0; i < n; i += 2 {
		if _, ok := m.Remove(math.Float64frombits(0x7ff8000000000000 | uint64(i))); !ok {
			t.Fatalf("Remove payload %d failed", i)
		}
	}
	for i := range n {
		v, ok := m.Get(math.Float64frombits(0x7ff8000000000000 | uint64(i)))
		if want := i%2 == 1; ok != want || (ok && v != float64(i)) {
			t.Fatalf("Get payload %d = (%v, %v), want present=%v", i, v, ok, want)
		}
	}
}

// Every generated Object<Value> map shares the template; spot-check one
// NaN/-0.0 round trip on each value flavour.
func TestObjectKeyFloatIdentityAllValueTypes(t *testing.T) {
	nan := math.Float64frombits(0x7ff8000000000001)
	negZero := math.Copysign(0, -1)
	check := func(name string, put func(float64), has func(float64) bool, n func() int) {
		put(nan)
		put(nan)
		put(0)
		put(negZero)
		if !has(nan) || !has(negZero) || !has(0) || n() != 3 {
			t.Errorf("%s: NaN/-0.0 identity broken (len %d)", name, n())
		}
	}
	a := NewObjectChar[float64]()
	check("Char", func(k float64) { a.Put(k, 'x') }, a.ContainsKey, a.Len)
	b := NewObjectFloat32[float64]()
	check("Float32", func(k float64) { b.Put(k, 1) }, b.ContainsKey, b.Len)
	c := NewObjectFloat64[float64]()
	check("Float64", func(k float64) { c.Put(k, 1) }, c.ContainsKey, c.Len)
	d := NewObjectInt8[float64]()
	check("Int8", func(k float64) { d.Put(k, 1) }, d.ContainsKey, d.Len)
	e := NewObjectInt16[float64]()
	check("Int16", func(k float64) { e.Put(k, 1) }, e.ContainsKey, e.Len)
	f := NewObjectInt32[float64]()
	check("Int32", func(k float64) { f.Put(k, 1) }, f.ContainsKey, f.Len)
	g := NewObjectInt64[float64]()
	check("Int64", func(k float64) { g.Put(k, 1) }, g.ContainsKey, g.Len)
}

// Random Put/Remove against a reference, for an int key (builtin == path) and
// a float key restricted to NaN payloads and signed zeros (bit-pattern path,
// reference keyed by bits). Guards the backward-shift deletion, which on
// 0462a09 could shift an entry before its ideal slot and lose it.
func TestObjectKeyRandomPutRemove(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	mi := NewObjectInt32[int]()
	refI := map[int]int32{}
	mf := NewObjectInt32[float64]()
	refF := map[uint64]int32{}
	floatKey := func(n int) float64 {
		switch n % 4 {
		case 0:
			return math.Copysign(0, -1)
		case 1:
			return 0
		default:
			return math.Float64frombits(0x7ff8000000000000 | uint64(n))
		}
	}
	for i := range 100000 {
		n := r.Intn(64)
		fk := floatKey(n)
		if r.Intn(2) == 0 {
			mi.Put(n, int32(i))
			refI[n] = int32(i)
			mf.Put(fk, int32(i))
			refF[math.Float64bits(fk)] = int32(i)
			continue
		}
		_, got := mi.Remove(n)
		_, want := refI[n]
		delete(refI, n)
		if got != want {
			t.Fatalf("int step %d: Remove(%d) = %v, want %v", i, n, got, want)
		}
		_, got = mf.Remove(fk)
		_, want = refF[math.Float64bits(fk)]
		delete(refF, math.Float64bits(fk))
		if got != want {
			t.Fatalf("float step %d: Remove(%v) = %v, want %v", i, fk, got, want)
		}
	}
	if mi.Len() != len(refI) || mf.Len() != len(refF) {
		t.Fatalf("Len = %d/%d, want %d/%d", mi.Len(), mf.Len(), len(refI), len(refF))
	}
}
