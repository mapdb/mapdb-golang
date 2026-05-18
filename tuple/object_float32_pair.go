
package tuple

import (
	"fmt"
)

// ObjectFloat32Pair is an immutable pair of (T, float32) where T is any type.
type ObjectFloat32Pair[T any] struct {
	one T
	two float32
}

// NewObjectFloat32Pair creates a new ObjectFloat32Pair.
func NewObjectFloat32Pair[T any](one T, two float32) ObjectFloat32Pair[T] {
	return ObjectFloat32Pair[T]{one: one, two: two}
}

// One returns the first element (object).
func (p ObjectFloat32Pair[T]) One() T {
	return p.one
}

// Two returns the second element (primitive).
func (p ObjectFloat32Pair[T]) Two() float32 {
	return p.two
}

// String returns a string representation.
func (p ObjectFloat32Pair[T]) String() string {
	return fmt.Sprintf("(%v, %v)", p.one, p.two)
}
