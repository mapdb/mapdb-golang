
package tuple

import (
	"fmt"
	"math"
)

// Float32Float64Pair is an immutable pair of (float32, float64).
type Float32Float64Pair struct {
	one float32
	two float64
}

// NewFloat32Float64Pair creates a new Float32Float64Pair.
func NewFloat32Float64Pair(one float32, two float64) Float32Float64Pair {
	return Float32Float64Pair{one: one, two: two}
}

// One returns the first element.
func (p Float32Float64Pair) One() float32 {
	return p.one
}

// Two returns the second element.
func (p Float32Float64Pair) Two() float64 {
	return p.two
}

// Equals returns true if both elements are equal.
func (p Float32Float64Pair) Equals(other Float32Float64Pair) bool {
	return math.Float32bits(p.one) == math.Float32bits(other.one) && math.Float64bits(p.two) == math.Float64bits(other.two)
}

// String returns a string representation.
func (p Float32Float64Pair) String() string {
	return fmt.Sprintf("(%v, %v)", p.one, p.two)
}

// CompareTo compares this pair to another. Compares first element first,
// then second element if first elements are equal.
// Returns negative if p < other, zero if equal, positive if p > other.
func (p Float32Float64Pair) CompareTo(other Float32Float64Pair) int {
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
func (p Float32Float64Pair) Swap() Float64Float32Pair {
	return NewFloat64Float32Pair(p.two, p.one)
}
