// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package seq

import (
	"container/heap"
	"iter"
	"slices"
	"sync"

	"github.com/mapdb/mapdb-golang/object"
)

// This file holds the ordering ops: the two eager-materializing methods
// (Sorted, Reversed), the re-runnability adapter (Cache), the OnEmpty fallback,
// and the k-way ordered merge that fronts the data pump.

// Sorted returns the elements of s ordered by cmp (cmp(a,b) < 0 means a orders
// first). It is eager: it consumes s once when called and returns a re-runnable
// Seq over the sorted result. O(n) memory, O(n log n) time. Ranging the result
// does not re-sort.
func (s Seq[T]) Sorted(cmp func(a, b T) int) Seq[T] {
	buf := s.ToSlice()
	slices.SortFunc(buf, cmp)
	return FromSlice(buf)
}

// Reversed returns the elements of s in reverse encounter order. Eager: consumes
// s once when called and returns a re-runnable Seq. O(n) memory.
func (s Seq[T]) Reversed() Seq[T] {
	buf := s.ToSlice()
	slices.Reverse(buf)
	return FromSlice(buf)
}

// OnEmpty returns s if it yields at least one element, otherwise the fallback.
// Lazy: it does not pre-run s — it forwards s's elements as they arrive and only
// switches to fallback if s turned out to be empty. O(1) memory.
func (s Seq[T]) OnEmpty(fallback Seq[T]) Seq[T] {
	return func(yield func(T) bool) {
		empty := true
		for v := range s {
			empty = false
			if !yield(v) {
				return
			}
		}
		if empty {
			for v := range fallback {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// Cache upgrades a single-shot Seq into a re-runnable one by memoizing. The first
// iteration that runs to COMPLETION materializes s into a slice; every later
// iteration replays that slice. A first iteration that breaks early caches
// nothing (a single-shot source is then burned, exactly as any single-shot Seq).
// O(n) memory once completed. The field access is mutex-guarded so concurrent
// replays are race-free; two racing FIRST iterations are not defined to share
// work — each runs the source — but do not corrupt the cache.
func (s Seq[T]) Cache() Seq[T] {
	var (
		mu     sync.Mutex
		cached []T
		done   bool
	)
	return func(yield func(T) bool) {
		mu.Lock()
		filled, snapshot := done, cached
		mu.Unlock()
		if filled {
			for _, v := range snapshot {
				if !yield(v) {
					return
				}
			}
			return
		}
		// Not yet cached: run the source, buffering as we forward.
		var buf []T
		complete := true
		for v := range s {
			buf = append(buf, v)
			if !yield(v) {
				complete = false
				break // consumer stopped early — do not cache a partial result
			}
		}
		if complete {
			mu.Lock()
			if !done {
				cached, done = buf, true
			}
			mu.Unlock()
		}
	}
}

// MergeSorted merges already-sorted inputs into one sorted Seq under cmp, via a
// k-way heap merge (one iter.Pull per input, torn down when iteration ends).
// Inputs MUST each be sorted by cmp; if not, the output order is undefined (use
// MustBeSorted to guard). Lazy and re-runnable when the inputs are; O(k) memory
// for k inputs. Stable: equal elements keep input order (lower input index first).
func MergeSorted[T any](cmp object.Comparator[T], ss ...Seq[T]) Seq[T] {
	return mergeSorted(cmp, false, ss)
}

// MergeSortedDistinct is MergeSorted that also collapses runs of equal elements
// (equal under cmp) to a single element. Assumes each input is itself sorted and
// distinct; across inputs it de-dups the merged stream. Lazy, O(k) memory.
func MergeSortedDistinct[T any](cmp object.Comparator[T], ss ...Seq[T]) Seq[T] {
	return mergeSorted(cmp, true, ss)
}

func mergeSorted[T any](cmp object.Comparator[T], distinct bool, ss []Seq[T]) Seq[T] {
	return func(yield func(T) bool) {
		nexts := make([]func() (T, bool), len(ss))
		stops := make([]func(), len(ss))
		for i, s := range ss {
			nexts[i], stops[i] = iter.Pull(iter.Seq[T](s))
		}
		defer func() {
			for _, stop := range stops {
				stop()
			}
		}()
		h := &mergeHeap[T]{cmp: cmp}
		for i, next := range nexts {
			if v, ok := next(); ok {
				h.items = append(h.items, mergeItem[T]{val: v, src: i})
			}
		}
		heap.Init(h)
		var prev T
		havePrev := false
		for h.Len() > 0 {
			top := h.items[0]
			if !distinct || !havePrev || cmp(top.val, prev) != 0 {
				if !yield(top.val) {
					return
				}
				prev, havePrev = top.val, true
			}
			if v, ok := nexts[top.src](); ok {
				h.items[0] = mergeItem[T]{val: v, src: top.src}
				heap.Fix(h, 0)
			} else {
				heap.Pop(h)
			}
		}
	}
}

// MustBeSorted returns s unchanged but panics during iteration if it observes an
// element out of cmp order (an element strictly less than its predecessor). Lazy,
// O(1) memory; use it to guard a MergeSorted input in tests or trusted pipelines.
func MustBeSorted[T any](s Seq[T], cmp object.Comparator[T]) Seq[T] {
	return func(yield func(T) bool) {
		var prev T
		havePrev := false
		for v := range s {
			if havePrev && cmp(v, prev) < 0 {
				panic("seq: MustBeSorted: sequence is not sorted")
			}
			prev, havePrev = v, true
			if !yield(v) {
				return
			}
		}
	}
}

// mergeItem is a heap entry: an element plus the index of the input it came from.
type mergeItem[T any] struct {
	val T
	src int
}

// mergeHeap is a min-heap on cmp, breaking ties by input index so the merge is
// stable (lower input index emitted first among equal elements).
type mergeHeap[T any] struct {
	items []mergeItem[T]
	cmp   object.Comparator[T]
}

func (h *mergeHeap[T]) Len() int { return len(h.items) }
func (h *mergeHeap[T]) Less(i, j int) bool {
	if c := h.cmp(h.items[i].val, h.items[j].val); c != 0 {
		return c < 0
	}
	return h.items[i].src < h.items[j].src
}
func (h *mergeHeap[T]) Swap(i, j int) { h.items[i], h.items[j] = h.items[j], h.items[i] }
func (h *mergeHeap[T]) Push(x any)    { h.items = append(h.items, x.(mergeItem[T])) }
func (h *mergeHeap[T]) Pop() any {
	old := h.items
	n := len(old)
	it := old[n-1]
	h.items = old[:n-1]
	return it
}
