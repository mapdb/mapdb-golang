package stack

import (
	"testing"
)

func TestInt32_PushPopPeek(t *testing.T) {
	s := NewInt32()
	s.Push(10)
	s.Push(20)
	s.Push(30)

	if s.Len() != 3 {
		t.Errorf("Size = %d, want 3", s.Len())
	}
	if p, ok := s.Peek(); !ok || p != 30 {
		t.Errorf("Peek = (%d, %v), want (30, true)", p, ok)
	}
	if v, ok := s.Pop(); !ok || v != 30 {
		t.Errorf("Pop = (%d, %v), want (30, true)", v, ok)
	}
	if v, ok := s.Pop(); !ok || v != 20 {
		t.Errorf("Pop = (%d, %v), want (20, true)", v, ok)
	}
	if s.Len() != 1 {
		t.Errorf("Size after 2 pops = %d, want 1", s.Len())
	}
}

func TestInt32_PopPeekEmpty(t *testing.T) {
	s := NewInt32()
	if _, ok := s.Pop(); ok {
		t.Error("Pop on empty stack should report not-ok")
	}
	if _, ok := s.Peek(); ok {
		t.Error("Peek on empty stack should report not-ok")
	}
}

func TestInt32_PeekAt(t *testing.T) {
	s := Int32Of(10, 20, 30) // 30 is top
	if v := s.PeekAt(0); v != 30 {
		t.Errorf("PeekAt(0) = %v, want 30", v)
	}
	if v := s.PeekAt(1); v != 20 {
		t.Errorf("PeekAt(1) = %v, want 20", v)
	}
	if v := s.PeekAt(2); v != 10 {
		t.Errorf("PeekAt(2) = %v, want 10", v)
	}
	assertPanics(t, func() { _ = s.PeekAt(99) })
}

func TestInt32_Contains(t *testing.T) {
	s := Int32Of(1, 2, 3)
	if !s.Contains(2) {
		t.Error("Contains(2) should be true")
	}
	if s.Contains(99) {
		t.Error("Contains(99) should be false")
	}
}

func TestInt32_All(t *testing.T) {
	s := Int32Of(10, 20, 30)
	var result []int32
	for v := range s.All() {
		result = append(result, v)
	}
	// All iterates top to bottom: 30, 20, 10
	if len(result) != 3 || result[0] != 30 || result[1] != 20 || result[2] != 10 {
		t.Errorf("All = %v, want [30, 20, 10]", result)
	}
}

func TestInt32_Select(t *testing.T) {
	s := Int32Of(1, 2, 3, 4, 5)
	evens := s.Select(func(v int32) bool { return v%2 == 0 })
	if evens.Len() != 2 {
		t.Errorf("Select evens size = %d, want 2", evens.Len())
	}
}

func TestInt32_Empty(t *testing.T) {
	s := NewInt32()
	if s.Len() != 0 {
		t.Error("New stack should be empty")
	}
}

func TestInt32_Clear(t *testing.T) {
	s := Int32Of(1, 2, 3)
	s.Clear()
	if s.Len() != 0 {
		t.Error("Stack should be empty after Clear")
	}
}

func TestInt32_ToImmutable(t *testing.T) {
	s := Int32Of(10, 20, 30)
	im := s.ToImmutable()
	if p, ok := im.Peek(); !ok || p != 30 {
		t.Errorf("Immutable Peek = (%d, %v), want (30, true)", p, ok)
	}
	if im.Len() != 3 {
		t.Errorf("Immutable Size = %d, want 3", im.Len())
	}
	// Mutating original should not affect immutable
	_, _ = s.Pop()
	if im.Len() != 3 {
		t.Errorf("Immutable Size after mutable Pop = %d, want 3", im.Len())
	}
}

func TestImmutableInt32_PersistentPush(t *testing.T) {
	im := NewImmutableInt32(10, 20)
	im2 := im.Push(30)
	if im.Len() != 2 {
		t.Errorf("Original size = %d, want 2", im.Len())
	}
	p, _ := im2.Peek()
	if im2.Len() != 3 || p != 30 {
		t.Errorf("New stack: size=%d peek=%d, want 3/30", im2.Len(), p)
	}
}

func TestImmutableInt32_PersistentPop(t *testing.T) {
	im := NewImmutableInt32(10, 20, 30)
	im2, top, ok := im.Pop()
	if !ok || top != 30 {
		t.Errorf("Pop = (%d, %v), want (30, true)", top, ok)
	}
	if im.Len() != 3 {
		t.Errorf("Original size = %d, want 3", im.Len())
	}
	p, _ := im2.Peek()
	if im2.Len() != 2 || p != 20 {
		t.Errorf("New stack: size=%d peek=%d, want 2/20", im2.Len(), p)
	}
}

func TestInt32_String(t *testing.T) {
	s := Int32Of(10, 20, 30)
	str := s.String()
	if str != "[30, 20, 10]" {
		t.Errorf("String = %q, want [30, 20, 10]", str)
	}
}
