package deque

import (
	"sync"
	"testing"
)

func TestSynchronizedFloat32_Generated_AddRemove(t *testing.T) {
	d := NewSynchronizedFloat32()
	d.AddLast(1.0)
	d.AddLast(2.0)
	d.AddFirst(3.0)
	if d.Len() != 3 {
		t.Errorf("Size = %d", d.Len())
	}
	v0, ok := d.RemoveFirst()
	if !ok || v0 != 3.0 {
		t.Errorf("RemoveFirst = (%v, %v)", v0, ok)
	}
	v1, ok := d.RemoveLast()
	if !ok || v1 != 2.0 {
		t.Errorf("RemoveLast = (%v, %v)", v1, ok)
	}
}

func TestSynchronizedFloat32_Generated_IsEmpty(t *testing.T) {
	d := NewSynchronizedFloat32()
	if d.Len() != 0 {
		t.Error("Should be empty")
	}
}

func TestSynchronizedFloat32_Generated_PeekContainsClear(t *testing.T) {
	d := NewSynchronizedFloat32()
	d.AddLast(1.0)
	if p, ok := d.PeekFirst(); !ok || p != 1.0 {
		t.Errorf("PeekFirst = (%v, %v)", p, ok)
	}
	if p, ok := d.PeekLast(); !ok || p != 1.0 {
		t.Errorf("PeekLast = (%v, %v)", p, ok)
	}
	if !d.Contains(1.0) {
		t.Error("Contains should be true")
	}
	d.Clear()
	if d.Len() != 0 {
		t.Error("Should be empty after Clear")
	}
}

func TestSynchronizedFloat32_Generated_ForEach(t *testing.T) {
	d := NewSynchronizedFloat32()
	d.AddLast(1.0)
	d.AddLast(2.0)
	count := 0
	d.ForEach(func(_ float32) { count++ })
	if count != 2 {
		t.Errorf("ForEach count = %d", count)
	}
}

func TestSynchronizedFloat32_Generated_ConcurrentAccess(t *testing.T) {
	d := NewSynchronizedFloat32()
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				d.AddLast(1.0)
				_ = d.Len()
				_, _ = d.RemoveFirst()
			}
		}()
	}
	wg.Wait()
}

func TestSynchronizedFloat32_Generated_String(t *testing.T) {
	d := NewSynchronizedFloat32()
	d.AddLast(1.0)
	if d.String() == "" {
		t.Error("empty")
	}
}
