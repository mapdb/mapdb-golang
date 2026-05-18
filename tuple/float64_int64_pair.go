
package tuple

import (
	"fmt"
	"math"
)

// Float64Int64Pair is an immutable pair of (float64, int64).
type Float64Int64Pair struct {
	one float64
	two int64
}

// NewFloat64Int64Pair creates a new Float64Int64Pair.
func NewFloat64Int64Pair(one float64, two int64) Float64Int64Pair {
	return Float64Int64Pair{one: one, two: two}
}

// One returns the first element.
func (p Float64Int64Pair) One() float64 {
	return p.one
}

// Two returns the second element.
func (p Float64Int64Pair) Two() int64 {
	return p.two
}

// Equals returns true if both elements are equal.
func (p Float64Int64Pair) Equals(other Float64Int64Pair) bool {
	return math.Float64bits(p.one) == math.Float64bits(other.one) && p.two == other.two
}

// String returns a string representation.
func (p Float64Int64Pair) String() string {
	return fmt.Sprintf("(%v, %v)", p.one, p.two)
}

// CompareTo compares this pair to another. Compares first element first,
// then second element if first elements are equal.
// Returns negative if p < other, zero if equal, positive if p > other.
func (p Float64Int64Pair) CompareTo(other Float64Int64Pair) int {
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
func (p Float64Int64Pair) Swap() Int64Float64Pair {
	return NewInt64Float64Pair(p.two, p.one)
}
