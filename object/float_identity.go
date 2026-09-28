// Copyright (c) 2026 Jan Kotek.
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package object

import (
	"iter"
	"math"
	"math/bits"
	"reflect"
	"unsafe"
)

// Float identity for the generic object/ collections.
//
// The collection spec (algorithms.md "NaN must hash and compare by bit
// pattern") requires float keys to use bit-pattern identity: NaN equals
// itself, distinct NaN payloads are distinct, and -0.0 != +0.0. Go's builtin
// map and == use IEEE equality instead (NaN != NaN, -0.0 == +0.0), so a
// builtin map[float64]V can neither find nor replace a NaN key and merges the
// two zeros.
//
// Every object/ hash collection therefore stores its keys in a keyIndex
// rather than a raw builtin map. For ordinary key types a keyIndex is exactly
// the builtin map it replaces. For a key type whose underlying type is
// float32 or float64 (including named types such as `type Celsius float64`)
// it is the open-addressing HashMapWithStrategy with a bit-pattern
// HashingStrategy. The choice is made once, when the index is first
// allocated. Composite keys that merely contain floats (structs, arrays,
// interfaces) keep Go's == semantics.

// floatKind reports reflect.Float32 or reflect.Float64 when T's underlying
// type is float32 or float64, and reflect.Invalid otherwise.
func floatKind[T any]() reflect.Kind {
	switch k := reflect.TypeFor[T]().Kind(); k {
	case reflect.Float32, reflect.Float64:
		return k
	}
	return reflect.Invalid
}

// float32Of reinterprets v as a float32. Only valid when floatKind[T]() is
// reflect.Float32 (T's underlying type is float32, so the layouts match).
func float32Of[T any](v T) float32 { return *(*float32)(unsafe.Pointer(&v)) }

// float64Of reinterprets v as a float64. Only valid when floatKind[T]() is
// reflect.Float64.
func float64Of[T any](v T) float64 { return *(*float64)(unsafe.Pointer(&v)) }

// fibonacciHash is the mapdb 64-bit Fibonacci hash (spec algorithms.md
// "Hash function": golden-ratio multiply by 0x9e3779b97f4a7c15, the constant
// every primitive hash map in this module uses) of a key's bit pattern,
// returned with its bits reversed.
//
// Fibonacci hashing takes the TOP bits of the product as the bucket index:
// they depend on every input bit, while the product's low bits depend only on
// the input's low bits. HashMapWithStrategy indexes by the LOW bits
// (hash & mask) and does not know its capacity, and float bit patterns of
// round values (1.0, 2.0, 0.5, ...) have long runs of zero low bits, so the
// low bits of the plain product would put them all in one bucket. Reversing
// the product moves its top k bits into the low k bits for every table size
// 2^k: the index is the Fibonacci top-bit index up to a fixed permutation of
// the buckets.
func fibonacciHash(b uint64) uint64 {
	return bits.Reverse64(b * 0x9e3779b97f4a7c15)
}

// floatBitsStrategy returns a bit-pattern HashingStrategy when K's underlying
// type is float32 or float64, and ok=false for every other K.
func floatBitsStrategy[K any]() (s HashingStrategy[K], ok bool) {
	switch floatKind[K]() {
	case reflect.Float32:
		return HashingStrategy[K]{
			HashCode: func(k K) uint64 { return fibonacciHash(uint64(math.Float32bits(float32Of(k)))) },
			Equals: func(a, b K) bool {
				return math.Float32bits(float32Of(a)) == math.Float32bits(float32Of(b))
			},
		}, true
	case reflect.Float64:
		return HashingStrategy[K]{
			HashCode: func(k K) uint64 { return fibonacciHash(math.Float64bits(float64Of(k))) },
			Equals: func(a, b K) bool {
				return math.Float64bits(float64Of(a)) == math.Float64bits(float64Of(b))
			},
		}, true
	}
	return HashingStrategy[K]{}, false
}

// keyIndex is the key -> value table behind the object/ hash collections: a
// builtin map for ordinary K, a bit-pattern strategy map for float K (see the
// file comment). The zero value is an empty index; it allocates on the first
// put. Iteration must not mutate the index (the float path's backward-shift
// deletion and resize are not range-safe); the public form of this rule is
// the package documentation's "Float keys" section, which every exported
// iteration method of a keyIndex-backed collection points to.
type keyIndex[K comparable, V any] struct {
	m map[K]V                    // non-float K
	f *HashMapWithStrategy[K, V] // float K; nil until allocated
}

// newKeyIndex returns an allocated index with room for capacity entries.
func newKeyIndex[K comparable, V any](capacity int) keyIndex[K, V] {
	if s, ok := floatBitsStrategy[K](); ok {
		return keyIndex[K, V]{f: NewHashMapWithStrategyCapacity[K, V](s, capacity)}
	}
	return keyIndex[K, V]{m: make(map[K]V, capacity)}
}

func (x *keyIndex[K, V]) get(k K) (V, bool) {
	if x.f != nil {
		return x.f.Get(k)
	}
	v, ok := x.m[k]
	return v, ok
}

func (x *keyIndex[K, V]) contains(k K) bool {
	_, ok := x.get(k)
	return ok
}

// put stores v under k and returns the previous value, if any.
func (x *keyIndex[K, V]) put(k K, v V) (V, bool) {
	if x.f == nil && x.m == nil {
		*x = newKeyIndex[K, V](0)
	}
	if x.f != nil {
		return x.f.Put(k, v)
	}
	old, existed := x.m[k]
	x.m[k] = v
	return old, existed
}

func (x *keyIndex[K, V]) remove(k K) (V, bool) {
	if x.f != nil {
		return x.f.Remove(k)
	}
	old, existed := x.m[k]
	if existed {
		delete(x.m, k)
	}
	return old, existed
}

func (x *keyIndex[K, V]) len() int {
	if x.f != nil {
		return x.f.Len()
	}
	return len(x.m)
}

func (x *keyIndex[K, V]) clear() {
	if x.f != nil {
		x.f.Clear()
		return
	}
	clear(x.m)
}

// all iterates the entries in unspecified order.
func (x *keyIndex[K, V]) all() iter.Seq2[K, V] {
	if x.f != nil {
		return x.f.All()
	}
	return func(yield func(K, V) bool) {
		for k, v := range x.m {
			if !yield(k, v) {
				return
			}
		}
	}
}

// sameKey is key equality as seen by keyIndex: bit-pattern equality for float
// K, == otherwise.
func sameKey[K comparable](a, b K) bool {
	switch floatKind[K]() {
	case reflect.Float32:
		return math.Float32bits(float32Of(a)) == math.Float32bits(float32Of(b))
	case reflect.Float64:
		return math.Float64bits(float64Of(a)) == math.Float64bits(float64Of(b))
	}
	return a == b
}
