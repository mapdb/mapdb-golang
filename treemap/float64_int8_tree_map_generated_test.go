package treemap

import "testing"

func TestFloat64Int8_Generated_PutGet(t *testing.T) {
	m := NewFloat64Int8()
	m.Put(3.0, 3)
	m.Put(1.0, 1)
	m.Put(2.0, 2)
	if m.Len() != 3 {
		t.Errorf("Size = %d", m.Len())
	}
	if v, ok := m.Get(2.0); !ok || v != 2 {
		t.Errorf("Get = (%v,%v)", v, ok)
	}
	if _, ok := m.Get(99.0); ok {
		t.Error("Get missing should be false")
	}
}
func TestFloat64Int8_Generated_Remove(t *testing.T) {
	m := NewFloat64Int8()
	m.Put(1.0, 1)
	m.Put(2.0, 2)
	old, ok := m.Remove(1.0)
	if !ok || old != 1 {
		t.Errorf("Remove = (%v,%v)", old, ok)
	}
	if m.Len() != 1 {
		t.Errorf("Size = %d", m.Len())
	}
}
func TestFloat64Int8_Generated_ContainsKey(t *testing.T) {
	m := NewFloat64Int8()
	m.Put(1.0, 1)
	if !m.ContainsKey(1.0) {
		t.Error("Should contain")
	}
	if m.ContainsKey(99.0) {
		t.Error("Should not contain")
	}
}
func TestFloat64Int8_Generated_MinMax(t *testing.T) {
	m := NewFloat64Int8()
	m.Put(3.0, 3)
	m.Put(1.0, 1)
	m.Put(2.0, 2)
	if k, _, ok := m.Min(); !ok || k != 1.0 {
		t.Errorf("Min key = %v", k)
	}
	if k, _, ok := m.Max(); !ok || k != 3.0 {
		t.Errorf("Max key = %v", k)
	}
}
func TestFloat64Int8_Generated_FloorCeiling(t *testing.T) {
	m := NewFloat64Int8()
	m.Put(1.0, 1)
	m.Put(3.0, 3)
	if k, _, ok := m.Floor(2.0); !ok || k != 1.0 {
		t.Errorf("Floor = %v", k)
	}
	if k, _, ok := m.Ceiling(2.0); !ok || k != 3.0 {
		t.Errorf("Ceiling = %v", k)
	}
}
func TestFloat64Int8_Generated_Clear(t *testing.T) {
	m := NewFloat64Int8()
	m.Put(1.0, 1)
	m.Clear()
	if m.Len() != 0 {
		t.Error("Should be empty")
	}
}
func TestFloat64Int8_Generated_All(t *testing.T) {
	m := NewFloat64Int8()
	m.Put(1.0, 1)
	m.Put(2.0, 2)
	count := 0
	for range m.All() {
		count++
	}
	if count != 2 {
		t.Errorf("All count = %d", count)
	}
}
func TestFloat64Int8_Generated_SortedOrder(t *testing.T) {
	m := NewFloat64Int8()
	m.Put(3.0, 3)
	m.Put(1.0, 1)
	m.Put(2.0, 2)
	var keys []float64
	for k := range m.Keys() {
		keys = append(keys, k)
	}
	for i := 1; i < len(keys); i++ {
		if keys[i] < keys[i-1] {
			t.Errorf("Keys not sorted at %d: %v", i, keys)
		}
	}
}
func TestFloat64Int8_Generated_Select(t *testing.T) {
	m := NewFloat64Int8()
	m.Put(1.0, 1)
	m.Put(2.0, 2)
	m.Put(3.0, 3)
	sel := m.Select(func(k float64, v int8) bool { return v > 1 })
	if sel.Len() != 2 {
		t.Errorf("Select size = %d", sel.Len())
	}
}
func TestFloat64Int8_Generated_String(t *testing.T) {
	m := NewFloat64Int8()
	m.Put(1.0, 1)
	if m.String() == "" {
		t.Error("empty")
	}
}
func TestFloat64Int8_Generated_HigherLower(t *testing.T) {
	m := NewFloat64Int8()
	m.Put(1.0, 1)
	m.Put(3.0, 3)
	m.Put(5.0, 5)
	// Higher(kv[0]) must skip kv[0] and return the next up.
	if k, _, ok := m.Higher(1.0); !ok || k != 3.0 {
		t.Errorf("Higher(1.0) = (%v, %v), want (3.0, true)", k, ok)
	}
	// Lower(kv[4]) must skip kv[4] itself.
	if k, _, ok := m.Lower(5.0); !ok || k != 3.0 {
		t.Errorf("Lower(5.0) = (%v, %v), want (3.0, true)", k, ok)
	}
	// At the extremes, there is no higher/lower.
	if _, _, ok := m.Higher(5.0); ok {
		t.Error("Higher of max should be false")
	}
	if _, _, ok := m.Lower(1.0); ok {
		t.Error("Lower of min should be false")
	}
}
func TestFloat64Int8_Generated_HeadTailSubMap(t *testing.T) {
	m := NewFloat64Int8()
	for i, k := range []float64{1.0, 2.0, 3.0, 4.0, 5.0} {
		v := []int8{1, 2, 3, 4, 5}[i]
		m.Put(k, v)
	}
	// HeadMap: keys < kv[2] → kv[0], kv[1]
	headCount := 0
	for range m.HeadMap(3.0) {
		headCount++
	}
	if headCount != 2 {
		t.Errorf("HeadMap count = %d, want 2", headCount)
	}
	// TailMap: keys >= kv[2] → kv[2], kv[3], kv[4]
	tailCount := 0
	for range m.TailMap(3.0) {
		tailCount++
	}
	if tailCount != 3 {
		t.Errorf("TailMap count = %d, want 3", tailCount)
	}
	// SubMap: [kv[1], kv[3]) → kv[1], kv[2]
	subCount := 0
	for range m.SubMap(2.0, 4.0) {
		subCount++
	}
	if subCount != 2 {
		t.Errorf("SubMap count = %d, want 2", subCount)
	}
}
func TestFloat64Int8_Generated_FirstLastEntry(t *testing.T) {
	m := NewFloat64Int8()
	m.Put(2.0, 2)
	m.Put(1.0, 1)
	m.Put(3.0, 3)
	if k, _, ok := m.FirstEntry(); !ok || k != 1.0 {
		t.Errorf("FirstEntry key = %v, want 1.0", k)
	}
	if k, _, ok := m.LastEntry(); !ok || k != 3.0 {
		t.Errorf("LastEntry key = %v, want 3.0", k)
	}
}
func TestFloat64Int8_Generated_PollFirstLastEntry(t *testing.T) {
	m := NewFloat64Int8()
	m.Put(1.0, 1)
	m.Put(2.0, 2)
	m.Put(3.0, 3)
	if k, _, ok := m.PollFirstEntry(); !ok || k != 1.0 {
		t.Errorf("PollFirstEntry key = %v", k)
	}
	if m.Len() != 2 {
		t.Errorf("Size after PollFirst = %d", m.Len())
	}
	if k, _, ok := m.PollLastEntry(); !ok || k != 3.0 {
		t.Errorf("PollLastEntry key = %v", k)
	}
	if m.Len() != 1 {
		t.Errorf("Size after PollLast = %d", m.Len())
	}
}
func TestFloat64Int8_Generated_DescendingOrder(t *testing.T) {
	m := NewFloat64Int8()
	m.Put(1.0, 1)
	m.Put(2.0, 2)
	m.Put(3.0, 3)
	var keys []float64
	for k := range m.DescendingKeys() {
		keys = append(keys, k)
	}
	for i := 1; i < len(keys); i++ {
		if keys[i] > keys[i-1] {
			t.Errorf("Descending not ordered at %d: %v", i, keys)
		}
	}
	// DescendingMap walks entries in the same order.
	mapCount := 0
	for range m.DescendingMap() {
		mapCount++
	}
	if mapCount != 3 {
		t.Errorf("DescendingMap count = %d", mapCount)
	}
}
