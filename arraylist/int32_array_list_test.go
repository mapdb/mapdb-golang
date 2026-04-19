package arraylist

import (
	"testing"
)

func TestInt32ArrayList_AddGet(t *testing.T) {
	l := NewInt32ArrayList()
	l.Add(10)
	l.Add(20)
	l.Add(30)

	if l.Size() != 3 {
		t.Errorf("Size() = %d, want 3", l.Size())
	}
	v, err := l.Get(1)
	if err != nil || v != 20 {
		t.Errorf("Get(1) = (%d, %v), want (20, nil)", v, err)
	}
}

func TestInt32ArrayList_GetOutOfBounds(t *testing.T) {
	l := Int32ArrayListOf(1, 2, 3)
	if _, err := l.Get(-1); err == nil {
		t.Error("Get(-1) should return an error")
	}
	if _, err := l.Get(3); err == nil {
		t.Error("Get(3) on size-3 list should return an error")
	}
}

func TestInt32ArrayList_Set(t *testing.T) {
	l := Int32ArrayListOf(1, 2, 3)
	old, err := l.Set(1, 99)
	if err != nil || old != 2 {
		t.Errorf("Set(1, 99) old = (%d, %v), want (2, nil)", old, err)
	}
	v, _ := l.Get(1)
	if v != 99 {
		t.Errorf("After Set: Get(1) = %d, want 99", v)
	}
}

func TestInt32ArrayList_Remove(t *testing.T) {
	l := Int32ArrayListOf(10, 20, 30)
	removed, err := l.RemoveAtIndex(1)
	if err != nil || removed != 20 || l.Size() != 2 {
		t.Errorf("RemoveAtIndex(1) = (%d, %v), size=%d", removed, err, l.Size())
	}
}

func TestInt32ArrayList_Contains(t *testing.T) {
	l := Int32ArrayListOf(1, 2, 3)
	if !l.Contains(2) {
		t.Error("Contains(2) should be true")
	}
	if l.Contains(99) {
		t.Error("Contains(99) should be false")
	}
}

func TestInt32ArrayList_Sort(t *testing.T) {
	l := Int32ArrayListOf(30, 10, 20)
	l.Sort()
	v0, _ := l.Get(0)
	v1, _ := l.Get(1)
	v2, _ := l.Get(2)
	if v0 != 10 || v1 != 20 || v2 != 30 {
		t.Errorf("After sort: %v", l.ToSlice())
	}
}

func TestInt32ArrayList_BinarySearch(t *testing.T) {
	l := Int32ArrayListOf(10, 20, 30, 40, 50)
	idx, found := l.BinarySearch(30)
	if !found || idx != 2 {
		t.Errorf("BinarySearch(30) = (%d, %v), want (2, true)", idx, found)
	}
	idx, found = l.BinarySearch(25)
	if found {
		t.Errorf("BinarySearch(25) = (%d, %v), want (_, false)", idx, found)
	}
}

func TestInt32ArrayList_All(t *testing.T) {
	l := Int32ArrayListOf(1, 2, 3)
	sum := int32(0)
	for v := range l.All() {
		sum += v
	}
	if sum != 6 {
		t.Errorf("All sum = %d, want 6", sum)
	}
}

func TestInt32ArrayList_Select(t *testing.T) {
	l := Int32ArrayListOf(1, 2, 3, 4, 5)
	evens := l.Select(func(v int32) bool { return v%2 == 0 })
	if evens.Size() != 2 {
		t.Errorf("Select evens size = %d, want 2", evens.Size())
	}
}

func TestInt32ArrayList_Sum(t *testing.T) {
	l := Int32ArrayListOf(1, 2, 3, 4, 5)
	if s := l.Sum(); s != 15 {
		t.Errorf("Sum = %d, want 15", s)
	}
}

func TestInt32ArrayList_MinMax(t *testing.T) {
	l := Int32ArrayListOf(30, 10, 50, 20)
	if min, ok := l.Min(); !ok || min != 10 {
		t.Errorf("Min = (%d, %v), want (10, true)", min, ok)
	}
	if max, ok := l.Max(); !ok || max != 50 {
		t.Errorf("Max = (%d, %v), want (50, true)", max, ok)
	}
}

func TestInt32ArrayList_Resize(t *testing.T) {
	l := NewInt32ArrayList()
	for i := int32(0); i < 1000; i++ {
		l.Add(i)
	}
	if l.Size() != 1000 {
		t.Errorf("Size = %d, want 1000", l.Size())
	}
	if v, err := l.Get(999); err != nil || v != 999 {
		t.Errorf("Get(999) = (%d, %v), want (999, nil)", v, err)
	}
}
