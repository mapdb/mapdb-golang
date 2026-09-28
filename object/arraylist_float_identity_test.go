package object

import (
	"math"
	"testing"
)

// ArrayList.Distinct/Contains/IndexOf/Remove and ArrayStack.Contains match
// float elements by bit pattern (spec algorithms.md "NaN must hash and
// compare by bit pattern"); other element types keep ==.

type kelvin float64

func convertFloats[F, T ~float32 | ~float64](in []F) []T {
	out := make([]T, len(in))
	for i, v := range in {
		out[i] = T(v)
	}
	return out
}

// checkListIdentity takes values that are pairwise distinct under the
// expected identity.
func checkListIdentity[T comparable](t *testing.T, vals []T) {
	t.Helper()
	doubled := append(append([]T{}, vals...), vals...)
	d := NewArrayListFrom(doubled...).Distinct()
	if d.Len() != len(vals) {
		t.Fatalf("Distinct Len = %d, want %d (%v)", d.Len(), len(vals), d)
	}
	for i, v := range vals {
		if !sameKey(d.Get(i), v) {
			t.Errorf("Distinct[%d] = %v, want %v", i, d.Get(i), v)
		}
	}
	l := NewArrayListFrom(vals...)
	for i, v := range vals {
		if !l.Contains(v) {
			t.Errorf("Contains(%v) = false", v)
		}
		if got := l.IndexOf(v); got != i {
			t.Errorf("IndexOf(%v) = %d, want %d", v, got, i)
		}
		if !NewArrayStackFrom(vals...).Contains(v) {
			t.Errorf("ArrayStack.Contains(%v) = false", v)
		}
		without := NewArrayListFrom(append(append([]T{}, vals[:i]...), vals[i+1:]...)...)
		if without.Contains(v) || without.IndexOf(v) != -1 {
			t.Errorf("list without %v still finds it", v)
		}
		if NewArrayStackFrom(without.ToSlice()...).Contains(v) {
			t.Errorf("ArrayStack without %v still finds it", v)
		}
		r := NewArrayListFrom(vals...)
		if !r.Remove(v) || r.Len() != len(vals)-1 || r.IndexOf(v) != -1 {
			t.Errorf("Remove(%v) removed the wrong element: %v", v, r)
		}
		if without.Remove(v) {
			t.Errorf("Remove(%v) on list without it = true", v)
		}
	}
}

func TestArrayListFloatIdentity(t *testing.T) {
	t.Run("float64", func(t *testing.T) { checkListIdentity(t, specials64().ladder()) })
	t.Run("float32", func(t *testing.T) { checkListIdentity(t, specials32().ladder()) })
	t.Run("named float32", func(t *testing.T) {
		checkListIdentity(t, convertFloats[float32, celsius](specials32().ladder()))
	})
	t.Run("named float64", func(t *testing.T) {
		checkListIdentity(t, convertFloats[float64, kelvin](specials64().ladder()))
	})
	t.Run("int unchanged", func(t *testing.T) { checkListIdentity(t, []int{0, -1, 1, math.MaxInt}) })
	t.Run("string unchanged", func(t *testing.T) { checkListIdentity(t, []string{"", "NaN", "-0"}) })
}

// Composite elements containing floats keep == (documented scope).
func TestArrayListCompositeKeepsEquality(t *testing.T) {
	type pair struct{ F float64 }
	l := NewArrayListFrom(pair{0}, pair{math.Copysign(0, -1)}, pair{math.NaN()})
	if got := l.Distinct().Len(); got != 2 {
		t.Fatalf("Distinct Len = %d, want 2 (±0 fields merge; NaN field unequal to itself)", got)
	}
	if l.Contains(pair{math.NaN()}) {
		t.Fatal("Contains(NaN field) = true, want false under ==")
	}
}
