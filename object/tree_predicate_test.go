// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package object

import "testing"

// These cover the predicate-query methods added to the comparator-backed tree
// types so they satisfy the (now T-any) Searchable/MapIterable hierarchy (11 §4).

func TestTreeSetPredicateQueries(t *testing.T) {
	s := NewTreeSet[int](NaturalComparator[int]())
	for _, v := range []int{2, 4, 6, 8} {
		s.Add(v)
	}
	even := func(x int) bool { return x%2 == 0 }
	big := func(x int) bool { return x > 5 }

	if !s.AnySatisfy(big) {
		t.Error("AnySatisfy(>5) = false, want true (6,8 present)")
	}
	if s.AnySatisfy(func(x int) bool { return x > 100 }) {
		t.Error("AnySatisfy(>100) = true, want false")
	}
	if !s.AllSatisfy(even) {
		t.Error("AllSatisfy(even) = false, want true")
	}
	if s.AllSatisfy(big) {
		t.Error("AllSatisfy(>5) = true, want false (2,4 present)")
	}
	if !s.NoneSatisfy(func(x int) bool { return x%2 == 1 }) {
		t.Error("NoneSatisfy(odd) = false, want true")
	}
	if s.NoneSatisfy(big) {
		t.Error("NoneSatisfy(>5) = true, want false")
	}

	// Empty set: All/None vacuously true, Any false.
	empty := NewTreeSet[int](NaturalComparator[int]())
	if empty.AnySatisfy(even) {
		t.Error("empty AnySatisfy = true, want false")
	}
	if !empty.AllSatisfy(even) || !empty.NoneSatisfy(even) {
		t.Error("empty AllSatisfy/NoneSatisfy = false, want true (vacuous)")
	}

	// Usable through the interface the relaxation unlocked.
	var set MutableSet[int] = s
	if !set.AnySatisfy(big) {
		t.Error("via MutableSet interface: AnySatisfy(>5) = false, want true")
	}
}

func TestTreeMapPredicateQueries(t *testing.T) {
	m := NewTreeMap[int, string](NaturalComparator[int]())
	m.Put(1, "a")
	m.Put(2, "b")
	m.Put(3, "c")

	if !m.AnySatisfy(func(k int, _ string) bool { return k == 2 }) {
		t.Error("AnySatisfy(k==2) = false, want true")
	}
	if !m.AllSatisfy(func(k int, _ string) bool { return k > 0 }) {
		t.Error("AllSatisfy(k>0) = false, want true")
	}
	if m.AllSatisfy(func(k int, _ string) bool { return k > 1 }) {
		t.Error("AllSatisfy(k>1) = true, want false (key 1 present)")
	}
	if !m.NoneSatisfy(func(_ int, v string) bool { return v == "z" }) {
		t.Error("NoneSatisfy(v==z) = false, want true")
	}

	empty := NewTreeMap[int, string](NaturalComparator[int]())
	if empty.AnySatisfy(func(int, string) bool { return true }) {
		t.Error("empty AnySatisfy = true, want false")
	}
	if !empty.AllSatisfy(func(int, string) bool { return false }) {
		t.Error("empty AllSatisfy = false, want true (vacuous)")
	}

	// Usable through the map interface the relaxation unlocked.
	var mi MapIterable[int, string] = m
	if mi.Len() != 3 {
		t.Errorf("via MapIterable interface: Len = %d, want 3", mi.Len())
	}
}
