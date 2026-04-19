// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package object

import (
	"testing"
)

func TestHashMap_NewEmpty(t *testing.T) {
	m := NewHashMap[string, int]()
	if m.Size() != 0 {
		t.Errorf("Size() = %d, want 0", m.Size())
	}
	if !m.IsEmpty() {
		t.Error("IsEmpty() = false, want true")
	}
}

func TestHashMap_Put(t *testing.T) {
	t.Run("new key", func(t *testing.T) {
		m := NewHashMap[string, int]()
		old, existed := m.Put("a", 1)
		if existed {
			t.Error("Put returned existed=true for new key")
		}
		if old != 0 {
			t.Errorf("Put returned old=%d, want 0", old)
		}
		if m.Size() != 1 {
			t.Errorf("Size() = %d, want 1", m.Size())
		}
	})

	t.Run("overwrite", func(t *testing.T) {
		m := NewHashMap[string, int]()
		m.Put("a", 1)
		old, existed := m.Put("a", 2)
		if !existed {
			t.Error("Put returned existed=false for existing key")
		}
		if old != 1 {
			t.Errorf("Put returned old=%d, want 1", old)
		}
		v, _ := m.Get("a")
		if v != 2 {
			t.Errorf("Get(a) = %d, want 2", v)
		}
		if m.Size() != 1 {
			t.Errorf("Size() = %d after overwrite, want 1", m.Size())
		}
	})
}

func TestHashMap_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		m := NewHashMap[string, int]()
		m.Put("x", 42)
		v, ok := m.Get("x")
		if !ok {
			t.Fatal("Get returned ok=false")
		}
		if v != 42 {
			t.Errorf("Get(x) = %d, want 42", v)
		}
	})

	t.Run("not found", func(t *testing.T) {
		m := NewHashMap[string, int]()
		v, ok := m.Get("missing")
		if ok {
			t.Error("Get returned ok=true for missing key")
		}
		if v != 0 {
			t.Errorf("Get zero value = %d, want 0", v)
		}
	})
}

func TestHashMap_Remove(t *testing.T) {
	m := NewHashMap[string, int]()
	m.Put("a", 1)
	m.Put("b", 2)

	old, ok := m.Remove("a")
	if !ok {
		t.Error("Remove(a) returned ok=false")
	}
	if old != 1 {
		t.Errorf("Remove(a) old = %d, want 1", old)
	}
	if m.Size() != 1 {
		t.Errorf("Size after Remove = %d, want 1", m.Size())
	}

	_, ok = m.Remove("missing")
	if ok {
		t.Error("Remove(missing) returned ok=true")
	}
}

func TestHashMap_ContainsKey(t *testing.T) {
	m := NewHashMap[string, int]()
	m.Put("a", 1)
	if !m.ContainsKey("a") {
		t.Error("ContainsKey(a) = false")
	}
	if m.ContainsKey("b") {
		t.Error("ContainsKey(b) = true")
	}
}

func TestHashMap_GetOrDefault(t *testing.T) {
	m := NewHashMap[string, int]()
	m.Put("a", 10)
	if v := m.GetOrDefault("a", 99); v != 10 {
		t.Errorf("GetOrDefault(a) = %d, want 10", v)
	}
	if v := m.GetOrDefault("missing", 99); v != 99 {
		t.Errorf("GetOrDefault(missing) = %d, want 99", v)
	}
}

func TestHashMap_ForEach(t *testing.T) {
	m := NewHashMap[string, int]()
	m.Put("a", 1)
	m.Put("b", 2)
	m.Put("c", 3)
	sum := 0
	m.ForEach(func(k string, v int) { sum += v })
	if sum != 6 {
		t.Errorf("ForEach sum = %d, want 6", sum)
	}
}

func TestHashMap_All(t *testing.T) {
	m := NewHashMap[string, int]()
	m.Put("a", 1)
	m.Put("b", 2)
	count := 0
	for k, v := range m.All() {
		if k == "" || v == 0 {
			t.Error("All yielded zero-value key or value")
		}
		count++
	}
	if count != 2 {
		t.Errorf("All yielded %d entries, want 2", count)
	}
}

func TestHashMap_Keys(t *testing.T) {
	m := NewHashMap[string, int]()
	m.Put("a", 1)
	m.Put("b", 2)
	keys := NewHashSet[string]()
	for k := range m.Keys() {
		keys.Add(k)
	}
	if keys.Size() != 2 || !keys.Contains("a") || !keys.Contains("b") {
		t.Errorf("Keys = %v, want {a, b}", keys)
	}
}

func TestHashMap_Values(t *testing.T) {
	m := NewHashMap[string, int]()
	m.Put("a", 1)
	m.Put("b", 2)
	sum := 0
	for v := range m.Values() {
		sum += v
	}
	if sum != 3 {
		t.Errorf("Values sum = %d, want 3", sum)
	}
}

func TestHashMap_AnySatisfy(t *testing.T) {
	m := NewHashMap[string, int]()
	m.Put("a", 1)
	m.Put("b", 5)
	if !m.AnySatisfy(func(k string, v int) bool { return v > 3 }) {
		t.Error("AnySatisfy(>3) = false")
	}
	if m.AnySatisfy(func(k string, v int) bool { return v > 10 }) {
		t.Error("AnySatisfy(>10) = true")
	}
}

func TestHashMap_AllSatisfy(t *testing.T) {
	m := NewHashMap[string, int]()
	m.Put("a", 2)
	m.Put("b", 4)
	if !m.AllSatisfy(func(k string, v int) bool { return v%2 == 0 }) {
		t.Error("AllSatisfy(even) = false")
	}
	m.Put("c", 3)
	if m.AllSatisfy(func(k string, v int) bool { return v%2 == 0 }) {
		t.Error("AllSatisfy(even) = true after adding odd")
	}
}

func TestHashMap_NoneSatisfy(t *testing.T) {
	m := NewHashMap[string, int]()
	m.Put("a", 1)
	m.Put("b", 3)
	if !m.NoneSatisfy(func(k string, v int) bool { return v%2 == 0 }) {
		t.Error("NoneSatisfy(even) = false")
	}
	m.Put("c", 2)
	if m.NoneSatisfy(func(k string, v int) bool { return v%2 == 0 }) {
		t.Error("NoneSatisfy(even) = true after adding even")
	}
}

func TestHashMap_Select(t *testing.T) {
	m := NewHashMap[string, int]()
	m.Put("a", 1)
	m.Put("b", 2)
	m.Put("c", 3)
	evens := m.Select(func(k string, v int) bool { return v%2 == 0 })
	if evens.Size() != 1 {
		t.Errorf("Select size = %d, want 1", evens.Size())
	}
	v, ok := evens.Get("b")
	if !ok || v != 2 {
		t.Errorf("Select missing b=2")
	}
}

func TestHashMap_Reject(t *testing.T) {
	m := NewHashMap[string, int]()
	m.Put("a", 1)
	m.Put("b", 2)
	m.Put("c", 3)
	odds := m.Reject(func(k string, v int) bool { return v%2 == 0 })
	if odds.Size() != 2 {
		t.Errorf("Reject size = %d, want 2", odds.Size())
	}
}

func TestHashMap_Detect(t *testing.T) {
	m := NewHashMap[string, int]()
	m.Put("a", 1)
	m.Put("b", 2)

	t.Run("found", func(t *testing.T) {
		k, v, ok := m.Detect(func(k string, v int) bool { return v == 2 })
		if !ok {
			t.Fatal("Detect returned false")
		}
		if k != "b" || v != 2 {
			t.Errorf("Detect = (%q, %d), want (b, 2)", k, v)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, _, ok := m.Detect(func(k string, v int) bool { return v == 99 })
		if ok {
			t.Error("Detect returned true for missing")
		}
	})
}

func TestHashMap_Count(t *testing.T) {
	m := NewHashMap[string, int]()
	m.Put("a", 1)
	m.Put("b", 2)
	m.Put("c", 3)
	m.Put("d", 4)
	n := m.Count(func(k string, v int) bool { return v%2 == 0 })
	if n != 2 {
		t.Errorf("Count(even) = %d, want 2", n)
	}
}

func TestHashMap_KeysToSlice(t *testing.T) {
	m := NewHashMap[string, int]()
	m.Put("x", 1)
	m.Put("y", 2)
	keys := m.KeysToSlice()
	if len(keys) != 2 {
		t.Errorf("KeysToSlice len = %d, want 2", len(keys))
	}
	found := make(map[string]bool)
	for _, k := range keys {
		found[k] = true
	}
	if !found["x"] || !found["y"] {
		t.Errorf("KeysToSlice = %v, want [x y]", keys)
	}
}

func TestHashMap_ValuesToSlice(t *testing.T) {
	m := NewHashMap[string, int]()
	m.Put("a", 10)
	m.Put("b", 20)
	vals := m.ValuesToSlice()
	if len(vals) != 2 {
		t.Errorf("ValuesToSlice len = %d, want 2", len(vals))
	}
	sum := 0
	for _, v := range vals {
		sum += v
	}
	if sum != 30 {
		t.Errorf("ValuesToSlice sum = %d, want 30", sum)
	}
}

func TestHashMap_Clear(t *testing.T) {
	m := NewHashMap[string, int]()
	m.Put("a", 1)
	m.Put("b", 2)
	m.Clear()
	if m.Size() != 0 {
		t.Errorf("Size after Clear = %d, want 0", m.Size())
	}
	if !m.IsEmpty() {
		t.Error("IsEmpty after Clear = false")
	}
}

func TestHashMap_String(t *testing.T) {
	m := NewHashMap[string, int]()
	m.Put("only", 1)
	s := m.String()
	if s != "{only: 1}" {
		t.Errorf("String() = %q, want %q", s, "{only: 1}")
	}
}

func TestHashMap_IntKeys(t *testing.T) {
	m := NewHashMap[int, string]()
	m.Put(1, "one")
	m.Put(2, "two")
	v, ok := m.Get(1)
	if !ok || v != "one" {
		t.Errorf("Get(1) = (%q, %v), want (one, true)", v, ok)
	}
}
