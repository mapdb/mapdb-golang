
package tuple

import (
	"fmt"
)

// Int64Int64Pair is an immutable pair of (int64, int64).
type Int64Int64Pair struct {
	one int64
	two int64
}

// NewInt64Int64Pair creates a new Int64Int64Pair.
func NewInt64Int64Pair(one int64, two int64) Int64Int64Pair {
	return Int64Int64Pair{one: one, two: two}
}

// One returns the first element.
func (p Int64Int64Pair) One() int64 {
	return p.one
}

// Two returns the second element.
func (p Int64Int64Pair) Two() int64 {
	return p.two
}

// Equals returns true if both elements are equal.
func (p Int64Int64Pair) Equals(other Int64Int64Pair) bool {
	return p.one == other.one && p.two == other.two
}

// String returns a string representation.
func (p Int64Int64Pair) String() string {
	return fmt.Sprintf("(%v, %v)", p.one, p.two)
}

// CompareTo compares this pair to another. Compares first element first,
// then second element if first elements are equal.
// Returns negative if p < other, zero if equal, positive if p > other.
func (p Int64Int64Pair) CompareTo(other Int64Int64Pair) int {
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
func (p Int64Int64Pair) Swap() Int64Int64Pair {
	return NewInt64Int64Pair(p.two, p.one)
}
