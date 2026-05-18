
package tuple

import (
	"fmt"
)

// Float32ObjectPair is an immutable pair of (float32, T) where T is any type.
type Float32ObjectPair[T any] struct {
	one float32
	two T
}

// NewFloat32ObjectPair creates a new Float32ObjectPair.
func NewFloat32ObjectPair[T any](one float32, two T) Float32ObjectPair[T] {
	return Float32ObjectPair[T]{one: one, two: two}
}

// One returns the first element (primitive).
func (p Float32ObjectPair[T]) One() float32 {
	return p.one
}

// Two returns the second element (object).
func (p Float32ObjectPair[T]) Two() T {
	return p.two
}

// String returns a string representation.
func (p Float32ObjectPair[T]) String() string {
	return fmt.Sprintf("(%v, %v)", p.one, p.two)
}
