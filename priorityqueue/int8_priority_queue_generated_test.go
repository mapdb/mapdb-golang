package priorityqueue

import (
	"testing"
)

func TestInt8_Generated_PushPopMin(t *testing.T) {
	q := NewInt8()
	q.Push(3)
	q.Push(1)
	q.Push(2)
	if q.Len() != 3 {
		t.Errorf("Size = %d, want 3", q.Len())
	}
	a, _ := q.Pop()
	b, _ := q.Pop()
	c, _ := q.Pop()
	if a > b || b > c {
		t.Errorf("Pop order not ascending: %v, %v, %v", a, b, c)
	}
	if _, ok := q.Pop(); ok {
		t.Errorf("Pop on empty: want empty")
	}
}

func TestInt8_Generated_Peek(t *testing.T) {
	q := Int8Of(3, 1, 2)
	got, ok := q.Peek()
	if !ok {
		t.Error("Peek on non-empty: want ok")
	}
	if got != 1 {
		t.Errorf("Peek = %v, want 1", got)
	}
}

func TestInt8_Generated_EmptyPeekPop(t *testing.T) {
	q := NewInt8()
	if _, ok := q.Peek(); ok {
		t.Errorf("Peek on empty: want empty")
	}
	if _, ok := q.Pop(); ok {
		t.Errorf("Pop on empty: want empty")
	}
	if q.Len() != 0 {
		t.Errorf("IsEmpty = false, want true")
	}
}

func TestInt8_Generated_ContainsClear(t *testing.T) {
	q := NewInt8()
	q.Push(1)
	if !q.Contains(1) {
		t.Errorf("Contains(1) = false, want true")
	}
	q.Clear()
	if q.Len() != 0 {
		t.Errorf("IsEmpty after Clear = false, want true")
	}
}

func TestInt8_Generated_DrainSorted(t *testing.T) {
	q := Int8Of(3, 1, 2)
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
