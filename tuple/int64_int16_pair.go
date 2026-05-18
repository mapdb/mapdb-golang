
package tuple

import (
	"fmt"
)

// Int64Int16Pair is an immutable pair of (int64, int16).
type Int64Int16Pair struct {
	one int64
	two int16
}

// NewInt64Int16Pair creates a new Int64Int16Pair.
func NewInt64Int16Pair(one int64, two int16) Int64Int16Pair {
	return Int64Int16Pair{one: one, two: two}
}

// One returns the first element.
func (p Int64Int16Pair) One() int64 {
	return p.one
}

// Two returns the second element.
func (p Int64Int16Pair) Two() int16 {
	return p.two
}

// Equals returns true if both elements are equal.
func (p Int64Int16Pair) Equals(other Int64Int16Pair) bool {
	return p.one == other.one && p.two == other.two
}

// String returns a string representation.
func (p Int64Int16Pair) String() string {
	return fmt.Sprintf("(%v, %v)", p.one, p.two)
}

// CompareTo compares this pair to another. Compares first element first,
// then second element if first elements are equal.
// Returns negative if p < other, zero if equal, positive if p > other.
func (p Int64Int16Pair) CompareTo(other Int64Int16Pair) int {
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
func (p Int64Int16Pair) Swap() Int16Int64Pair {
	return NewInt16Int64Pair(p.two, p.one)
}
