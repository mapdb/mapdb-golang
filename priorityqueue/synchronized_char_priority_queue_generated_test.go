package priorityqueue

import (
	"sync"
	"testing"
)

func TestSynchronizedChar_Generated_PushPopPeek(t *testing.T) {
	q := NewSynchronizedChar()
	q.Push(3)
	q.Push(1)
	q.Push(2)
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

func TestSynchronizedChar_Generated_EmptyPeekPop(t *testing.T) {
	q := NewSynchronizedChar()
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

func TestSynchronizedChar_Generated_ContainsClear(t *testing.T) {
	q := NewSynchronizedChar()
	q.Push(1)
	if !q.Contains(1) {
		t.Error("Contains should be true")
	}
	q.Clear()
	if q.Len() != 0 {
		t.Error("IsEmpty after Clear")
	}
}

func TestSynchronizedChar_Generated_DrainSorted(t *testing.T) {
	q := NewSynchronizedChar()
	q.Push(3)
	q.Push(1)
	q.Push(2)
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

func TestSynchronizedChar_Generated_ConcurrentAccess(t *testing.T) {
	q := NewSynchronizedChar()
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				q.Push(1)
				_ = q.Len()
				_, _ = q.Pop()
			}
		}()
	}
	wg.Wait()
}

func TestSynchronizedChar_Generated_String(t *testing.T) {
	q := NewSynchronizedChar()
	q.Push(1)
	if q.String() == "" {
		t.Error("empty")
	}
}
