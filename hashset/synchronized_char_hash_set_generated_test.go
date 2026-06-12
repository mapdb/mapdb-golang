package hashset

import "testing"

func TestSynchronizedChar_Generated_AddContains(t *testing.T) {
	s := NewSynchronizedChar()
	s.Add(1)
	s.Add(2)
	if s.Len() != 2 {
		t.Errorf("Size = %d", s.Len())
	}
	if !s.Contains(1) {
		t.Error("Contains should be true")
	}
	if s.Contains(99) {
		t.Error("Contains should be false")
	}
}
func TestSynchronizedChar_Generated_AddDuplicate(t *testing.T) {
	s := NewSynchronizedChar()
	s.Add(1)
	added := s.Add(1)
	if added {
		t.Error("Duplicate add should return false")
	}
	if s.Len() != 1 {
		t.Errorf("Size = %d", s.Len())
	}
}
func TestSynchronizedChar_Generated_Remove(t *testing.T) {
	s := NewSynchronizedChar()
	s.Add(1)
	s.Add(2)
	if !s.Remove(1) {
		t.Error("Remove should return true")
	}
	if s.Contains(1) {
		t.Error("Should not contain after remove")
	}
}
func TestSynchronizedChar_Generated_Clear(t *testing.T) {
	s := NewSynchronizedChar()
	s.Add(1)
	s.Add(2)
	s.Clear()
	if s.Len() != 0 {
		t.Error("Should be empty")
	}
}
func TestSynchronizedChar_Generated_All(t *testing.T) {
	s := NewSynchronizedChar()
	s.Add(1)
	s.Add(2)
	count := 0
	for range s.All() {
		count++
	}
	if count != 2 {
		t.Errorf("All count = %d", count)
	}
}
func TestSynchronizedChar_Generated_ToSlice(t *testing.T) {
	s := NewSynchronizedChar()
	s.Add(1)
	s.Add(2)
	if len(s.ToSlice()) != 2 {
		t.Error("wrong len")
	}
}
func TestSynchronizedChar_Generated_String(t *testing.T) {
	s := NewSynchronizedChar()
	s.Add(1)
	if s.String() == "" {
		t.Error("empty")
	}
}
