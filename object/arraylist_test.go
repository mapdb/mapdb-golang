// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package object

import (
	"testing"
)

func TestArrayList_NewEmpty(t *testing.T) {
	list := NewArrayList[int]()
	if list.Len() != 0 {
		t.Errorf("Size() = %d, want 0", list.Len())
	}
	if list.Len() != 0 {
		t.Error("IsEmpty() = false, want true")
	}
}

func TestArrayList_NewArrayListFrom(t *testing.T) {
	list := NewArrayListFrom(1, 2, 3)
	if list.Len() != 3 {
		t.Errorf("Size() = %d, want 3", list.Len())
	}
	v := list.Get(0)
	if v != 1 {
		t.Errorf("Get(0) = %d, want 1", v)
	}
}

func TestArrayList_NewArrayListFrom_String(t *testing.T) {
	list := NewArrayListFrom("a", "b", "c")
	if list.Len() != 3 {
		t.Errorf("Size() = %d, want 3", list.Len())
	}
	v := list.Get(1)
	if v != "b" {
		t.Errorf("Get(1) = %q, want %q", v, "b")
	}
}

func TestArrayList_Add(t *testing.T) {
	list := NewArrayList[int]()
	list.Add(10)
	list.Add(20)
	if list.Len() != 2 {
		t.Errorf("Size() = %d, want 2", list.Len())
	}
	if list.Len() == 0 {
		t.Error("IsEmpty() = true, want false")
	}
}

func TestArrayList_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		list := NewArrayListFrom(10, 20, 30)
		v := list.Get(1)
		if v != 20 {
			t.Errorf("Get(1) = %d, want 20", v)
		}
	})

	t.Run("out of bounds negative", func(t *testing.T) {
		list := NewArrayListFrom(10, 20)
		assertPanics(t, func() { list.Get(-1) })
	})

	t.Run("out of bounds high", func(t *testing.T) {
		list := NewArrayListFrom(10, 20)
		assertPanics(t, func() { list.Get(5) })
	})
}

func TestArrayList_Set(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		list := NewArrayListFrom(10, 20, 30)
		old := list.Set(1, 99)
		if old != 20 {
			t.Errorf("Set returned old = %d, want 20", old)
		}
		v := list.Get(1)
		if v != 99 {
			t.Errorf("after Set, Get(1) = %d, want 99", v)
		}
	})

	t.Run("out of bounds", func(t *testing.T) {
		list := NewArrayListFrom(10)
		assertPanics(t, func() { list.Set(5, 99) })
	})
}

func TestArrayList_Contains(t *testing.T) {
	list := NewArrayListFrom(1, 2, 3)
	if !list.Contains(2) {
		t.Error("Contains(2) = false, want true")
	}
	if list.Contains(99) {
		t.Error("Contains(99) = true, want false")
	}
}

func TestArrayList_IndexOf(t *testing.T) {
	list := NewArrayListFrom("a", "b", "c", "b")
	if got := list.IndexOf("b"); got != 1 {
		t.Errorf("IndexOf(b) = %d, want 1", got)
	}
	if got := list.IndexOf("z"); got != -1 {
		t.Errorf("IndexOf(z) = %d, want -1", got)
	}
}

func TestArrayList_ForEach(t *testing.T) {
	list := NewArrayListFrom(1, 2, 3)
	sum := 0
	list.ForEach(func(v int) { sum += v })
	if sum != 6 {
		t.Errorf("ForEach sum = %d, want 6", sum)
	}
}

func TestArrayList_All(t *testing.T) {
	list := NewArrayListFrom(10, 20, 30)
	var collected []int
	for v := range list.All() {
		collected = append(collected, v)
	}
	if len(collected) != 3 {
		t.Errorf("All yielded %d elements, want 3", len(collected))
	}
	if collected[0] != 10 || collected[1] != 20 || collected[2] != 30 {
		t.Errorf("All yielded %v, want [10 20 30]", collected)
	}
}

func TestArrayList_AnySatisfy(t *testing.T) {
	list := NewArrayListFrom(1, 2, 3, 4)
	if !list.AnySatisfy(func(v int) bool { return v > 3 }) {
		t.Error("AnySatisfy(>3) = false, want true")
	}
	if list.AnySatisfy(func(v int) bool { return v > 10 }) {
		t.Error("AnySatisfy(>10) = true, want false")
	}
}

func TestArrayList_AllSatisfy(t *testing.T) {
	list := NewArrayListFrom(2, 4, 6)
	even := func(v int) bool { return v%2 == 0 }
	if !list.AllSatisfy(even) {
		t.Error("AllSatisfy(even) = false, want true")
	}
	list.Add(3)
	if list.AllSatisfy(even) {
		t.Error("AllSatisfy(even) = true after adding 3, want false")
	}
}

func TestArrayList_NoneSatisfy(t *testing.T) {
	list := NewArrayListFrom(1, 3, 5)
	even := func(v int) bool { return v%2 == 0 }
	if !list.NoneSatisfy(even) {
		t.Error("NoneSatisfy(even) = false, want true")
	}
	list.Add(2)
	if list.NoneSatisfy(even) {
		t.Error("NoneSatisfy(even) = true after adding 2, want false")
	}
}

func TestArrayList_Select(t *testing.T) {
	list := NewArrayListFrom(1, 2, 3, 4, 5)
	evens := list.Select(func(v int) bool { return v%2 == 0 })
	if evens.Len() != 2 {
		t.Errorf("Select size = %d, want 2", evens.Len())
	}
	if !evens.Contains(2) || !evens.Contains(4) {
		t.Errorf("Select result = %v, want [2, 4]", evens)
	}
}

func TestArrayList_Reject(t *testing.T) {
	list := NewArrayListFrom(1, 2, 3, 4, 5)
	odds := list.Reject(func(v int) bool { return v%2 == 0 })
	if odds.Len() != 3 {
		t.Errorf("Reject size = %d, want 3", odds.Len())
	}
}

func TestArrayList_Detect(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		list := NewArrayListFrom(1, 2, 3, 4)
		v, ok := list.Detect(func(v int) bool { return v > 2 })
		if !ok {
			t.Fatal("Detect returned false, want true")
		}
		if v != 3 {
			t.Errorf("Detect = %d, want 3", v)
		}
	})

	t.Run("not found", func(t *testing.T) {
		list := NewArrayListFrom(1, 2)
		_, ok := list.Detect(func(v int) bool { return v > 10 })
		if ok {
			t.Error("Detect returned true, want false")
		}
	})
}

func TestArrayList_Count(t *testing.T) {
	list := NewArrayListFrom(1, 2, 3, 4, 5, 6)
	n := list.Count(func(v int) bool { return v%2 == 0 })
	if n != 3 {
		t.Errorf("Count(even) = %d, want 3", n)
	}
}

func TestArrayList_InjectInto(t *testing.T) {
	list := NewArrayListFrom(1, 2, 3, 4)
	sum := list.InjectInto(0, func(acc any, v int) any {
		return acc.(int) + v
	})
	if sum.(int) != 10 {
		t.Errorf("InjectInto sum = %d, want 10", sum.(int))
	}
}

func TestArrayList_Sort(t *testing.T) {
	list := NewArrayListFrom(3, 1, 4, 1, 5)
	list.Sort(func(a, b int) bool { return a < b })
	expected := []int{1, 1, 3, 4, 5}
	got := list.ToSlice()
	for i, v := range expected {
		if got[i] != v {
			t.Errorf("Sort: index %d = %d, want %d", i, got[i], v)
		}
	}
}

func TestArrayList_Reversed(t *testing.T) {
	list := NewArrayListFrom(1, 2, 3)
	rev := list.Reversed()
	got := rev.ToSlice()
	if got[0] != 3 || got[1] != 2 || got[2] != 1 {
		t.Errorf("Reversed = %v, want [3 2 1]", got)
	}
	// original unchanged
	orig := list.ToSlice()
	if orig[0] != 1 {
		t.Error("Reversed mutated original list")
	}
}

func TestArrayList_Distinct(t *testing.T) {
	list := NewArrayListFrom(1, 2, 2, 3, 1)
	d := list.Distinct()
	if d.Len() != 3 {
		t.Errorf("Distinct size = %d, want 3", d.Len())
	}
	// first occurrence order
	got := d.ToSlice()
	if got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Errorf("Distinct = %v, want [1 2 3]", got)
	}
}

func TestArrayList_Remove(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		list := NewArrayListFrom(1, 2, 3, 2)
		ok := list.Remove(2)
		if !ok {
			t.Error("Remove(2) = false, want true")
		}
		if list.Len() != 3 {
			t.Errorf("Size after Remove = %d, want 3", list.Len())
		}
		// should remove first occurrence
		v := list.Get(1)
		if v != 3 {
			t.Errorf("after Remove(2), index 1 = %d, want 3", v)
		}
	})

	t.Run("not found", func(t *testing.T) {
		list := NewArrayListFrom(1, 2, 3)
		ok := list.Remove(99)
		if ok {
			t.Error("Remove(99) = true, want false")
		}
	})
}

func TestArrayList_Clear(t *testing.T) {
	list := NewArrayListFrom(1, 2, 3)
	list.Clear()
	if list.Len() != 0 {
		t.Errorf("Size after Clear = %d, want 0", list.Len())
	}
	if list.Len() != 0 {
		t.Error("IsEmpty after Clear = false, want true")
	}
}

func TestArrayList_String(t *testing.T) {
	list := NewArrayListFrom(1, 2, 3)
	s := list.String()
	if s != "[1, 2, 3]" {
		t.Errorf("String() = %q, want %q", s, "[1, 2, 3]")
	}
}

func TestArrayList_String_StringType(t *testing.T) {
	list := NewArrayListFrom("hello", "world")
	s := list.String()
	if s != "[hello, world]" {
		t.Errorf("String() = %q, want %q", s, "[hello, world]")
	}
}

func TestArrayList_ToSlice(t *testing.T) {
	list := NewArrayListFrom(1, 2, 3)
	s := list.ToSlice()
	if len(s) != 3 {
		t.Fatalf("ToSlice len = %d, want 3", len(s))
	}
	// mutating slice should not affect list
	s[0] = 999
	v := list.Get(0)
	if v != 1 {
		t.Error("ToSlice did not return a copy")
	}
}

func TestArrayList_StringTypes(t *testing.T) {
	list := NewArrayList[string]()
	list.Add("foo")
	list.Add("bar")
	list.Add("baz")

	if !list.Contains("bar") {
		t.Error("Contains(bar) = false")
	}
	if list.Contains("qux") {
		t.Error("Contains(qux) = true")
	}
	if list.IndexOf("baz") != 2 {
		t.Errorf("IndexOf(baz) = %d, want 2", list.IndexOf("baz"))
	}

	selected := list.Select(func(s string) bool { return len(s) == 3 })
	if selected.Len() != 3 {
		t.Errorf("Select 3-char strings size = %d, want 3", selected.Len())
	}
}
