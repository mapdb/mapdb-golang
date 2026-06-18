// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package object

import (
	"fmt"
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
	if s.Len() != 1 {
		t.Fatalf("expected 1, got %d", s.Len())
	}
	if !s.Contains("hElLo") {
		t.Fatal("expected case-insensitive contains")
	}
	s.Remove("HELLO")
	if s.Len() != 0 {
		t.Fatal("expected empty after remove")
	}
}

func TestCaseInsensitiveHashMap(t *testing.T) {
	m := NewHashMapWithStrategy[string, int](CaseInsensitiveHashingStrategy())
	m.Put("Content-Type", 1)
	m.Put("content-type", 2) // should overwrite
	if m.Len() != 1 {
		t.Fatalf("expected 1, got %d", m.Len())
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

	if s.Len() != 2 {
		t.Fatalf("expected 2 (by name), got %d", s.Len())
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
	if m.Len() != 1 {
		t.Fatalf("expected 1, got %d", m.Len())
	}
	v, _ := m.Get(Person{"Alice", 0, ""})
	if v != "second" {
		t.Fatalf("expected 'second', got %v", v)
	}
}

// TestByField_StringContent exercises the core hashmap invariant
// (a == b ⇒ hash(a) == hash(b)) for ByField with a string field.
// The two equal-content strings are built from different sources,
// so the underlying byte-header pointers differ. Before the fix
// that used unsafe.Sizeof on the string header, this test failed.
func TestByField_StringContent(t *testing.T) {
	strategy := ByField(func(p Person) string { return p.Name })
	m := NewHashMapWithStrategy[Person, string](strategy)

	// Two Persons with equal-content names from different sources.
	p1 := Person{Name: "alice", Age: 30}
	p2 := Person{Name: fmt.Sprintf("ali%s", "ce"), Age: 40}

	if p1.Name != p2.Name {
		t.Fatalf("precondition: names should compare equal (%q vs %q)", p1.Name, p2.Name)
	}

	m.Put(p1, "first")
	if !m.ContainsKey(p2) {
		t.Fatalf("ContainsKey must find equal-content string built differently")
	}
	v, ok := m.Get(p2)
	if !ok || v != "first" {
		t.Fatalf("Get must return inserted value regardless of string source; got (%q, %v)", v, ok)
	}

	m.Put(p2, "second") // should overwrite
	if m.Len() != 1 {
		t.Fatalf("expected overwrite, size=%d", m.Len())
	}
}

// TestByField_StructWithString covers a struct field that contains
// a string (not just a plain string). hash/maphash.Comparable must
// recurse into the struct and respect == content semantics.
func TestByField_StructWithString(t *testing.T) {
	type addr struct {
		Street string
		ZIP    int
	}
	type user struct {
		Name string
		Home addr
	}
	strategy := ByField(func(u user) addr { return u.Home })
	s := NewHashSetWithStrategy(strategy)

	// Build the two addr values via different string constructions.
	a := addr{Street: "Main St", ZIP: 10001}
	b := addr{Street: fmt.Sprint("Main ", "St"), ZIP: 10001}

	if a != b {
		t.Fatalf("precondition: addrs should be == (%+v vs %+v)", a, b)
	}

	s.Add(user{Name: "u1", Home: a})
	if !s.Contains(user{Name: "u2", Home: b}) {
		t.Fatalf("Contains must find struct-field-equal user built with differently-sourced strings")
	}
}

// TestByField_NumericRoundTrip is a sanity check for the common
// numeric case — the maphash.Comparable path must behave for plain
// numeric fields the same way the old unsafe path did (no regression).
func TestByField_NumericRoundTrip(t *testing.T) {
	strategy := ByField(func(p Person) int { return p.Age })
	m := NewHashMapWithStrategy[Person, string](strategy)
	m.Put(Person{Name: "A", Age: 20}, "twenty")
	m.Put(Person{Name: "B", Age: 30}, "thirty")

	v, ok := m.Get(Person{Name: "ignored", Age: 20})
	if !ok || v != "twenty" {
		t.Fatalf("expected 'twenty', got (%q, %v)", v, ok)
	}
	if m.Len() != 2 {
		t.Fatalf("expected 2 entries, got %d", m.Len())
	}
}

func TestHashSetWithStrategyOperations(t *testing.T) {
	s := NewHashSetWithStrategy(StringHashingStrategy())
	s.Add("a")
	s.Add("b")
	s.Add("c")

	sel := s.Select(func(v string) bool { return v != "b" })
	if sel.Len() != 2 {
		t.Fatalf("select: expected 2, got %d", sel.Len())
	}

	rej := s.Reject(func(v string) bool { return v == "a" })
	if rej.Len() != 2 {
		t.Fatalf("reject: expected 2, got %d", rej.Len())
	}
}

func TestHashSetWithStrategyResize(t *testing.T) {
	s := NewHashSetWithStrategy(StringHashingStrategy())
	for i := 0; i < 1000; i++ {
		s.Add(string(rune('a'+i%26)) + string(rune('0'+i/26)))
	}
	if s.Len() < 100 { // many collisions but should still work
		t.Fatalf("expected many elements, got %d", s.Len())
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

	if m.Len() != 3 {
		t.Fatalf("expected 3, got %d", m.Len())
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
	if m.Len() != 1 {
		t.Fatalf("expected 1, got %d", m.Len())
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
	if m.Len() != 50 {
		t.Fatalf("expected 50, got %d", m.Len())
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

	if s.Len() != 3 {
		t.Fatalf("expected 3, got %d", s.Len())
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
	if s.Len() != 25 {
		t.Fatalf("expected 25, got %d", s.Len())
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
	evens := s.SelectWhere(func(v int) bool { return v%2 == 0 })
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
	if m.Len() != 0 {
		t.Fatal("expected empty after clear")
	}
}

func TestTreeSetStress(t *testing.T) {
	s := NewTreeSet[int](NaturalComparator[int]())
	for i := 999; i >= 0; i-- {
		s.Add(i)
	}
	if s.Len() != 1000 {
		t.Fatalf("expected 1000, got %d", s.Len())
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
	if s.Len() != 0 {
		t.Fatalf("expected empty, got %d", s.Len())
	}
}
