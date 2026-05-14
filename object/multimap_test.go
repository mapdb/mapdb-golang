// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.

package object

import (
	"slices"
	"testing"
)

// ── HashMultimap ──────────────────────────────────────────────────────

func TestHashMultimap_PutAndGet(t *testing.T) {
	m := NewHashMultimap[string, int]()
	m.Put("a", 1)
	m.Put("a", 2)
	m.Put("b", 3)

	if !slices.Equal(m.Get("a"), []int{1, 2}) {
		t.Errorf("Get(a) = %v, want [1 2]", m.Get("a"))
	}
	if !slices.Equal(m.Get("b"), []int{3}) {
		t.Errorf("Get(b) = %v, want [3]", m.Get("b"))
	}
	if m.Size() != 3 {
		t.Errorf("Size = %d, want 3", m.Size())
	}
	if m.SizeDistinct() != 2 {
		t.Errorf("SizeDistinct = %d, want 2", m.SizeDistinct())
	}
}

func TestHashMultimap_PutAll(t *testing.T) {
	m := NewHashMultimap[string, int]()
	m.PutAll("a", 1, 2, 3)
	m.PutAll("a", 4) // appends, not overwrites
	if !slices.Equal(m.Get("a"), []int{1, 2, 3, 4}) {
		t.Errorf("got %v", m.Get("a"))
	}
	if m.Size() != 4 {
		t.Errorf("Size = %d, want 4", m.Size())
	}
}

func TestHashMultimap_ContainsKey(t *testing.T) {
	m := NewHashMultimap[string, int]()
	m.Put("a", 1)
	if !m.ContainsKey("a") {
		t.Error("ContainsKey(a) should be true")
	}
	if m.ContainsKey("missing") {
		t.Error("ContainsKey(missing) should be false")
	}
}

func TestHashMultimap_RemoveKey(t *testing.T) {
	m := NewHashMultimap[string, int]()
	m.Put("a", 1)
	m.Put("a", 2)
	m.Put("b", 3)
	removed := m.RemoveKey("a")
	if !slices.Equal(removed, []int{1, 2}) {
		t.Errorf("RemoveKey(a) = %v, want [1 2]", removed)
	}
	if m.ContainsKey("a") {
		t.Error("ContainsKey(a) after RemoveKey should be false")
	}
	if m.Size() != 1 {
		t.Errorf("Size = %d, want 1", m.Size())
	}
}

func TestHashMultimap_RemoveMatching(t *testing.T) {
	m := NewHashMultimap[string, int]()
	m.PutAll("nums", 1, 2, 3, 2, 4)
	n := m.RemoveMatching("nums", 2, func(a, b int) bool { return a == b })
	if n != 2 {
		t.Errorf("removed %d, want 2", n)
	}
	if !slices.Equal(m.Get("nums"), []int{1, 3, 4}) {
		t.Errorf("after remove: %v", m.Get("nums"))
	}
}

func TestHashMultimap_RemoveMatching_AllValuesGone(t *testing.T) {
	m := NewHashMultimap[string, int]()
	m.PutAll("k", 1, 1, 1)
	n := m.RemoveMatching("k", 1, func(a, b int) bool { return a == b })
	if n != 3 {
		t.Errorf("removed %d, want 3", n)
	}
	if m.ContainsKey("k") {
		t.Error("key with all values removed should be deleted")
	}
}

func TestHashMultimap_Clear(t *testing.T) {
	m := NewHashMultimap[string, int]()
	m.Put("a", 1)
	m.Clear()
	if !m.IsEmpty() {
		t.Error("should be empty after Clear")
	}
}

func TestHashMultimap_Iteration(t *testing.T) {
	m := NewHashMultimap[string, int]()
	m.Put("a", 1)
	m.Put("a", 2)
	m.Put("b", 3)

	total := 0
	for range m.All() {
		total++
	}
	if total != 3 {
		t.Errorf("All yielded %d pairs, want 3", total)
	}

	distinct := 0
	for range m.Keys() {
		distinct++
	}
	if distinct != 2 {
		t.Errorf("Keys yielded %d, want 2", distinct)
	}

	values := 0
	for range m.Values() {
		values++
	}
	if values != 3 {
		t.Errorf("Values yielded %d, want 3", values)
	}
}

func TestHashMultimap_ToMap_IsDefensive(t *testing.T) {
	m := NewHashMultimap[string, int]()
	m.PutAll("a", 1, 2)
	cp := m.ToMap()
	cp["a"] = append(cp["a"], 99)
	if slices.Contains(m.Get("a"), 99) {
		t.Error("ToMap should return a defensive copy")
	}
}

// ── TreeMultimap ──────────────────────────────────────────────────────

func TestTreeMultimap_KeysSorted(t *testing.T) {
	m := NewTreeMultimap[string, int](NaturalComparator[string]())
	m.Put("banana", 2)
	m.Put("apple", 1)
	m.Put("cherry", 3)
	m.Put("apple", 4)

	got := make([]string, 0, 3)
	for k := range m.Keys() {
		got = append(got, k)
	}
	if !slices.Equal(got, []string{"apple", "banana", "cherry"}) {
		t.Errorf("Keys in sorted order = %v", got)
	}
}

func TestTreeMultimap_WithinKeyInsertionOrder(t *testing.T) {
	m := NewTreeMultimap[int, string](NaturalComparator[int]())
	m.Put(1, "first")
	m.Put(1, "second")
	m.Put(1, "third")
	if !slices.Equal(m.Get(1), []string{"first", "second", "third"}) {
		t.Errorf("within-key order broken: %v", m.Get(1))
	}
}

func TestTreeMultimap_Size(t *testing.T) {
	m := NewTreeMultimap[int, int](NaturalComparator[int]())
	m.PutAll(1, 10, 20, 30)
	m.PutAll(2, 40)
	if m.Size() != 4 {
		t.Errorf("Size = %d, want 4", m.Size())
	}
	if m.SizeDistinct() != 2 {
		t.Errorf("SizeDistinct = %d, want 2", m.SizeDistinct())
	}
}

func TestTreeMultimap_RemoveKey(t *testing.T) {
	m := NewTreeMultimap[int, int](NaturalComparator[int]())
	m.PutAll(1, 10, 20)
	m.Put(2, 30)
	removed := m.RemoveKey(1)
	if !slices.Equal(removed, []int{10, 20}) {
		t.Errorf("RemoveKey(1) = %v", removed)
	}
	if m.Size() != 1 {
		t.Errorf("Size after = %d, want 1", m.Size())
	}
}

func TestTreeMultimap_RemoveMatching(t *testing.T) {
	m := NewTreeMultimap[int, int](NaturalComparator[int]())
	m.PutAll(1, 10, 20, 10, 30)
	n := m.RemoveMatching(1, 10, func(a, b int) bool { return a == b })
	if n != 2 {
		t.Errorf("removed %d, want 2", n)
	}
	if !slices.Equal(m.Get(1), []int{20, 30}) {
		t.Errorf("got %v", m.Get(1))
	}
}

func TestTreeMultimap_Iteration(t *testing.T) {
	m := NewTreeMultimap[int, string](NaturalComparator[int]())
	m.Put(2, "b")
	m.Put(1, "a1")
	m.Put(1, "a2")

	type kv struct {
		k int
		v string
	}
	var got []kv
	for k, v := range m.All() {
		got = append(got, kv{k, v})
	}
	want := []kv{{1, "a1"}, {1, "a2"}, {2, "b"}}
	if !slices.Equal(got, want) {
		t.Errorf("All order = %v, want %v", got, want)
	}
}

func TestTreeMultimap_ReverseComparator(t *testing.T) {
	m := NewTreeMultimap[int, string](ReverseComparator[int]())
	m.Put(1, "a")
	m.Put(3, "c")
	m.Put(2, "b")

	var keys []int
	for k := range m.Keys() {
		keys = append(keys, k)
	}
	if !slices.Equal(keys, []int{3, 2, 1}) {
		t.Errorf("descending keys = %v", keys)
	}
}
