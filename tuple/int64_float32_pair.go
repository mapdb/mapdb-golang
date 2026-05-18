
package tuple

import (
	"fmt"
	"math"
)

// Int64Float32Pair is an immutable pair of (int64, float32).
type Int64Float32Pair struct {
	one int64
	two float32
}

// NewInt64Float32Pair creates a new Int64Float32Pair.
func NewInt64Float32Pair(one int64, two float32) Int64Float32Pair {
	return Int64Float32Pair{one: one, two: two}
}

// One returns the first element.
func (p Int64Float32Pair) One() int64 {
	return p.one
}

// Two returns the second element.
func (p Int64Float32Pair) Two() float32 {
	return p.two
}

// Equals returns true if both elements are equal.
func (p Int64Float32Pair) Equals(other Int64Float32Pair) bool {
	return p.one == other.one && math.Float32bits(p.two) == math.Float32bits(other.two)
}

// String returns a string representation.
func (p Int64Float32Pair) String() string {
	return fmt.Sprintf("(%v, %v)", p.one, p.two)
}

// CompareTo compares this pair to another. Compares first element first,
// then second element if first elements are equal.
// Returns negative if p < other, zero if equal, positive if p > other.
func (p Int64Float32Pair) CompareTo(other Int64Float32Pair) int {
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
func (p Int64Float32Pair) Swap() Float32Int64Pair {
	return NewFloat32Int64Pair(p.two, p.one)
}
