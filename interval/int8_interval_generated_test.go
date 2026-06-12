package interval

import (
	"testing"
)

func TestNewInt8_Ascending(t *testing.T) {
	iv := Int8FromTo(1, 5)
	if iv.Len() != 5 {
		t.Errorf("Size() = %d, want 5", iv.Len())
	}
	got := iv.ToSlice()
	want := []int8{1, 2, 3, 4, 5}
	if len(got) != len(want) {
		t.Fatalf("ToSlice() len = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("ToSlice()[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestNewInt8_Descending(t *testing.T) {
	iv := Int8FromTo(5, 1)
	if iv.Len() != 5 {
		t.Errorf("Size() = %d, want 5", iv.Len())
	}
	got := iv.ToSlice()
	want := []int8{5, 4, 3, 2, 1}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("ToSlice()[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestNewInt8_ByStep(t *testing.T) {
	iv := NewInt8(0, 10, 2)
	if iv.Len() != 6 {
		t.Errorf("Size() = %d, want 6", iv.Len())
	}
	got := iv.ToSlice()
	want := []int8{0, 2, 4, 6, 8, 10}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("ToSlice()[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestNewInt8_NegativeStep(t *testing.T) {
	iv := NewInt8(10, 0, -3)
	got := iv.ToSlice()
	want := []int8{10, 7, 4, 1}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestNewInt8_SingleElement(t *testing.T) {
	iv := Int8FromTo(3, 3)
	if iv.Len() != 1 {
		t.Errorf("Size() = %d, want 1", iv.Len())
	}
	got := iv.ToSlice()
	if got[0] != 3 {
		t.Errorf("ToSlice()[0] = %v, want 3", got[0])
	}
}

func TestInt8_Contains(t *testing.T) {
	iv := NewInt8(0, 10, 2)
	if !iv.Contains(0) {
		t.Error("Contains(0) = false")
	}
	if !iv.Contains(4) {
		t.Error("Contains(4) = false")
	}
	if !iv.Contains(10) {
		t.Error("Contains(10) = false")
	}
	if iv.Contains(3) {
		t.Error("Contains(3) = true")
	}
	if iv.Contains(11) {
		t.Error("Contains(11) = true")
	}
}

func TestInt8_BoundaryDoesNotWrap(t *testing.T) {
	iv := Int8FromTo(int8(126), int8(127))
	got := iv.ToSlice()
	want := []int8{126, 127}
	if len(got) != len(want) {
		t.Fatalf("ToSlice() len = %d, want %d; got %v", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("ToSlice()[%d] = %v, want %v", i, got[i], want[i])
		}
	}

	desc := Int8FromTo(int8(-127), int8(-128))
	got = desc.ToSlice()
	want = []int8{-127, -128}
	if len(got) != len(want) {
		t.Fatalf("descending ToSlice() len = %d, want %d; got %v", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("descending ToSlice()[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestInt8_Get(t *testing.T) {
	iv := Int8FromTo(1, 5)
	v := iv.Get(0)
	if v != 1 {
		t.Errorf("Get(0) = %v", v)
	}
	v = iv.Get(4)
	if v != 5 {
		t.Errorf("Get(4) = %v", v)
	}
	assertPanics(t, func() { iv.Get(5) })
}

func TestInt8_Reversed(t *testing.T) {
	iv := Int8FromTo(1, 5)
	rev := iv.Reversed()
	got := rev.ToSlice()
	want := []int8{5, 4, 3, 2, 1}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestInt8_ReversedMinimumStepPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Reversed should panic for minimum step")
		}
	}()
	iv := NewInt8(0, int8(-1<<7), int8(-1<<7))
	_ = iv.Reversed()
}

func TestInt8_OneTo(t *testing.T) {
	iv := Int8OneTo(3)
	got := iv.ToSlice()
	want := []int8{1, 2, 3}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestInt8_ZeroTo(t *testing.T) {
	iv := Int8ZeroTo(3)
	got := iv.ToSlice()
	want := []int8{0, 1, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestInt8_IsEmpty(t *testing.T) {
	iv := Int8FromTo(1, 5)
	if iv.Len() == 0 {
		t.Error("IsEmpty() = true for non-empty interval")
	}
}

func TestInt8_AnySatisfy(t *testing.T) {
	iv := Int8FromTo(1, 5)
	if !iv.AnySatisfy(func(v int8) bool { return v == 3 }) {
		t.Error("AnySatisfy should find 3")
	}
	if iv.AnySatisfy(func(v int8) bool { return v == 99 }) {
		t.Error("AnySatisfy should not find 99")
	}
}

func TestInt8_AllSatisfy(t *testing.T) {
	iv := Int8FromTo(1, 5)
	if !iv.AllSatisfy(func(v int8) bool { return v > 0 }) {
		t.Error("AllSatisfy should be true for v > 0")
	}
	if iv.AllSatisfy(func(v int8) bool { return v > 3 }) {
		t.Error("AllSatisfy should be false for v > 3")
	}
}

func TestInt8_NoneSatisfy(t *testing.T) {
	iv := Int8FromTo(1, 5)
	if !iv.NoneSatisfy(func(v int8) bool { return v > 10 }) {
		t.Error("NoneSatisfy should be true for v > 10")
	}
}

func TestInt8_String(t *testing.T) {
	iv := Int8FromTo(1, 3)
	s := iv.String()
	if s == "" {
		t.Error("String() is empty")
	}
}

func TestInt8_ZeroStepPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewInt8 with step=0 should panic")
		}
	}()
	NewInt8(1, 5, 0)
}

func TestInt8_ForEach(t *testing.T) {
	iv := Int8FromTo(1, 3)
	sum := int8(0)
	iv.ForEach(func(v int8) { sum += v })
	if sum != 6 {
		t.Errorf("ForEach sum = %v, want 6", sum)
	}
}
