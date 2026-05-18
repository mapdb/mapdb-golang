
package tuple

import (
	"fmt"
)

// ObjectFloat64Pair is an immutable pair of (T, float64) where T is any type.
type ObjectFloat64Pair[T any] struct {
	one T
	two float64
}

// NewObjectFloat64Pair creates a new ObjectFloat64Pair.
func NewObjectFloat64Pair[T any](one T, two float64) ObjectFloat64Pair[T] {
	return ObjectFloat64Pair[T]{one: one, two: two}
}

// One returns the first element (object).
func (p ObjectFloat64Pair[T]) One() T {
	return p.one
}

// Two returns the second element (primitive).
func (p ObjectFloat64Pair[T]) Two() float64 {
	return p.two
}

// String returns a string representation.
func (p ObjectFloat64Pair[T]) String() string {
	return fmt.Sprintf("(%v, %v)", p.one, p.two)
}
