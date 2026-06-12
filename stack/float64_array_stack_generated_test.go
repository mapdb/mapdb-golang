package stack

import (
	"testing"
)

func TestFloat64_Generated_PushPeekPop(t *testing.T) {
	s := NewFloat64()
	s.Push(1.0)
	s.Push(2.0)
	s.Push(3.0)

	if s.Len() != 3 {
		t.Errorf("Size() = %d, want 3", s.Len())
	}

	if top, ok := s.Peek(); !ok || top != 3.0 {
		t.Errorf("Peek() = (%v, %v), want (3.0, true)", top, ok)
	}

	if val, ok := s.Pop(); !ok || val != 3.0 {
		t.Errorf("Pop() = (%v, %v), want (3.0, true)", val, ok)
	}
	if s.Len() != 2 {
		t.Errorf("Size after pop = %d, want 2", s.Len())
	}
}

func TestFloat64_Generated_PopPeekEmpty(t *testing.T) {
	s := NewFloat64()
	if _, ok := s.Pop(); ok {
		t.Error("Pop on empty stack should report not-ok")
	}
	if _, ok := s.Peek(); ok {
		t.Error("Peek on empty stack should report not-ok")
	}
}

func TestFloat64_Generated_IsEmpty(t *testing.T) {
	s := NewFloat64()
	if s.Len() != 0 {
		t.Error("New stack should be empty")
	}
	s.Push(1.0)
	if s.Len() == 0 {
		t.Error("Stack with element should not be empty")
	}
}

func TestFloat64_Generated_Clear(t *testing.T) {
	s := NewFloat64()
	s.Push(1.0)
	s.Push(2.0)
	s.Clear()
	if s.Len() != 0 {
		t.Errorf("After Clear: size=%d, empty=%v", s.Len(), s.Len() == 0)
	}
}

func TestFloat64_Generated_Contains(t *testing.T) {
	s := NewFloat64()
	s.Push(1.0)
	s.Push(2.0)
	if !s.Contains(1.0) {
		t.Error("Contains(1.0) should be true")
	}
	if s.Contains(3.0) {
		t.Error("Contains(3.0) should be false")
	}
}

func TestFloat64_Generated_All(t *testing.T) {
	s := NewFloat64()
	s.Push(1.0)
	s.Push(2.0)
	s.Push(3.0)
	count := 0
	for range s.All() {
		count++
	}
	if count != 3 {
		t.Errorf("All count = %d, want 3", count)
	}
}

func TestFloat64_Generated_ForEach(t *testing.T) {
	s := NewFloat64()
	s.Push(1.0)
	s.Push(2.0)
	sum := float64(0)
	s.ForEach(func(v float64) {
		sum += v
	})
	expected := float64(1.0) + float64(2.0)
	if sum != expected {
		t.Errorf("ForEach sum = %v, want %v", sum, expected)
	}
}

func TestFloat64_Generated_ToSlice(t *testing.T) {
	s := NewFloat64()
	s.Push(1.0)
	s.Push(2.0)
	s.Push(3.0)
	sl := s.ToSlice()
	if len(sl) != 3 {
		t.Errorf("ToSlice len = %d, want 3", len(sl))
	}
}

func TestFloat64_Generated_LIFO_Order(t *testing.T) {
	s := NewFloat64()
	s.Push(1.0)
	s.Push(2.0)
	s.Push(3.0)

	val1, _ := s.Pop()
	val2, _ := s.Pop()
	val3, _ := s.Pop()
	if val1 != 3.0 || val2 != 2.0 || val3 != 1.0 {
		t.Errorf("LIFO order: got %v, %v, %v", val1, val2, val3)
	}
}

func TestFloat64_Generated_String(t *testing.T) {
	s := NewFloat64()
	s.Push(1.0)
	if s.String() == "" {
		t.Error("String should not be empty")
	}
}

func TestFloat64_Generated_PeekAtPanics(t *testing.T) {
	s := Float64Of(1, 2, 3)
	assertPanics(t, func() { _ = s.PeekAt(99) })
}
