
package priority_queue

import (
	"fmt"
	"math"
	"strings"
)

// Float32PriorityQueue is a min-heap priority queue of float32 values.
// O(log n) Push/Pop, O(1) Peek.
type Float32PriorityQueue struct {
	items []float32
}

// NewFloat32PriorityQueue creates a new empty Float32PriorityQueue.
func NewFloat32PriorityQueue() *Float32PriorityQueue {
	return &Float32PriorityQueue{items: make([]float32, 0, 16)}
}

// Float32PriorityQueueOf creates a new Float32PriorityQueue and heapifies the given values in O(n).
func Float32PriorityQueueOf(values ...float32) *Float32PriorityQueue {
	q := &Float32PriorityQueue{items: make([]float32, len(values))}
	copy(q.items, values)
	if len(q.items) > 1 {
		for i := len(q.items)/2 - 1; i >= 0; i-- {
			q.siftDown(i)
		}
	}
	return q
}

// Push adds a value to the heap. O(log n).
func (q *Float32PriorityQueue) Push(value float32) {
	q.items = append(q.items, value)
	q.siftUp(len(q.items) - 1)
}

// Pop removes and returns the smallest element, or an error if empty. O(log n).
func (q *Float32PriorityQueue) Pop() (float32, error) {
	if len(q.items) == 0 {
		return 0.0, fmt.Errorf("Float32PriorityQueue: Pop on empty queue")
	}
	top := q.items[0]
	last := len(q.items) - 1
	q.items[0] = q.items[last]
	q.items = q.items[:last]
	if len(q.items) > 0 {
		q.siftDown(0)
	}
	return top, nil
}

// Peek returns the smallest element without removing it, or an error if empty.
func (q *Float32PriorityQueue) Peek() (float32, error) {
	if len(q.items) == 0 {
		return 0.0, fmt.Errorf("Float32PriorityQueue: Peek on empty queue")
	}
	return q.items[0], nil
}

// Size returns the number of elements in the queue.
func (q *Float32PriorityQueue) Size() int { return len(q.items) }

// IsEmpty returns true if the queue has no elements.
func (q *Float32PriorityQueue) IsEmpty() bool { return len(q.items) == 0 }

// Clear removes all elements.
func (q *Float32PriorityQueue) Clear() { q.items = q.items[:0] }

// Contains returns true if the queue contains the given value. O(n).
func (q *Float32PriorityQueue) Contains(value float32) bool {
	for _, v := range q.items {
		if math.Float32bits(v) == math.Float32bits(value) {
			return true
		}
	}
	return false
}

// ToSlice returns a copy of the internal heap array (NOT sorted).
func (q *Float32PriorityQueue) ToSlice() []float32 {
	out := make([]float32, len(q.items))
	copy(out, q.items)
	return out
}

// DrainSorted pops all elements in ascending order, consuming the queue.
func (q *Float32PriorityQueue) DrainSorted() []float32 {
	out := make([]float32, 0, len(q.items))
	for len(q.items) > 0 {
		v, _ := q.Pop()
		out = append(out, v)
	}
	return out
}

// String returns a string representation in heap-array order.
func (q *Float32PriorityQueue) String() string {
	if len(q.items) == 0 {
		return "[]"
	}
	var sb strings.Builder
	sb.WriteString("[")
	for i, v := range q.items {
		if i > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v", v)
	}
	sb.WriteString("]")
	return sb.String()
}

func (q *Float32PriorityQueue) less(a, b int) bool {
	// Bit-tiebreak comparator: NaN compares as greatest, so it sinks to the
	// bottom of the min-heap and drains last instead of first (raw `<` returns
	// false for any NaN comparison, corrupting heap order).
	return cmpFloat32(q.items[a], q.items[b]) < 0
}

func (q *Float32PriorityQueue) siftUp(start int) {
	i := start
	for i > 0 {
		parent := (i - 1) / 2
		if q.less(i, parent) {
			q.items[i], q.items[parent] = q.items[parent], q.items[i]
			i = parent
		} else {
			break
		}
	}
}

func (q *Float32PriorityQueue) siftDown(start int) {
	i := start
	n := len(q.items)
	for {
		left := 2*i + 1
		if left >= n {
			break
		}
		right := left + 1
		best := left
		if right < n && q.less(right, left) {
			best = right
		}
		if q.less(best, i) {
			q.items[best], q.items[i] = q.items[i], q.items[best]
			i = best
		} else {
			break
		}
	}
}
