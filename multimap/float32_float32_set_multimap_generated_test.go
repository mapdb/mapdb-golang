package multimap

import (
	"testing"
)

func TestFloat32Float32Set_Generated_PutGet(t *testing.T) {
	m := NewFloat32Float32Set()
	m.Put(1.0, 1.0)
	m.Put(1.0, 2.0)
	m.Put(2.0, 3.0)

	vals := m.Get(1.0)
	if len(vals) != 2 {
		t.Errorf("Get(1.0) len = %d, want 2", len(vals))
	}
	if len(m.Get(2.0)) != 1 {
		t.Errorf("Get(2.0) len = %d, want 1", len(m.Get(2.0)))
	}
	if len(m.Get(99.0)) != 0 {
		t.Errorf("Get(99.0) len = %d, want 0", len(m.Get(99.0)))
	}
	if m.Len() != 3 {
		t.Errorf("Size() = %d, want 3", m.Len())
	}
	if m.KeysCount() != 2 {
		t.Errorf("KeysCount() = %d, want 2", m.KeysCount())
	}
}

func TestFloat32Float32Set_Generated_GetAll(t *testing.T) {
	m := NewFloat32Float32Set()
	m.Put(1.0, 1.0)
	m.Put(1.0, 2.0)
	vals := m.GetAll(1.0)
	if len(vals) != 2 {
		t.Errorf("GetAll(1.0) len = %d, want 2", len(vals))
	}
}

func TestFloat32Float32Set_Generated_RemoveAll(t *testing.T) {
	m := NewFloat32Float32Set()
	m.Put(1.0, 1.0)
	m.Put(1.0, 2.0)
	m.Put(2.0, 3.0)
	removed := m.RemoveAll(1.0)
	if len(removed) != 2 {
		t.Errorf("RemoveAll returned %d values, want 2", len(removed))
	}
	if m.Len() != 1 {
		t.Errorf("Size() = %d, want 1", m.Len())
	}
	if m.KeysCount() != 1 {
		t.Errorf("KeysCount() = %d, want 1", m.KeysCount())
	}
	if m.ContainsKey(1.0) {
		t.Errorf("ContainsKey(1.0) = true after RemoveAll, want false")
	}
}

func TestFloat32Float32Set_Generated_ContainsKey(t *testing.T) {
	m := NewFloat32Float32Set()
	m.Put(1.0, 1.0)
	if !m.ContainsKey(1.0) {
		t.Errorf("ContainsKey(1.0) = false, want true")
	}
	if m.ContainsKey(99.0) {
		t.Errorf("ContainsKey(99.0) = true, want false")
	}
}

func TestFloat32Float32Set_Generated_ContainsKeyValue(t *testing.T) {
	m := NewFloat32Float32Set()
	m.Put(1.0, 1.0)
	m.Put(1.0, 2.0)
	if !m.ContainsKeyValue(1.0, 1.0) {
		t.Errorf("ContainsKeyValue(1.0, 1.0) = false, want true")
	}
	if m.ContainsKeyValue(1.0, 3.0) {
		t.Errorf("ContainsKeyValue(1.0, 3.0) = true, want false")
	}
}

func TestFloat32Float32Set_Generated_Clear(t *testing.T) {
	m := NewFloat32Float32Set()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)
	m.Clear()
	if m.Len() != 0 {
		t.Errorf("IsEmpty() = false after Clear, want true")
	}
	if m.Len() != 0 {
		t.Errorf("Size() = %d after Clear, want 0", m.Len())
	}
}

func TestFloat32Float32Set_Generated_ForEach(t *testing.T) {
	m := NewFloat32Float32Set()
	m.Put(1.0, 1.0)
	m.Put(1.0, 2.0)
	m.Put(2.0, 3.0)
	count := 0
	m.ForEach(func(_ float32, _ float32) {
		count++
	})
	if count != 3 {
		t.Errorf("ForEach visited %d pairs, want 3", count)
	}
}

func TestFloat32Float32Set_Generated_ForEachKeyValues(t *testing.T) {
	m := NewFloat32Float32Set()
	m.Put(1.0, 1.0)
	m.Put(1.0, 2.0)
	m.Put(2.0, 3.0)
	keyCount := 0
	m.ForEachKeyValues(func(_ float32, vals []float32) {
		keyCount++
		if len(vals) == 0 {
			t.Errorf("ForEachKeyValues passed empty slice")
		}
	})
	if keyCount != 2 {
		t.Errorf("ForEachKeyValues visited %d keys, want 2", keyCount)
	}
}

func TestFloat32Float32Set_Generated_SelectReject(t *testing.T) {
	m := NewFloat32Float32Set()
	m.Put(1.0, 1.0)
	m.Put(1.0, 2.0)
	m.Put(2.0, 3.0)
	sel := m.Select(func(_ float32, v float32) bool { return v == 1.0 })
	if sel.Len() != 1 {
		t.Errorf("Select size = %d, want 1", sel.Len())
	}
	rej := m.Reject(func(_ float32, v float32) bool { return v == 1.0 })
	if rej.Len() != 2 {
		t.Errorf("Reject size = %d, want 2", rej.Len())
	}
}

func TestFloat32Float32Set_Generated_KeysValues(t *testing.T) {
	m := NewFloat32Float32Set()
	m.Put(1.0, 1.0)
	m.Put(1.0, 2.0)
	m.Put(2.0, 3.0)
	keys := m.Keys()
	if len(keys) != 2 {
		t.Errorf("Keys() len = %d, want 2", len(keys))
	}
	vals := m.Values()
	if len(vals) != 3 {
		t.Errorf("Values() len = %d, want 3", len(vals))
	}
}

func TestFloat32Float32Set_Generated_Equals(t *testing.T) {
	m1 := NewFloat32Float32Set()
	m1.Put(1.0, 1.0)
	m1.Put(1.0, 2.0)
	m2 := NewFloat32Float32Set()
	m2.Put(1.0, 1.0)
	m2.Put(1.0, 2.0)
	if !m1.Equals(m2) {
		t.Errorf("Equals() = false, want true")
	}
}

func TestFloat32Float32Set_Generated_String(t *testing.T) {
	m := NewFloat32Float32Set()
	m.Put(1.0, 1.0)
	s := m.String()
	if len(s) == 0 {
		t.Errorf("String() is empty")
	}
}

func TestFloat32Float32Set_Generated_FluentAPI(t *testing.T) {
	m := NewFloat32Float32Set()
	m.PutReturning(1.0, 1.0).PutReturning(1.0, 2.0).PutReturning(2.0, 3.0)
	if m.Len() != 3 {
		t.Errorf("Size() = %d, want 3", m.Len())
	}
	m.RemoveKeyReturning(1.0)
	if m.Len() != 1 {
		t.Errorf("Size() = %d after RemoveKeyReturning, want 1", m.Len())
	}
}
