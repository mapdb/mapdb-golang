
package tuple

import (
	"fmt"
	"math"
)

// Int16Float32Pair is an immutable pair of (int16, float32).
type Int16Float32Pair struct {
	one int16
	two float32
}

// NewInt16Float32Pair creates a new Int16Float32Pair.
func NewInt16Float32Pair(one int16, two float32) Int16Float32Pair {
	return Int16Float32Pair{one: one, two: two}
}

// One returns the first element.
func (p Int16Float32Pair) One() int16 {
	return p.one
}

// Two returns the second element.
func (p Int16Float32Pair) Two() float32 {
	return p.two
}

// Equals returns true if both elements are equal.
func (p Int16Float32Pair) Equals(other Int16Float32Pair) bool {
	return p.one == other.one && math.Float32bits(p.two) == math.Float32bits(other.two)
}

// String returns a string representation.
func (p Int16Float32Pair) String() string {
	return fmt.Sprintf("(%v, %v)", p.one, p.two)
}

// CompareTo compares this pair to another. Compares first element first,
// then second element if first elements are equal.
// Returns negative if p < other, zero if equal, positive if p > other.
func (p Int16Float32Pair) CompareTo(other Int16Float32Pair) int {
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
func (p Int16Float32Pair) Swap() Float32Int16Pair {
	return NewFloat32Int16Pair(p.two, p.one)
}
