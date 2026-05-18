
package tuple

import (
	"fmt"
	"math"
)

// Int64Float64Pair is an immutable pair of (int64, float64).
type Int64Float64Pair struct {
	one int64
	two float64
}

// NewInt64Float64Pair creates a new Int64Float64Pair.
func NewInt64Float64Pair(one int64, two float64) Int64Float64Pair {
	return Int64Float64Pair{one: one, two: two}
}

// One returns the first element.
func (p Int64Float64Pair) One() int64 {
	return p.one
}

// Two returns the second element.
func (p Int64Float64Pair) Two() float64 {
	return p.two
}

// Equals returns true if both elements are equal.
func (p Int64Float64Pair) Equals(other Int64Float64Pair) bool {
	return p.one == other.one && math.Float64bits(p.two) == math.Float64bits(other.two)
}

// String returns a string representation.
func (p Int64Float64Pair) String() string {
	return fmt.Sprintf("(%v, %v)", p.one, p.two)
}

// CompareTo compares this pair to another. Compares first element first,
// then second element if first elements are equal.
// Returns negative if p < other, zero if equal, positive if p > other.
func (p Int64Float64Pair) CompareTo(other Int64Float64Pair) int {
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
func (p Int64Float64Pair) Swap() Float64Int64Pair {
	return NewFloat64Int64Pair(p.two, p.one)
}
