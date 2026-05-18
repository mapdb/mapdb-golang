
package stack

import "testing"

func TestSynchronizedInt16ArrayStack_Generated_PushPopPeek(t *testing.T) {
	s := NewSynchronizedInt16ArrayStack()
	s.Push(1)
	s.Push(2)
	s.Push(3)
	if s.Size() != 3 {
		t.Errorf("Size = %d", s.Size())
	}
	if p, err := s.Peek(); err != nil || p != 3 {
		t.Errorf("Peek = (%v, %v)", p, err)
	}
	val, err := s.Pop()
	if err != nil || val != 3 {
		t.Errorf("Pop = (%v, %v)", val, err)
	}
	if s.Size() != 2 {
		t.Errorf("Size after pop = %d", s.Size())
	}
}
func TestSynchronizedInt16ArrayStack_Generated_IsEmpty(t *testing.T) {
	s := NewSynchronizedInt16ArrayStack()
	if !s.IsEmpty() {
		t.Error("Should be empty")
	}
}
func TestSynchronizedInt16ArrayStack_Generated_Contains(t *testing.T) {
	s := NewSynchronizedInt16ArrayStack()
	s.Push(1)
	if !s.Contains(1) {
		t.Error("Contains should be true")
	}
	if s.Contains(2) {
		t.Error("Contains should be false")
	}
}
func TestSynchronizedInt16ArrayStack_Generated_Clear(t *testing.T) {
	s := NewSynchronizedInt16ArrayStack()
	s.Push(1)
	s.Clear()
	if !s.IsEmpty() {
		t.Error("Should be empty")
	}
}
func TestSynchronizedInt16ArrayStack_Generated_All(t *testing.T) {
	s := NewSynchronizedInt16ArrayStack()
	s.Push(1)
	s.Push(2)
	count := 0
	for range s.All() {
		count++
	}
	if count != 2 {
		t.Errorf("All count = %d", count)
	}
}
func TestSynchronizedInt16ArrayStack_Generated_String(t *testing.T) {
	s := NewSynchronizedInt16ArrayStack()
	s.Push(1)
	if s.String() == "" {
		t.Error("empty")
	}
}
