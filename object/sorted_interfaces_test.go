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

// Exercise the tree types THROUGH the SortedSet / NavigableMap interfaces added
// in the hierarchy-v2 work — navigation and range views must be reachable and
// correct via the interface, not only the concrete type.

func TestSortedSetInterface(t *testing.T) {
	ts := NewTreeSet[int](NaturalComparator[int]())
	for _, v := range []int{10, 20, 30, 40} {
		ts.Add(v)
	}
	var s SortedSet[int] = ts

	if v, ok := s.Min(); !ok || v != 10 {
		t.Errorf("Min = %d,%v want 10,true", v, ok)
	}
	if v, ok := s.Max(); !ok || v != 40 {
		t.Errorf("Max = %d,%v want 40,true", v, ok)
	}
	if v, ok := s.Floor(25); !ok || v != 20 {
		t.Errorf("Floor(25) = %d,%v want 20,true", v, ok)
	}
	if v, ok := s.Ceiling(25); !ok || v != 30 {
		t.Errorf("Ceiling(25) = %d,%v want 30,true", v, ok)
	}
	if v, ok := s.Lower(20); !ok || v != 10 {
		t.Errorf("Lower(20) = %d,%v want 10,true", v, ok)
	}
	if v, ok := s.Higher(20); !ok || v != 30 {
		t.Errorf("Higher(20) = %d,%v want 30,true", v, ok)
	}
	if r := s.Rank(30); r != 2 {
		t.Errorf("Rank(30) = %d, want 2", r)
	}
	if v, ok := s.Select(2); !ok || v != 30 {
		t.Errorf("Select(2) = %d,%v want 30,true", v, ok)
	}
}

func TestNavigableMapInterface(t *testing.T) {
	tm := NewTreeMap[int, string](NaturalComparator[int]())
	for _, k := range []int{1, 2, 3, 4, 5} {
		tm.Put(k, string(rune('a'+k-1)))
	}
	var m NavigableMap[int, string] = tm

	// Endpoints (SortedMap surface).
	if k, v, ok := m.Min(); !ok || k != 1 || v != "a" {
		t.Errorf("Min = %d,%q,%v want 1,a,true", k, v, ok)
	}
	if k, _, ok := m.Max(); !ok || k != 5 {
		t.Errorf("Max key = %d,%v want 5,true", k, ok)
	}
	// Point navigation.
	if k, _, ok := m.Floor(3); !ok || k != 3 {
		t.Errorf("Floor(3) key = %d,%v want 3,true", k, ok)
	}
	if k, _, ok := m.Higher(3); !ok || k != 4 {
		t.Errorf("Higher(3) key = %d,%v want 4,true", k, ok)
	}
	// Order statistics.
	if r := m.Rank(4); r != 3 {
		t.Errorf("Rank(4) = %d, want 3", r)
	}
	if k, _, ok := m.SelectEntry(0); !ok || k != 1 {
		t.Errorf("SelectEntry(0) key = %d,%v want 1,true", k, ok)
	}
	// Range views (half-open, ascending).
	var sub []int
	for k := range m.SubMap(2, 4) {
		sub = append(sub, k)
	}
	if !slices.Equal(sub, []int{2, 3}) {
		t.Errorf("SubMap(2,4) keys = %v, want [2 3]", sub)
	}
	// Descending keys.
	var desc []int
	for k := range m.DescendingKeys() {
		desc = append(desc, k)
	}
	if !slices.Equal(desc, []int{5, 4, 3, 2, 1}) {
		t.Errorf("DescendingKeys = %v, want [5 4 3 2 1]", desc)
	}
}
