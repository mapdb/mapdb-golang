package hashmap

import (
	"testing"
)

func TestInt32Int64_Empty(t *testing.T) {
	m := NewInt32Int64()
	if m.Len() != 0 {
		t.Error("New map should be empty")
	}
	if _, ok := m.Get(1); ok {
		t.Error("Get on empty map should return false")
	}
	if _, ok := m.Remove(1); ok {
		t.Error("Remove on empty map should return false")
	}
	if m.String() != "{}" {
		t.Errorf("String = %q, want {}", m.String())
	}
}

func TestInt32Int64_SingleElement(t *testing.T) {
	m := NewInt32Int64()
	m.Put(42, 100)
	if m.Len() != 1 {
		t.Errorf("Size = %d, want 1", m.Len())
	}
	m.Remove(42)
	if m.Len() != 0 {
		t.Error("Should be empty after removing sole element")
	}
}

func TestInt32Int64_RemoveAll(t *testing.T) {
	m := NewInt32Int64()
	for i := int32(0); i < 100; i++ {
		m.Put(i, int64(i))
	}
	for i := int32(0); i < 100; i++ {
		m.Remove(i)
	}
	if m.Len() != 0 {
		t.Errorf("Should be empty after removing all, size=%d", m.Len())
	}
}

func TestInt32Int64_AddToValue(t *testing.T) {
	m := NewInt32Int64()
	v := m.AddToValue(1, 10)
	if v != 10 {
		t.Errorf("AddToValue(1, 10) on empty = %d, want 10", v)
	}
	v = m.AddToValue(1, 5)
	if v != 15 {
		t.Errorf("AddToValue(1, 5) = %d, want 15", v)
	}
}

func TestInt32Int64_UpdateValue(t *testing.T) {
	m := NewInt32Int64()
	v := m.UpdateValue(1, 0, func(old int64) int64 { return old + 10 })
	if v != 10 {
		t.Errorf("UpdateValue on absent = %d, want 10", v)
	}
	v = m.UpdateValue(1, 0, func(old int64) int64 { return old * 2 })
	if v != 20 {
		t.Errorf("UpdateValue on present = %d, want 20", v)
	}
}

func TestInt32Int64_PutReturningRemoveKeyReturning(t *testing.T) {
	m := NewInt32Int64()
	m.PutReturning(1, 10).PutReturning(2, 20).PutReturning(3, 30)
	if m.Len() != 3 {
		t.Errorf("Size after PutReturning = %d, want 3", m.Len())
	}
	m.RemoveKeyReturning(2)
	if m.Len() != 2 || m.ContainsKey(2) {
		t.Errorf("After RemoveKeyReturning(2): size=%d contains=%v", m.Len(), m.ContainsKey(2))
	}
}

func TestInt32Int64_SumOfValues(t *testing.T) {
	m := NewInt32Int64()
	m.Put(1, 10)
	m.Put(2, 20)
	m.Put(3, 30)
	if s := m.SumOfValues(); s != 60 {
		t.Errorf("SumOfValues = %d, want 60", s)
	}
}

func TestInt32Int64_InjectInto(t *testing.T) {
	m := NewInt32Int64()
	m.Put(1, 10)
	m.Put(2, 20)
	sum := m.InjectInto(0, func(acc int64, k int32, v int64) int64 { return acc + v })
	if sum != 30 {
		t.Errorf("InjectInto sum = %d, want 30", sum)
	}
}

func TestInt32Int64_BreakFromIterator(t *testing.T) {
	m := NewInt32Int64()
	for i := int32(0); i < 100; i++ {
		m.Put(i, int64(i))
	}
	count := 0
	for range m.All() {
		count++
		if count == 5 {
			break
		}
	}
	if count != 5 {
		t.Errorf("Break from iterator: count=%d, want 5", count)
	}
}

func TestSynchronizedInt32Int64_ConcurrentAccess(t *testing.T) {
	m := NewSynchronizedInt32Int64()
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := int32(0); i < 1000; i++ {
			m.Put(i, int64(i*10))
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 1000; i++ {
			m.Len()
			m.Get(int32(i))
		}
		done <- true
	}()

	<-done
	<-done
	// If we get here without panic/race, the mutex is working
	if m.Len() != 1000 {
		t.Errorf("Size = %d, want 1000", m.Len())
	}
}
