
package bag

import (
	"testing"
)

func TestFloat64HashBag_Generated_AddOccurrences(t *testing.T) {
	b := NewFloat64HashBag()
	b.Add(1.0)
	b.Add(1.0)
	b.Add(2.0)

	if b.OccurrencesOf(1.0) != 2 {
		t.Errorf("OccurrencesOf(1.0) = %d, want 2", b.OccurrencesOf(1.0))
	}
	if b.OccurrencesOf(2.0) != 1 {
		t.Errorf("OccurrencesOf(2.0) = %d, want 1", b.OccurrencesOf(2.0))
	}
	if b.Size() != 3 {
		t.Errorf("Size = %d, want 3", b.Size())
	}
	if b.SizeDistinct() != 2 {
		t.Errorf("SizeDistinct = %d, want 2", b.SizeDistinct())
	}
}

func TestFloat64HashBag_Generated_Of(t *testing.T) {
	b := Float64HashBagOf(1.0, 1.0, 2.0)
	if b.Size() != 3 {
		t.Errorf("Of: Size = %d, want 3", b.Size())
	}
	if b.OccurrencesOf(1.0) != 2 {
		t.Errorf("Of: OccurrencesOf(1.0) = %d, want 2", b.OccurrencesOf(1.0))
	}
}

func TestFloat64HashBag_Generated_Remove(t *testing.T) {
	b := Float64HashBagOf(1.0, 1.0, 1.0, 2.0)
	b.Remove(1.0)
	if b.OccurrencesOf(1.0) != 2 {
		t.Errorf("After Remove: occurrences = %d, want 2", b.OccurrencesOf(1.0))
	}
}

func TestFloat64HashBag_Generated_RemoveAll(t *testing.T) {
	b := Float64HashBagOf(1.0, 1.0, 1.0, 2.0)
	b.RemoveAll(1.0)
	if b.Contains(1.0) {
		t.Error("After RemoveAll: should not contain 1.0")
	}
	if b.Size() != 1 {
		t.Errorf("After RemoveAll: Size = %d, want 1", b.Size())
	}
}

func TestFloat64HashBag_Generated_Contains(t *testing.T) {
	b := Float64HashBagOf(1.0, 2.0)
	if !b.Contains(1.0) {
		t.Error("Contains(1.0) should be true")
	}
	if b.Contains(99.0) {
		t.Error("Contains(99.0) should be false")
	}
}

func TestFloat64HashBag_Generated_IsEmpty(t *testing.T) {
	b := NewFloat64HashBag()
	if !b.IsEmpty() {
		t.Error("New bag should be empty")
	}
	b.Add(1.0)
	if b.IsEmpty() {
		t.Error("Bag with element should not be empty")
	}
}

func TestFloat64HashBag_Generated_Clear(t *testing.T) {
	b := Float64HashBagOf(1.0, 2.0)
	b.Clear()
	if b.Size() != 0 || !b.IsEmpty() {
		t.Errorf("After Clear: size=%d, empty=%v", b.Size(), b.IsEmpty())
	}
}

func TestFloat64HashBag_Generated_ForEachWithOccurrences(t *testing.T) {
	b := Float64HashBagOf(1.0, 1.0, 2.0, 2.0, 2.0)
	total := 0
	b.ForEachWithOccurrences(func(v float64, count int) {
		total += count
	})
	if total != 5 {
		t.Errorf("Total occurrences = %d, want 5", total)
	}
}

func TestFloat64HashBag_Generated_TopOccurrences(t *testing.T) {
	b := Float64HashBagOf(1.0, 1.0, 1.0, 2.0, 2.0, 3.0)
	top := b.TopOccurrences(1)
	if len(top) != 1 || top[0].Count != 3 {
		t.Errorf("TopOccurrences(1) = %v", top)
	}
}

func TestFloat64HashBag_Generated_All(t *testing.T) {
	b := Float64HashBagOf(1.0, 1.0, 2.0)
	count := 0
	for range b.All() {
		count++
	}
	// All() yields each element including duplicates
	if count != 3 {
		t.Errorf("All count = %d, want 3", count)
	}
}

func TestFloat64HashBag_Generated_Select(t *testing.T) {
	b := Float64HashBagOf(1.0, 2.0, 3.0)
	selected := b.Select(func(v float64) bool { return v > 1.0 })
	if selected.Size() != 2 {
		t.Errorf("Select size = %d, want 2", selected.Size())
	}
}

func TestFloat64HashBag_Generated_AnySatisfy(t *testing.T) {
	b := Float64HashBagOf(1.0, 2.0, 3.0)
	if !b.AnySatisfy(func(v float64) bool { return v == 2.0 }) {
		t.Error("AnySatisfy should be true")
	}
	if b.AnySatisfy(func(v float64) bool { return v == 99.0 }) {
		t.Error("AnySatisfy should be false")
	}
}

func TestFloat64HashBag_Generated_String(t *testing.T) {
	b := Float64HashBagOf(1.0)
	if b.String() == "" {
		t.Error("String should not be empty")
	}
}
