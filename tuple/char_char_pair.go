
package tuple

import (
	"fmt"
)

// CharCharPair is an immutable pair of (uint16, uint16).
type CharCharPair struct {
	one uint16
	two uint16
}

// NewCharCharPair creates a new CharCharPair.
func NewCharCharPair(one uint16, two uint16) CharCharPair {
	return CharCharPair{one: one, two: two}
}

// One returns the first element.
func (p CharCharPair) One() uint16 {
	return p.one
}

// Two returns the second element.
func (p CharCharPair) Two() uint16 {
	return p.two
}

// Equals returns true if both elements are equal.
func (p CharCharPair) Equals(other CharCharPair) bool {
	return p.one == other.one && p.two == other.two
}

// String returns a string representation.
func (p CharCharPair) String() string {
	return fmt.Sprintf("(%v, %v)", p.one, p.two)
}

// CompareTo compares this pair to another. Compares first element first,
// then second element if first elements are equal.
// Returns negative if p < other, zero if equal, positive if p > other.
func (p CharCharPair) CompareTo(other CharCharPair) int {
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
func (p CharCharPair) Swap() CharCharPair {
	return NewCharCharPair(p.two, p.one)
}
