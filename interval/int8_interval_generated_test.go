
package interval

import (
	"testing"
)

func TestNewInt8Interval_Ascending(t *testing.T) {
	iv := Int8IntervalFromTo(1, 5)
	if iv.Size() != 5 {
		t.Errorf("Size() = %d, want 5", iv.Size())
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

func TestNewInt8Interval_Descending(t *testing.T) {
	iv := Int8IntervalFromTo(5, 1)
	if iv.Size() != 5 {
		t.Errorf("Size() = %d, want 5", iv.Size())
	}
	got := iv.ToSlice()
	want := []int8{5, 4, 3, 2, 1}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("ToSlice()[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestNewInt8Interval_ByStep(t *testing.T) {
	iv := NewInt8Interval(0, 10, 2)
	if iv.Size() != 6 {
		t.Errorf("Size() = %d, want 6", iv.Size())
	}
	got := iv.ToSlice()
	want := []int8{0, 2, 4, 6, 8, 10}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("ToSlice()[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestNewInt8Interval_NegativeStep(t *testing.T) {
	iv := NewInt8Interval(10, 0, -3)
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

func TestNewInt8Interval_SingleElement(t *testing.T) {
	iv := Int8IntervalFromTo(3, 3)
	if iv.Size() != 1 {
		t.Errorf("Size() = %d, want 1", iv.Size())
	}
	got := iv.ToSlice()
	if got[0] != 3 {
		t.Errorf("ToSlice()[0] = %v, want 3", got[0])
	}
}

func TestInt8Interval_Contains(t *testing.T) {
	iv := NewInt8Interval(0, 10, 2)
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

func TestInt8Interval_BoundaryDoesNotWrap(t *testing.T) {
	iv := Int8IntervalFromTo(int8(126), int8(127))
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

	desc := Int8IntervalFromTo(int8(-127), int8(-128))
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

func TestInt8Interval_Get(t *testing.T) {
	iv := Int8IntervalFromTo(1, 5)
	v, err := iv.Get(0)
	if err != nil || v != 1 {
		t.Errorf("Get(0) = %v, %v", v, err)
	}
	v, err = iv.Get(4)
	if err != nil || v != 5 {
		t.Errorf("Get(4) = %v, %v", v, err)
	}
	_, err = iv.Get(5)
	if err == nil {
		t.Error("Get(5) should return error")
	}
}

func TestInt8Interval_Reversed(t *testing.T) {
	iv := Int8IntervalFromTo(1, 5)
	rev := iv.Reversed()
	got := rev.ToSlice()
	want := []int8{5, 4, 3, 2, 1}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestInt8Interval_ReversedMinimumStepPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Reversed should panic for minimum step")
		}
	}()
	iv := NewInt8Interval(0, int8(-1<<7), int8(-1<<7))
	_ = iv.Reversed()
}

func TestInt8Interval_OneTo(t *testing.T) {
	iv := Int8IntervalOneTo(3)
	got := iv.ToSlice()
	want := []int8{1, 2, 3}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestInt8Interval_ZeroTo(t *testing.T) {
	iv := Int8IntervalZeroTo(3)
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

func TestInt8Interval_IsEmpty(t *testing.T) {
	iv := Int8IntervalFromTo(1, 5)
	if iv.IsEmpty() {
		t.Error("IsEmpty() = true for non-empty interval")
	}
}

func TestInt8Interval_AnySatisfy(t *testing.T) {
	iv := Int8IntervalFromTo(1, 5)
	if !iv.AnySatisfy(func(v int8) bool { return v == 3 }) {
		t.Error("AnySatisfy should find 3")
	}
	if iv.AnySatisfy(func(v int8) bool { return v == 99 }) {
		t.Error("AnySatisfy should not find 99")
	}
}

func TestInt8Interval_AllSatisfy(t *testing.T) {
	iv := Int8IntervalFromTo(1, 5)
	if !iv.AllSatisfy(func(v int8) bool { return v > 0 }) {
		t.Error("AllSatisfy should be true for v > 0")
	}
	if iv.AllSatisfy(func(v int8) bool { return v > 3 }) {
		t.Error("AllSatisfy should be false for v > 3")
	}
}

func TestInt8Interval_NoneSatisfy(t *testing.T) {
	iv := Int8IntervalFromTo(1, 5)
	if !iv.NoneSatisfy(func(v int8) bool { return v > 10 }) {
		t.Error("NoneSatisfy should be true for v > 10")
	}
}

func TestInt8Interval_String(t *testing.T) {
	iv := Int8IntervalFromTo(1, 3)
	s := iv.String()
	if s == "" {
		t.Error("String() is empty")
	}
}

func TestInt8Interval_ZeroStepPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewInt8Interval with step=0 should panic")
		}
	}()
	NewInt8Interval(1, 5, 0)
}

func TestInt8Interval_ForEach(t *testing.T) {
	iv := Int8IntervalFromTo(1, 3)
	sum := int8(0)
	iv.ForEach(func(v int8) { sum += v })
	if sum != 6 {
		t.Errorf("ForEach sum = %v, want 6", sum)
	}
}
