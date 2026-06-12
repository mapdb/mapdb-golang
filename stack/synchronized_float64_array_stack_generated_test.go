package stack

import "testing"

func TestSynchronizedFloat64_Generated_PushPopPeek(t *testing.T) {
	s := NewSynchronizedFloat64()
	s.Push(1.0)
	s.Push(2.0)
	s.Push(3.0)
	if s.Len() != 3 {
		t.Errorf("Size = %d", s.Len())
	}
	if p, ok := s.Peek(); !ok || p != 3.0 {
		t.Errorf("Peek = (%v, %v)", p, ok)
	}
	val, ok := s.Pop()
	if !ok || val != 3.0 {
		t.Errorf("Pop = (%v, %v)", val, ok)
	}
	if s.Len() != 2 {
		t.Errorf("Size after pop = %d", s.Len())
	}
}
func TestSynchronizedFloat64_Generated_IsEmpty(t *testing.T) {
	s := NewSynchronizedFloat64()
	if s.Len() != 0 {
		t.Error("Should be empty")
	}
}
func TestSynchronizedFloat64_Generated_Contains(t *testing.T) {
	s := NewSynchronizedFloat64()
	s.Push(1.0)
	if !s.Contains(1.0) {
		t.Error("Contains should be true")
	}
	if s.Contains(2.0) {
		t.Error("Contains should be false")
	}
}
func TestSynchronizedFloat64_Generated_Clear(t *testing.T) {
	s := NewSynchronizedFloat64()
	s.Push(1.0)
	s.Clear()
	if s.Len() != 0 {
		t.Error("Should be empty")
	}
}
func TestSynchronizedFloat64_Generated_All(t *testing.T) {
	s := NewSynchronizedFloat64()
	s.Push(1.0)
	s.Push(2.0)
	count := 0
	for range s.All() {
		count++
	}
	if count != 2 {
		t.Errorf("All count = %d", count)
	}
}
func TestSynchronizedFloat64_Generated_String(t *testing.T) {
	s := NewSynchronizedFloat64()
	s.Push(1.0)
	if s.String() == "" {
		t.Error("empty")
	}
}

func TestSynchronizedFloat64_Generated_PeekAtPanics(t *testing.T) {
	s := NewSynchronizedFloat64From(Float64Of(1, 2, 3))
	assertPanics(t, func() { _ = s.PeekAt(99) })
}
