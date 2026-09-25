package sentinelhashmap

import (
	"math"
	"testing"
	"time"
)

func boundedInt32MapOperation(t *testing.T, operation func() (int32, bool)) (int32, bool) {
	t.Helper()
	type result struct {
		value int32
		ok    bool
	}
	done := make(chan result, 1)
	go func() {
		value, ok := operation()
		done <- result{value, ok}
	}()
	select {
	case got := <-done:
		return got.value, got.ok
	case <-time.After(2 * time.Second):
		t.Fatal("map operation did not return within 2 seconds")
		return 0, false
	}
}

// The old count-only resize rule allowed these 11 removals and five inserts to
// leave no empty physical slot in a capacity-16 table, despite Len() being 5.
func tombstoneChurnInt32() *Int32Int32 {
	m := NewInt32Int32WithCapacity(16)
	for _, key := range []int32{18, 25, 50, 57, 82, 89, 114, 121, 146, 153, 178} {
		m.Put(key, key)
	}
	for _, key := range []int32{18, 25, 50, 57, 82, 89, 114, 121, 146, 153, 178} {
		m.Remove(key)
	}
	for _, key := range []int32{19, 10, 12, 21, 11} {
		m.Put(key, key)
	}
	return m
}

func TestInt32Int32TombstoneChurn(t *testing.T) {
	for _, operation := range []string{"get", "remove", "put"} {
		t.Run(operation, func(t *testing.T) {
			m := tombstoneChurnInt32()
			if got := m.Len(); got != 5 {
				t.Fatalf("Len() = %d, want 5", got)
			}
			if got := len(m.keys); got != 16 {
				t.Fatalf("capacity = %d, want 16 after churn", got)
			}
			switch operation {
			case "get":
				if value, ok := boundedInt32MapOperation(t, func() (int32, bool) { return m.Get(16) }); ok || value != 0 {
					t.Fatalf("Get(missing) = (%d, %v)", value, ok)
				}
			case "remove":
				if value, ok := boundedInt32MapOperation(t, func() (int32, bool) { return m.Remove(16) }); ok || value != 0 {
					t.Fatalf("Remove(missing) = (%d, %v)", value, ok)
				}
			case "put":
				if value, ok := boundedInt32MapOperation(t, func() (int32, bool) { return m.Put(16, 160) }); ok || value != 0 {
					t.Fatalf("Put(new) = (%d, %v)", value, ok)
				}
				if value, ok := m.Get(16); !ok || value != 160 {
					t.Fatalf("Get(new) = (%d, %v)", value, ok)
				}
			}
			for _, key := range []int32{19, 10, 12, 21, 11} {
				if value, ok := m.Get(key); !ok || value != key {
					t.Fatalf("Get(%d) = (%d, %v)", key, value, ok)
				}
			}
		})
	}
}

func TestInt32Int32FullPhysicalTable(t *testing.T) {
	m := NewInt32Int32WithCapacity(16)
	m.Put(0, 100)
	m.Put(1, 101)
	m.Put(2, 102)
	for i, key := range m.keys {
		if key == int32Int32EmptyKey {
			m.keys[i] = int32Int32RemovedKey
			m.tombstones++
		}
	}
	if m.tombstones != 15 {
		t.Fatalf("tombstones = %d, want 15", m.tombstones)
	}
	if value, ok := boundedInt32MapOperation(t, func() (int32, bool) { return m.Get(16) }); ok || value != 0 {
		t.Fatalf("Get(missing) = (%d, %v)", value, ok)
	}
	if value, ok := boundedInt32MapOperation(t, func() (int32, bool) { return m.Remove(16) }); ok || value != 0 {
		t.Fatalf("Remove(missing) = (%d, %v)", value, ok)
	}
	if value, ok := boundedInt32MapOperation(t, func() (int32, bool) { return m.Put(16, 160) }); ok || value != 0 {
		t.Fatalf("Put(new) = (%d, %v)", value, ok)
	}
	for _, entry := range []struct{ key, value int32 }{
		{0, 100}, {1, 101}, {2, 102}, {16, 160},
	} {
		if value, ok := m.Get(entry.key); !ok || value != entry.value {
			t.Fatalf("Get(%d) = (%d, %v), want %d", entry.key, value, ok, entry.value)
		}
	}
	if m.tombstones != 0 || len(m.keys) != 16 {
		t.Fatalf("rehash left tombstones=%d capacity=%d", m.tombstones, len(m.keys))
	}
}

func TestInt32Int32UpdatePastTombstone(t *testing.T) {
	m := NewInt32Int32WithCapacity(16)
	m.Put(18, 180)
	m.Put(25, 250) // Same bucket as 18; the tombstone stays before this key.
	m.Remove(18)
	if old, ok := m.Put(25, 251); !ok || old != 250 {
		t.Fatalf("Put(existing) = (%d, %v)", old, ok)
	}
	if m.Len() != 1 || m.tombstones != 1 {
		t.Fatalf("update changed size=%d tombstones=%d", m.Len(), m.tombstones)
	}
	if value, ok := m.Get(25); !ok || value != 251 {
		t.Fatalf("Get(updated) = (%d, %v)", value, ok)
	}
}

func TestInt32Int32NearLimitChurnAmortizesRehash(t *testing.T) {
	for _, scenario := range []struct {
		name        string
		cycles      int
		replacement func(int) (int32, int32)
		maxRebuilds int
	}{
		{"same key", 1000, func(int) (int32, int32) { return 2, 2 }, 0},
		{"different keys", 1024, func(i int) (int32, int32) { return int32(2 + i), int32(5000 + i) }, 4},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			const capacity = 4096
			const regularCount = capacity*3/4 - 1
			m := NewInt32Int32WithCapacity(capacity)
			m.Put(0, 100)
			m.Put(1, 101)
			want := make(map[int32]int32, regularCount)
			for key := int32(2); key < int32(regularCount+2); key++ {
				m.Put(key, key)
				want[key] = key
			}
			if len(m.keys) != capacity || m.Len() != regularCount+2 {
				t.Fatalf("initial capacity=%d Len=%d", len(m.keys), m.Len())
			}
			previousKeys := m.keys
			rebuilds := 0
			for i := 0; i < scenario.cycles; i++ {
				oldKey, newKey := scenario.replacement(i)
				if old, ok := m.Remove(oldKey); !ok || old != want[oldKey] {
					t.Fatalf("cycle %d Remove(%d) = (%d, %v)", i, oldKey, old, ok)
				}
				delete(want, oldKey)
				if old, ok := m.Put(newKey, newKey); ok || old != 0 {
					t.Fatalf("cycle %d Put(%d) = (%d, %v)", i, newKey, old, ok)
				}
				want[newKey] = newKey
				if &m.keys[0] != &previousKeys[0] {
					rebuilds++
					previousKeys = m.keys
				}
				if m.Len() != regularCount+2 || len(m.keys) != capacity {
					t.Fatalf("cycle %d capacity=%d Len=%d", i, len(m.keys), m.Len())
				}
			}
			if rebuilds > scenario.maxRebuilds {
				t.Fatalf("%d rebuilds in %d cycles, want at most %d", rebuilds, scenario.cycles, scenario.maxRebuilds)
			}
			for key, expected := range want {
				if value, ok := m.Get(key); !ok || value != expected {
					t.Fatalf("Get(%d) = (%d, %v), want %d", key, value, ok, expected)
				}
			}
			for _, entry := range []struct{ key, value int32 }{{0, 100}, {1, 101}} {
				if value, ok := m.Get(entry.key); !ok || value != entry.value {
					t.Fatalf("Get(%d) = (%d, %v), want %d", entry.key, value, ok, entry.value)
				}
			}
		})
	}
}

func TestInt8Int64TombstonesAndSentinels(t *testing.T) {
	m := NewInt8Int64WithCapacity(16)
	m.Put(0, 100)
	m.Put(1, 101)
	for key := int8(2); key < 11; key++ {
		m.Put(key, int64(key))
		m.Remove(key)
	}
	for key := int8(11); key < 20; key++ {
		m.Put(key, int64(key)*10)
	}
	if got := len(m.keys); got != 16 {
		t.Fatalf("capacity grew under churn: %d", got)
	}
	if value, ok := m.Get(0); !ok || value != 100 {
		t.Fatalf("Get(0) = (%d, %v)", value, ok)
	}
	if value, ok := m.Get(1); !ok || value != 101 {
		t.Fatalf("Get(1) = (%d, %v)", value, ok)
	}
	if value, ok := m.Remove(19); !ok || value != 190 {
		t.Fatalf("Remove(19) = (%d, %v)", value, ok)
	}
	if got := m.Len(); got != 10 {
		t.Fatalf("Len() = %d, want 10", got)
	}
	m.Clear()
	if m.Len() != 0 {
		t.Fatalf("Clear left size=%d", m.Len())
	}
	if value, ok := m.Put(20, 200); ok || value != 0 {
		t.Fatalf("Put after Clear = (%d, %v)", value, ok)
	}
}

func TestFloat64Int32TombstonesAndSpecialKeys(t *testing.T) {
	m := NewFloat64Int32WithCapacity(16)
	negativeZero := math.Copysign(0, -1)
	nan := math.Float64frombits(0x7ff8000000000001)
	for _, entry := range []struct {
		key   float64
		value int32
	}{
		{0, 10}, {negativeZero, 20}, {1, 30}, {nan, 40},
	} {
		m.Put(entry.key, entry.value)
	}
	for key := float64(2); key < 11; key++ {
		m.Put(key, int32(key))
		m.Remove(key)
	}
	for key := float64(11); key < 19; key++ {
		m.Put(key, int32(key))
	}
	for _, entry := range []struct {
		key   float64
		value int32
	}{
		{0, 10}, {negativeZero, 20}, {1, 30}, {nan, 40},
	} {
		if value, ok := m.Get(entry.key); !ok || value != entry.value {
			t.Fatalf("Get(%v) = (%d, %v), want %d", entry.key, value, ok, entry.value)
		}
	}
	if value, ok := m.Remove(nan); !ok || value != 40 {
		t.Fatalf("Remove(NaN) = (%d, %v)", value, ok)
	}
	if value, ok := m.Get(1000); ok || value != 0 {
		t.Fatalf("Get(missing) = (%d, %v)", value, ok)
	}
}
