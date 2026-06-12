// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package object

import (
	"testing"
)

func TestHashBiMap_NewEmpty(t *testing.T) {
	bm := NewHashBiMap[string, int]()
	if bm.Len() != 0 {
		t.Errorf("Size() = %d, want 0", bm.Len())
	}
	if bm.Len() != 0 {
		t.Error("IsEmpty() = false, want true")
	}
}

func TestHashBiMap_Put(t *testing.T) {
	t.Run("new entry", func(t *testing.T) {
		bm := NewHashBiMap[string, int]()
		old, existed := bm.Put("a", 1)
		if existed {
			t.Error("Put returned existed=true for new key")
		}
		if old != 0 {
			t.Errorf("Put returned old=%d, want 0", old)
		}
		if bm.Len() != 1 {
			t.Errorf("Size() = %d, want 1", bm.Len())
		}
	})

	t.Run("overwrite same key", func(t *testing.T) {
		bm := NewHashBiMap[string, int]()
		bm.Put("a", 1)
		old, existed := bm.Put("a", 2)
		if !existed {
			t.Error("Put returned existed=false for existing key")
		}
		if old != 1 {
			t.Errorf("Put returned old=%d, want 1", old)
		}
		v, ok := bm.Get("a")
		if !ok || v != 2 {
			t.Errorf("Get(a) = (%d, %v), want (2, true)", v, ok)
		}
		// old value 1 should no longer be in inverse
		if bm.ContainsValue(1) {
			t.Error("ContainsValue(1) = true after overwrite, want false")
		}
	})
}

func TestHashBiMap_BijectionInvariant(t *testing.T) {
	t.Run("existing value under new key removes old key", func(t *testing.T) {
		bm := NewHashBiMap[string, int]()
		bm.Put("a", 1)
		bm.Put("b", 2)
		// Now put value 1 under key "c" — should remove key "a"
		bm.Put("c", 1)

		if bm.ContainsKey("a") {
			t.Error("key 'a' should have been removed (bijection enforcement)")
		}
		v, ok := bm.Get("c")
		if !ok || v != 1 {
			t.Errorf("Get(c) = (%d, %v), want (1, true)", v, ok)
		}
		k, ok := bm.GetInverse(1)
		if !ok || k != "c" {
			t.Errorf("GetInverse(1) = (%q, %v), want (c, true)", k, ok)
		}
		if bm.Len() != 2 {
			t.Errorf("Size() = %d, want 2", bm.Len())
		}
	})

	t.Run("same key same value is no-op", func(t *testing.T) {
		bm := NewHashBiMap[string, int]()
		bm.Put("a", 1)
		old, existed := bm.Put("a", 1)
		if !existed {
			t.Error("expected existed=true for same key-value re-put")
		}
		if old != 1 {
			t.Errorf("old = %d, want 1", old)
		}
		if bm.Len() != 1 {
			t.Errorf("Size() = %d, want 1", bm.Len())
		}
	})

	t.Run("complex bijection scenario", func(t *testing.T) {
		bm := NewHashBiMap[int, int]()
		bm.Put(1, 10)
		bm.Put(2, 20)
		bm.Put(3, 30)

		// Put key=4, value=20 should remove key=2
		bm.Put(4, 20)
		if bm.ContainsKey(2) {
			t.Error("key 2 should be removed")
		}
		if bm.Len() != 3 {
			t.Errorf("Size() = %d, want 3", bm.Len())
		}
		k, ok := bm.GetInverse(20)
		if !ok || k != 4 {
			t.Errorf("GetInverse(20) = (%d, %v), want (4, true)", k, ok)
		}
	})
}

func TestHashBiMap_Get(t *testing.T) {
	bm := NewHashBiMap[string, int]()
	bm.Put("x", 42)
	v, ok := bm.Get("x")
	if !ok || v != 42 {
		t.Errorf("Get(x) = (%d, %v), want (42, true)", v, ok)
	}
	_, ok = bm.Get("missing")
	if ok {
		t.Error("Get(missing) = true, want false")
	}
}

func TestHashBiMap_GetInverse(t *testing.T) {
	bm := NewHashBiMap[string, int]()
	bm.Put("hello", 99)
	k, ok := bm.GetInverse(99)
	if !ok || k != "hello" {
		t.Errorf("GetInverse(99) = (%q, %v), want (hello, true)", k, ok)
	}
	_, ok = bm.GetInverse(0)
	if ok {
		t.Error("GetInverse(0) = true, want false")
	}
}

func TestHashBiMap_ContainsKey(t *testing.T) {
	bm := NewHashBiMap[string, int]()
	bm.Put("a", 1)
	if !bm.ContainsKey("a") {
		t.Error("ContainsKey(a) = false")
	}
	if bm.ContainsKey("b") {
		t.Error("ContainsKey(b) = true")
	}
}

func TestHashBiMap_ContainsValue(t *testing.T) {
	bm := NewHashBiMap[string, int]()
	bm.Put("a", 1)
	if !bm.ContainsValue(1) {
		t.Error("ContainsValue(1) = false")
	}
	if bm.ContainsValue(99) {
		t.Error("ContainsValue(99) = true")
	}
}

func TestHashBiMap_Remove(t *testing.T) {
	bm := NewHashBiMap[string, int]()
	bm.Put("a", 1)
	bm.Put("b", 2)

	v, ok := bm.Remove("a")
	if !ok || v != 1 {
		t.Errorf("Remove(a) = (%d, %v), want (1, true)", v, ok)
	}
	if bm.ContainsKey("a") {
		t.Error("ContainsKey(a) after Remove = true")
	}
	if bm.ContainsValue(1) {
		t.Error("ContainsValue(1) after Remove = true")
	}
	if bm.Len() != 1 {
		t.Errorf("Size after Remove = %d, want 1", bm.Len())
	}

	_, ok = bm.Remove("missing")
	if ok {
		t.Error("Remove(missing) = true, want false")
	}
}

func TestHashBiMap_RemoveInverse(t *testing.T) {
	bm := NewHashBiMap[string, int]()
	bm.Put("a", 1)
	bm.Put("b", 2)

	k, ok := bm.RemoveInverse(2)
	if !ok || k != "b" {
		t.Errorf("RemoveInverse(2) = (%q, %v), want (b, true)", k, ok)
	}
	if bm.ContainsKey("b") {
		t.Error("ContainsKey(b) after RemoveInverse = true")
	}
	if bm.ContainsValue(2) {
		t.Error("ContainsValue(2) after RemoveInverse = true")
	}

	_, ok = bm.RemoveInverse(99)
	if ok {
		t.Error("RemoveInverse(99) = true, want false")
	}
}

func TestHashBiMap_Inverse(t *testing.T) {
	bm := NewHashBiMap[string, int]()
	bm.Put("a", 1)
	bm.Put("b", 2)

	inv := bm.Inverse()
	if inv.Len() != 2 {
		t.Errorf("Inverse size = %d, want 2", inv.Len())
	}

	// In the inverse, keys are int, values are string
	v, ok := inv.Get(1)
	if !ok || v != "a" {
		t.Errorf("Inverse.Get(1) = (%q, %v), want (a, true)", v, ok)
	}
	v, ok = inv.Get(2)
	if !ok || v != "b" {
		t.Errorf("Inverse.Get(2) = (%q, %v), want (b, true)", v, ok)
	}

	// Inverse of inverse should give original values
	k, ok := inv.GetInverse("a")
	if !ok || k != 1 {
		t.Errorf("Inverse.GetInverse(a) = (%d, %v), want (1, true)", k, ok)
	}

	// Modifying inverse should not affect original
	inv.Put(3, "c")
	if bm.Len() != 2 {
		t.Error("modifying inverse affected original")
	}
}

func TestHashBiMap_ForEach(t *testing.T) {
	bm := NewHashBiMap[string, int]()
	bm.Put("a", 1)
	bm.Put("b", 2)
	sum := 0
	bm.ForEach(func(k string, v int) { sum += v })
	if sum != 3 {
		t.Errorf("ForEach sum = %d, want 3", sum)
	}
}

func TestHashBiMap_All(t *testing.T) {
	bm := NewHashBiMap[string, int]()
	bm.Put("a", 1)
	bm.Put("b", 2)
	count := 0
	for k, v := range bm.All() {
		if k == "" || v == 0 {
			t.Error("All yielded zero values")
		}
		count++
	}
	if count != 2 {
		t.Errorf("All yielded %d entries, want 2", count)
	}
}

func TestHashBiMap_Keys(t *testing.T) {
	bm := NewHashBiMap[string, int]()
	bm.Put("x", 1)
	bm.Put("y", 2)
	keys := NewHashSet[string]()
	for k := range bm.Keys() {
		keys.Add(k)
	}
	if keys.Len() != 2 || !keys.Contains("x") || !keys.Contains("y") {
		t.Errorf("Keys = %v, want {x, y}", keys)
	}
}

func TestHashBiMap_Values(t *testing.T) {
	bm := NewHashBiMap[string, int]()
	bm.Put("a", 10)
	bm.Put("b", 20)
	sum := 0
	for v := range bm.Values() {
		sum += v
	}
	if sum != 30 {
		t.Errorf("Values sum = %d, want 30", sum)
	}
}

func TestHashBiMap_KeysToSlice(t *testing.T) {
	bm := NewHashBiMap[string, int]()
	bm.Put("a", 1)
	bm.Put("b", 2)
	keys := bm.KeysToSlice()
	if len(keys) != 2 {
		t.Errorf("KeysToSlice len = %d, want 2", len(keys))
	}
}

func TestHashBiMap_ValuesToSlice(t *testing.T) {
	bm := NewHashBiMap[string, int]()
	bm.Put("a", 1)
	bm.Put("b", 2)
	vals := bm.ValuesToSlice()
	if len(vals) != 2 {
		t.Errorf("ValuesToSlice len = %d, want 2", len(vals))
	}
}

func TestHashBiMap_AnySatisfy(t *testing.T) {
	bm := NewHashBiMap[string, int]()
	bm.Put("a", 1)
	bm.Put("b", 5)
	if !bm.AnySatisfy(func(k string, v int) bool { return v > 3 }) {
		t.Error("AnySatisfy(>3) = false")
	}
	if bm.AnySatisfy(func(k string, v int) bool { return v > 10 }) {
		t.Error("AnySatisfy(>10) = true")
	}
}

func TestHashBiMap_AllSatisfy(t *testing.T) {
	bm := NewHashBiMap[string, int]()
	bm.Put("a", 2)
	bm.Put("b", 4)
	if !bm.AllSatisfy(func(k string, v int) bool { return v%2 == 0 }) {
		t.Error("AllSatisfy(even) = false")
	}
	bm.Put("c", 3)
	if bm.AllSatisfy(func(k string, v int) bool { return v%2 == 0 }) {
		t.Error("AllSatisfy(even) = true after adding odd")
	}
}

func TestHashBiMap_NoneSatisfy(t *testing.T) {
	bm := NewHashBiMap[string, int]()
	bm.Put("a", 1)
	bm.Put("b", 3)
	if !bm.NoneSatisfy(func(k string, v int) bool { return v%2 == 0 }) {
		t.Error("NoneSatisfy(even) = false")
	}
	bm.Put("c", 2)
	if bm.NoneSatisfy(func(k string, v int) bool { return v%2 == 0 }) {
		t.Error("NoneSatisfy(even) = true after adding even")
	}
}

func TestHashBiMap_Clear(t *testing.T) {
	bm := NewHashBiMap[string, int]()
	bm.Put("a", 1)
	bm.Put("b", 2)
	bm.Clear()
	if bm.Len() != 0 {
		t.Errorf("Size after Clear = %d, want 0", bm.Len())
	}
	if bm.Len() != 0 {
		t.Error("IsEmpty after Clear = false")
	}
	if bm.ContainsKey("a") {
		t.Error("ContainsKey(a) after Clear = true")
	}
	if bm.ContainsValue(1) {
		t.Error("ContainsValue(1) after Clear = true")
	}
}

func TestHashBiMap_String(t *testing.T) {
	bm := NewHashBiMap[string, int]()
	bm.Put("only", 1)
	s := bm.String()
	// format: {BiMap: only↔1}
	expected := "{BiMap: only\u2194\u0031}"
	if s != expected {
		t.Errorf("String() = %q, want %q", s, expected)
	}
}

func TestHashBiMap_IntToString(t *testing.T) {
	bm := NewHashBiMap[int, string]()
	bm.Put(1, "one")
	bm.Put(2, "two")

	v, ok := bm.Get(1)
	if !ok || v != "one" {
		t.Errorf("Get(1) = (%q, %v), want (one, true)", v, ok)
	}
	k, ok := bm.GetInverse("two")
	if !ok || k != 2 {
		t.Errorf("GetInverse(two) = (%d, %v), want (2, true)", k, ok)
	}
}
