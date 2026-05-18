
package priority_queue

import (
	"sync"
	"testing"
)

func TestSynchronizedFloat32PriorityQueue_Generated_PushPopPeek(t *testing.T) {
	q := NewSynchronizedFloat32PriorityQueue()
	q.Push(3.0)
	q.Push(1.0)
	q.Push(2.0)
	if q.Size() != 3 {
		t.Errorf("Size = %d", q.Size())
	}
	a, _ := q.Pop()
	b, _ := q.Pop()
	c, _ := q.Pop()
	if a > b || b > c {
		t.Errorf("not ascending: %v %v %v", a, b, c)
	}
}

func TestSynchronizedFloat32PriorityQueue_Generated_EmptyPeekPop(t *testing.T) {
	q := NewSynchronizedFloat32PriorityQueue()
	if _, err := q.Peek(); err == nil {
		t.Error("Peek on empty: want error")
	}
	if _, err := q.Pop(); err == nil {
		t.Error("Pop on empty: want error")
	}
	if !q.IsEmpty() {
		t.Error("IsEmpty should be true")
	}
}

func TestSynchronizedFloat32PriorityQueue_Generated_ContainsClear(t *testing.T) {
	q := NewSynchronizedFloat32PriorityQueue()
	q.Push(1.0)
	if !q.Contains(1.0) {
		t.Error("Contains should be true")
	}
	q.Clear()
	if !q.IsEmpty() {
		t.Error("IsEmpty after Clear")
	}
}

func TestSynchronizedFloat32PriorityQueue_Generated_DrainSorted(t *testing.T) {
	q := NewSynchronizedFloat32PriorityQueue()
	q.Push(3.0)
	q.Push(1.0)
	q.Push(2.0)
	sorted := q.DrainSorted()
	if len(sorted) != 3 {
		t.Errorf("DrainSorted len = %d", len(sorted))
	}
	for i := 1; i < len(sorted); i++ {
		if sorted[i-1] > sorted[i] {
			t.Errorf("not ascending at %d", i)
		}
	}
}

func TestSynchronizedFloat32PriorityQueue_Generated_ConcurrentAccess(t *testing.T) {
	q := NewSynchronizedFloat32PriorityQueue()
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				q.Push(1.0)
				_ = q.Size()
				_, _ = q.Pop()
			}
		}()
	}
	wg.Wait()
}

func TestSynchronizedFloat32PriorityQueue_Generated_String(t *testing.T) {
	q := NewSynchronizedFloat32PriorityQueue()
	q.Push(1.0)
	if q.String() == "" {
		t.Error("empty")
	}
}
