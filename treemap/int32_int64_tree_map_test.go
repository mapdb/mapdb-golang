package treemap

import (
	"testing"
)

func TestInt32Int64_PutGet(t *testing.T) {
	m := NewInt32Int64()
	m.Put(3, 30)
	m.Put(1, 10)
	m.Put(2, 20)

	if v, ok := m.Get(2); !ok || v != 20 {
		t.Errorf("Get(2) = (%d, %v), want (20, true)", v, ok)
	}
	if m.Len() != 3 {
		t.Errorf("Size = %d, want 3", m.Len())
	}
}

func TestInt32Int64_SortedIteration(t *testing.T) {
	m := NewInt32Int64()
	m.Put(50, 500)
	m.Put(10, 100)
	m.Put(30, 300)
	m.Put(20, 200)
	m.Put(40, 400)

	var keys []int32
	for k, _ := range m.All() {
		keys = append(keys, k)
	}
	for i := 1; i < len(keys); i++ {
		if keys[i] <= keys[i-1] {
			t.Fatalf("Keys not sorted: %v", keys)
		}
	}
}

func TestInt32Int64_MinMax(t *testing.T) {
	m := NewInt32Int64()
	m.Put(30, 300)
	m.Put(10, 100)
	m.Put(50, 500)

	k, v, ok := m.Min()
	if !ok || k != 10 || v != 100 {
		t.Errorf("Min = (%d, %d, %v), want (10, 100, true)", k, v, ok)
	}
	k, v, ok = m.Max()
	if !ok || k != 50 || v != 500 {
		t.Errorf("Max = (%d, %d, %v), want (50, 500, true)", k, v, ok)
	}
}

func TestInt32Int64_FloorCeiling(t *testing.T) {
	m := NewInt32Int64()
	m.Put(10, 100)
	m.Put(20, 200)
	m.Put(30, 300)

	k, v, ok := m.Floor(25)
	if !ok || k != 20 || v != 200 {
		t.Errorf("Floor(25) = (%d, %d, %v), want (20, 200, true)", k, v, ok)
	}
	k, v, ok = m.Ceiling(25)
	if !ok || k != 30 || v != 300 {
		t.Errorf("Ceiling(25) = (%d, %d, %v), want (30, 300, true)", k, v, ok)
	}
}

func TestInt32Int64_Remove(t *testing.T) {
	m := NewInt32Int64()
	for i := int32(1); i <= 20; i++ {
		m.Put(i, int64(i*10))
	}
	for i := int32(1); i <= 20; i += 2 {
		m.Remove(i)
	}
	if m.Len() != 10 {
		t.Errorf("Size after removes = %d, want 10", m.Len())
	}
	// Verify remaining keys are even and sorted
	prev := int32(0)
	for k, _ := range m.All() {
		if k%2 != 0 {
			t.Errorf("Odd key %d should have been removed", k)
		}
		if k <= prev {
			t.Errorf("Keys not sorted: %d after %d", k, prev)
		}
		prev = k
	}
}

func TestInt32Int64_RangeKeys(t *testing.T) {
	m := NewInt32Int64()
	for i := int32(1); i <= 10; i++ {
		m.Put(i, int64(i*10))
	}
	var keys []int32
	for k, _ := range m.RangeKeys(3, 7) {
		keys = append(keys, k)
	}
	if len(keys) != 4 { // 3, 4, 5, 6
		t.Errorf("RangeKeys(3,7) = %v, want [3,4,5,6]", keys)
	}
}

func TestInt32Int64_LargeInsertDelete(t *testing.T) {
	m := NewInt32Int64()
	for i := int32(0); i < 1000; i++ {
		m.Put(i, int64(i))
	}
	if m.Len() != 1000 {
		t.Fatalf("Size = %d, want 1000", m.Len())
	}
	for i := int32(0); i < 500; i++ {
		m.Remove(i)
	}
	if m.Len() != 500 {
		t.Fatalf("Size after 500 removes = %d, want 500", m.Len())
	}
	// Verify sorted
	prev := int32(-1)
	for k, _ := range m.All() {
		if k <= prev {
			t.Fatalf("Not sorted: %d after %d", k, prev)
		}
		prev = k
	}
}
