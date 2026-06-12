package arraylist

import (
	"testing"
)

func TestInt32_AddGet(t *testing.T) {
	l := NewInt32()
	l.Add(10)
	l.Add(20)
	l.Add(30)

	if l.Len() != 3 {
		t.Errorf("Size() = %d, want 3", l.Len())
	}
	v := l.Get(1)
	if v != 20 {
		t.Errorf("Get(1) = %d, want 20", v)
	}
}

func TestInt32_GetOutOfBounds(t *testing.T) {
	l := Int32Of(1, 2, 3)
	assertPanics(t, func() { l.Get(-1) })
	assertPanics(t, func() { l.Get(3) })
}

func TestInt32_Set(t *testing.T) {
	l := Int32Of(1, 2, 3)
	old := l.Set(1, 99)
	if old != 2 {
		t.Errorf("Set(1, 99) old = %d, want 2", old)
	}
	v := l.Get(1)
	if v != 99 {
		t.Errorf("After Set: Get(1) = %d, want 99", v)
	}
}

func TestInt32_Remove(t *testing.T) {
	l := Int32Of(10, 20, 30)
	removed := l.RemoveAtIndex(1)
	if removed != 20 || l.Len() != 2 {
		t.Errorf("RemoveAtIndex(1) = %d, size=%d", removed, l.Len())
	}
}

func TestInt32_Contains(t *testing.T) {
	l := Int32Of(1, 2, 3)
	if !l.Contains(2) {
		t.Error("Contains(2) should be true")
	}
	if l.Contains(99) {
		t.Error("Contains(99) should be false")
	}
}

func TestInt32_Sort(t *testing.T) {
	l := Int32Of(30, 10, 20)
	l.Sort()
	v0 := l.Get(0)
	v1 := l.Get(1)
	v2 := l.Get(2)
	if v0 != 10 || v1 != 20 || v2 != 30 {
		t.Errorf("After sort: %v", l.ToSlice())
	}
}

func TestInt32_BinarySearch(t *testing.T) {
	l := Int32Of(10, 20, 30, 40, 50)
	idx, found := l.BinarySearch(30)
	if !found || idx != 2 {
		t.Errorf("BinarySearch(30) = (%d, %v), want (2, true)", idx, found)
	}
	idx, found = l.BinarySearch(25)
	if found {
		t.Errorf("BinarySearch(25) = (%d, %v), want (_, false)", idx, found)
	}
}

func TestInt32_All(t *testing.T) {
	l := Int32Of(1, 2, 3)
	sum := int32(0)
	for v := range l.All() {
		sum += v
	}
	if sum != 6 {
		t.Errorf("All sum = %d, want 6", sum)
	}
}

func TestInt32_Select(t *testing.T) {
	l := Int32Of(1, 2, 3, 4, 5)
	evens := l.Select(func(v int32) bool { return v%2 == 0 })
	if evens.Len() != 2 {
		t.Errorf("Select evens size = %d, want 2", evens.Len())
	}
}

func TestInt32_Sum(t *testing.T) {
	l := Int32Of(1, 2, 3, 4, 5)
	if s := l.Sum(); s != 15 {
		t.Errorf("Sum = %d, want 15", s)
	}
}

func TestInt32_MinMax(t *testing.T) {
	l := Int32Of(30, 10, 50, 20)
	if min, ok := l.Min(); !ok || min != 10 {
		t.Errorf("Min = (%d, %v), want (10, true)", min, ok)
	}
	if max, ok := l.Max(); !ok || max != 50 {
		t.Errorf("Max = (%d, %v), want (50, true)", max, ok)
	}
}

func TestInt32_Resize(t *testing.T) {
	l := NewInt32()
	for i := int32(0); i < 1000; i++ {
		l.Add(i)
	}
	if l.Len() != 1000 {
		t.Errorf("Size = %d, want 1000", l.Len())
	}
	if v := l.Get(999); v != 999 {
		t.Errorf("Get(999) = %d, want 999", v)
	}
}
