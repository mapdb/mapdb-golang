package priorityqueue

import "testing"

// TestInt32Of_IsTheHeapPump documents that Of already performs an O(n) Floyd
// heapify (the priority queue's bulk-load path — no separate pump is needed) and
// produces a valid min-heap whose drain order is fully sorted.
func TestInt32Of_IsTheHeapPump(t *testing.T) {
	q := Int32Of(5, 1, 9, 3, 7, 2, 8, 4, 6, 0)
	if q.Len() != 10 {
		t.Fatalf("len %d", q.Len())
	}
	prev := int32(-1)
	for q.Len() > 0 {
		v, ok := q.Pop()
		if !ok {
			t.Fatal("unexpected empty")
		}
		if v < prev {
			t.Fatalf("not sorted: %d after %d", v, prev)
		}
		prev = v
	}
}
