package bag

import "testing"

func TestSynchronizedHashFloat64_Generated_AddOccurrences(t *testing.T) {
	s := NewSynchronizedHashFloat64()
	s.Add(1.0)
	s.Add(1.0)
	s.Add(2.0)
	if s.OccurrencesOf(1.0) != 2 {
		t.Errorf("OccurrencesOf = %d", s.OccurrencesOf(1.0))
	}
	if s.Len() != 3 {
		t.Errorf("Size = %d", s.Len())
	}
	if s.SizeDistinct() != 2 {
		t.Errorf("SizeDistinct = %d", s.SizeDistinct())
	}
}
func TestSynchronizedHashFloat64_Generated_Remove(t *testing.T) {
	s := NewSynchronizedHashFloat64()
	s.Add(1.0)
	s.Add(1.0)
	s.Add(2.0)
	if !s.Remove(1.0) {
		t.Error("Remove should return true")
	}
	if s.OccurrencesOf(1.0) != 1 {
		t.Errorf("After remove: %d", s.OccurrencesOf(1.0))
	}
}
func TestSynchronizedHashFloat64_Generated_Contains(t *testing.T) {
	s := NewSynchronizedHashFloat64()
	s.Add(1.0)
	if !s.Contains(1.0) {
		t.Error("Contains should be true")
	}
}
func TestSynchronizedHashFloat64_Generated_Clear(t *testing.T) {
	s := NewSynchronizedHashFloat64()
	s.Add(1.0)
	s.Clear()
	if s.Len() != 0 {
		t.Error("Should be empty")
	}
}
func TestSynchronizedHashFloat64_Generated_All(t *testing.T) {
	s := NewSynchronizedHashFloat64()
	s.Add(1.0)
	s.Add(1.0)
	s.Add(2.0)
	count := 0
	for range s.All() {
		count++
	}
	if count != 3 {
		t.Errorf("All count = %d", count)
	}
}
func TestSynchronizedHashFloat64_Generated_String(t *testing.T) {
	s := NewSynchronizedHashFloat64()
	s.Add(1.0)
	if s.String() == "" {
		t.Error("empty")
	}
}
