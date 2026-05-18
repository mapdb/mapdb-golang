
package stack

import (
	"testing"
)

func TestCharArrayStack_Generated_PushPeekPop(t *testing.T) {
	s := NewCharArrayStack()
	s.Push(1)
	s.Push(2)
	s.Push(3)

	if s.Size() != 3 {
		t.Errorf("Size() = %d, want 3", s.Size())
	}

	if top, err := s.Peek(); err != nil || top != 3 {
		t.Errorf("Peek() = (%v, %v), want (3, nil)", top, err)
	}

	if val, err := s.Pop(); err != nil || val != 3 {
		t.Errorf("Pop() = (%v, %v), want (3, nil)", val, err)
	}
	if s.Size() != 2 {
		t.Errorf("Size after pop = %d, want 2", s.Size())
	}
}

func TestCharArrayStack_Generated_PopPeekEmpty(t *testing.T) {
	s := NewCharArrayStack()
	if _, err := s.Pop(); err == nil {
		t.Error("Pop on empty stack should return an error")
	}
	if _, err := s.Peek(); err == nil {
		t.Error("Peek on empty stack should return an error")
	}
}

func TestCharArrayStack_Generated_IsEmpty(t *testing.T) {
	s := NewCharArrayStack()
	if !s.IsEmpty() {
		t.Error("New stack should be empty")
	}
	s.Push(1)
	if s.IsEmpty() {
		t.Error("Stack with element should not be empty")
	}
}

func TestCharArrayStack_Generated_Clear(t *testing.T) {
	s := NewCharArrayStack()
	s.Push(1)
	s.Push(2)
	s.Clear()
	if s.Size() != 0 || !s.IsEmpty() {
		t.Errorf("After Clear: size=%d, empty=%v", s.Size(), s.IsEmpty())
	}
}

func TestCharArrayStack_Generated_Contains(t *testing.T) {
	s := NewCharArrayStack()
	s.Push(1)
	s.Push(2)
	if !s.Contains(1) {
		t.Error("Contains(1) should be true")
	}
	if s.Contains(3) {
		t.Error("Contains(3) should be false")
	}
}

func TestCharArrayStack_Generated_All(t *testing.T) {
	s := NewCharArrayStack()
	s.Push(1)
	s.Push(2)
	s.Push(3)
	count := 0
	for range s.All() {
		count++
	}
	if count != 3 {
		t.Errorf("All count = %d, want 3", count)
	}
}

func TestCharArrayStack_Generated_ForEach(t *testing.T) {
	s := NewCharArrayStack()
	s.Push(1)
	s.Push(2)
	sum := uint16(0)
	s.ForEach(func(v uint16) {
		sum += v
	})
	expected := uint16(1) + uint16(2)
	if sum != expected {
		t.Errorf("ForEach sum = %v, want %v", sum, expected)
	}
}

func TestCharArrayStack_Generated_ToSlice(t *testing.T) {
	s := NewCharArrayStack()
	s.Push(1)
	s.Push(2)
	s.Push(3)
	sl := s.ToSlice()
	if len(sl) != 3 {
		t.Errorf("ToSlice len = %d, want 3", len(sl))
	}
}

func TestCharArrayStack_Generated_LIFO_Order(t *testing.T) {
	s := NewCharArrayStack()
	s.Push(1)
	s.Push(2)
	s.Push(3)

	val1, _ := s.Pop()
	val2, _ := s.Pop()
	val3, _ := s.Pop()
	if val1 != 3 || val2 != 2 || val3 != 1 {
		t.Errorf("LIFO order: got %v, %v, %v", val1, val2, val3)
	}
}

func TestCharArrayStack_Generated_String(t *testing.T) {
	s := NewCharArrayStack()
	s.Push(1)
	if s.String() == "" {
		t.Error("String should not be empty")
	}
}
