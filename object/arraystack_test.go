// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package object

import (
	"testing"
)

func TestArrayStack_NewEmpty(t *testing.T) {
	s := NewArrayStack[int]()
	if s.Size() != 0 {
		t.Errorf("Size() = %d, want 0", s.Size())
	}
	if !s.IsEmpty() {
		t.Error("IsEmpty() = false, want true")
	}
}

func TestArrayStack_NewArrayStackFrom(t *testing.T) {
	// last element is top
	s := NewArrayStackFrom(1, 2, 3)
	if s.Size() != 3 {
		t.Errorf("Size() = %d, want 3", s.Size())
	}
	top, err := s.Peek()
	if err != nil {
		t.Fatalf("Peek error: %v", err)
	}
	if top != 3 {
		t.Errorf("Peek() = %d, want 3 (last element is top)", top)
	}
}

func TestArrayStack_Push(t *testing.T) {
	s := NewArrayStack[int]()
	s.Push(10)
	s.Push(20)
	if s.Size() != 2 {
		t.Errorf("Size() = %d, want 2", s.Size())
	}
	top, _ := s.Peek()
	if top != 20 {
		t.Errorf("Peek() = %d, want 20", top)
	}
}

func TestArrayStack_Pop(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s := NewArrayStackFrom(1, 2, 3)
		v, err := s.Pop()
		if err != nil {
			t.Fatalf("Pop error: %v", err)
		}
		if v != 3 {
			t.Errorf("Pop() = %d, want 3", v)
		}
		if s.Size() != 2 {
			t.Errorf("Size after Pop = %d, want 2", s.Size())
		}

		v, _ = s.Pop()
		if v != 2 {
			t.Errorf("second Pop() = %d, want 2", v)
		}
	})

	t.Run("empty stack error", func(t *testing.T) {
		s := NewArrayStack[int]()
		_, err := s.Pop()
		if err == nil {
			t.Error("Pop on empty stack: expected error")
		}
	})
}

func TestArrayStack_Peek(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s := NewArrayStackFrom(10, 20)
		v, err := s.Peek()
		if err != nil {
			t.Fatalf("Peek error: %v", err)
		}
		if v != 20 {
			t.Errorf("Peek() = %d, want 20", v)
		}
		// Peek should not remove element
		if s.Size() != 2 {
			t.Errorf("Size after Peek = %d, want 2", s.Size())
		}
	})

	t.Run("empty stack error", func(t *testing.T) {
		s := NewArrayStack[int]()
		_, err := s.Peek()
		if err == nil {
			t.Error("Peek on empty stack: expected error")
		}
	})
}

func TestArrayStack_PeekAt(t *testing.T) {
	s := NewArrayStackFrom(10, 20, 30) // top is 30

	t.Run("top", func(t *testing.T) {
		v, err := s.PeekAt(0)
		if err != nil {
			t.Fatalf("PeekAt(0) error: %v", err)
		}
		if v != 30 {
			t.Errorf("PeekAt(0) = %d, want 30", v)
		}
	})

	t.Run("middle", func(t *testing.T) {
		v, err := s.PeekAt(1)
		if err != nil {
			t.Fatalf("PeekAt(1) error: %v", err)
		}
		if v != 20 {
			t.Errorf("PeekAt(1) = %d, want 20", v)
		}
	})

	t.Run("bottom", func(t *testing.T) {
		v, err := s.PeekAt(2)
		if err != nil {
			t.Fatalf("PeekAt(2) error: %v", err)
		}
		if v != 10 {
			t.Errorf("PeekAt(2) = %d, want 10", v)
		}
	})

	t.Run("out of bounds", func(t *testing.T) {
		_, err := s.PeekAt(5)
		if err == nil {
			t.Error("PeekAt(5) expected error")
		}
	})

	t.Run("negative index", func(t *testing.T) {
		_, err := s.PeekAt(-1)
		if err == nil {
			t.Error("PeekAt(-1) expected error")
		}
	})
}

func TestArrayStack_Contains(t *testing.T) {
	s := NewArrayStackFrom(1, 2, 3)
	if !s.Contains(2) {
		t.Error("Contains(2) = false")
	}
	if s.Contains(99) {
		t.Error("Contains(99) = true")
	}
}

func TestArrayStack_ForEach_TopToBottom(t *testing.T) {
	s := NewArrayStackFrom(1, 2, 3) // top is 3
	var order []int
	s.ForEach(func(v int) { order = append(order, v) })
	expected := []int{3, 2, 1}
	if len(order) != 3 {
		t.Fatalf("ForEach yielded %d elements, want 3", len(order))
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("ForEach order[%d] = %d, want %d", i, order[i], v)
		}
	}
}

func TestArrayStack_All_TopToBottom(t *testing.T) {
	s := NewArrayStackFrom(10, 20, 30) // top is 30
	var order []int
	for v := range s.All() {
		order = append(order, v)
	}
	expected := []int{30, 20, 10}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("All order[%d] = %d, want %d", i, order[i], v)
		}
	}
}

func TestArrayStack_ToSlice_TopToBottom(t *testing.T) {
	s := NewArrayStackFrom(1, 2, 3) // top is 3
	sl := s.ToSlice()
	expected := []int{3, 2, 1}
	if len(sl) != 3 {
		t.Fatalf("ToSlice len = %d, want 3", len(sl))
	}
	for i, v := range expected {
		if sl[i] != v {
			t.Errorf("ToSlice[%d] = %d, want %d", i, sl[i], v)
		}
	}
}

func TestArrayStack_Clear(t *testing.T) {
	s := NewArrayStackFrom(1, 2, 3)
	s.Clear()
	if s.Size() != 0 {
		t.Errorf("Size after Clear = %d, want 0", s.Size())
	}
	if !s.IsEmpty() {
		t.Error("IsEmpty after Clear = false")
	}
}

func TestArrayStack_String(t *testing.T) {
	s := NewArrayStackFrom(1, 2, 3) // top is 3
	str := s.String()
	if str != "[top: 3, 2, 1]" {
		t.Errorf("String() = %q, want %q", str, "[top: 3, 2, 1]")
	}
}

func TestArrayStack_StringType(t *testing.T) {
	s := NewArrayStack[string]()
	s.Push("hello")
	s.Push("world")
	top, _ := s.Peek()
	if top != "world" {
		t.Errorf("Peek() = %q, want %q", top, "world")
	}
	v, _ := s.Pop()
	if v != "world" {
		t.Errorf("Pop() = %q, want %q", v, "world")
	}
	if s.Size() != 1 {
		t.Errorf("Size after Pop = %d, want 1", s.Size())
	}
}

func TestArrayStack_AnySatisfy(t *testing.T) {
	s := NewArrayStackFrom(1, 2, 3)
	if !s.AnySatisfy(func(v int) bool { return v == 2 }) {
		t.Error("AnySatisfy(==2) = false")
	}
	if s.AnySatisfy(func(v int) bool { return v > 10 }) {
		t.Error("AnySatisfy(>10) = true")
	}
}

func TestArrayStack_AllSatisfy(t *testing.T) {
	s := NewArrayStackFrom(2, 4, 6)
	if !s.AllSatisfy(func(v int) bool { return v%2 == 0 }) {
		t.Error("AllSatisfy(even) = false")
	}
	s.Push(3)
	if s.AllSatisfy(func(v int) bool { return v%2 == 0 }) {
		t.Error("AllSatisfy(even) = true after pushing 3")
	}
}

func TestArrayStack_NoneSatisfy(t *testing.T) {
	s := NewArrayStackFrom(1, 3, 5)
	if !s.NoneSatisfy(func(v int) bool { return v%2 == 0 }) {
		t.Error("NoneSatisfy(even) = false")
	}
	s.Push(2)
	if s.NoneSatisfy(func(v int) bool { return v%2 == 0 }) {
		t.Error("NoneSatisfy(even) = true after pushing 2")
	}
}
