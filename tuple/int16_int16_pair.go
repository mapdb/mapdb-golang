
package tuple

import (
	"fmt"
)

// Int16Int16Pair is an immutable pair of (int16, int16).
type Int16Int16Pair struct {
	one int16
	two int16
}

// NewInt16Int16Pair creates a new Int16Int16Pair.
func NewInt16Int16Pair(one int16, two int16) Int16Int16Pair {
	return Int16Int16Pair{one: one, two: two}
}

// One returns the first element.
func (p Int16Int16Pair) One() int16 {
	return p.one
}

// Two returns the second element.
func (p Int16Int16Pair) Two() int16 {
	return p.two
}

// Equals returns true if both elements are equal.
func (p Int16Int16Pair) Equals(other Int16Int16Pair) bool {
	return p.one == other.one && p.two == other.two
}

// String returns a string representation.
func (p Int16Int16Pair) String() string {
	return fmt.Sprintf("(%v, %v)", p.one, p.two)
}

// CompareTo compares this pair to another. Compares first element first,
// then second element if first elements are equal.
// Returns negative if p < other, zero if equal, positive if p > other.
func (p Int16Int16Pair) CompareTo(other Int16Int16Pair) int {
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
func (p Int16Int16Pair) Swap() Int16Int16Pair {
	return NewInt16Int16Pair(p.two, p.one)
}
