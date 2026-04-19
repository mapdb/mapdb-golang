package hashmap

import (
	"testing"
)

func TestInt32Int64HashMap_PutGet(t *testing.T) {
	m := NewInt32Int64HashMap()
	m.Put(1, 100)
	m.Put(2, 200)
	m.Put(3, 300)

	if v, ok := m.Get(1); !ok || v != 100 {
		t.Errorf("Get(1) = (%d, %v), want (100, true)", v, ok)
	}
	if v, ok := m.Get(99); ok || v != 0 {
		t.Errorf("Get(99) = (%d, %v), want (0, false)", v, ok)
	}
	if m.Size() != 3 {
		t.Errorf("Size() = %d, want 3", m.Size())
	}
}

func TestInt32Int64HashMap_PutOverwrite(t *testing.T) {
	m := NewInt32Int64HashMap()
	m.Put(1, 100)
	old, existed := m.Put(1, 200)
	if !existed || old != 100 {
		t.Errorf("Put overwrite = (%d, %v), want (100, true)", old, existed)
	}
	if v, _ := m.Get(1); v != 200 {
		t.Errorf("Get(1) after overwrite = %d, want 200", v)
	}
}

func TestInt32Int64HashMap_Remove(t *testing.T) {
	m := NewInt32Int64HashMap()
	m.Put(1, 100)
	m.Put(2, 200)

	old, ok := m.Remove(1)
	if !ok || old != 100 {
		t.Errorf("Remove(1) = (%d, %v), want (100, true)", old, ok)
	}
	if m.Size() != 1 {
		t.Errorf("Size after remove = %d, want 1", m.Size())
	}
	if m.ContainsKey(1) {
		t.Error("ContainsKey(1) should be false after remove")
	}
}

func TestInt32Int64HashMap_All(t *testing.T) {
	m := NewInt32Int64HashMap()
	m.Put(1, 10)
	m.Put(2, 20)

	count := 0
	sum := int64(0)
	for k, v := range m.All() {
		count++
		sum += int64(k) + v
	}
	if count != 2 || sum != 33 {
		t.Errorf("All: count=%d sum=%d, want count=2 sum=33", count, sum)
	}
}

func TestInt32Int64HashMap_SelectReject(t *testing.T) {
	m := NewInt32Int64HashMap()
	m.Put(1, 10)
	m.Put(2, 20)
	m.Put(3, 30)

	selected := m.Select(func(k int32, v int64) bool { return v > 15 })
	if selected.Size() != 2 {
		t.Errorf("Select size = %d, want 2", selected.Size())
	}

	rejected := m.Reject(func(k int32, v int64) bool { return v > 15 })
	if rejected.Size() != 1 {
		t.Errorf("Reject size = %d, want 1", rejected.Size())
	}
}

func TestInt32Int64HashMap_Entry(t *testing.T) {
	m := NewInt32Int64HashMap()
	v := m.Entry(1).OrInsert(100)
	if v != 100 {
		t.Errorf("OrInsert = %d, want 100", v)
	}
	v = m.Entry(1).OrInsert(200)
	if v != 100 {
		t.Errorf("OrInsert existing = %d, want 100 (original)", v)
	}
}

func TestInt32Int64HashMap_Resize(t *testing.T) {
	m := NewInt32Int64HashMap()
	for i := int32(0); i < 1000; i++ {
		m.Put(i, int64(i*10))
	}
	if m.Size() != 1000 {
		t.Errorf("Size = %d, want 1000", m.Size())
	}
	for i := int32(0); i < 1000; i++ {
		v, ok := m.Get(i)
		if !ok || v != int64(i*10) {
			t.Fatalf("Get(%d) = (%d, %v), want (%d, true)", i, v, ok, i*10)
		}
	}
}

func TestInt32Int64HashMap_GetOrDefault(t *testing.T) {
	m := NewInt32Int64HashMap()
	m.Put(1, 100)
	if v := m.GetOrDefault(1, -1); v != 100 {
		t.Errorf("GetOrDefault(1, -1) = %d, want 100", v)
	}
	if v := m.GetOrDefault(99, -1); v != -1 {
		t.Errorf("GetOrDefault(99, -1) = %d, want -1", v)
	}
}

func TestInt32Int64HashMap_Clear(t *testing.T) {
	m := NewInt32Int64HashMap()
	m.Put(1, 100)
	m.Put(2, 200)
	m.Clear()
	if m.Size() != 0 || !m.IsEmpty() {
		t.Errorf("After Clear: size=%d, empty=%v", m.Size(), m.IsEmpty())
	}
}

func TestInt32Int64HashMap_Predicates(t *testing.T) {
	m := NewInt32Int64HashMap()
	m.Put(1, 10)
	m.Put(2, 20)

	if !m.AnySatisfy(func(k int32, v int64) bool { return v == 20 }) {
		t.Error("AnySatisfy should be true")
	}
	if m.AllSatisfy(func(k int32, v int64) bool { return v > 15 }) {
		t.Error("AllSatisfy should be false")
	}
	if !m.NoneSatisfy(func(k int32, v int64) bool { return v > 100 }) {
		t.Error("NoneSatisfy should be true")
	}
}
