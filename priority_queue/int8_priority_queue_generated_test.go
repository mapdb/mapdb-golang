
package priority_queue

import (
	"testing"
)

func TestInt8PriorityQueue_Generated_PushPopMin(t *testing.T) {
	q := NewInt8PriorityQueue()
	q.Push(3)
	q.Push(1)
	q.Push(2)
	if q.Size() != 3 {
		t.Errorf("Size = %d, want 3", q.Size())
	}
	a, _ := q.Pop()
	b, _ := q.Pop()
	c, _ := q.Pop()
	if a > b || b > c {
		t.Errorf("Pop order not ascending: %v, %v, %v", a, b, c)
	}
	if _, err := q.Pop(); err == nil {
		t.Errorf("Pop on empty: want error")
	}
}

func TestInt8PriorityQueue_Generated_Peek(t *testing.T) {
	q := Int8PriorityQueueOf(3, 1, 2)
	got, err := q.Peek()
	if err != nil {
		t.Errorf("Peek err = %v", err)
	}
	if got != 1 {
		t.Errorf("Peek = %v, want 1", got)
	}
}

func TestInt8PriorityQueue_Generated_EmptyPeekPop(t *testing.T) {
	q := NewInt8PriorityQueue()
	if _, err := q.Peek(); err == nil {
		t.Errorf("Peek on empty: want error")
	}
	if _, err := q.Pop(); err == nil {
		t.Errorf("Pop on empty: want error")
	}
	if !q.IsEmpty() {
		t.Errorf("IsEmpty = false, want true")
	}
}

func TestInt8PriorityQueue_Generated_ContainsClear(t *testing.T) {
	q := NewInt8PriorityQueue()
	q.Push(1)
	if !q.Contains(1) {
		t.Errorf("Contains(1) = false, want true")
	}
	q.Clear()
	if !q.IsEmpty() {
		t.Errorf("IsEmpty after Clear = false, want true")
	}
}

func TestInt8PriorityQueue_Generated_DrainSorted(t *testing.T) {
	q := Int8PriorityQueueOf(3, 1, 2)
	sorted := q.DrainSorted()
	if len(sorted) != 3 {
		t.Errorf("DrainSorted len = %d, want 3", len(sorted))
	}
	for i := 1; i < len(sorted); i++ {
		if sorted[i-1] > sorted[i] {
			t.Errorf("DrainSorted not ascending at %d: %v, %v", i, sorted[i-1], sorted[i])
		}
	}
}
