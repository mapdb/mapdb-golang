
package deque

import (
	"sync"
	"testing"
)

func TestSynchronizedInt8ArrayDeque_Generated_AddRemove(t *testing.T) {
	d := NewSynchronizedInt8ArrayDeque()
	d.AddLast(1)
	d.AddLast(2)
	d.AddFirst(3)
	if d.Size() != 3 {
		t.Errorf("Size = %d", d.Size())
	}
	v0, err := d.RemoveFirst()
	if err != nil || v0 != 3 {
		t.Errorf("RemoveFirst = (%v, %v)", v0, err)
	}
	v1, err := d.RemoveLast()
	if err != nil || v1 != 2 {
		t.Errorf("RemoveLast = (%v, %v)", v1, err)
	}
}

func TestSynchronizedInt8ArrayDeque_Generated_IsEmpty(t *testing.T) {
	d := NewSynchronizedInt8ArrayDeque()
	if !d.IsEmpty() {
		t.Error("Should be empty")
	}
}

func TestSynchronizedInt8ArrayDeque_Generated_PeekContainsClear(t *testing.T) {
	d := NewSynchronizedInt8ArrayDeque()
	d.AddLast(1)
	if p, err := d.PeekFirst(); err != nil || p != 1 {
		t.Errorf("PeekFirst = (%v, %v)", p, err)
	}
	if p, err := d.PeekLast(); err != nil || p != 1 {
		t.Errorf("PeekLast = (%v, %v)", p, err)
	}
	if !d.Contains(1) {
		t.Error("Contains should be true")
	}
	d.Clear()
	if !d.IsEmpty() {
		t.Error("Should be empty after Clear")
	}
}

func TestSynchronizedInt8ArrayDeque_Generated_ForEach(t *testing.T) {
	d := NewSynchronizedInt8ArrayDeque()
	d.AddLast(1)
	d.AddLast(2)
	count := 0
	d.ForEach(func(_ int8) { count++ })
	if count != 2 {
		t.Errorf("ForEach count = %d", count)
	}
}

func TestSynchronizedInt8ArrayDeque_Generated_ConcurrentAccess(t *testing.T) {
	d := NewSynchronizedInt8ArrayDeque()
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				d.AddLast(1)
				_ = d.Size()
				_, _ = d.RemoveFirst()
			}
		}()
	}
	wg.Wait()
}

func TestSynchronizedInt8ArrayDeque_Generated_String(t *testing.T) {
	d := NewSynchronizedInt8ArrayDeque()
	d.AddLast(1)
	if d.String() == "" {
		t.Error("empty")
	}
}
