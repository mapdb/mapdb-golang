package seq

import (
	"slices"
	"testing"
)

func TestEnumerate(t *testing.T) {
	var keys []int
	var vals []string
	for i, v := range Of("a", "b", "c").Enumerate() {
		keys = append(keys, i)
		vals = append(vals, v)
	}
	if !slices.Equal(keys, []int{0, 1, 2}) || !slices.Equal(vals, []string{"a", "b", "c"}) {
		t.Errorf("Enumerate = %v/%v", keys, vals)
	}
}

func TestSeq2KeysValuesSwap(t *testing.T) {
	s := Of(10, 20, 30).Enumerate() // (0,10) (1,20) (2,30)
	if got := s.Keys().ToSlice(); !slices.Equal(got, []int{0, 1, 2}) {
		t.Errorf("Keys = %v", got)
	}
	if got := s.Values().ToSlice(); !slices.Equal(got, []int{10, 20, 30}) {
		t.Errorf("Values = %v", got)
	}
	// Swap turns (idx,val) into (val,idx); collect the swapped keys.
	if got := s.Swap().Keys().ToSlice(); !slices.Equal(got, []int{10, 20, 30}) {
		t.Errorf("Swap.Keys = %v", got)
	}
}

func TestSeq2Filter(t *testing.T) {
	s := Of(1, 2, 3, 4).Enumerate().Filter(func(i, v int) bool { return v%2 == 0 })
	if got := s.Values().ToSlice(); !slices.Equal(got, []int{2, 4}) {
		t.Errorf("Seq2.Filter = %v", got)
	}
}

func TestSeq2ReRunnable(t *testing.T) {
	s := Of(1, 2, 3).Enumerate().Filter(func(i, v int) bool { return v > 1 })
	for r := 0; r < 2; r++ {
		if got := s.Values().ToSlice(); !slices.Equal(got, []int{2, 3}) {
			t.Errorf("run %d: %v", r, got)
		}
	}
}

func TestMap2MapKeysMapValues(t *testing.T) {
	base := Of(1, 2, 3).Enumerate() // (0,1)(1,2)(2,3)
	mv := MapValues(base, func(_ int, v int) int { return v * 10 })
	if got := mv.Values().ToSlice(); !slices.Equal(got, []int{10, 20, 30}) {
		t.Errorf("MapValues = %v", got)
	}
	mk := MapKeys(base, func(k int, _ int) int { return k + 100 })
	if got := mk.Keys().ToSlice(); !slices.Equal(got, []int{100, 101, 102}) {
		t.Errorf("MapKeys = %v", got)
	}
	m2 := Map2(base, func(k, v int) (string, int) { return "", k + v })
	if got := m2.Values().ToSlice(); !slices.Equal(got, []int{1, 3, 5}) {
		t.Errorf("Map2 = %v", got)
	}
}

func TestFromMapAndToMap(t *testing.T) {
	src := map[string]int{"a": 1, "b": 2, "c": 3}
	got, err := ToMap(FromMap(src), ErrorOnDuplicate)
	if err != nil {
		t.Fatalf("ToMap err = %v", err)
	}
	if len(got) != 3 || got["a"] != 1 || got["b"] != 2 || got["c"] != 3 {
		t.Errorf("round-trip = %v", got)
	}
}

func TestToMapDuplicatePolicies(t *testing.T) {
	dup := func() Seq2[string, int] {
		return func(yield func(string, int) bool) {
			pairs := []struct {
				k string
				v int
			}{{"x", 1}, {"x", 2}, {"y", 3}}
			for _, p := range pairs {
				if !yield(p.k, p.v) {
					return
				}
			}
		}
	}
	if _, err := ToMap(dup(), ErrorOnDuplicate); err == nil {
		t.Error("ErrorOnDuplicate should error on repeated key")
	}
	if m, err := ToMap(dup(), KeepFirst); err != nil || m["x"] != 1 {
		t.Errorf("KeepFirst = %v,%v", m, err)
	}
	if m, err := ToMap(dup(), KeepLast); err != nil || m["x"] != 2 {
		t.Errorf("KeepLast = %v,%v", m, err)
	}
}

func TestGroupBy2(t *testing.T) {
	// pairs (parity, value); regroup by whether value > 2
	s := Of(1, 2, 3, 4).Enumerate()
	g := GroupBy2(s, func(_ int, v int) bool { return v > 2 })
	if !slices.Equal(g[false], []int{1, 2}) || !slices.Equal(g[true], []int{3, 4}) {
		t.Errorf("GroupBy2 = %v", g)
	}
}
