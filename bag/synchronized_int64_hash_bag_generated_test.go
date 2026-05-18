
package bag

import "testing"

func TestSynchronizedInt64HashBag_Generated_AddOccurrences(t *testing.T) {
	s := NewSynchronizedInt64HashBag()
	s.Add(1)
	s.Add(1)
	s.Add(2)
	if s.OccurrencesOf(1) != 2 {
		t.Errorf("OccurrencesOf = %d", s.OccurrencesOf(1))
	}
	if s.Size() != 3 {
		t.Errorf("Size = %d", s.Size())
	}
	if s.SizeDistinct() != 2 {
		t.Errorf("SizeDistinct = %d", s.SizeDistinct())
	}
}
func TestSynchronizedInt64HashBag_Generated_Remove(t *testing.T) {
	s := NewSynchronizedInt64HashBag()
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
func TestSynchronizedInt64HashBag_Generated_Contains(t *testing.T) {
	s := NewSynchronizedInt64HashBag()
	s.Add(1)
	if !s.Contains(1) {
		t.Error("Contains should be true")
	}
}
func TestSynchronizedInt64HashBag_Generated_Clear(t *testing.T) {
	s := NewSynchronizedInt64HashBag()
	s.Add(1)
	s.Clear()
	if !s.IsEmpty() {
		t.Error("Should be empty")
	}
}
func TestSynchronizedInt64HashBag_Generated_All(t *testing.T) {
	s := NewSynchronizedInt64HashBag()
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
func TestSynchronizedInt64HashBag_Generated_String(t *testing.T) {
	s := NewSynchronizedInt64HashBag()
	s.Add(1)
	if s.String() == "" {
		t.Error("empty")
	}
}
