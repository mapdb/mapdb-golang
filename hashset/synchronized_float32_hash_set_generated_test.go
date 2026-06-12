package hashset

import "testing"

func TestSynchronizedFloat32_Generated_AddContains(t *testing.T) {
	s := NewSynchronizedFloat32()
	s.Add(1.0)
	s.Add(2.0)
	if s.Len() != 2 {
		t.Errorf("Size = %d", s.Len())
	}
	if !s.Contains(1.0) {
		t.Error("Contains should be true")
	}
	if s.Contains(99.0) {
		t.Error("Contains should be false")
	}
}
func TestSynchronizedFloat32_Generated_AddDuplicate(t *testing.T) {
	s := NewSynchronizedFloat32()
	s.Add(1.0)
	added := s.Add(1.0)
	if added {
		t.Error("Duplicate add should return false")
	}
	if s.Len() != 1 {
		t.Errorf("Size = %d", s.Len())
	}
}
func TestSynchronizedFloat32_Generated_Remove(t *testing.T) {
	s := NewSynchronizedFloat32()
	s.Add(1.0)
	s.Add(2.0)
	if !s.Remove(1.0) {
		t.Error("Remove should return true")
	}
	if s.Contains(1.0) {
		t.Error("Should not contain after remove")
	}
}
func TestSynchronizedFloat32_Generated_Clear(t *testing.T) {
	s := NewSynchronizedFloat32()
	s.Add(1.0)
	s.Add(2.0)
	s.Clear()
	if s.Len() != 0 {
		t.Error("Should be empty")
	}
}
func TestSynchronizedFloat32_Generated_All(t *testing.T) {
	s := NewSynchronizedFloat32()
	s.Add(1.0)
	s.Add(2.0)
	count := 0
	for range s.All() {
		count++
	}
	if count != 2 {
		t.Errorf("All count = %d", count)
	}
}
func TestSynchronizedFloat32_Generated_ToSlice(t *testing.T) {
	s := NewSynchronizedFloat32()
	s.Add(1.0)
	s.Add(2.0)
	if len(s.ToSlice()) != 2 {
		t.Error("wrong len")
	}
}
func TestSynchronizedFloat32_Generated_String(t *testing.T) {
	s := NewSynchronizedFloat32()
	s.Add(1.0)
	if s.String() == "" {
		t.Error("empty")
	}
}
