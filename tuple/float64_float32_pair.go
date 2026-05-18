
package tuple

import (
	"fmt"
	"math"
)

// Float64Float32Pair is an immutable pair of (float64, float32).
type Float64Float32Pair struct {
	one float64
	two float32
}

// NewFloat64Float32Pair creates a new Float64Float32Pair.
func NewFloat64Float32Pair(one float64, two float32) Float64Float32Pair {
	return Float64Float32Pair{one: one, two: two}
}

// One returns the first element.
func (p Float64Float32Pair) One() float64 {
	return p.one
}

// Two returns the second element.
func (p Float64Float32Pair) Two() float32 {
	return p.two
}

// Equals returns true if both elements are equal.
func (p Float64Float32Pair) Equals(other Float64Float32Pair) bool {
	return math.Float64bits(p.one) == math.Float64bits(other.one) && math.Float32bits(p.two) == math.Float32bits(other.two)
}

// String returns a string representation.
func (p Float64Float32Pair) String() string {
	return fmt.Sprintf("(%v, %v)", p.one, p.two)
}

// CompareTo compares this pair to another. Compares first element first,
// then second element if first elements are equal.
// Returns negative if p < other, zero if equal, positive if p > other.
func (p Float64Float32Pair) CompareTo(other Float64Float32Pair) int {
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
func (p Float64Float32Pair) Swap() Float32Float64Pair {
	return NewFloat32Float64Pair(p.two, p.one)
}
