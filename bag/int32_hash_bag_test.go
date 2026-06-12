package bag

import (
	"testing"
)

func TestHashInt32_AddOccurrences(t *testing.T) {
	b := NewHashInt32()
	b.Add(1)
	b.Add(1)
	b.Add(2)

	if b.OccurrencesOf(1) != 2 {
		t.Errorf("OccurrencesOf(1) = %d, want 2", b.OccurrencesOf(1))
	}
	if b.OccurrencesOf(2) != 1 {
		t.Errorf("OccurrencesOf(2) = %d, want 1", b.OccurrencesOf(2))
	}
	if b.Len() != 3 {
		t.Errorf("Size = %d, want 3", b.Len())
	}
	if b.SizeDistinct() != 2 {
		t.Errorf("SizeDistinct = %d, want 2", b.SizeDistinct())
	}
}

func TestHashInt32_Remove(t *testing.T) {
	b := HashInt32Of(1, 1, 1, 2)
	b.Remove(1)
	if b.OccurrencesOf(1) != 2 {
		t.Errorf("After Remove(1): occurrences = %d, want 2", b.OccurrencesOf(1))
	}
	b.RemoveAll(1)
	if b.Contains(1) {
		t.Error("After RemoveAll(1): should not contain 1")
	}
}

func TestHashInt32_ForEachWithOccurrences(t *testing.T) {
	b := HashInt32Of(1, 1, 2, 2, 2)
	total := 0
	b.ForEachWithOccurrences(func(v int32, count int) {
		total += count
	})
	if total != 5 {
		t.Errorf("Total occurrences = %d, want 5", total)
	}
}

func TestHashInt32_TopOccurrences(t *testing.T) {
	b := HashInt32Of(1, 1, 1, 2, 2, 3)
	top := b.TopOccurrences(1)
	if len(top) != 1 || top[0].Count != 3 {
		t.Errorf("TopOccurrences(1) = %v", top)
	}
}
