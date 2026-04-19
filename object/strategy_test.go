// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package object

import (
	"slices"
	"strings"
	"testing"
)

// ── HashingStrategy tests ─────────────────────────────────────────────

func TestCaseInsensitiveHashSet(t *testing.T) {
	s := NewHashSetWithStrategy(CaseInsensitiveHashingStrategy())
	s.Add("Hello")
	s.Add("hello") // should be duplicate
	s.Add("HELLO") // should be duplicate
	if s.Size() != 1 {
		t.Fatalf("expected 1, got %d", s.Size())
	}
	if !s.Contains("hElLo") {
		t.Fatal("expected case-insensitive contains")
	}
	s.Remove("HELLO")
	if s.Size() != 0 {
		t.Fatal("expected empty after remove")
	}
}

func TestCaseInsensitiveHashMap(t *testing.T) {
	m := NewHashMapWithStrategy[string, int](CaseInsensitiveHashingStrategy())
	m.Put("Content-Type", 1)
	m.Put("content-type", 2) // should overwrite
	if m.Size() != 1 {
		t.Fatalf("expected 1, got %d", m.Size())
	}
	v, ok := m.Get("CONTENT-TYPE")
	if !ok || v != 2 {
		t.Fatalf("expected 2, got %v", v)
	}
}

type Person struct {
	Name string
	Age  int
	City string
}

func TestByFieldHashSet(t *testing.T) {
	strategy := ByFieldString(func(p Person) string { return p.Name })
	s := NewHashSetWithStrategy(strategy)
	s.Add(Person{"Alice", 30, "NYC"})
	s.Add(Person{"Alice", 25, "LA"}) // same name → duplicate
	s.Add(Person{"Bob", 30, "NYC"})

	if s.Size() != 2 {
		t.Fatalf("expected 2 (by name), got %d", s.Size())
	}
	if !s.Contains(Person{"Alice", 99, "Mars"}) {
		t.Fatal("should find Alice by name regardless of other fields")
	}
}

func TestByFieldHashMap(t *testing.T) {
	strategy := ByFieldString(func(p Person) string { return p.Name })
	m := NewHashMapWithStrategy[Person, string](strategy)
	m.Put(Person{"Alice", 30, "NYC"}, "first")
	m.Put(Person{"Alice", 25, "LA"}, "second") // overwrites by name
	if m.Size() != 1 {
		t.Fatalf("expected 1, got %d", m.Size())
	}
	v, _ := m.Get(Person{"Alice", 0, ""})
	if v != "second" {
		t.Fatalf("expected 'second', got %v", v)
	}
}

func TestHashSetWithStrategyOperations(t *testing.T) {
	s := NewHashSetWithStrategy(StringHashingStrategy())
	s.Add("a")
	s.Add("b")
	s.Add("c")

	sel := s.Select(func(v string) bool { return v != "b" })
	if sel.Size() != 2 {
		t.Fatalf("select: expected 2, got %d", sel.Size())
	}

	rej := s.Reject(func(v string) bool { return v == "a" })
	if rej.Size() != 2 {
		t.Fatalf("reject: expected 2, got %d", rej.Size())
	}
}

func TestHashSetWithStrategyResize(t *testing.T) {
	s := NewHashSetWithStrategy(StringHashingStrategy())
	for i := 0; i < 1000; i++ {
		s.Add(string(rune('a'+i%26)) + string(rune('0'+i/26)))
	}
	if s.Size() < 100 { // many collisions but should still work
		t.Fatalf("expected many elements, got %d", s.Size())
	}
}

func TestHashMapWithStrategyGetOrDefault(t *testing.T) {
	m := NewHashMapWithStrategy[string, int](StringHashingStrategy())
	m.Put("a", 1)
	if v := m.GetOrDefault("a", 99); v != 1 {
		t.Fatalf("expected 1, got %d", v)
	}
	if v := m.GetOrDefault("missing", 99); v != 99 {
		t.Fatalf("expected 99, got %d", v)
	}
}

// ── TreeMap / TreeSet tests ───────────────────────────────────────────

func TestTreeMapBasic(t *testing.T) {
	m := NewTreeMap[string, int](NaturalComparator[string]())
	m.Put("banana", 2)
	m.Put("apple", 1)
	m.Put("cherry", 3)

	if m.Size() != 3 {
		t.Fatalf("expected 3, got %d", m.Size())
	}
	v, ok := m.Get("apple")
	if !ok || v != 1 {
		t.Fatalf("expected 1, got %v", v)
	}

	// Iteration should be sorted
	keys := make([]string, 0, 3)
	m.ForEach(func(k string, v int) { keys = append(keys, k) })
	if !slices.Equal(keys, []string{"apple", "banana", "cherry"}) {
		t.Fatalf("expected sorted order, got %v", keys)
	}
}

func TestTreeMapOverwrite(t *testing.T) {
	m := NewTreeMap[int, string](NaturalComparator[int]())
	m.Put(1, "one")
	old, existed := m.Put(1, "ONE")
	if !existed || old != "one" {
		t.Fatalf("expected overwrite, got %v/%v", old, existed)
	}
	if m.Size() != 1 {
		t.Fatalf("expected 1, got %d", m.Size())
	}
}

func TestTreeMapRemove(t *testing.T) {
	m := NewTreeMap[int, int](NaturalComparator[int]())
	for i := 0; i < 100; i++ {
		m.Put(i, i*10)
	}
	for i := 0; i < 100; i += 2 {
		m.Remove(i)
	}
	if m.Size() != 50 {
		t.Fatalf("expected 50, got %d", m.Size())
	}
	// Check remaining are odd
	m.ForEach(func(k, v int) {
		if k%2 == 0 {
			t.Fatalf("even key %d should have been removed", k)
		}
	})
}

func TestTreeMapMinMax(t *testing.T) {
	m := NewTreeMap[int, string](NaturalComparator[int]())
	m.Put(5, "five")
	m.Put(1, "one")
	m.Put(9, "nine")

	k, _, ok := m.Min()
	if !ok || k != 1 {
		t.Fatalf("min: expected 1, got %v", k)
	}
	k, _, ok = m.Max()
	if !ok || k != 9 {
		t.Fatalf("max: expected 9, got %v", k)
	}
}

func TestTreeMapReverseComparator(t *testing.T) {
	m := NewTreeMap[int, int](ReverseComparator[int]())
	m.Put(1, 10)
	m.Put(3, 30)
	m.Put(2, 20)

	keys := make([]int, 0, 3)
	m.ForEach(func(k, v int) { keys = append(keys, k) })
	if !slices.Equal(keys, []int{3, 2, 1}) {
		t.Fatalf("expected reverse order [3,2,1], got %v", keys)
	}
}

func TestTreeMapByFieldComparator(t *testing.T) {
	m := NewTreeMap[Person, string](ComparatorByField(func(p Person) string { return p.Name }))
	m.Put(Person{"Charlie", 30, "NYC"}, "c")
	m.Put(Person{"Alice", 25, "LA"}, "a")
	m.Put(Person{"Bob", 35, "SF"}, "b")

	keys := make([]string, 0, 3)
	m.ForEach(func(k Person, v string) { keys = append(keys, k.Name) })
	if !slices.Equal(keys, []string{"Alice", "Bob", "Charlie"}) {
		t.Fatalf("expected alphabetical by name, got %v", keys)
	}
}

func TestTreeSetBasic(t *testing.T) {
	s := NewTreeSet[int](NaturalComparator[int]())
	s.Add(3)
	s.Add(1)
	s.Add(2)
	s.Add(1) // duplicate

	if s.Size() != 3 {
		t.Fatalf("expected 3, got %d", s.Size())
	}

	got := s.ToSlice()
	if !slices.Equal(got, []int{1, 2, 3}) {
		t.Fatalf("expected sorted [1,2,3], got %v", got)
	}
}

func TestTreeSetMinMax(t *testing.T) {
	s := NewTreeSet[string](NaturalComparator[string]())
	s.Add("banana")
	s.Add("apple")
	s.Add("cherry")

	min, ok := s.Min()
	if !ok || min != "apple" {
		t.Fatalf("min: expected apple, got %v", min)
	}
	max, ok := s.Max()
	if !ok || max != "cherry" {
		t.Fatalf("max: expected cherry, got %v", max)
	}
}

func TestTreeSetRemove(t *testing.T) {
	s := NewTreeSet[int](NaturalComparator[int]())
	for i := 0; i < 50; i++ {
		s.Add(i)
	}
	for i := 0; i < 50; i += 2 {
		s.Remove(i)
	}
	if s.Size() != 25 {
		t.Fatalf("expected 25, got %d", s.Size())
	}
	if s.Contains(0) || s.Contains(2) {
		t.Fatal("even values should be removed")
	}
	if !s.Contains(1) || !s.Contains(3) {
		t.Fatal("odd values should remain")
	}
}

func TestTreeSetSelectReject(t *testing.T) {
	s := NewTreeSet[int](NaturalComparator[int]())
	for i := 1; i <= 5; i++ {
		s.Add(i)
	}
	evens := s.Select(func(v int) bool { return v%2 == 0 })
	if !slices.Equal(evens.ToSlice(), []int{2, 4}) {
		t.Fatalf("expected [2,4], got %v", evens.ToSlice())
	}
}

func TestReversed(t *testing.T) {
	// Reverse an arbitrary comparator (not just natural ordering).
	byAge := ComparatorByField(func(p Person) int { return p.Age })
	byAgeDesc := Reversed(byAge)

	s := NewTreeSet[Person](byAgeDesc)
	s.Add(Person{"A", 20, ""})
	s.Add(Person{"B", 30, ""})
	s.Add(Person{"C", 10, ""})

	ages := make([]int, 0)
	s.ForEach(func(p Person) { ages = append(ages, p.Age) })
	if !slices.Equal(ages, []int{30, 20, 10}) {
		t.Fatalf("expected descending ages [30,20,10], got %v", ages)
	}
}

func TestComparatorByFieldWith(t *testing.T) {
	// Sort persons by name case-insensitively.
	ciStr := Comparator[string](func(a, b string) int {
		la, lb := strings.ToLower(a), strings.ToLower(b)
		if la < lb {
			return -1
		} else if la > lb {
			return 1
		}
		return 0
	})
	byNameCI := ComparatorByFieldWith(func(p Person) string { return p.Name }, ciStr)

	s := NewTreeSet[Person](byNameCI)
	s.Add(Person{"bob", 0, ""})
	s.Add(Person{"Alice", 0, ""})
	s.Add(Person{"CAROL", 0, ""})

	names := make([]string, 0)
	s.ForEach(func(p Person) { names = append(names, p.Name) })
	if !slices.Equal(names, []string{"Alice", "bob", "CAROL"}) {
		t.Fatalf("expected case-insensitive alphabetical, got %v", names)
	}
}

func TestThenComparing(t *testing.T) {
	byAge := ComparatorByField(func(p Person) int { return p.Age })
	byName := ComparatorByField(func(p Person) string { return p.Name })
	cmp := ThenComparing(byAge, byName)

	s := NewTreeSet[Person](cmp)
	s.Add(Person{"Charlie", 30, ""})
	s.Add(Person{"Alice", 30, ""})
	s.Add(Person{"Bob", 25, ""})

	names := make([]string, 0, 3)
	s.ForEach(func(p Person) { names = append(names, p.Name) })
	// Bob(25) < Alice(30) < Charlie(30) — age first, then name
	if !slices.Equal(names, []string{"Bob", "Alice", "Charlie"}) {
		t.Fatalf("expected [Bob, Alice, Charlie], got %v", names)
	}
}

func TestTreeMapClear(t *testing.T) {
	m := NewTreeMap[int, int](NaturalComparator[int]())
	m.Put(1, 1)
	m.Put(2, 2)
	m.Clear()
	if !m.IsEmpty() {
		t.Fatal("expected empty after clear")
	}
}

func TestTreeSetStress(t *testing.T) {
	s := NewTreeSet[int](NaturalComparator[int]())
	for i := 999; i >= 0; i-- {
		s.Add(i)
	}
	if s.Size() != 1000 {
		t.Fatalf("expected 1000, got %d", s.Size())
	}
	// Verify sorted
	prev := -1
	for v := range s.All() {
		if v <= prev {
			t.Fatalf("not sorted: %d after %d", v, prev)
		}
		prev = v
	}
	// Remove all
	for i := 0; i < 1000; i++ {
		s.Remove(i)
	}
	if !s.IsEmpty() {
		t.Fatalf("expected empty, got %d", s.Size())
	}
}
