// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package object

import (
	"testing"
)

func TestHashBag_NewEmpty(t *testing.T) {
	b := NewHashBag[int]()
	if b.Size() != 0 {
		t.Errorf("Size() = %d, want 0", b.Size())
	}
	if !b.IsEmpty() {
		t.Error("IsEmpty() = false, want true")
	}
	if b.SizeDistinct() != 0 {
		t.Errorf("SizeDistinct() = %d, want 0", b.SizeDistinct())
	}
}

func TestHashBag_NewHashBagFrom(t *testing.T) {
	b := NewHashBagFrom("a", "b", "a", "c", "a")
	if b.Size() != 5 {
		t.Errorf("Size() = %d, want 5", b.Size())
	}
	if b.SizeDistinct() != 3 {
		t.Errorf("SizeDistinct() = %d, want 3", b.SizeDistinct())
	}
	if b.OccurrencesOf("a") != 3 {
		t.Errorf("OccurrencesOf(a) = %d, want 3", b.OccurrencesOf("a"))
	}
}

func TestHashBag_Add(t *testing.T) {
	b := NewHashBag[int]()
	b.Add(5)
	b.Add(5)
	b.Add(10)
	if b.Size() != 3 {
		t.Errorf("Size() = %d, want 3", b.Size())
	}
	if b.OccurrencesOf(5) != 2 {
		t.Errorf("OccurrencesOf(5) = %d, want 2", b.OccurrencesOf(5))
	}
	if b.OccurrencesOf(10) != 1 {
		t.Errorf("OccurrencesOf(10) = %d, want 1", b.OccurrencesOf(10))
	}
}

func TestHashBag_AddOccurrences(t *testing.T) {
	b := NewHashBag[string]()
	b.AddOccurrences("x", 5)
	if b.OccurrencesOf("x") != 5 {
		t.Errorf("OccurrencesOf(x) = %d, want 5", b.OccurrencesOf("x"))
	}
	if b.Size() != 5 {
		t.Errorf("Size() = %d, want 5", b.Size())
	}

	// zero/negative occurrences should be ignored
	b.AddOccurrences("y", 0)
	b.AddOccurrences("z", -1)
	if b.SizeDistinct() != 1 {
		t.Errorf("SizeDistinct() = %d, want 1 (zero/neg ignored)", b.SizeDistinct())
	}
}

func TestHashBag_OccurrencesOf(t *testing.T) {
	b := NewHashBagFrom(1, 2, 2, 3, 3, 3)
	if b.OccurrencesOf(1) != 1 {
		t.Errorf("OccurrencesOf(1) = %d, want 1", b.OccurrencesOf(1))
	}
	if b.OccurrencesOf(3) != 3 {
		t.Errorf("OccurrencesOf(3) = %d, want 3", b.OccurrencesOf(3))
	}
	if b.OccurrencesOf(99) != 0 {
		t.Errorf("OccurrencesOf(99) = %d, want 0", b.OccurrencesOf(99))
	}
}

func TestHashBag_Remove(t *testing.T) {
	b := NewHashBagFrom(1, 1, 1, 2)

	t.Run("removes one occurrence", func(t *testing.T) {
		ok := b.Remove(1)
		if !ok {
			t.Error("Remove(1) = false, want true")
		}
		if b.OccurrencesOf(1) != 2 {
			t.Errorf("OccurrencesOf(1) after remove = %d, want 2", b.OccurrencesOf(1))
		}
		if b.Size() != 3 {
			t.Errorf("Size after remove = %d, want 3", b.Size())
		}
	})

	t.Run("removes last occurrence cleans up", func(t *testing.T) {
		b.Remove(2)
		if b.Contains(2) {
			t.Error("Contains(2) = true after removing only occurrence")
		}
		if b.SizeDistinct() != 1 {
			t.Errorf("SizeDistinct() = %d, want 1", b.SizeDistinct())
		}
	})

	t.Run("absent element", func(t *testing.T) {
		ok := b.Remove(99)
		if ok {
			t.Error("Remove(99) = true, want false")
		}
	})
}

func TestHashBag_Contains(t *testing.T) {
	b := NewHashBagFrom(1, 2, 3)
	if !b.Contains(2) {
		t.Error("Contains(2) = false")
	}
	if b.Contains(99) {
		t.Error("Contains(99) = true")
	}
}

func TestHashBag_ForEachWithOccurrences(t *testing.T) {
	b := NewHashBagFrom("a", "a", "b")
	seen := make(map[string]int)
	b.ForEachWithOccurrences(func(v string, count int) {
		seen[v] = count
	})
	if seen["a"] != 2 {
		t.Errorf("ForEachWithOccurrences: a = %d, want 2", seen["a"])
	}
	if seen["b"] != 1 {
		t.Errorf("ForEachWithOccurrences: b = %d, want 1", seen["b"])
	}
}

func TestHashBag_TopOccurrences(t *testing.T) {
	b := NewHashBagFrom(1, 1, 1, 2, 2, 3)
	top := b.TopOccurrences(2)
	if len(top) != 2 {
		t.Fatalf("TopOccurrences(2) len = %d, want 2", len(top))
	}
	// first should be the most frequent
	if top[0].Value != 1 || top[0].Count != 3 {
		t.Errorf("TopOccurrences[0] = {%d, %d}, want {1, 3}", top[0].Value, top[0].Count)
	}
	if top[1].Value != 2 || top[1].Count != 2 {
		t.Errorf("TopOccurrences[1] = {%d, %d}, want {2, 2}", top[1].Value, top[1].Count)
	}
}

func TestHashBag_BottomOccurrences(t *testing.T) {
	b := NewHashBagFrom(1, 1, 1, 2, 2, 3)
	bottom := b.BottomOccurrences(1)
	if len(bottom) != 1 {
		t.Fatalf("BottomOccurrences(1) len = %d, want 1", len(bottom))
	}
	if bottom[0].Value != 3 || bottom[0].Count != 1 {
		t.Errorf("BottomOccurrences[0] = {%d, %d}, want {3, 1}", bottom[0].Value, bottom[0].Count)
	}
}

func TestHashBag_TopOccurrences_MoreThanDistinct(t *testing.T) {
	b := NewHashBagFrom(1, 2)
	top := b.TopOccurrences(10)
	if len(top) != 2 {
		t.Errorf("TopOccurrences(10) len = %d, want 2 (capped)", len(top))
	}
}

func TestHashBag_Select(t *testing.T) {
	b := NewHashBagFrom(1, 1, 2, 2, 2, 3)
	evens := b.Select(func(v int) bool { return v%2 == 0 })
	if evens.OccurrencesOf(2) != 3 {
		t.Errorf("Select: OccurrencesOf(2) = %d, want 3", evens.OccurrencesOf(2))
	}
	if evens.SizeDistinct() != 1 {
		t.Errorf("Select: SizeDistinct = %d, want 1", evens.SizeDistinct())
	}
	if evens.Size() != 3 {
		t.Errorf("Select: Size = %d, want 3", evens.Size())
	}
}

func TestHashBag_Reject(t *testing.T) {
	b := NewHashBagFrom(1, 1, 2, 3, 3, 3)
	notOdd := b.Reject(func(v int) bool { return v%2 != 0 })
	if notOdd.SizeDistinct() != 1 {
		t.Errorf("Reject: SizeDistinct = %d, want 1", notOdd.SizeDistinct())
	}
	if notOdd.OccurrencesOf(2) != 1 {
		t.Errorf("Reject: OccurrencesOf(2) = %d, want 1", notOdd.OccurrencesOf(2))
	}
}

func TestHashBag_Clear(t *testing.T) {
	b := NewHashBagFrom(1, 2, 3)
	b.Clear()
	if b.Size() != 0 {
		t.Errorf("Size after Clear = %d, want 0", b.Size())
	}
	if b.SizeDistinct() != 0 {
		t.Errorf("SizeDistinct after Clear = %d, want 0", b.SizeDistinct())
	}
	if !b.IsEmpty() {
		t.Error("IsEmpty after Clear = false")
	}
}

func TestHashBag_All(t *testing.T) {
	b := NewHashBagFrom(1, 1, 2)
	count := 0
	for range b.All() {
		count++
	}
	// All yields each element once per occurrence
	if count != 3 {
		t.Errorf("All yielded %d elements, want 3", count)
	}
}

func TestHashBag_AnySatisfy(t *testing.T) {
	b := NewHashBagFrom(1, 2, 3)
	if !b.AnySatisfy(func(v int) bool { return v == 2 }) {
		t.Error("AnySatisfy(==2) = false")
	}
	if b.AnySatisfy(func(v int) bool { return v > 10 }) {
		t.Error("AnySatisfy(>10) = true")
	}
}

func TestHashBag_AllSatisfy(t *testing.T) {
	b := NewHashBagFrom(2, 4, 6)
	if !b.AllSatisfy(func(v int) bool { return v%2 == 0 }) {
		t.Error("AllSatisfy(even) = false")
	}
	b.Add(3)
	if b.AllSatisfy(func(v int) bool { return v%2 == 0 }) {
		t.Error("AllSatisfy(even) = true after adding 3")
	}
}

func TestHashBag_NoneSatisfy(t *testing.T) {
	b := NewHashBagFrom(1, 3, 5)
	if !b.NoneSatisfy(func(v int) bool { return v%2 == 0 }) {
		t.Error("NoneSatisfy(even) = false")
	}
	b.Add(2)
	if b.NoneSatisfy(func(v int) bool { return v%2 == 0 }) {
		t.Error("NoneSatisfy(even) = true after adding 2")
	}
}

func TestHashBag_ToSlice(t *testing.T) {
	b := NewHashBagFrom(1, 1, 2)
	sl := b.ToSlice()
	if len(sl) != 3 {
		t.Errorf("ToSlice len = %d, want 3", len(sl))
	}
	// count occurrences in slice
	counts := make(map[int]int)
	for _, v := range sl {
		counts[v]++
	}
	if counts[1] != 2 {
		t.Errorf("ToSlice: count of 1 = %d, want 2", counts[1])
	}
	if counts[2] != 1 {
		t.Errorf("ToSlice: count of 2 = %d, want 1", counts[2])
	}
}

func TestHashBag_String(t *testing.T) {
	b := NewHashBagFrom(42)
	s := b.String()
	// single element: {42x1} or similar
	if s != "{42\u00d71}" {
		t.Errorf("String() = %q, want %q", s, "{42\u00d71}")
	}
}
