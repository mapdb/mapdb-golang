package deque

import (
	"sync"
	"testing"
)

func TestSynchronizedInt16_Generated_AddRemove(t *testing.T) {
	d := NewSynchronizedInt16()
	d.AddLast(1)
	d.AddLast(2)
	d.AddFirst(3)
	if d.Len() != 3 {
		t.Errorf("Size = %d", d.Len())
	}
	v0, ok := d.RemoveFirst()
	if !ok || v0 != 3 {
		t.Errorf("RemoveFirst = (%v, %v)", v0, ok)
	}
	v1, ok := d.RemoveLast()
	if !ok || v1 != 2 {
		t.Errorf("RemoveLast = (%v, %v)", v1, ok)
	}
}

func TestSynchronizedInt16_Generated_IsEmpty(t *testing.T) {
	d := NewSynchronizedInt16()
	if d.Len() != 0 {
		t.Error("Should be empty")
	}
}

func TestSynchronizedInt16_Generated_PeekContainsClear(t *testing.T) {
	d := NewSynchronizedInt16()
	d.AddLast(1)
	if p, ok := d.PeekFirst(); !ok || p != 1 {
		t.Errorf("PeekFirst = (%v, %v)", p, ok)
	}
	if p, ok := d.PeekLast(); !ok || p != 1 {
		t.Errorf("PeekLast = (%v, %v)", p, ok)
	}
	if !d.Contains(1) {
		t.Error("Contains should be true")
	}
	d.Clear()
	if d.Len() != 0 {
		t.Error("Should be empty after Clear")
	}
}

func TestSynchronizedInt16_Generated_ForEach(t *testing.T) {
	d := NewSynchronizedInt16()
	d.AddLast(1)
	d.AddLast(2)
	count := 0
	d.ForEach(func(_ int16) { count++ })
	if count != 2 {
		t.Errorf("ForEach count = %d", count)
	}
}

func TestSynchronizedInt16_Generated_ConcurrentAccess(t *testing.T) {
	d := NewSynchronizedInt16()
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				d.AddLast(1)
				_ = d.Len()
				_, _ = d.RemoveFirst()
			}
		}()
	}
	wg.Wait()
}

func TestSynchronizedInt16_Generated_String(t *testing.T) {
	d := NewSynchronizedInt16()
	d.AddLast(1)
	if d.String() == "" {
		t.Error("empty")
	}
}
