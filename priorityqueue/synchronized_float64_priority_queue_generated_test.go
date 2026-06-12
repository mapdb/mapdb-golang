package priorityqueue

import (
	"sync"
	"testing"
)

func TestSynchronizedFloat64_Generated_PushPopPeek(t *testing.T) {
	q := NewSynchronizedFloat64()
	q.Push(3.0)
	q.Push(1.0)
	q.Push(2.0)
	if q.Len() != 3 {
		t.Errorf("Size = %d", q.Len())
	}
	a, _ := q.Pop()
	b, _ := q.Pop()
	c, _ := q.Pop()
	if a > b || b > c {
		t.Errorf("not ascending: %v %v %v", a, b, c)
	}
}

func TestSynchronizedFloat64_Generated_EmptyPeekPop(t *testing.T) {
	q := NewSynchronizedFloat64()
	if _, ok := q.Peek(); ok {
		t.Error("Peek on empty: want empty")
	}
	if _, ok := q.Pop(); ok {
		t.Error("Pop on empty: want empty")
	}
	if q.Len() != 0 {
		t.Error("IsEmpty should be true")
	}
}

func TestSynchronizedFloat64_Generated_ContainsClear(t *testing.T) {
	q := NewSynchronizedFloat64()
	q.Push(1.0)
	if !q.Contains(1.0) {
		t.Error("Contains should be true")
	}
	q.Clear()
	if q.Len() != 0 {
		t.Error("IsEmpty after Clear")
	}
}

func TestSynchronizedFloat64_Generated_DrainSorted(t *testing.T) {
	q := NewSynchronizedFloat64()
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

func TestSynchronizedFloat64_Generated_ConcurrentAccess(t *testing.T) {
	q := NewSynchronizedFloat64()
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				q.Push(1.0)
				_ = q.Len()
				_, _ = q.Pop()
			}
		}()
	}
	wg.Wait()
}

func TestSynchronizedFloat64_Generated_String(t *testing.T) {
	q := NewSynchronizedFloat64()
	q.Push(1.0)
	if q.String() == "" {
		t.Error("empty")
	}
}
