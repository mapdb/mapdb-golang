package sentinelhashmap

import (
	"testing"
)

func TestInt32Int64_PutGet(t *testing.T) {
	m := NewInt32Int64()
	m.Put(0, 100) // sentinel key 0
	m.Put(1, 200) // sentinel key 1
	m.Put(2, 300) // regular key

	if v, ok := m.Get(0); !ok || v != 100 {
		t.Errorf("Get(0) = (%d, %v), want (100, true)", v, ok)
	}
	if v, ok := m.Get(1); !ok || v != 200 {
		t.Errorf("Get(1) = (%d, %v), want (200, true)", v, ok)
	}
	if v, ok := m.Get(2); !ok || v != 300 {
		t.Errorf("Get(2) = (%d, %v), want (300, true)", v, ok)
	}
	if m.Len() != 3 {
		t.Errorf("Size() = %d, want 3", m.Len())
	}
}

func TestInt32Int64_RemoveSentinels(t *testing.T) {
	m := NewInt32Int64()
	m.Put(0, 100)
	m.Put(1, 200)
	m.Put(5, 500)

	old, ok := m.Remove(0)
	if !ok || old != 100 {
		t.Errorf("Remove(0) = (%d, %v), want (100, true)", old, ok)
	}
	if m.ContainsKey(0) {
		t.Error("ContainsKey(0) should be false after remove")
	}
	if m.Len() != 2 {
		t.Errorf("Size = %d, want 2", m.Len())
	}
}

func TestInt32Int64_All(t *testing.T) {
	m := NewInt32Int64()
	m.Put(0, 10)
	m.Put(1, 20)
	m.Put(5, 50)

	count := 0
	for range m.All() {
		count++
	}
	if count != 3 {
		t.Errorf("All count = %d, want 3", count)
	}
}

func TestInt32Int64_Resize(t *testing.T) {
	m := NewInt32Int64()
	for i := int32(0); i < 100; i++ {
		m.Put(i, int64(i*10))
	}
	if m.Len() != 100 {
		t.Errorf("Size = %d, want 100", m.Len())
	}
	for i := int32(0); i < 100; i++ {
		v, ok := m.Get(i)
		if !ok || v != int64(i*10) {
			t.Fatalf("Get(%d) = (%d, %v), want (%d, true)", i, v, ok, i*10)
		}
	}
}
