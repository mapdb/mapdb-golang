package stack

import (
	"testing"
)

func TestInt32ArrayStack_PushPopPeek(t *testing.T) {
	s := NewInt32ArrayStack()
	s.Push(10)
	s.Push(20)
	s.Push(30)

	if s.Size() != 3 {
		t.Errorf("Size = %d, want 3", s.Size())
	}
	if p, err := s.Peek(); err != nil || p != 30 {
		t.Errorf("Peek = (%d, %v), want (30, nil)", p, err)
	}
	if v, err := s.Pop(); err != nil || v != 30 {
		t.Errorf("Pop = (%d, %v), want (30, nil)", v, err)
	}
	if v, err := s.Pop(); err != nil || v != 20 {
		t.Errorf("Pop = (%d, %v), want (20, nil)", v, err)
	}
	if s.Size() != 1 {
		t.Errorf("Size after 2 pops = %d, want 1", s.Size())
	}
}

func TestInt32ArrayStack_PopPeekEmpty(t *testing.T) {
	s := NewInt32ArrayStack()
	if _, err := s.Pop(); err == nil {
		t.Error("Pop on empty stack should return an error")
	}
	if _, err := s.Peek(); err == nil {
		t.Error("Peek on empty stack should return an error")
	}
}

func TestInt32ArrayStack_PeekAt(t *testing.T) {
	s := Int32ArrayStackOf(10, 20, 30) // 30 is top
	if v, err := s.PeekAt(0); err != nil || v != 30 {
		t.Errorf("PeekAt(0) = (%d, %v), want (30, nil)", v, err)
	}
	if v, err := s.PeekAt(1); err != nil || v != 20 {
		t.Errorf("PeekAt(1) = (%d, %v), want (20, nil)", v, err)
	}
	if v, err := s.PeekAt(2); err != nil || v != 10 {
		t.Errorf("PeekAt(2) = (%d, %v), want (10, nil)", v, err)
	}
	if _, err := s.PeekAt(99); err == nil {
		t.Error("PeekAt(99) should return an error")
	}
}

func TestInt32ArrayStack_Contains(t *testing.T) {
	s := Int32ArrayStackOf(1, 2, 3)
	if !s.Contains(2) {
		t.Error("Contains(2) should be true")
	}
	if s.Contains(99) {
		t.Error("Contains(99) should be false")
	}
}

func TestInt32ArrayStack_All(t *testing.T) {
	s := Int32ArrayStackOf(10, 20, 30)
	var result []int32
	for v := range s.All() {
		result = append(result, v)
	}
	// All iterates top to bottom: 30, 20, 10
	if len(result) != 3 || result[0] != 30 || result[1] != 20 || result[2] != 10 {
		t.Errorf("All = %v, want [30, 20, 10]", result)
	}
}

func TestInt32ArrayStack_Select(t *testing.T) {
	s := Int32ArrayStackOf(1, 2, 3, 4, 5)
	evens := s.Select(func(v int32) bool { return v%2 == 0 })
	if evens.Size() != 2 {
		t.Errorf("Select evens size = %d, want 2", evens.Size())
	}
}

func TestInt32ArrayStack_Empty(t *testing.T) {
	s := NewInt32ArrayStack()
	if !s.IsEmpty() {
		t.Error("New stack should be empty")
	}
}

func TestInt32ArrayStack_Clear(t *testing.T) {
	s := Int32ArrayStackOf(1, 2, 3)
	s.Clear()
	if !s.IsEmpty() {
		t.Error("Stack should be empty after Clear")
	}
}

func TestInt32ArrayStack_ToImmutable(t *testing.T) {
	s := Int32ArrayStackOf(10, 20, 30)
	im := s.ToImmutable()
	if p, err := im.Peek(); err != nil || p != 30 {
		t.Errorf("Immutable Peek = (%d, %v), want (30, nil)", p, err)
	}
	if im.Size() != 3 {
		t.Errorf("Immutable Size = %d, want 3", im.Size())
	}
	// Mutating original should not affect immutable
	_, _ = s.Pop()
	if im.Size() != 3 {
		t.Errorf("Immutable Size after mutable Pop = %d, want 3", im.Size())
	}
}

func TestImmutableInt32ArrayStack_PersistentPush(t *testing.T) {
	im := NewImmutableInt32ArrayStack(10, 20)
	im2 := im.Push(30)
	if im.Size() != 2 {
		t.Errorf("Original size = %d, want 2", im.Size())
	}
	p, _ := im2.Peek()
	if im2.Size() != 3 || p != 30 {
		t.Errorf("New stack: size=%d peek=%d, want 3/30", im2.Size(), p)
	}
}

func TestImmutableInt32ArrayStack_PersistentPop(t *testing.T) {
	im := NewImmutableInt32ArrayStack(10, 20, 30)
	im2, top, err := im.Pop()
	if err != nil || top != 30 {
		t.Errorf("Pop = (%d, %v), want (30, nil)", top, err)
	}
	if im.Size() != 3 {
		t.Errorf("Original size = %d, want 3", im.Size())
	}
	p, _ := im2.Peek()
	if im2.Size() != 2 || p != 20 {
		t.Errorf("New stack: size=%d peek=%d, want 2/20", im2.Size(), p)
	}
}

func TestInt32ArrayStack_String(t *testing.T) {
	s := Int32ArrayStackOf(10, 20, 30)
	str := s.String()
	if str != "[30, 20, 10]" {
		t.Errorf("String = %q, want [30, 20, 10]", str)
	}
}
