
package tuple

import (
	"fmt"
)

// Int8Int32Pair is an immutable pair of (int8, int32).
type Int8Int32Pair struct {
	one int8
	two int32
}

// NewInt8Int32Pair creates a new Int8Int32Pair.
func NewInt8Int32Pair(one int8, two int32) Int8Int32Pair {
	return Int8Int32Pair{one: one, two: two}
}

// One returns the first element.
func (p Int8Int32Pair) One() int8 {
	return p.one
}

// Two returns the second element.
func (p Int8Int32Pair) Two() int32 {
	return p.two
}

// Equals returns true if both elements are equal.
func (p Int8Int32Pair) Equals(other Int8Int32Pair) bool {
	return p.one == other.one && p.two == other.two
}

// String returns a string representation.
func (p Int8Int32Pair) String() string {
	return fmt.Sprintf("(%v, %v)", p.one, p.two)
}

// CompareTo compares this pair to another. Compares first element first,
// then second element if first elements are equal.
// Returns negative if p < other, zero if equal, positive if p > other.
func (p Int8Int32Pair) CompareTo(other Int8Int32Pair) int {
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
func (p Int8Int32Pair) Swap() Int32Int8Pair {
	return NewInt32Int8Pair(p.two, p.one)
}
