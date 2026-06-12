// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package object

import (
	"testing"
)

func TestHashSet_NewEmpty(t *testing.T) {
	s := NewHashSet[int]()
	if s.Len() != 0 {
		t.Errorf("Size() = %d, want 0", s.Len())
	}
	if s.Len() != 0 {
		t.Error("IsEmpty() = false, want true")
	}
}

func TestHashSet_NewHashSetFrom(t *testing.T) {
	s := NewHashSetFrom(1, 2, 3, 2)
	if s.Len() != 3 {
		t.Errorf("Size() = %d, want 3 (duplicates ignored)", s.Len())
	}
}

func TestHashSet_Add(t *testing.T) {
	t.Run("returns true for new element", func(t *testing.T) {
		s := NewHashSet[int]()
		if !s.Add(1) {
			t.Error("Add(1) = false, want true")
		}
		if s.Len() != 1 {
			t.Errorf("Size() = %d, want 1", s.Len())
		}
	})

	t.Run("returns false for duplicate", func(t *testing.T) {
		s := NewHashSetFrom(1, 2)
		if s.Add(1) {
			t.Error("Add(1) = true, want false (duplicate)")
		}
		if s.Len() != 2 {
			t.Errorf("Size() = %d, want 2", s.Len())
		}
	})
}

func TestHashSet_Remove(t *testing.T) {
	t.Run("returns true when present", func(t *testing.T) {
		s := NewHashSetFrom(1, 2, 3)
		if !s.Remove(2) {
			t.Error("Remove(2) = false, want true")
		}
		if s.Len() != 2 {
			t.Errorf("Size() = %d, want 2", s.Len())
		}
		if s.Contains(2) {
			t.Error("Contains(2) after Remove = true")
		}
	})

	t.Run("returns false when absent", func(t *testing.T) {
		s := NewHashSetFrom(1, 2)
		if s.Remove(99) {
			t.Error("Remove(99) = true, want false")
		}
	})
}

func TestHashSet_Contains(t *testing.T) {
	s := NewHashSetFrom(10, 20, 30)
	if !s.Contains(20) {
		t.Error("Contains(20) = false")
	}
	if s.Contains(99) {
		t.Error("Contains(99) = true")
	}
}

func TestHashSet_Union(t *testing.T) {
	a := NewHashSetFrom(1, 2, 3)
	b := NewHashSetFrom(3, 4, 5)
	u := a.Union(b)
	if u.Len() != 5 {
		t.Errorf("Union size = %d, want 5", u.Len())
	}
	for _, v := range []int{1, 2, 3, 4, 5} {
		if !u.Contains(v) {
			t.Errorf("Union missing %d", v)
		}
	}
}

func TestHashSet_Intersect(t *testing.T) {
	a := NewHashSetFrom(1, 2, 3, 4)
	b := NewHashSetFrom(3, 4, 5, 6)
	inter := a.Intersect(b)
	if inter.Len() != 2 {
		t.Errorf("Intersect size = %d, want 2", inter.Len())
	}
	if !inter.Contains(3) || !inter.Contains(4) {
		t.Error("Intersect missing expected elements")
	}
}

func TestHashSet_Difference(t *testing.T) {
	a := NewHashSetFrom(1, 2, 3, 4)
	b := NewHashSetFrom(3, 4, 5)
	diff := a.Difference(b)
	if diff.Len() != 2 {
		t.Errorf("Difference size = %d, want 2", diff.Len())
	}
	if !diff.Contains(1) || !diff.Contains(2) {
		t.Error("Difference missing expected elements")
	}
}

func TestHashSet_SymmetricDifference(t *testing.T) {
	a := NewHashSetFrom(1, 2, 3)
	b := NewHashSetFrom(2, 3, 4)
	sd := a.SymmetricDifference(b)
	if sd.Len() != 2 {
		t.Errorf("SymmetricDifference size = %d, want 2", sd.Len())
	}
	if !sd.Contains(1) || !sd.Contains(4) {
		t.Error("SymmetricDifference missing expected elements")
	}
	if sd.Contains(2) || sd.Contains(3) {
		t.Error("SymmetricDifference should not contain common elements")
	}
}

func TestHashSet_Select(t *testing.T) {
	s := NewHashSetFrom(1, 2, 3, 4, 5)
	evens := s.Select(func(v int) bool { return v%2 == 0 })
	if evens.Len() != 2 {
		t.Errorf("Select size = %d, want 2", evens.Len())
	}
	if !evens.Contains(2) || !evens.Contains(4) {
		t.Error("Select missing expected even elements")
	}
}

func TestHashSet_Reject(t *testing.T) {
	s := NewHashSetFrom(1, 2, 3, 4, 5)
	odds := s.Reject(func(v int) bool { return v%2 == 0 })
	if odds.Len() != 3 {
		t.Errorf("Reject size = %d, want 3", odds.Len())
	}
}

func TestHashSet_Detect(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		s := NewHashSetFrom(10, 20, 30)
		v, ok := s.Detect(func(v int) bool { return v == 20 })
		if !ok {
			t.Fatal("Detect returned false")
		}
		if v != 20 {
			t.Errorf("Detect = %d, want 20", v)
		}
	})

	t.Run("not found", func(t *testing.T) {
		s := NewHashSetFrom(10, 20)
		_, ok := s.Detect(func(v int) bool { return v == 99 })
		if ok {
			t.Error("Detect returned true, want false")
		}
	})
}

func TestHashSet_Count(t *testing.T) {
	s := NewHashSetFrom(1, 2, 3, 4, 5, 6)
	n := s.Count(func(v int) bool { return v%2 == 0 })
	if n != 3 {
		t.Errorf("Count(even) = %d, want 3", n)
	}
}

func TestHashSet_ForEach(t *testing.T) {
	s := NewHashSetFrom(1, 2, 3)
	sum := 0
	s.ForEach(func(v int) { sum += v })
	if sum != 6 {
		t.Errorf("ForEach sum = %d, want 6", sum)
	}
}

func TestHashSet_All(t *testing.T) {
	s := NewHashSetFrom(10, 20, 30)
	collected := NewHashSet[int]()
	for v := range s.All() {
		collected.Add(v)
	}
	if collected.Len() != 3 {
		t.Errorf("All yielded %d elements, want 3", collected.Len())
	}
}

func TestHashSet_AnySatisfy(t *testing.T) {
	s := NewHashSetFrom(1, 2, 3)
	if !s.AnySatisfy(func(v int) bool { return v == 2 }) {
		t.Error("AnySatisfy(==2) = false, want true")
	}
	if s.AnySatisfy(func(v int) bool { return v > 10 }) {
		t.Error("AnySatisfy(>10) = true, want false")
	}
}

func TestHashSet_AllSatisfy(t *testing.T) {
	s := NewHashSetFrom(2, 4, 6)
	if !s.AllSatisfy(func(v int) bool { return v%2 == 0 }) {
		t.Error("AllSatisfy(even) = false, want true")
	}
	s.Add(3)
	if s.AllSatisfy(func(v int) bool { return v%2 == 0 }) {
		t.Error("AllSatisfy(even) = true after adding 3")
	}
}

func TestHashSet_NoneSatisfy(t *testing.T) {
	s := NewHashSetFrom(1, 3, 5)
	if !s.NoneSatisfy(func(v int) bool { return v%2 == 0 }) {
		t.Error("NoneSatisfy(even) = false, want true")
	}
	s.Add(2)
	if s.NoneSatisfy(func(v int) bool { return v%2 == 0 }) {
		t.Error("NoneSatisfy(even) = true after adding 2")
	}
}

func TestHashSet_Clear(t *testing.T) {
	s := NewHashSetFrom(1, 2, 3)
	s.Clear()
	if s.Len() != 0 {
		t.Errorf("Size after Clear = %d, want 0", s.Len())
	}
	if s.Len() != 0 {
		t.Error("IsEmpty after Clear = false")
	}
}

func TestHashSet_String(t *testing.T) {
	s := NewHashSetFrom(42)
	str := s.String()
	if str != "{42}" {
		t.Errorf("String() = %q, want %q", str, "{42}")
	}
}

func TestHashSet_ToSlice(t *testing.T) {
	s := NewHashSetFrom(1, 2, 3)
	sl := s.ToSlice()
	if len(sl) != 3 {
		t.Errorf("ToSlice len = %d, want 3", len(sl))
	}
	// all elements present (order undefined)
	seen := make(map[int]bool)
	for _, v := range sl {
		seen[v] = true
	}
	for _, v := range []int{1, 2, 3} {
		if !seen[v] {
			t.Errorf("ToSlice missing %d", v)
		}
	}
}

func TestHashSet_StringType(t *testing.T) {
	s := NewHashSet[string]()
	s.Add("hello")
	s.Add("world")
	s.Add("hello") // duplicate
	if s.Len() != 2 {
		t.Errorf("Size() = %d, want 2", s.Len())
	}
	if !s.Contains("hello") || !s.Contains("world") {
		t.Error("missing expected string elements")
	}

	diff := NewHashSetFrom("hello", "world", "foo")
	result := diff.Difference(s)
	if result.Len() != 1 || !result.Contains("foo") {
		t.Errorf("Difference unexpected result: %v", result)
	}
}
