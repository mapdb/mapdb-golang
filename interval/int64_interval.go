package interval

import (
	"fmt"
	"iter"
	"strings"
)

// Int64Interval is a virtual collection representing a range of int64 values
// [from, to] with a given step. No elements are materialised in memory.
type Int64Interval struct {
	from int64
	to   int64
	step int64
}

// NewInt64Interval creates an interval from `from` to `to` (inclusive) with the
// given step. Panics if step is zero or if the step direction doesn't match
// the from/to direction.
func NewInt64Interval(from, to, step int64) *Int64Interval {
	if step == 0 {
		panic("Int64Interval: step must not be zero")
	}
	if from < to && step < 0 {
		panic("Int64Interval: step must be positive when from < to")
	}
	if from > to && step > 0 {
		panic("Int64Interval: step must be negative when from > to")
	}
	return &Int64Interval{from: from, to: to, step: step}
}

// Int64IntervalFromTo creates an interval from `from` to `to` (inclusive) with
// step 1 (ascending) or -1 (descending).
func Int64IntervalFromTo(from, to int64) *Int64Interval {
	var step int64 = 1
	if from > to {
		step = -1
	}
	return &Int64Interval{from: from, to: to, step: step}
}

// Int64IntervalOneTo creates an interval from 1 to `to` (inclusive).
func Int64IntervalOneTo(to int64) *Int64Interval {
	return Int64IntervalFromTo(1, to)
}

// Int64IntervalZeroTo creates an interval from 0 to `to` (inclusive).
func Int64IntervalZeroTo(to int64) *Int64Interval {
	return Int64IntervalFromTo(0, to)
}

// From returns the start of the interval.
func (iv *Int64Interval) From() int64 { return iv.from }

// To returns the end of the interval (inclusive).
func (iv *Int64Interval) To() int64 { return iv.to }

// Step returns the step.
func (iv *Int64Interval) Step() int64 { return iv.step }

// Size returns the number of elements in the interval.
func (iv *Int64Interval) Size() int {
	if (iv.step > 0 && iv.from > iv.to) || (iv.step < 0 && iv.from < iv.to) {
		return 0
	}
	count := iv.distance()/iv.absStep() + 1
	maxInt := uint64(^uint(0) >> 1)
	if count > maxInt {
		return int(maxInt)
	}
	return int(count)
}

// IsEmpty returns true if the interval contains no elements.
func (iv *Int64Interval) IsEmpty() bool { return iv.Size() == 0 }

// Contains returns true if the interval contains the given value.
func (iv *Int64Interval) Contains(value int64) bool {
	if iv.step > 0 {
		return value >= iv.from && value <= iv.to && (uint64(int64(value))-uint64(int64(iv.from)))%iv.absStep() == 0
	}
	return value <= iv.from && value >= iv.to && (uint64(int64(iv.from))-uint64(int64(value)))%iv.absStep() == 0
}

// Get returns the element at the given index, or an error if out of bounds.
func (iv *Int64Interval) Get(index int) (int64, error) {
	if index < 0 || index >= iv.Size() {
		return 0, fmt.Errorf("Int64Interval: index out of bounds: %d (size %d)", index, iv.Size())
	}
	return int64(int64(iv.from) + int64(iv.step)*int64(index)), nil
}

func (iv *Int64Interval) absStep() uint64 {
	step := int64(iv.step)
	if step < 0 {
		return uint64(^step) + 1
	}
	return uint64(step)
}

func (iv *Int64Interval) distance() uint64 {
	if iv.step > 0 {
		return uint64(int64(iv.to)) - uint64(int64(iv.from))
	}
	return uint64(int64(iv.from)) - uint64(int64(iv.to))
}

// All returns an iter.Seq that yields elements in order.
func (iv *Int64Interval) All() iter.Seq[int64] {
	return func(yield func(int64) bool) {
		size := iv.Size()
		for i := 0; i < size; i++ {
			value, err := iv.Get(i)
			if err != nil {
				return
			}
			if !yield(value) {
				return
			}
		}
	}
}

// ForEach calls the given function for each element in order.
func (iv *Int64Interval) ForEach(f func(int64)) {
	for v := range iv.All() {
		f(v)
	}
}

// AnySatisfy returns true if any element satisfies the predicate.
func (iv *Int64Interval) AnySatisfy(predicate func(int64) bool) bool {
	for v := range iv.All() {
		if predicate(v) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (iv *Int64Interval) AllSatisfy(predicate func(int64) bool) bool {
	for v := range iv.All() {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (iv *Int64Interval) NoneSatisfy(predicate func(int64) bool) bool {
	for v := range iv.All() {
		if predicate(v) {
			return false
		}
	}
	return true
}

// ToSlice returns all elements as a slice.
func (iv *Int64Interval) ToSlice() []int64 {
	n := iv.Size()
	result := make([]int64, 0, n)
	for v := range iv.All() {
		result = append(result, v)
	}
	return result
}

// Reversed returns a new interval with elements in reverse order.
func (iv *Int64Interval) Reversed() *Int64Interval {
	if iv.step == int64(-1<<63) {
		panic("Int64Interval: cannot reverse interval with minimum step")
	}
	return &Int64Interval{from: iv.to, to: iv.from, step: -iv.step}
}

// String returns a string representation of the interval.
func (iv *Int64Interval) String() string {
	n := iv.Size()
	if n == 0 {
		return "[]"
	}
	var sb strings.Builder
	sb.WriteString("[")
	first := true
	for v := range iv.All() {
		if !first {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v", v)
		first = false
	}
	sb.WriteString("]")
	return sb.String()
}
