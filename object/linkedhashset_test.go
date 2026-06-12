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

func TestLinkedHashSetBasic(t *testing.T) {
	s := NewLinkedHashSet[int]()
	if s.Len() != 0 {
		t.Fatal("expected empty")
	}
	if !s.Add(1) {
		t.Fatal("expected true on first add")
	}
	if !s.Add(2) {
		t.Fatal("expected true")
	}
	if s.Add(1) {
		t.Fatal("expected false on duplicate")
	}
	if s.Len() != 2 {
		t.Fatalf("expected 2, got %d", s.Len())
	}
	if !s.Contains(1) {
		t.Fatal("expected contains 1")
	}
	if s.Contains(99) {
		t.Fatal("expected not contains 99")
	}
}

func TestLinkedHashSetInsertionOrder(t *testing.T) {
	s := NewLinkedHashSetFrom(3, 1, 4, 1, 5, 9) // 1 is duplicate
	expected := []int{3, 1, 4, 5, 9}
	got := s.ToSlice()
	if !slices.Equal(got, expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestLinkedHashSetRemove(t *testing.T) {
	s := NewLinkedHashSetFrom(1, 2, 3, 4)
	if !s.Remove(2) {
		t.Fatal("expected true")
	}
	if s.Remove(2) {
		t.Fatal("expected false on second remove")
	}
	expected := []int{1, 3, 4}
	if !slices.Equal(s.ToSlice(), expected) {
		t.Fatalf("expected %v, got %v", expected, s.ToSlice())
	}

	// Remove head
	s.Remove(1)
	if !slices.Equal(s.ToSlice(), []int{3, 4}) {
		t.Fatalf("expected [3,4], got %v", s.ToSlice())
	}

	// Remove tail
	s.Remove(4)
	if !slices.Equal(s.ToSlice(), []int{3}) {
		t.Fatalf("expected [3], got %v", s.ToSlice())
	}

	// Remove last element
	s.Remove(3)
	if s.Len() != 0 {
		t.Fatal("expected empty")
	}
}

func TestLinkedHashSetClear(t *testing.T) {
	s := NewLinkedHashSetFrom(1, 2, 3)
	s.Clear()
	if s.Len() != 0 {
		t.Fatal("expected empty after clear")
	}
}

func TestLinkedHashSetIteration(t *testing.T) {
	s := NewLinkedHashSetFrom(5, 3, 1, 4, 2)
	var got []int
	for v := range s.All() {
		got = append(got, v)
	}
	if !slices.Equal(got, []int{5, 3, 1, 4, 2}) {
		t.Fatalf("expected [5,3,1,4,2], got %v", got)
	}
}

func TestLinkedHashSetSetOperations(t *testing.T) {
	a := NewLinkedHashSetFrom(1, 2, 3)
	b := NewLinkedHashSetFrom(2, 3, 4)

	u := a.Union(b)
	if u.Len() != 4 {
		t.Fatalf("union expected 4, got %d", u.Len())
	}
	if !slices.Equal(u.ToSlice(), []int{1, 2, 3, 4}) {
		t.Fatalf("union order expected [1,2,3,4], got %v", u.ToSlice())
	}

	inter := a.Intersect(b)
	if inter.Len() != 2 {
		t.Fatalf("intersect expected 2, got %d", inter.Len())
	}

	diff := a.Difference(b)
	if diff.Len() != 1 {
		t.Fatalf("diff expected 1, got %d", diff.Len())
	}
	if !diff.Contains(1) {
		t.Fatal("diff should contain 1")
	}

	sym := a.SymmetricDifference(b)
	if sym.Len() != 2 {
		t.Fatalf("symdiff expected 2, got %d", sym.Len())
	}
}

func TestLinkedHashSetFunctional(t *testing.T) {
	s := NewLinkedHashSetFrom(1, 2, 3, 4, 5)
	if !s.AnySatisfy(func(v int) bool { return v > 4 }) {
		t.Fatal("expected any > 4")
	}
	if !s.AllSatisfy(func(v int) bool { return v > 0 }) {
		t.Fatal("expected all > 0")
	}
	if !s.NoneSatisfy(func(v int) bool { return v > 10 }) {
		t.Fatal("expected none > 10")
	}
	if c := s.Count(func(v int) bool { return v%2 == 0 }); c != 2 {
		t.Fatalf("expected count 2, got %d", c)
	}
}

func TestLinkedHashSetSelectReject(t *testing.T) {
	s := NewLinkedHashSetFrom(1, 2, 3, 4, 5)
	evens := s.Select(func(v int) bool { return v%2 == 0 })
	if !slices.Equal(evens.ToSlice(), []int{2, 4}) {
		t.Fatalf("expected [2,4], got %v", evens.ToSlice())
	}
	odds := s.Reject(func(v int) bool { return v%2 == 0 })
	if !slices.Equal(odds.ToSlice(), []int{1, 3, 5}) {
		t.Fatalf("expected [1,3,5], got %v", odds.ToSlice())
	}
}

func TestLinkedHashSetDetect(t *testing.T) {
	s := NewLinkedHashSetFrom(1, 2, 3)
	v, found := s.Detect(func(val int) bool { return val > 1 })
	if !found || v != 2 {
		t.Fatalf("expected 2, got %v", v)
	}
}

func TestLinkedHashSetString(t *testing.T) {
	s := NewLinkedHashSetFrom(1, 2, 3)
	str := s.String()
	if str != "{1, 2, 3}" {
		t.Fatalf("unexpected string: %s", str)
	}
}
