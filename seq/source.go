// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package seq

import "iter"

// Numeric is the constraint for the numeric sources and reductions (Range, Sum,
// …): every built-in integer and floating-point kind.
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// From adopts a standard library iter.Seq[T] as a Seq[T]. Zero-cost; the result
// has the same re-runnability as s.
func From[T any](s iter.Seq[T]) Seq[T] { return Seq[T](s) }

// Of returns a re-runnable Seq over the given values. O(1) memory (the values
// are already materialized in the variadic slice). ⟨EC: of⟩
func Of[T any](vs ...T) Seq[T] { return FromSlice(vs) }

// FromSlice returns a re-runnable Seq over the elements of s in order. The slice
// is not copied; mutating it after the call is visible to later iterations.
func FromSlice[T any](s []T) Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range s {
			if !yield(v) {
				return
			}
		}
	}
}

// Range returns the half-open sequence lo, lo+1, …, hi-1. Empty if lo >= hi.
// Lazy, O(1) memory.
func Range[T Numeric](lo, hi T) Seq[T] {
	return func(yield func(T) bool) {
		for v := lo; v < hi; v += 1 {
			if !yield(v) {
				return
			}
		}
	}
}

// RangeStep returns lo, lo+step, lo+2*step, … up to but not including hi. With a
// positive step it ascends while v < hi; with a negative step it descends while
// v > hi. Lazy, O(1) memory. Panics if step == 0.
//
// Advancing is overflow-safe: if v+step would wrap past the type's bounds (or,
// for floats, make no progress because step is too small to change v), the
// sequence stops rather than looping forever. So RangeStep[int8](126, 127, 2)
// yields just 126.
func RangeStep[T Numeric](lo, hi, step T) Seq[T] {
	if step == 0 {
		panic("seq: RangeStep step must be non-zero")
	}
	return func(yield func(T) bool) {
		v := lo
		if step > 0 {
			for v < hi {
				if !yield(v) {
					return
				}
				next := v + step
				if next <= v { // overflow wrap-around, or no float progress
					return
				}
				v = next
			}
		} else {
			for v > hi {
				if !yield(v) {
					return
				}
				next := v + step
				if next >= v { // underflow wrap-around, or no float progress
					return
				}
				v = next
			}
		}
	}
}

// Iterate returns the infinite sequence seed, next(seed), next(next(seed)), ….
// Lazy, O(1) memory; bound it with Take / TakeWhile / First.
func Iterate[T any](seed T, next func(T) T) Seq[T] {
	return func(yield func(T) bool) {
		for v := seed; ; v = next(v) {
			if !yield(v) {
				return
			}
		}
	}
}

// Generate returns the infinite sequence f(), f(), f(), …. Lazy, O(1) memory;
// bound it with Take / TakeWhile / First.
func Generate[T any](f func() T) Seq[T] {
	return func(yield func(T) bool) {
		for {
			if !yield(f()) {
				return
			}
		}
	}
}

// Repeat returns the infinite sequence v, v, v, …. Lazy, O(1) memory.
func Repeat[T any](v T) Seq[T] {
	return func(yield func(T) bool) {
		for {
			if !yield(v) {
				return
			}
		}
	}
}

// RepeatN returns v repeated n times (empty for n <= 0). Lazy, O(1) memory.
func RepeatN[T any](v T, n int) Seq[T] {
	return func(yield func(T) bool) {
		for i := 0; i < n; i++ {
			if !yield(v) {
				return
			}
		}
	}
}
