// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package object

import (
	"slices"
	"testing"
)

func TestLinkedHashMapBasic(t *testing.T) {
	m := NewLinkedHashMap[string, int]()
	if !m.IsEmpty() {
		t.Fatal("expected empty")
	}

	_, existed := m.Put("a", 1)
	if existed {
		t.Fatal("should not exist")
	}
	m.Put("b", 2)
	m.Put("c", 3)

	if m.Size() != 3 {
		t.Fatalf("expected 3, got %d", m.Size())
	}
	if v, ok := m.Get("b"); !ok || v != 2 {
		t.Fatalf("expected 2, got %v", v)
	}
	if !m.ContainsKey("a") {
		t.Fatal("expected contains a")
	}
}

func TestLinkedHashMapInsertionOrder(t *testing.T) {
	m := NewLinkedHashMap[string, int]()
	m.Put("c", 3)
	m.Put("a", 1)
	m.Put("b", 2)

	keys := m.KeysToSlice()
	expected := []string{"c", "a", "b"}
	if !slices.Equal(keys, expected) {
		t.Fatalf("expected %v, got %v", expected, keys)
	}

	vals := m.ValuesToSlice()
	expectedVals := []int{3, 1, 2}
	if !slices.Equal(vals, expectedVals) {
		t.Fatalf("expected %v, got %v", expectedVals, vals)
	}
}

func TestLinkedHashMapOverwrite(t *testing.T) {
	m := NewLinkedHashMap[string, int]()
	m.Put("a", 1)
	m.Put("b", 2)
	m.Put("c", 3)
	old, existed := m.Put("b", 20)
	if !existed || old != 2 {
		t.Fatalf("expected old=2, got %v", old)
	}
	// Order should be preserved (b stays in position)
	keys := m.KeysToSlice()
	expected := []string{"a", "b", "c"}
	if !slices.Equal(keys, expected) {
		t.Fatalf("expected %v, got %v", expected, keys)
	}
	if v, _ := m.Get("b"); v != 20 {
		t.Fatalf("expected 20, got %v", v)
	}
}

func TestLinkedHashMapRemove(t *testing.T) {
	m := NewLinkedHashMap[string, int]()
	m.Put("a", 1)
	m.Put("b", 2)
	m.Put("c", 3)

	v, ok := m.Remove("b")
	if !ok || v != 2 {
		t.Fatalf("expected removed 2, got %v", v)
	}
	keys := m.KeysToSlice()
	expected := []string{"a", "c"}
	if !slices.Equal(keys, expected) {
		t.Fatalf("expected %v, got %v", expected, keys)
	}

	// Remove head
	m.Remove("a")
	keys = m.KeysToSlice()
	if !slices.Equal(keys, []string{"c"}) {
		t.Fatalf("expected [c], got %v", keys)
	}

	// Remove tail (last element)
	m.Remove("c")
	if !m.IsEmpty() {
		t.Fatal("expected empty")
	}
}

func TestLinkedHashMapClear(t *testing.T) {
	m := NewLinkedHashMap[int, int]()
	m.Put(1, 10)
	m.Put(2, 20)
	m.Clear()
	if !m.IsEmpty() {
		t.Fatal("expected empty after clear")
	}
}

func TestLinkedHashMapIteration(t *testing.T) {
	m := NewLinkedHashMap[int, int]()
	m.Put(3, 30)
	m.Put(1, 10)
	m.Put(2, 20)

	var keys []int
	var vals []int
	for k, v := range m.All() {
		keys = append(keys, k)
		vals = append(vals, v)
	}
	if !slices.Equal(keys, []int{3, 1, 2}) {
		t.Fatalf("expected [3,1,2], got %v", keys)
	}
	if !slices.Equal(vals, []int{30, 10, 20}) {
		t.Fatalf("expected [30,10,20], got %v", vals)
	}
}

func TestLinkedHashMapFunctional(t *testing.T) {
	m := NewLinkedHashMap[string, int]()
	m.Put("x", 1)
	m.Put("y", 2)
	m.Put("z", 3)

	if !m.AnySatisfy(func(_ string, v int) bool { return v > 2 }) {
		t.Fatal("expected any > 2")
	}
	if !m.AllSatisfy(func(_ string, v int) bool { return v > 0 }) {
		t.Fatal("expected all > 0")
	}
	if !m.NoneSatisfy(func(_ string, v int) bool { return v > 10 }) {
		t.Fatal("expected none > 10")
	}
	if c := m.Count(func(_ string, v int) bool { return v%2 == 0 }); c != 1 {
		t.Fatalf("expected count 1, got %d", c)
	}
}

func TestLinkedHashMapSelectReject(t *testing.T) {
	m := NewLinkedHashMap[int, int]()
	m.Put(1, 10)
	m.Put(2, 20)
	m.Put(3, 30)

	big := m.Select(func(_ int, v int) bool { return v > 15 })
	if big.Size() != 2 {
		t.Fatalf("expected 2, got %d", big.Size())
	}
	// Verify order preserved
	if !slices.Equal(big.KeysToSlice(), []int{2, 3}) {
		t.Fatalf("expected [2,3], got %v", big.KeysToSlice())
	}

	small := m.Reject(func(_ int, v int) bool { return v > 15 })
	if small.Size() != 1 {
		t.Fatalf("expected 1, got %d", small.Size())
	}
}

func TestLinkedHashMapDetect(t *testing.T) {
	m := NewLinkedHashMap[string, int]()
	m.Put("a", 1)
	m.Put("b", 2)
	m.Put("c", 3)

	k, v, found := m.Detect(func(_ string, val int) bool { return val > 1 })
	if !found || k != "b" || v != 2 {
		t.Fatalf("expected b:2, got %v:%v", k, v)
	}
}

func TestLinkedHashMapString(t *testing.T) {
	m := NewLinkedHashMap[string, int]()
	m.Put("a", 1)
	m.Put("b", 2)
	s := m.String()
	if s != "{a: 1, b: 2}" {
		t.Fatalf("unexpected string: %s", s)
	}
}
