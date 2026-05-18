
package tuple

import (
	"fmt"
)

// Int64Int32Pair is an immutable pair of (int64, int32).
type Int64Int32Pair struct {
	one int64
	two int32
}

// NewInt64Int32Pair creates a new Int64Int32Pair.
func NewInt64Int32Pair(one int64, two int32) Int64Int32Pair {
	return Int64Int32Pair{one: one, two: two}
}

// One returns the first element.
func (p Int64Int32Pair) One() int64 {
	return p.one
}

// Two returns the second element.
func (p Int64Int32Pair) Two() int32 {
	return p.two
}

// Equals returns true if both elements are equal.
func (p Int64Int32Pair) Equals(other Int64Int32Pair) bool {
	return p.one == other.one && p.two == other.two
}

// String returns a string representation.
func (p Int64Int32Pair) String() string {
	return fmt.Sprintf("(%v, %v)", p.one, p.two)
}

// CompareTo compares this pair to another. Compares first element first,
// then second element if first elements are equal.
// Returns negative if p < other, zero if equal, positive if p > other.
func (p Int64Int32Pair) CompareTo(other Int64Int32Pair) int {
	if p.one < other.one {
		return -1
	}
	if p.one > other.one {
		return 1
	}
	if p.two < other.two {
		return -1
	}
	if p.two > other.two {
		return 1
	}
	return 0
}

// Swap returns a new pair with elements swapped: (two, one).
func (p Int64Int32Pair) Swap() Int32Int64Pair {
	return NewInt32Int64Pair(p.two, p.one)
}
