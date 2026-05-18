
package tuple

import (
	"fmt"
	"math"
)

// Float32Int32Pair is an immutable pair of (float32, int32).
type Float32Int32Pair struct {
	one float32
	two int32
}

// NewFloat32Int32Pair creates a new Float32Int32Pair.
func NewFloat32Int32Pair(one float32, two int32) Float32Int32Pair {
	return Float32Int32Pair{one: one, two: two}
}

// One returns the first element.
func (p Float32Int32Pair) One() float32 {
	return p.one
}

// Two returns the second element.
func (p Float32Int32Pair) Two() int32 {
	return p.two
}

// Equals returns true if both elements are equal.
func (p Float32Int32Pair) Equals(other Float32Int32Pair) bool {
	return math.Float32bits(p.one) == math.Float32bits(other.one) && p.two == other.two
}

// String returns a string representation.
func (p Float32Int32Pair) String() string {
	return fmt.Sprintf("(%v, %v)", p.one, p.two)
}

// CompareTo compares this pair to another. Compares first element first,
// then second element if first elements are equal.
// Returns negative if p < other, zero if equal, positive if p > other.
func (p Float32Int32Pair) CompareTo(other Float32Int32Pair) int {
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
func (p Float32Int32Pair) Swap() Int32Float32Pair {
	return NewInt32Float32Pair(p.two, p.one)
}
