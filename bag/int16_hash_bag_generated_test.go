
package bag

import (
	"testing"
)

func TestInt16HashBag_Generated_AddOccurrences(t *testing.T) {
	b := NewInt16HashBag()
	b.Add(1)
	b.Add(1)
	b.Add(2)

	if b.OccurrencesOf(1) != 2 {
		t.Errorf("OccurrencesOf(1) = %d, want 2", b.OccurrencesOf(1))
	}
	if b.OccurrencesOf(2) != 1 {
		t.Errorf("OccurrencesOf(2) = %d, want 1", b.OccurrencesOf(2))
	}
	if b.Size() != 3 {
		t.Errorf("Size = %d, want 3", b.Size())
	}
	if b.SizeDistinct() != 2 {
		t.Errorf("SizeDistinct = %d, want 2", b.SizeDistinct())
	}
}

func TestInt16HashBag_Generated_Of(t *testing.T) {
	b := Int16HashBagOf(1, 1, 2)
	if b.Size() != 3 {
		t.Errorf("Of: Size = %d, want 3", b.Size())
	}
	if b.OccurrencesOf(1) != 2 {
		t.Errorf("Of: OccurrencesOf(1) = %d, want 2", b.OccurrencesOf(1))
	}
}

func TestInt16HashBag_Generated_Remove(t *testing.T) {
	b := Int16HashBagOf(1, 1, 1, 2)
	b.Remove(1)
	if b.OccurrencesOf(1) != 2 {
		t.Errorf("After Remove: occurrences = %d, want 2", b.OccurrencesOf(1))
	}
}

func TestInt16HashBag_Generated_RemoveAll(t *testing.T) {
	b := Int16HashBagOf(1, 1, 1, 2)
	b.RemoveAll(1)
	if b.Contains(1) {
		t.Error("After RemoveAll: should not contain 1")
	}
	if b.Size() != 1 {
		t.Errorf("After RemoveAll: Size = %d, want 1", b.Size())
	}
}

func TestInt16HashBag_Generated_Contains(t *testing.T) {
	b := Int16HashBagOf(1, 2)
	if !b.Contains(1) {
		t.Error("Contains(1) should be true")
	}
	if b.Contains(99) {
		t.Error("Contains(99) should be false")
	}
}

func TestInt16HashBag_Generated_IsEmpty(t *testing.T) {
	b := NewInt16HashBag()
	if !b.IsEmpty() {
		t.Error("New bag should be empty")
	}
	b.Add(1)
	if b.IsEmpty() {
		t.Error("Bag with element should not be empty")
	}
}

func TestInt16HashBag_Generated_Clear(t *testing.T) {
	b := Int16HashBagOf(1, 2)
	b.Clear()
	if b.Size() != 0 || !b.IsEmpty() {
		t.Errorf("After Clear: size=%d, empty=%v", b.Size(), b.IsEmpty())
	}
}

func TestInt16HashBag_Generated_ForEachWithOccurrences(t *testing.T) {
	b := Int16HashBagOf(1, 1, 2, 2, 2)
	total := 0
	b.ForEachWithOccurrences(func(v int16, count int) {
		total += count
	})
	if total != 5 {
		t.Errorf("Total occurrences = %d, want 5", total)
	}
}

func TestInt16HashBag_Generated_TopOccurrences(t *testing.T) {
	b := Int16HashBagOf(1, 1, 1, 2, 2, 3)
	top := b.TopOccurrences(1)
	if len(top) != 1 || top[0].Count != 3 {
		t.Errorf("TopOccurrences(1) = %v", top)
	}
}

func TestInt16HashBag_Generated_All(t *testing.T) {
	b := Int16HashBagOf(1, 1, 2)
	count := 0
	for range b.All() {
		count++
	}
	// All() yields each element including duplicates
	if count != 3 {
		t.Errorf("All count = %d, want 3", count)
	}
}

func TestInt16HashBag_Generated_Select(t *testing.T) {
	b := Int16HashBagOf(1, 2, 3)
	selected := b.Select(func(v int16) bool { return v > 1 })
	if selected.Size() != 2 {
		t.Errorf("Select size = %d, want 2", selected.Size())
	}
}

func TestInt16HashBag_Generated_AnySatisfy(t *testing.T) {
	b := Int16HashBagOf(1, 2, 3)
	if !b.AnySatisfy(func(v int16) bool { return v == 2 }) {
		t.Error("AnySatisfy should be true")
	}
	if b.AnySatisfy(func(v int16) bool { return v == 99 }) {
		t.Error("AnySatisfy should be false")
	}
}

func TestInt16HashBag_Generated_String(t *testing.T) {
	b := Int16HashBagOf(1)
	if b.String() == "" {
		t.Error("String should not be empty")
	}
}
