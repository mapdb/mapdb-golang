
package tuple

import (
	"fmt"
)

// Int64Int8Pair is an immutable pair of (int64, int8).
type Int64Int8Pair struct {
	one int64
	two int8
}

// NewInt64Int8Pair creates a new Int64Int8Pair.
func NewInt64Int8Pair(one int64, two int8) Int64Int8Pair {
	return Int64Int8Pair{one: one, two: two}
}

// One returns the first element.
func (p Int64Int8Pair) One() int64 {
	return p.one
}

// Two returns the second element.
func (p Int64Int8Pair) Two() int8 {
	return p.two
}

// Equals returns true if both elements are equal.
func (p Int64Int8Pair) Equals(other Int64Int8Pair) bool {
	return p.one == other.one && p.two == other.two
}

// String returns a string representation.
func (p Int64Int8Pair) String() string {
	return fmt.Sprintf("(%v, %v)", p.one, p.two)
}

// CompareTo compares this pair to another. Compares first element first,
// then second element if first elements are equal.
// Returns negative if p < other, zero if equal, positive if p > other.
func (p Int64Int8Pair) CompareTo(other Int64Int8Pair) int {
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
func (p Int64Int8Pair) Swap() Int8Int64Pair {
	return NewInt8Int64Pair(p.two, p.one)
}
