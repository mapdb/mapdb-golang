// Package floatid is the float bit-pattern identity shared by the generic
// (type-parameterised) collections outside package object.
//
// The collection spec (algorithms.md "NaN must hash and compare by bit
// pattern") requires float keys and elements to compare and hash by their
// IEEE 754 bit pattern: NaN equals itself, distinct NaN payloads are
// distinct, and -0.0 != +0.0. Go's == and builtin map use IEEE equality
// instead. A generic T whose underlying type is float32 or float64 (named
// types such as `type Celsius float64` included) is detected at run time and
// routed through the helpers below; every other T keeps Go's ==. Composite
// types that merely contain floats (structs, arrays, interfaces) keep ==.
//
// This mirrors object/float_identity.go, which package object keeps
// private.
package floatid

import (
	"math"
	"math/bits"
	"reflect"
	"unsafe"
)

// Kind reports reflect.Float32 or reflect.Float64 when T's underlying type is
// float32 or float64, and reflect.Invalid otherwise.
func Kind[T any]() reflect.Kind {
	switch k := reflect.TypeFor[T]().Kind(); k {
	case reflect.Float32, reflect.Float64:
		return k
	}
	return reflect.Invalid
}

// IsFloat reports whether T's underlying type is float32 or float64.
func IsFloat[T any]() bool { return Kind[T]() != reflect.Invalid }

// Bits returns v's IEEE 754 bit pattern (float32 zero-extended). kind must be
// Kind[T]() and must not be reflect.Invalid.
func Bits[T any](kind reflect.Kind, v T) uint64 {
	if kind == reflect.Float32 {
		return uint64(math.Float32bits(*(*float32)(unsafe.Pointer(&v))))
	}
	return math.Float64bits(*(*float64)(unsafe.Pointer(&v)))
}

// Same is value identity: bit-pattern equality when T is a float type, ==
// otherwise.
func Same[T comparable](a, b T) bool {
	if k := Kind[T](); k != reflect.Invalid {
		return Bits(k, a) == Bits(k, b)
	}
	return a == b
}

// SameKind is Same with Kind[T]() precomputed by the caller (hot loops).
func SameKind[T comparable](kind reflect.Kind, a, b T) bool {
	if kind != reflect.Invalid {
		return Bits(kind, a) == Bits(kind, b)
	}
	return a == b
}

// Hash is the mapdb 64-bit Fibonacci hash (spec algorithms.md "Hash
// function", multiplier 0x9e3779b97f4a7c15) of a float bit pattern, with the
// product's bits reversed so that a table indexing by the LOW bits
// (hash & mask) sees the Fibonacci top-bit index (up to a fixed bucket
// permutation). Same technique as object/float_identity.go fibonacciHash.
func Hash(b uint64) uint64 {
	return bits.Reverse64(b * 0x9e3779b97f4a7c15)
}
