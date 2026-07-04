// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package seq

import "iter"

// Seq is a lazy, chainable sequence: a defined function type over iter.Seq[T].
// It ranges directly (for v := range s) and carries the fluent, type-preserving
// method set below. Type-changing operations (Map, FlatMap, Distinct, …) are
// free functions returning Seq so a pipeline keeps flowing.
type Seq[T any] iter.Seq[T]

// Std releases s as a standard library iter.Seq[T]. Zero-cost.
func (s Seq[T]) Std() iter.Seq[T] { return iter.Seq[T](s) }

// Filter returns the elements of s that satisfy pred. Lazy, O(1) memory,
// preserves re-runnability. ⟨EC: select⟩
func (s Seq[T]) Filter(pred func(T) bool) Seq[T] {
	return func(yield func(T) bool) {
		for v := range s {
			if pred(v) && !yield(v) {
				return
			}
		}
	}
}

// FilterNot returns the elements of s that do not satisfy pred. Lazy, O(1)
// memory, preserves re-runnability. ⟨EC: reject⟩
func (s Seq[T]) FilterNot(pred func(T) bool) Seq[T] {
	return func(yield func(T) bool) {
		for v := range s {
			if !pred(v) && !yield(v) {
				return
			}
		}
	}
}

// Take returns at most the first n elements of s. Lazy, O(1) memory. For n <= 0
// it yields nothing and — importantly — does not pull a single element from s,
// so Of(...).Take(0) over an infinite source terminates.
func (s Seq[T]) Take(n int) Seq[T] {
	return func(yield func(T) bool) {
		if n <= 0 {
			return
		}
		i := 0
		for v := range s {
			if !yield(v) {
				return
			}
			if i++; i >= n {
				return
			}
		}
	}
}

// TakeWhile returns the leading elements of s for which pred holds, stopping at
// (and not yielding) the first element that fails. Lazy, O(1) memory.
func (s Seq[T]) TakeWhile(pred func(T) bool) Seq[T] {
	return func(yield func(T) bool) {
		for v := range s {
			if !pred(v) || !yield(v) {
				return
			}
		}
	}
}

// Drop skips the first n elements of s and yields the rest. Lazy, O(1) memory.
// For n <= 0 it yields all of s.
func (s Seq[T]) Drop(n int) Seq[T] {
	return func(yield func(T) bool) {
		i := 0
		for v := range s {
			if i < n {
				i++
				continue
			}
			if !yield(v) {
				return
			}
		}
	}
}

// DropWhile skips the leading elements of s for which pred holds and yields the
// rest, starting at the first element that fails pred. Lazy, O(1) memory.
func (s Seq[T]) DropWhile(pred func(T) bool) Seq[T] {
	return func(yield func(T) bool) {
		dropping := true
		for v := range s {
			if dropping {
				if pred(v) {
					continue
				}
				dropping = false
			}
			if !yield(v) {
				return
			}
		}
	}
}

// Peek returns a Seq that calls f on each element as it passes, then yields it
// unchanged. Useful for observing a pipeline (logging, counting) without
// consuming it. Lazy, O(1) memory. ⟨EC: tap⟩
func (s Seq[T]) Peek(f func(T)) Seq[T] {
	return func(yield func(T) bool) {
		for v := range s {
			f(v)
			if !yield(v) {
				return
			}
		}
	}
}

// Concat returns s followed by each of others, in order. Lazy, O(1) memory.
// ⟨EC: concatenate⟩
func (s Seq[T]) Concat(others ...Seq[T]) Seq[T] {
	return func(yield func(T) bool) {
		for v := range s {
			if !yield(v) {
				return
			}
		}
		for _, o := range others {
			for v := range o {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// Cycle repeats s endlessly: its elements, then again from the start, forever.
// Lazy, O(1) memory, and single-shot-unsafe — s must be re-runnable (ranging it
// more than once must reproduce it). Pair with Take to bound it.
func (s Seq[T]) Cycle() Seq[T] {
	return func(yield func(T) bool) {
		for {
			empty := true
			for v := range s {
				empty = false
				if !yield(v) {
					return
				}
			}
			if empty {
				return // never loop forever on an empty source
			}
		}
	}
}

// Chunk groups s into consecutive slices of up to n elements (the final chunk
// may be shorter). It returns iter.Seq[[]T], not Seq[[]T]: a method on Seq[T]
// returning Seq[[]T] fails to compile ("instantiation cycle", T instantiated as
// []T), and chunking is normally terminal anyway. Wrap with seq.From if you need
// to keep chaining. Lazy in the source, O(n) per chunk. Panics if n <= 0.
// ⟨EC: chunk⟩
func (s Seq[T]) Chunk(n int) iter.Seq[[]T] {
	if n <= 0 {
		panic("seq: Chunk size must be positive")
	}
	return func(yield func([]T) bool) {
		batch := make([]T, 0, n)
		for v := range s {
			batch = append(batch, v)
			if len(batch) == n {
				if !yield(batch) {
					return
				}
				batch = make([]T, 0, n)
			}
		}
		if len(batch) > 0 {
			yield(batch)
		}
	}
}
