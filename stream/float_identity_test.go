package stream

import (
	"math"
	"slices"
	"testing"
)

// Distinct and Contains match float elements by bit pattern (spec
// algorithms.md "NaN must hash and compare by bit pattern"); other element
// types keep ==.

type celsius float64
type single float32

func edges64() []float64 {
	return []float64{
		math.Float64frombits(0x7ff8000000000000),
		math.Float64frombits(0x7ff8000000000001),
		math.Float64frombits(0xfff8000000000000),
		math.Float64frombits(0xfff8000000000001),
		0, math.Copysign(0, -1), math.Inf(1), math.Inf(-1), 1.5,
	}
}

func edges32() []float32 {
	return []float32{
		math.Float32frombits(0x7fc00000),
		math.Float32frombits(0x7fc00001),
		math.Float32frombits(0xffc00000),
		math.Float32frombits(0xffc00001),
		0, math.Float32frombits(0x80000000),
		float32(math.Inf(1)), float32(math.Inf(-1)), 1.5,
	}
}

func convert[F, T ~float32 | ~float64](in []F) []T {
	out := make([]T, len(in))
	for i, v := range in {
		out[i] = T(v)
	}
	return out
}

// checkSeqIdentity takes values that are pairwise distinct under the
// expected identity; same reports that identity.
func checkSeqIdentity[T comparable](t *testing.T, vals []T, same func(a, b T) bool) {
	t.Helper()
	doubled := append(append([]T{}, vals...), vals...)
	got := slices.Collect(Distinct(slices.Values(doubled)))
	if len(got) != len(vals) {
		t.Fatalf("Distinct len = %d, want %d (%v)", len(got), len(vals), got)
	}
	for i := range vals {
		if !same(got[i], vals[i]) {
			t.Errorf("Distinct[%d] = %v, want %v", i, got[i], vals[i])
		}
	}
	for i, v := range vals {
		if !Contains(slices.Values(vals), v) {
			t.Errorf("Contains(%v) = false", v)
		}
		without := append(append([]T{}, vals[:i]...), vals[i+1:]...)
		if Contains(slices.Values(without), v) {
			t.Errorf("Contains(%v) on seq without it = true", v)
		}
	}
}

func TestStreamFloatIdentity(t *testing.T) {
	bits64 := func(a, b float64) bool { return math.Float64bits(a) == math.Float64bits(b) }
	bits32 := func(a, b float32) bool { return math.Float32bits(a) == math.Float32bits(b) }
	t.Run("float64", func(t *testing.T) { checkSeqIdentity(t, edges64(), bits64) })
	t.Run("float32", func(t *testing.T) { checkSeqIdentity(t, edges32(), bits32) })
	t.Run("named float64", func(t *testing.T) {
		checkSeqIdentity(t, convert[float64, celsius](edges64()), func(a, b celsius) bool { return bits64(float64(a), float64(b)) })
	})
	t.Run("named float32", func(t *testing.T) {
		checkSeqIdentity(t, convert[float32, single](edges32()), func(a, b single) bool { return bits32(float32(a), float32(b)) })
	})
	t.Run("int unchanged", func(t *testing.T) {
		checkSeqIdentity(t, []int{0, -1, 1, math.MaxInt}, func(a, b int) bool { return a == b })
	})
}

// GroupByToMap and ToMap return builtin maps by contract, so float keys get
// Go == identity there (documented limitation); GroupBy keeps bit-pattern
// keys.
func TestStreamBuiltinMapFloatKeysDocumented(t *testing.T) {
	nan := math.Float64frombits(0x7ff8000000000000)
	negZero := math.Copysign(0, -1)
	vals := []float64{nan, nan, 0, negZero}
	id := func(v float64) float64 { return v }

	if g := GroupBy(slices.Values(vals), id); g.SizeDistinct() != 3 || g.Len() != 4 {
		t.Fatalf("GroupBy: SizeDistinct=%d Len=%d, want 3/4", g.SizeDistinct(), g.Len())
	}
	// Builtin map: each NaN is its own unreachable group, ±0 merge.
	if m := GroupByToMap(slices.Values(vals), id); len(m) != 3 || len(m[0]) != 2 {
		t.Fatalf("GroupByToMap: len=%d len(m[0])=%d, want 3/2 (documented builtin identity)", len(m), len(m[0]))
	}
	pairs := func(yield func(float64, int) bool) {
		for i, v := range vals {
			if !yield(v, i) {
				return
			}
		}
	}
	if m := ToMap(pairs); len(m) != 3 || m[0] != 3 {
		t.Fatalf("ToMap: len=%d m[0]=%d, want 3/3 (documented builtin identity)", len(m), m[0])
	}
}
