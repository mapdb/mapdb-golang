// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.

package object

import (
	"slices"
	"testing"
)

func TestTreeSet_Navigation(t *testing.T) {
	s := NewTreeSet[int](NaturalComparator[int]())
	for _, v := range []int{10, 20, 30} {
		s.Add(v)
	}
	if v, ok := s.Floor(25); !ok || v != 20 {
		t.Errorf("Floor(25) = (%d,%v), want 20", v, ok)
	}
	if v, ok := s.Ceiling(25); !ok || v != 30 {
		t.Errorf("Ceiling(25) = (%d,%v), want 30", v, ok)
	}
	if v, ok := s.Floor(10); !ok || v != 10 {
		t.Errorf("Floor(10) = (%d,%v), want 10", v, ok)
	}
	if _, ok := s.Lower(10); ok {
		t.Error("Lower(10) should be false")
	}
	if _, ok := s.Higher(30); ok {
		t.Error("Higher(30) should be false")
	}
	if v, ok := s.First(); !ok || v != 10 {
		t.Errorf("First = (%d,%v), want 10", v, ok)
	}
	if v, ok := s.Last(); !ok || v != 30 {
		t.Errorf("Last = (%d,%v), want 30", v, ok)
	}
}

func TestTreeSet_PollEmptyAndSingle(t *testing.T) {
	s := NewTreeSet[int](NaturalComparator[int]())
	if _, ok := s.PollFirst(); ok {
		t.Error("PollFirst on empty should be false")
	}
	if _, ok := s.PollLast(); ok {
		t.Error("PollLast on empty should be false")
	}
	s.Add(7)
	if v, ok := s.PollFirst(); !ok || v != 7 {
		t.Errorf("PollFirst = (%d,%v), want 7", v, ok)
	}
	if _, ok := s.PollFirst(); ok {
		t.Error("PollFirst on now-empty should be false")
	}
}

func TestTreeSet_SubSetCopyPreservesReverseComparator(t *testing.T) {
	s := NewTreeSet[int](ReverseComparator[int]())
	for _, v := range []int{10, 20, 30, 40, 50} {
		s.Add(v)
	}
	if got := s.ToSlice(); !slices.Equal(got, []int{50, 40, 30, 20, 10}) {
		t.Fatalf("source order = %v, want descending", got)
	}
	sub := s.SubSetCopy(40, 10) // {40,30,20} in reverse order
	if got := sub.ToSlice(); !slices.Equal(got, []int{40, 30, 20}) {
		t.Errorf("SubSetCopy = %v, want [40 30 20] (reverse order preserved)", got)
	}
	// Independence both directions.
	sub.Remove(30)
	if !s.Contains(30) {
		t.Error("original must still contain 30 after mutating snapshot")
	}
	s.Remove(40)
	if !sub.Contains(40) {
		t.Error("snapshot must still contain 40 after mutating original")
	}
}
