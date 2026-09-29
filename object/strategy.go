// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package object

import (
	"cmp"
	"hash/maphash"
	"reflect"
	"strings"
	"unicode"
)

//go:generate go run ../internal/codegen object

// ── HashingStrategy ───────────────────────────────────────────────────

// HashingStrategy externalises identity for hash-based collections.
// Instead of relying on the element's own hash/equals, the collection
// uses the strategy. This enables case-insensitive keys, identity by
// extracted field, etc.
//
// This is the Go equivalent of Eclipse Collections' HashingStrategy<T>.
type HashingStrategy[T any] struct {
	HashCode func(T) uint64
	Equals   func(T, T) bool
}

// StringHashingStrategy returns the default hashing strategy for strings.
func StringHashingStrategy() HashingStrategy[string] {
	seed := maphash.MakeSeed()
	return HashingStrategy[string]{
		HashCode: func(s string) uint64 { return maphash.String(seed, s) },
		Equals:   func(a, b string) bool { return a == b },
	}
}

// CaseInsensitiveHashingStrategy returns a hashing strategy for strings
// that ignores case. "Hello" and "hello" are considered equal.
func CaseInsensitiveHashingStrategy() HashingStrategy[string] {
	seed := maphash.MakeSeed()
	return HashingStrategy[string]{
		HashCode: func(s string) uint64 { return maphash.String(seed, simpleFoldCanonical(s)) },
		Equals:   func(a, b string) bool { return strings.EqualFold(a, b) },
	}
}

// simpleFoldCanonical chooses the smallest rune in each Unicode simple-fold
// orbit. EqualFold compares by those orbits; lowercasing alone does not (for
// example, "s" and "ſ" or "σ" and "ς" would hash differently).
func simpleFoldCanonical(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		canonical := r
		for folded := unicode.SimpleFold(r); folded != r; folded = unicode.SimpleFold(folded) {
			if folded < canonical {
				canonical = folded
			}
		}
		b.WriteRune(canonical)
	}
	return b.String()
}

// ByField returns a hashing strategy that hashes and compares by an
// extracted field. Works for any comparable field type — strings,
// numeric types, bools, pointers, channels, interfaces, and structs
// or fixed-size arrays composed thereof.
//
//	strategy := ByField(func(p Person) string { return p.Name })
//
// The hash respects Go's == semantics: if extract(a) == extract(b)
// then the hashes are equal. This is guaranteed by hash/maphash's
// Comparable, which dispatches on the runtime representation rather
// than the in-memory byte pattern (so two strings with different
// backing pointers but equal content hash identically).
func ByField[T any, F comparable](extract func(T) F) HashingStrategy[T] {
	seed := maphash.MakeSeed()
	return HashingStrategy[T]{
		HashCode: func(v T) uint64 { return maphash.Comparable(seed, extract(v)) },
		Equals:   func(a, b T) bool { return extract(a) == extract(b) },
	}
}

// ByFieldString returns a hashing strategy that hashes and compares by
// a string field. More efficient than ByField for string extractions.
func ByFieldString[T any](extract func(T) string) HashingStrategy[T] {
	seed := maphash.MakeSeed()
	return HashingStrategy[T]{
		HashCode: func(v T) uint64 { return maphash.String(seed, extract(v)) },
		Equals:   func(a, b T) bool { return extract(a) == extract(b) },
	}
}

// ── Comparator ────────────────────────────────────────────────────────

// Comparator defines an ordering between two values.
// Returns negative if a < b, zero if a == b, positive if a > b.
//
// This is the Go equivalent of Java's Comparator<T>.
type Comparator[T any] func(a, b T) int

// NaturalComparator returns a comparator that uses the natural ordering
// of ordered types (numbers, strings).
//
// For float32/float64 T (including named types with a float underlying
// type) the natural ordering is the IEEE 754 totalOrder mandated by the
// collection spec, not Go's < / cmp.Compare:
//
//	-NaN < -Inf < negative finite < -0.0 < +0.0 < positive finite < +Inf < +NaN
//
// -0.0 and +0.0 are distinct, and NaNs with different payloads are distinct
// and ordered by payload, so a TreeSet/TreeMap keeps all of them.
func NaturalComparator[T cmp.Ordered]() Comparator[T] {
	switch floatKind[T]() {
	case reflect.Float32:
		return func(a, b T) int { return cmpFloat32(float32Of(a), float32Of(b)) }
	case reflect.Float64:
		return func(a, b T) int { return cmpFloat64(float64Of(a), float64Of(b)) }
	}
	return func(a, b T) int { return cmp.Compare(a, b) }
}

// ReverseComparator returns the exact reverse of NaturalComparator
// (for floats: the reversed IEEE 754 totalOrder).
func ReverseComparator[T cmp.Ordered]() Comparator[T] {
	natural := NaturalComparator[T]()
	return func(a, b T) int { return natural(b, a) }
}

// ComparatorByField returns a comparator that orders by an extracted field,
// using NaturalComparator ordering on the field (so float fields use the
// IEEE 754 totalOrder).
//
//	cmp := ComparatorByField(func(p Person) string { return p.Name })
func ComparatorByField[T any, F cmp.Ordered](extract func(T) F) Comparator[T] {
	natural := NaturalComparator[F]()
	return func(a, b T) int { return natural(extract(a), extract(b)) }
}

// ReverseComparatorByField returns a comparator that orders by an extracted field in reverse.
func ReverseComparatorByField[T any, F cmp.Ordered](extract func(T) F) Comparator[T] {
	natural := NaturalComparator[F]()
	return func(a, b T) int { return natural(extract(b), extract(a)) }
}

// ThenComparing chains two comparators: uses the second when the first returns zero.
func ThenComparing[T any](primary, secondary Comparator[T]) Comparator[T] {
	return func(a, b T) int {
		if r := primary(a, b); r != 0 {
			return r
		}
		return secondary(a, b)
	}
}

// Reversed returns a comparator that reverses the given one. Works on any
// comparator (unlike ReverseComparator which requires cmp.Ordered).
//
//	byName := ComparatorByField(func(p Person) string { return p.Name })
//	byNameDesc := Reversed(byName)
func Reversed[T any](c Comparator[T]) Comparator[T] {
	return func(a, b T) int { return c(b, a) }
}

// ComparatorByFieldWith returns a comparator that orders by an extracted
// field using a custom sub-comparator (instead of natural ordering).
// Useful for e.g. sorting by a string field case-insensitively.
//
//	byNameCI := ComparatorByFieldWith(
//	    func(p Person) string { return p.Name },
//	    func(a, b string) int { return strings.Compare(strings.ToLower(a), strings.ToLower(b)) },
//	)
func ComparatorByFieldWith[T any, F any](extract func(T) F, sub Comparator[F]) Comparator[T] {
	return func(a, b T) int { return sub(extract(a), extract(b)) }
}
