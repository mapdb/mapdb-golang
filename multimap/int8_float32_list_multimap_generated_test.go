
package multimap

import (
	"testing"
)

func TestInt8Float32ListMultimap_Generated_PutGet(t *testing.T) {
	m := NewInt8Float32ListMultimap()
	m.Put(1, 1.0)
	m.Put(1, 2.0)
	m.Put(2, 3.0)

	vals := m.Get(1)
	if len(vals) != 2 {
		t.Errorf("Get(1) len = %d, want 2", len(vals))
	}
	if len(m.Get(2)) != 1 {
		t.Errorf("Get(2) len = %d, want 1", len(m.Get(2)))
	}
	if len(m.Get(99)) != 0 {
		t.Errorf("Get(99) len = %d, want 0", len(m.Get(99)))
	}
	if m.Size() != 3 {
		t.Errorf("Size() = %d, want 3", m.Size())
	}
	if m.KeysCount() != 2 {
		t.Errorf("KeysCount() = %d, want 2", m.KeysCount())
	}
}

func TestInt8Float32ListMultimap_Generated_GetAll(t *testing.T) {
	m := NewInt8Float32ListMultimap()
	m.Put(1, 1.0)
	m.Put(1, 2.0)
	vals := m.GetAll(1)
	if len(vals) != 2 {
		t.Errorf("GetAll(1) len = %d, want 2", len(vals))
	}
}

func TestInt8Float32ListMultimap_Generated_GetDoesNotAlias(t *testing.T) {
	m := NewInt8Float32ListMultimap()
	m.Put(1, 1.0)
	m.Put(1, 2.0)
	vals := m.Get(1)
	vals[0] = 3.0
	if m.ContainsKeyValue(1, 3.0) {
		t.Error("mutating Get result changed stored values")
	}
}

func TestInt8Float32ListMultimap_Generated_RemoveAll(t *testing.T) {
	m := NewInt8Float32ListMultimap()
	m.Put(1, 1.0)
	m.Put(1, 2.0)
	m.Put(2, 3.0)
	removed := m.RemoveAll(1)
	if len(removed) != 2 {
		t.Errorf("RemoveAll returned %d values, want 2", len(removed))
	}
	if m.Size() != 1 {
		t.Errorf("Size() = %d, want 1", m.Size())
	}
	if m.KeysCount() != 1 {
		t.Errorf("KeysCount() = %d, want 1", m.KeysCount())
	}
	if m.ContainsKey(1) {
		t.Errorf("ContainsKey(1) = true after RemoveAll, want false")
	}
}

func TestInt8Float32ListMultimap_Generated_ContainsKey(t *testing.T) {
	m := NewInt8Float32ListMultimap()
	m.Put(1, 1.0)
	if !m.ContainsKey(1) {
		t.Errorf("ContainsKey(1) = false, want true")
	}
	if m.ContainsKey(99) {
		t.Errorf("ContainsKey(99) = true, want false")
	}
}

func TestInt8Float32ListMultimap_Generated_ContainsKeyValue(t *testing.T) {
	m := NewInt8Float32ListMultimap()
	m.Put(1, 1.0)
	m.Put(1, 2.0)
	if !m.ContainsKeyValue(1, 1.0) {
		t.Errorf("ContainsKeyValue(1, 1.0) = false, want true")
	}
	if m.ContainsKeyValue(1, 3.0) {
		t.Errorf("ContainsKeyValue(1, 3.0) = true, want false")
	}
}

func TestInt8Float32ListMultimap_Generated_Clear(t *testing.T) {
	m := NewInt8Float32ListMultimap()
	m.Put(1, 1.0)
	m.Put(2, 2.0)
	m.Clear()
	if !m.IsEmpty() {
		t.Errorf("IsEmpty() = false after Clear, want true")
	}
	if m.Size() != 0 {
		t.Errorf("Size() = %d after Clear, want 0", m.Size())
	}
}

func TestInt8Float32ListMultimap_Generated_ForEach(t *testing.T) {
	m := NewInt8Float32ListMultimap()
	m.Put(1, 1.0)
	m.Put(1, 2.0)
	m.Put(2, 3.0)
	count := 0
	m.ForEach(func(_ int8, _ float32) {
		count++
	})
	if count != 3 {
		t.Errorf("ForEach visited %d pairs, want 3", count)
	}
}

func TestInt8Float32ListMultimap_Generated_ForEachKeyValues(t *testing.T) {
	m := NewInt8Float32ListMultimap()
	m.Put(1, 1.0)
	m.Put(1, 2.0)
	m.Put(2, 3.0)
	keyCount := 0
	m.ForEachKeyValues(func(_ int8, vals []float32) {
		keyCount++
		if len(vals) == 0 {
			t.Errorf("ForEachKeyValues passed empty slice")
		}
		vals[0] = 3.0
	})
	if keyCount != 2 {
		t.Errorf("ForEachKeyValues visited %d keys, want 2", keyCount)
	}
	if m.ContainsKeyValue(1, 3.0) {
		t.Error("mutating ForEachKeyValues slice changed stored values")
	}
}

func TestInt8Float32ListMultimap_Generated_SelectReject(t *testing.T) {
	m := NewInt8Float32ListMultimap()
	m.Put(1, 1.0)
	m.Put(1, 2.0)
	m.Put(2, 3.0)
	sel := m.Select(func(_ int8, v float32) bool { return v == 1.0 })
	if sel.Size() != 1 {
		t.Errorf("Select size = %d, want 1", sel.Size())
	}
	rej := m.Reject(func(_ int8, v float32) bool { return v == 1.0 })
	if rej.Size() != 2 {
		t.Errorf("Reject size = %d, want 2", rej.Size())
	}
}

func TestInt8Float32ListMultimap_Generated_KeysValues(t *testing.T) {
	m := NewInt8Float32ListMultimap()
	m.Put(1, 1.0)
	m.Put(1, 2.0)
	m.Put(2, 3.0)
	keys := m.Keys()
	if len(keys) != 2 {
		t.Errorf("Keys() len = %d, want 2", len(keys))
	}
	vals := m.Values()
	if len(vals) != 3 {
		t.Errorf("Values() len = %d, want 3", len(vals))
	}
}

func TestInt8Float32ListMultimap_Generated_Equals(t *testing.T) {
	m1 := NewInt8Float32ListMultimap()
	m1.Put(1, 1.0)
	m1.Put(1, 2.0)
	m2 := NewInt8Float32ListMultimap()
	m2.Put(1, 1.0)
	m2.Put(1, 2.0)
	if !m1.Equals(m2) {
		t.Errorf("Equals() = false, want true")
	}
}

func TestInt8Float32ListMultimap_Generated_String(t *testing.T) {
	m := NewInt8Float32ListMultimap()
	m.Put(1, 1.0)
	s := m.String()
	if len(s) == 0 {
		t.Errorf("String() is empty")
	}
}

func TestInt8Float32ListMultimap_Generated_FluentAPI(t *testing.T) {
	m := NewInt8Float32ListMultimap()
	m.WithKeyValue(1, 1.0).WithKeyValue(1, 2.0).WithKeyValue(2, 3.0)
	if m.Size() != 3 {
		t.Errorf("Size() = %d, want 3", m.Size())
	}
	m.WithoutKey(1)
	if m.Size() != 1 {
		t.Errorf("Size() = %d after WithoutKey, want 1", m.Size())
	}
}
