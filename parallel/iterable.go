// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Parallel views — the Go analogue of Eclipse Collections'
// RichIterable.asParallel(...) / ParallelIterable.
//
// Java's ParallelIterable is the work-stealing view: a Spliterator handed to a
// ForkJoinPool that recursively splits and steals work between threads. Go has
// no work-stealing pool, so this implementation is NOT true work stealing: it
// is goroutine-batch-backed, reusing the fixed-section batch machinery in
// parallel.go (splitBatches + one goroutine per batch). Terminal operations
// mirror the RichIterable surface used elsewhere in the family.

package parallel

import "sync"

// ParallelIterable is a parallel view over a slice. Construct it with
// AsParallel.
//
// It is goroutine-batch-backed rather than work-stealing: each terminal method
// splits the underlying slice into fixed contiguous batches (see splitBatches)
// and runs one goroutine per batch, falling back to sequential execution below
// minForkSize. Select/Reject/Collect restore input order; ForEach and the
// short-circuiting predicates make no ordering guarantee across batches.
type ParallelIterable[T any] struct {
	data        []T
	minForkSize int
	taskCount   int
}

// AsParallel returns a parallel view over data using the package defaults
// (DefaultMinForkSize and (NCPU+1)*2 tasks capped at 200).
func AsParallel[T any](data []T) ParallelIterable[T] {
	return ParallelIterable[T]{
		data:        data,
		minForkSize: DefaultMinForkSize,
		taskCount:   defaultTaskCount(),
	}
}

// AsParallelWith returns a parallel view over data with an explicit
// minForkSize and taskCount.
func AsParallelWith[T any](data []T, minForkSize, taskCount int) ParallelIterable[T] {
	return ParallelIterable[T]{
		data:        data,
		minForkSize: minForkSize,
		taskCount:   taskCount,
	}
}

// Len returns the number of elements in the view.
func (p ParallelIterable[T]) Len() int {
	return len(p.data)
}

// IsEmpty reports whether the view has no elements.
func (p ParallelIterable[T]) IsEmpty() bool {
	return len(p.data) == 0
}

// ForEach applies f to every element in parallel batches. Order within a batch
// is sequential; order across batches is unspecified.
func (p ParallelIterable[T]) ForEach(f func(T)) {
	ForEachWith(p.data, f, p.minForkSize, p.taskCount)
}

// Select returns the elements satisfying predicate, preserving input order.
func (p ParallelIterable[T]) Select(predicate func(T) bool) []T {
	return SelectWith(p.data, predicate, p.minForkSize, p.taskCount)
}

// Reject returns the elements NOT satisfying predicate, preserving input order.
func (p ParallelIterable[T]) Reject(predicate func(T) bool) []T {
	return RejectWith(p.data, predicate, p.minForkSize, p.taskCount)
}

// Count returns the number of elements satisfying predicate.
func (p ParallelIterable[T]) Count(predicate func(T) bool) int {
	return CountWith(p.data, predicate, p.minForkSize, p.taskCount)
}

// AnySatisfy reports whether any element satisfies predicate. It short-circuits
// across batches once a match is found.
func (p ParallelIterable[T]) AnySatisfy(predicate func(T) bool) bool {
	return AnySatisfyWith(p.data, predicate, p.minForkSize, p.taskCount)
}

// AllSatisfy reports whether every element satisfies predicate. It is vacuously
// true for an empty view.
func (p ParallelIterable[T]) AllSatisfy(predicate func(T) bool) bool {
	return !AnySatisfyWith(p.data, func(v T) bool { return !predicate(v) }, p.minForkSize, p.taskCount)
}

// NoneSatisfy reports whether no element satisfies predicate.
func (p ParallelIterable[T]) NoneSatisfy(predicate func(T) bool) bool {
	return !p.AnySatisfy(predicate)
}

// Detect returns an element satisfying predicate and true, or the zero value
// and false if none match. Which matching element is returned is unspecified.
// It short-circuits across batches once a match is found.
func (p ParallelIterable[T]) Detect(predicate func(T) bool) (T, bool) {
	return detectWith(p.data, predicate, p.minForkSize, p.taskCount)
}

// ParallelCollect maps every element of p through transform in parallel,
// preserving input order in the result.
//
// It is a free function rather than a method because Go methods cannot
// introduce their own type parameter (the result type R), and is named to
// avoid colliding with the batch-level Collect in this package.
func ParallelCollect[T, R any](p ParallelIterable[T], transform func(T) R) []R {
	return CollectWith(p.data, transform, p.minForkSize, p.taskCount)
}

// detectWith finds an element satisfying predicate using the fixed-batch
// machinery, short-circuiting across batches via a shared found flag.
func detectWith[T any](data []T, predicate func(T) bool, minForkSize, taskCount int) (T, bool) {
	var zero T
	n := len(data)
	if n == 0 {
		return zero, false
	}
	if n < minForkSize || taskCount <= 1 {
		for _, v := range data {
			if predicate(v) {
				return v, true
			}
		}
		return zero, false
	}

	batches := splitBatches(n, taskCount)
	results := make([]T, len(batches))
	hits := make([]bool, len(batches))
	var done sync.Mutex
	stopped := false
	var wg sync.WaitGroup
	wg.Add(len(batches))
	for idx, b := range batches {
		go func(i, lo, hi int) {
			defer wg.Done()
			for j := lo; j < hi; j++ {
				done.Lock()
				s := stopped
				done.Unlock()
				if s {
					return
				}
				if predicate(data[j]) {
					results[i] = data[j]
					hits[i] = true
					done.Lock()
					stopped = true
					done.Unlock()
					return
				}
			}
		}(idx, b.lo, b.hi)
	}
	wg.Wait()

	// Prefer the earliest batch that found a match so the result is stable.
	for i := range hits {
		if hits[i] {
			return results[i], true
		}
	}
	return zero, false
}

// ParallelSum returns the sum of all elements in p computed in parallel
// batches.
//
// It is a free function rather than a method because it needs a numeric type
// constraint that the generic ParallelIterable[T] does not carry, and is named
// to avoid colliding with the batch-level Sum in this package.
func ParallelSum[T interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
}](p ParallelIterable[T]) T {
	return SumWith(p.data, p.minForkSize, p.taskCount)
}
