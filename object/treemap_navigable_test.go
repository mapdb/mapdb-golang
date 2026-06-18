// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.

package object

import (
	"slices"
	"testing"
)

func buildTestTreeMap() *TreeMap[int, string] {
	m := NewTreeMap[int, string](NaturalComparator[int]())
	m.Put(10, "a")
	m.Put(20, "b")
	m.Put(30, "c")
	m.Put(40, "d")
	m.Put(50, "e")
	return m
}

func TestTreeMap_FloorCeiling(t *testing.T) {
	m := buildTestTreeMap()
	// Floor(25) = 20
	if k, _, ok := m.Floor(25); !ok || k != 20 {
		t.Errorf("Floor(25) = (%v, %v), want 20", k, ok)
	}
	// Floor on exact match returns that key.
	if k, _, ok := m.Floor(30); !ok || k != 30 {
		t.Errorf("Floor(30) = (%v, %v), want 30", k, ok)
	}
	// Ceiling(25) = 30
	if k, _, ok := m.Ceiling(25); !ok || k != 30 {
		t.Errorf("Ceiling(25) = (%v, %v), want 30", k, ok)
	}
	// Ceiling below min is min.
	if k, _, ok := m.Ceiling(1); !ok || k != 10 {
		t.Errorf("Ceiling(1) = (%v, %v), want 10", k, ok)
	}
	// Floor above max is max.
	if k, _, ok := m.Floor(999); !ok || k != 50 {
		t.Errorf("Floor(999) = (%v, %v), want 50", k, ok)
	}
	// Floor below min is false.
	if _, _, ok := m.Floor(1); ok {
		t.Error("Floor(1) should be false on range [10,50]")
	}
}

func TestTreeMap_HigherLower(t *testing.T) {
	m := buildTestTreeMap()
	// Higher(30) skips 30 → 40
	if k, _, ok := m.Higher(30); !ok || k != 40 {
		t.Errorf("Higher(30) = (%v, %v), want 40", k, ok)
	}
	// Lower(30) skips 30 → 20
	if k, _, ok := m.Lower(30); !ok || k != 20 {
		t.Errorf("Lower(30) = (%v, %v), want 20", k, ok)
	}
	// Higher at max = false
	if _, _, ok := m.Higher(50); ok {
		t.Error("Higher(max) should be false")
	}
	// Lower at min = false
	if _, _, ok := m.Lower(10); ok {
		t.Error("Lower(min) should be false")
	}
}

func TestTreeMap_HeadMap(t *testing.T) {
	m := buildTestTreeMap()
	var keys []int
	for k := range m.HeadMap(30) {
		keys = append(keys, k)
	}
	if !slices.Equal(keys, []int{10, 20}) {
		t.Errorf("HeadMap(30) keys = %v", keys)
	}
}

func TestTreeMap_TailMap(t *testing.T) {
	m := buildTestTreeMap()
	var keys []int
	for k := range m.TailMap(30) {
		keys = append(keys, k)
	}
	if !slices.Equal(keys, []int{30, 40, 50}) {
		t.Errorf("TailMap(30) keys = %v", keys)
	}
}

func TestTreeMap_SubMap(t *testing.T) {
	m := buildTestTreeMap()
	var keys []int
	for k := range m.SubMap(20, 40) {
		keys = append(keys, k)
	}
	if !slices.Equal(keys, []int{20, 30}) {
		t.Errorf("SubMap(20,40) keys = %v", keys)
	}
}

func TestTreeMap_FirstLastEntry(t *testing.T) {
	m := buildTestTreeMap()
	if k, _, ok := m.FirstEntry(); !ok || k != 10 {
		t.Errorf("FirstEntry = %v, want 10", k)
	}
	if k, _, ok := m.LastEntry(); !ok || k != 50 {
		t.Errorf("LastEntry = %v, want 50", k)
	}
}

func TestTreeMap_PollFirstEntry(t *testing.T) {
	m := buildTestTreeMap()
	if k, _, ok := m.PollFirstEntry(); !ok || k != 10 {
		t.Errorf("PollFirstEntry = %v, want 10", k)
	}
	if m.ContainsKey(10) {
		t.Error("10 should be removed after PollFirstEntry")
	}
	if m.Len() != 4 {
		t.Errorf("Size after PollFirst = %d, want 4", m.Len())
	}
}

func TestTreeMap_PollLastEntry(t *testing.T) {
	m := buildTestTreeMap()
	if k, _, ok := m.PollLastEntry(); !ok || k != 50 {
		t.Errorf("PollLastEntry = %v, want 50", k)
	}
	if m.ContainsKey(50) {
		t.Error("50 should be removed")
	}
}

func TestTreeMap_PollEmpty(t *testing.T) {
	m := NewTreeMap[int, int](NaturalComparator[int]())
	if _, _, ok := m.PollFirstEntry(); ok {
		t.Error("PollFirstEntry on empty should be false")
	}
	if _, _, ok := m.PollLastEntry(); ok {
		t.Error("PollLastEntry on empty should be false")
	}
}

func TestTreeMap_DescendingKeys(t *testing.T) {
	m := buildTestTreeMap()
	var keys []int
	for k := range m.DescendingKeys() {
		keys = append(keys, k)
	}
	if !slices.Equal(keys, []int{50, 40, 30, 20, 10}) {
		t.Errorf("DescendingKeys = %v", keys)
	}
}

func TestTreeMap_DescendingMap(t *testing.T) {
	m := buildTestTreeMap()
	type kv struct {
		K int
		V string
	}
	var got []kv
	for k, v := range m.DescendingMap() {
		got = append(got, kv{k, v})
	}
	want := []kv{{50, "e"}, {40, "d"}, {30, "c"}, {20, "b"}, {10, "a"}}
	if !slices.Equal(got, want) {
		t.Errorf("DescendingMap = %v", got)
	}
}

func TestTreeMap_SubMapCopyPreservesReverseComparator(t *testing.T) {
	// SubMapCopy must keep the source ordering (reverse), not reset to natural,
	// and be an independent snapshot.
	m := NewTreeMap[int, int](ReverseComparator[int]())
	for _, k := range []int{10, 20, 30, 40, 50} {
		m.Put(k, k*10)
	}
	// Source iterates descending under the reverse comparator.
	var src []int
	for k := range m.Keys() {
		src = append(src, k)
	}
	if !slices.Equal(src, []int{50, 40, 30, 20, 10}) {
		t.Fatalf("source order = %v, want descending", src)
	}
	// Under a reverse comparator [from, to) means from >= key > to, i.e. the
	// half-open window in comparator order starting at 40 up to (not incl.) 10.
	sub := m.SubMapCopy(40, 10) // {40,30,20} in reverse order
	var subKeys []int
	for k := range sub.Keys() {
		subKeys = append(subKeys, k)
	}
	if !slices.Equal(subKeys, []int{40, 30, 20}) {
		t.Errorf("SubMapCopy keys = %v, want [40 30 20] (reverse order preserved)", subKeys)
	}
	// Independence: mutating the snapshot does not touch the original.
	sub.Remove(30)
	if !m.ContainsKey(30) {
		t.Error("original must still contain 30 after mutating snapshot")
	}
	sub.Put(99, 990)
	if m.ContainsKey(99) {
		t.Error("original must not gain 99 from snapshot")
	}
	// And vice versa.
	m.Remove(40)
	if !sub.ContainsKey(40) {
		t.Error("snapshot must still contain 40 after mutating original")
	}
}

func TestTreeMap_HeadTailSubMap_EarlyBreak(t *testing.T) {
	// Breaking out of iter.Seq2 must unwind correctly for each view.
	m := buildTestTreeMap()
	count := 0
	for range m.TailMap(20) {
		count++
		if count == 2 {
			break
		}
	}
	if count != 2 {
		t.Errorf("TailMap early-break count = %d", count)
	}
}
