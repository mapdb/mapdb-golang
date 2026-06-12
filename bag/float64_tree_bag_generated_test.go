package bag

import (
	"testing"
)

func TestTreeFloat64_Generated_AddOccurrences(t *testing.T) {
	b := NewTreeFloat64()
	b.Add(1.0)
	b.Add(1.0)
	b.Add(2.0)

	if b.OccurrencesOf(1.0) != 2 {
		t.Errorf("OccurrencesOf(1.0) = %d, want 2", b.OccurrencesOf(1.0))
	}
	if b.OccurrencesOf(2.0) != 1 {
		t.Errorf("OccurrencesOf(2.0) = %d, want 1", b.OccurrencesOf(2.0))
	}
	if b.Len() != 3 {
		t.Errorf("Size = %d, want 3", b.Len())
	}
	if b.SizeDistinct() != 2 {
		t.Errorf("SizeDistinct = %d, want 2", b.SizeDistinct())
	}
}

func TestTreeFloat64_Generated_Of(t *testing.T) {
	b := TreeFloat64Of(1.0, 1.0, 2.0)
	if b.Len() != 3 {
		t.Errorf("Of: Size = %d, want 3", b.Len())
	}
	if b.OccurrencesOf(1.0) != 2 {
		t.Errorf("Of: OccurrencesOf(1.0) = %d, want 2", b.OccurrencesOf(1.0))
	}
}

func TestTreeFloat64_Generated_Remove(t *testing.T) {
	b := TreeFloat64Of(1.0, 1.0, 1.0, 2.0)
	b.Remove(1.0)
	if b.OccurrencesOf(1.0) != 2 {
		t.Errorf("After Remove: occurrences = %d, want 2", b.OccurrencesOf(1.0))
	}
}

func TestTreeFloat64_Generated_RemoveAll(t *testing.T) {
	b := TreeFloat64Of(1.0, 1.0, 1.0, 2.0)
	b.RemoveAll(1.0)
	if b.Contains(1.0) {
		t.Error("After RemoveAll: should not contain 1.0")
	}
	if b.Len() != 1 {
		t.Errorf("After RemoveAll: Size = %d, want 1", b.Len())
	}
}

func TestTreeFloat64_Generated_Contains(t *testing.T) {
	b := TreeFloat64Of(1.0, 2.0)
	if !b.Contains(1.0) {
		t.Error("Contains(1.0) should be true")
	}
	if b.Contains(99.0) {
		t.Error("Contains(99.0) should be false")
	}
}

func TestTreeFloat64_Generated_IsEmpty(t *testing.T) {
	b := NewTreeFloat64()
	if b.Len() != 0 {
		t.Error("New bag should be empty")
	}
	b.Add(1.0)
	if b.Len() == 0 {
		t.Error("Bag with element should not be empty")
	}
}

func TestTreeFloat64_Generated_Clear(t *testing.T) {
	b := TreeFloat64Of(1.0, 2.0)
	b.Clear()
	if b.Len() != 0 {
		t.Errorf("After Clear: size=%d, empty=%v", b.Len(), b.Len() == 0)
	}
}

func TestTreeFloat64_Generated_SortedIteration(t *testing.T) {
	b := TreeFloat64Of(3.0, 1.0, 2.0)
	slice := b.ToSortedSlice()
	if len(slice) != 3 {
		t.Fatalf("ToSortedSlice len = %d, want 3", len(slice))
	}
	for i := 1; i < len(slice); i++ {
		if slice[i] < slice[i-1] {
			t.Errorf("ToSortedSlice not sorted: %v >= %v at index %d", slice[i-1], slice[i], i)
		}
	}
}

func TestTreeFloat64_Generated_MinMax(t *testing.T) {
	b := TreeFloat64Of(3.0, 1.0, 2.0)
	minVal, minOk := b.Min()
	if !minOk || minVal != 1.0 {
		t.Errorf("Min = %v, %v; want 1.0, true", minVal, minOk)
	}
	maxVal, maxOk := b.Max()
	if !maxOk || maxVal != 3.0 {
		t.Errorf("Max = %v, %v; want 3.0, true", maxVal, maxOk)
	}

	empty := NewTreeFloat64()
	_, emptyMinOk := empty.Min()
	if emptyMinOk {
		t.Error("Min on empty bag should return false")
	}
	_, emptyMaxOk := empty.Max()
	if emptyMaxOk {
		t.Error("Max on empty bag should return false")
	}
}

func TestTreeFloat64_Generated_ForEachWithOccurrences(t *testing.T) {
	b := TreeFloat64Of(1.0, 1.0, 2.0, 2.0, 2.0)
	total := 0
	b.ForEachWithOccurrences(func(v float64, count int) {
		total += count
	})
	if total != 5 {
		t.Errorf("Total occurrences = %d, want 5", total)
	}
}

func TestTreeFloat64_Generated_TopOccurrences(t *testing.T) {
	b := TreeFloat64Of(1.0, 1.0, 1.0, 2.0, 2.0, 3.0)
	top := b.TopOccurrences(1)
	if len(top) != 1 || top[0].Count != 3 {
		t.Errorf("TopOccurrences(1) = %v", top)
	}
}

func TestTreeFloat64_Generated_Select(t *testing.T) {
	b := TreeFloat64Of(1.0, 2.0, 3.0)
	selected := b.Select(func(v float64) bool { return v > 1.0 })
	if selected.Len() != 2 {
		t.Errorf("Select size = %d, want 2", selected.Len())
	}
}

func TestTreeFloat64_Generated_AnySatisfy(t *testing.T) {
	b := TreeFloat64Of(1.0, 2.0, 3.0)
	if !b.AnySatisfy(func(v float64) bool { return v == 2.0 }) {
		t.Error("AnySatisfy should be true")
	}
	if b.AnySatisfy(func(v float64) bool { return v == 99.0 }) {
		t.Error("AnySatisfy should be false")
	}
}

func TestTreeFloat64_Generated_String(t *testing.T) {
	b := TreeFloat64Of(1.0)
	str := b.String()
	if str == "" {
		t.Error("String should not be empty")
	}
	empty := NewTreeFloat64()
	if empty.String() != "{}" {
		t.Errorf("Empty String = %q, want {}", empty.String())
	}
}

func TestTreeFloat64_Generated_Equals(t *testing.T) {
	b1 := TreeFloat64Of(1.0, 1.0, 2.0)
	b2 := TreeFloat64Of(1.0, 1.0, 2.0)
	if !b1.Equals(b2) {
		t.Error("Equal bags should be Equals")
	}
	b3 := TreeFloat64Of(1.0, 2.0)
	if b1.Equals(b3) {
		t.Error("Different bags should not be Equals")
	}
}
