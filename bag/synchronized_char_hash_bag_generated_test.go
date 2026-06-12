package bag

import "testing"

func TestSynchronizedHashChar_Generated_AddOccurrences(t *testing.T) {
	s := NewSynchronizedHashChar()
	s.Add(1)
	s.Add(1)
	s.Add(2)
	if s.OccurrencesOf(1) != 2 {
		t.Errorf("OccurrencesOf = %d", s.OccurrencesOf(1))
	}
	if s.Len() != 3 {
		t.Errorf("Size = %d", s.Len())
	}
	if s.SizeDistinct() != 2 {
		t.Errorf("SizeDistinct = %d", s.SizeDistinct())
	}
}
func TestSynchronizedHashChar_Generated_Remove(t *testing.T) {
	s := NewSynchronizedHashChar()
	s.Add(1)
	s.Add(1)
	s.Add(2)
	if !s.Remove(1) {
		t.Error("Remove should return true")
	}
	if s.OccurrencesOf(1) != 1 {
		t.Errorf("After remove: %d", s.OccurrencesOf(1))
	}
}
func TestSynchronizedHashChar_Generated_Contains(t *testing.T) {
	s := NewSynchronizedHashChar()
	s.Add(1)
	if !s.Contains(1) {
		t.Error("Contains should be true")
	}
}
func TestSynchronizedHashChar_Generated_Clear(t *testing.T) {
	s := NewSynchronizedHashChar()
	s.Add(1)
	s.Clear()
	if s.Len() != 0 {
		t.Error("Should be empty")
	}
}
func TestSynchronizedHashChar_Generated_All(t *testing.T) {
	s := NewSynchronizedHashChar()
	s.Add(1)
	s.Add(1)
	s.Add(2)
	count := 0
	for range s.All() {
		count++
	}
	if count != 3 {
		t.Errorf("All count = %d", count)
	}
}
func TestSynchronizedHashChar_Generated_String(t *testing.T) {
	s := NewSynchronizedHashChar()
	s.Add(1)
	if s.String() == "" {
		t.Error("empty")
	}
}
