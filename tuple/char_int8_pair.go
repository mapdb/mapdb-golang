
package tuple

import (
	"fmt"
)

// CharInt8Pair is an immutable pair of (uint16, int8).
type CharInt8Pair struct {
	one uint16
	two int8
}

// NewCharInt8Pair creates a new CharInt8Pair.
func NewCharInt8Pair(one uint16, two int8) CharInt8Pair {
	return CharInt8Pair{one: one, two: two}
}

// One returns the first element.
func (p CharInt8Pair) One() uint16 {
	return p.one
}

// Two returns the second element.
func (p CharInt8Pair) Two() int8 {
	return p.two
}

// Equals returns true if both elements are equal.
func (p CharInt8Pair) Equals(other CharInt8Pair) bool {
	return p.one == other.one && p.two == other.two
}

// String returns a string representation.
func (p CharInt8Pair) String() string {
	return fmt.Sprintf("(%v, %v)", p.one, p.two)
}

// CompareTo compares this pair to another. Compares first element first,
// then second element if first elements are equal.
// Returns negative if p < other, zero if equal, positive if p > other.
func (p CharInt8Pair) CompareTo(other CharInt8Pair) int {
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
func (p CharInt8Pair) Swap() Int8CharPair {
	return NewInt8CharPair(p.two, p.one)
}
