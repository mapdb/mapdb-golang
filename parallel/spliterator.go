// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Splittable iterators — the Go analogue of java.util.Spliterator.
//
// A Spliterator traverses elements like an iterator but can also *split*
// itself (TrySplit), handing back a new spliterator that covers a portion of
// the remaining elements. This is the decomposition primitive Java parallel
// streams build on; it describes how work can be divided and performs no
// parallel execution by itself.
//
// This part of the package is pure stdlib.

package parallel

// Spliterator characteristic flags. The bit values match
// java.util.Spliterator so that ported code reads identically.
const (
	// Distinct means elements are distinct (no duplicates) — e.g. a set source.
	Distinct uint = 0x0000_0001
	// Sorted means elements follow a defined sort order (see also Ordered).
	Sorted uint = 0x0000_0004
	// Ordered means encounter order is meaningful and preserved across splits.
	Ordered uint = 0x0000_0010
	// Sized means EstimateSize is an exact element count before traversal.
	Sized uint = 0x0000_0040
	// NonNull means no element is null (always true for non-pointer sources).
	NonNull uint = 0x0000_0100
	// Immutable means the source will not be structurally modified during traversal.
	Immutable uint = 0x0000_0400
	// Concurrent means the source may be safely modified concurrently by other goroutines.
	Concurrent uint = 0x0000_1000
	// Subsized means every spliterator produced by TrySplit is itself Sized.
	Subsized uint = 0x0000_4000
)

// Spliterator is a traversable, splittable source of elements.
//
// It mirrors java.util.Spliterator: TryAdvance consumes one element, TrySplit
// partitions the remainder, and Characteristics advertises ordering/size
// guarantees.
type Spliterator[T any] interface {
	// TryAdvance passes the next element to action and returns true if an
	// element remained; otherwise it returns false and does nothing.
	TryAdvance(action func(T)) bool

	// TrySplit attempts to split off a prefix of the remaining elements into a
	// new spliterator, leaving the suffix in the receiver (the Java
	// convention). The returned bool is false when the source is too small to
	// divide usefully, in which case the receiver is unchanged and should be
	// traversed sequentially.
	TrySplit() (Spliterator[T], bool)

	// EstimateSize estimates the number of elements that would be traversed by
	// ForEachRemaining. It is exact when the Sized characteristic is reported.
	EstimateSize() int

	// Characteristics is the bitwise-OR of this spliterator's characteristic flags.
	Characteristics() uint

	// ForEachRemaining traverses every remaining element, applying action to each.
	ForEachRemaining(action func(T))
}

// HasCharacteristics reports whether all of the given characteristic bits are
// set on sp.
func HasCharacteristics[T any](sp Spliterator[T], flags uint) bool {
	return sp.Characteristics()&flags == flags
}

// ExactSize returns the exact remaining count and true if sp is Sized,
// otherwise it returns 0 and false.
func ExactSize[T any](sp Spliterator[T]) (int, bool) {
	if HasCharacteristics(sp, Sized) {
		return sp.EstimateSize(), true
	}
	return 0, false
}

// SliceSpliterator is a Spliterator over a slice.
//
// Splitting is O(1): the backing slice is halved at its midpoint. It reports
// Ordered | Sized | Subsized since slices have an exact, stable length.
type SliceSpliterator[T any] struct {
	slice []T
}

// NewSliceSpliterator returns a spliterator covering the whole of slice.
func NewSliceSpliterator[T any](slice []T) *SliceSpliterator[T] {
	return &SliceSpliterator[T]{slice: slice}
}

// Remainder returns the not-yet-traversed remainder, as a slice.
func (s *SliceSpliterator[T]) Remainder() []T {
	return s.slice
}

// TryAdvance passes the next element to action and returns true, or returns
// false when the slice is exhausted.
func (s *SliceSpliterator[T]) TryAdvance(action func(T)) bool {
	if len(s.slice) == 0 {
		return false
	}
	first := s.slice[0]
	s.slice = s.slice[1:]
	action(first)
	return true
}

// TrySplit splits the remaining slice at its midpoint, returning the prefix as
// a new spliterator and keeping the suffix in the receiver. It returns false
// (leaving the receiver unchanged) when fewer than two elements remain.
func (s *SliceSpliterator[T]) TrySplit() (Spliterator[T], bool) {
	n := len(s.slice)
	if n < 2 {
		return nil, false
	}
	mid := n / 2
	prefix := s.slice[:mid]
	s.slice = s.slice[mid:]
	return &SliceSpliterator[T]{slice: prefix}, true
}

// EstimateSize returns the exact number of remaining elements.
func (s *SliceSpliterator[T]) EstimateSize() int {
	return len(s.slice)
}

// Characteristics returns Ordered | Sized | Subsized.
func (s *SliceSpliterator[T]) Characteristics() uint {
	return Ordered | Sized | Subsized
}

// ForEachRemaining applies action to every remaining element, in order, and
// exhausts the spliterator.
func (s *SliceSpliterator[T]) ForEachRemaining(action func(T)) {
	for _, v := range s.slice {
		action(v)
	}
	s.slice = nil
}
