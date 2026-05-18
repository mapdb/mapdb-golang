
package hashmap

import (
	"testing"
)

func TestCharFloat32HashBiMap_Generated_PutGet(t *testing.T) {
	m := NewCharFloat32HashBiMap()
	m.Put(1, 1.0)
	m.Put(2, 2.0)
	m.Put(3, 3.0)

	if v, ok := m.Get(1); !ok || v != 1.0 {
		t.Errorf("Get(1) = (%v, %v), want (1.0, true)", v, ok)
	}
	if v, ok := m.Get(99); ok || v != 0.0 {
		t.Errorf("Get(99) = (%v, %v), want (0.0, false)", v, ok)
	}
	if m.Size() != 3 {
		t.Errorf("Size() = %d, want 3", m.Size())
	}
}

func TestCharFloat32HashBiMap_Generated_GetKey(t *testing.T) {
	m := NewCharFloat32HashBiMap()
	m.Put(1, 1.0)
	m.Put(2, 2.0)

	if k, ok := m.GetKey(1.0); !ok || k != 1 {
		t.Errorf("GetKey(1.0) = (%v, %v), want (1, true)", k, ok)
	}
	if k, ok := m.GetKey(99.0); ok || k != 0 {
		t.Errorf("GetKey(99.0) = (%v, %v), want (0, false)", k, ok)
	}
}

func TestCharFloat32HashBiMap_Generated_PutOverwriteKey(t *testing.T) {
	m := NewCharFloat32HashBiMap()
	m.Put(1, 1.0)
	old, existed := m.Put(1, 2.0)
	if !existed || old != 1.0 {
		t.Errorf("Put overwrite = (%v, %v), want (1.0, true)", old, existed)
	}
	// Forward should map to new value
	if v, _ := m.Get(1); v != 2.0 {
		t.Errorf("Get after overwrite = %v, want 2.0", v)
	}
	// Old value reverse mapping should be gone
	if _, ok := m.GetKey(1.0); ok {
		t.Error("GetKey for old value should be false after overwrite")
	}
	// New value reverse mapping should exist
	if k, ok := m.GetKey(2.0); !ok || k != 1 {
		t.Errorf("GetKey(2.0) = (%v, %v), want (1, true)", k, ok)
	}
}

func TestCharFloat32HashBiMap_Generated_PutOverwriteValue(t *testing.T) {
	m := NewCharFloat32HashBiMap()
	m.Put(1, 1.0)
	m.Put(2, 2.0)
	// Put a new key with an existing value — should evict the old key
	m.Put(3, 1.0)

	// Old key should no longer exist
	if m.ContainsKey(1) {
		t.Error("Old key should be evicted when its value is reassigned to a new key")
	}
	// New key should map to the value
	if v, ok := m.Get(3); !ok || v != 1.0 {
		t.Errorf("Get(3) = (%v, %v), want (1.0, true)", v, ok)
	}
	// Reverse should map value to new key
	if k, ok := m.GetKey(1.0); !ok || k != 3 {
		t.Errorf("GetKey(1.0) = (%v, %v), want (3, true)", k, ok)
	}
	if m.Size() != 2 {
		t.Errorf("Size() = %d, want 2", m.Size())
	}
}

func TestCharFloat32HashBiMap_Generated_Remove(t *testing.T) {
	m := NewCharFloat32HashBiMap()
	m.Put(1, 1.0)
	m.Put(2, 2.0)

	old, ok := m.Remove(1)
	if !ok || old != 1.0 {
		t.Errorf("Remove(1) = (%v, %v), want (1.0, true)", old, ok)
	}
	if m.Size() != 1 {
		t.Errorf("Size after remove = %d, want 1", m.Size())
	}
	// Both directions should be cleared
	if m.ContainsKey(1) {
		t.Error("ContainsKey should be false after Remove")
	}
	if m.ContainsValue(1.0) {
		t.Error("ContainsValue should be false after Remove")
	}

	_, ok = m.Remove(99)
	if ok {
		t.Error("Remove missing key should return false")
	}
}

func TestCharFloat32HashBiMap_Generated_RemoveValue(t *testing.T) {
	m := NewCharFloat32HashBiMap()
	m.Put(1, 1.0)
	m.Put(2, 2.0)

	oldKey, ok := m.RemoveValue(1.0)
	if !ok || oldKey != 1 {
		t.Errorf("RemoveValue(1.0) = (%v, %v), want (1, true)", oldKey, ok)
	}
	if m.Size() != 1 {
		t.Errorf("Size after RemoveValue = %d, want 1", m.Size())
	}
	if m.ContainsKey(1) {
		t.Error("ContainsKey should be false after RemoveValue")
	}
	if m.ContainsValue(1.0) {
		t.Error("ContainsValue should be false after RemoveValue")
	}

	_, ok = m.RemoveValue(99.0)
	if ok {
		t.Error("RemoveValue missing value should return false")
	}
}

func TestCharFloat32HashBiMap_Generated_ContainsKey(t *testing.T) {
	m := NewCharFloat32HashBiMap()
	m.Put(1, 1.0)
	if !m.ContainsKey(1) {
		t.Error("ContainsKey(1) should be true")
	}
	if m.ContainsKey(99) {
		t.Error("ContainsKey(99) should be false")
	}
}

func TestCharFloat32HashBiMap_Generated_ContainsValue(t *testing.T) {
	m := NewCharFloat32HashBiMap()
	m.Put(1, 1.0)
	if !m.ContainsValue(1.0) {
		t.Error("ContainsValue(1.0) should be true")
	}
	if m.ContainsValue(99.0) {
		t.Error("ContainsValue(99.0) should be false")
	}
}

func TestCharFloat32HashBiMap_Generated_IsEmpty(t *testing.T) {
	m := NewCharFloat32HashBiMap()
	if !m.IsEmpty() {
		t.Error("New bi-map should be empty")
	}
	m.Put(1, 1.0)
	if m.IsEmpty() {
		t.Error("Bi-map with entry should not be empty")
	}
}

func TestCharFloat32HashBiMap_Generated_Clear(t *testing.T) {
	m := NewCharFloat32HashBiMap()
	m.Put(1, 1.0)
	m.Put(2, 2.0)
	m.Clear()
	if m.Size() != 0 || !m.IsEmpty() {
		t.Errorf("After Clear: size=%d, empty=%v", m.Size(), m.IsEmpty())
	}
	if m.ContainsKey(1) {
		t.Error("ContainsKey should be false after Clear")
	}
	if m.ContainsValue(1.0) {
		t.Error("ContainsValue should be false after Clear")
	}
}

func TestCharFloat32HashBiMap_Generated_ForEach(t *testing.T) {
	m := NewCharFloat32HashBiMap()
	m.Put(1, 1.0)
	m.Put(2, 2.0)

	count := 0
	m.ForEach(func(k uint16, v float32) {
		count++
	})
	if count != 2 {
		t.Errorf("ForEach count = %d, want 2", count)
	}
}

func TestCharFloat32HashBiMap_Generated_Keys(t *testing.T) {
	m := NewCharFloat32HashBiMap()
	m.Put(1, 1.0)
	m.Put(2, 2.0)

	count := 0
	for range m.Keys() {
		count++
	}
	if count != 2 {
		t.Errorf("Keys count = %d, want 2", count)
	}
}

func TestCharFloat32HashBiMap_Generated_Values(t *testing.T) {
	m := NewCharFloat32HashBiMap()
	m.Put(1, 1.0)
	m.Put(2, 2.0)

	count := 0
	for range m.Values() {
		count++
	}
	if count != 2 {
		t.Errorf("Values count = %d, want 2", count)
	}
}

func TestCharFloat32HashBiMap_Generated_Inverse(t *testing.T) {
	m := NewCharFloat32HashBiMap()
	m.Put(1, 1.0)
	m.Put(2, 2.0)

	inv := m.Inverse()
	if inv.Size() != 2 {
		t.Errorf("Inverse size = %d, want 2", inv.Size())
	}
	if v, ok := inv.Get(1.0); !ok || v != 1 {
		t.Errorf("Inverse.Get(1.0) = (%v, %v), want (1, true)", v, ok)
	}
}

func TestCharFloat32HashBiMap_Generated_Equals(t *testing.T) {
	m1 := NewCharFloat32HashBiMap()
	m1.Put(1, 1.0)
	m1.Put(2, 2.0)

	m2 := NewCharFloat32HashBiMap()
	m2.Put(2, 2.0)
	m2.Put(1, 1.0)

	if !m1.Equals(m2) {
		t.Error("Equal bi-maps should be equal")
	}

	m3 := NewCharFloat32HashBiMap()
	m3.Put(1, 1.0)
	if m1.Equals(m3) {
		t.Error("Different bi-maps should not be equal")
	}
}

func TestCharFloat32HashBiMap_Generated_String(t *testing.T) {
	m := NewCharFloat32HashBiMap()
	if m.String() != "{}" {
		t.Errorf("Empty bi-map String() = %q, want \"{}\"", m.String())
	}
	m.Put(1, 1.0)
	if m.String() == "" {
		t.Error("String should not be empty for non-empty bi-map")
	}
}

func TestCharFloat32HashBiMap_Generated_BijectiveInvariant(t *testing.T) {
	m := NewCharFloat32HashBiMap()
	m.Put(1, 1.0)
	m.Put(2, 2.0)
	m.Put(3, 3.0)

	// Verify bijective: every key maps to a unique value and vice versa
	m.ForEach(func(k uint16, v float32) {
		if rk, ok := m.GetKey(v); !ok || rk != k {
			t.Errorf("Bijective invariant broken: key=%v, value=%v, reverse key=%v", k, v, rk)
		}
	})
}
