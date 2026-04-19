// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Package parallel provides batch-parallel operations on slices.
//
// Unlike element-by-element parallelism (e.g. Spliterator), these utilities
// split the input into contiguous batches. Each batch runs in its own
// goroutine, accumulates results locally, and the results are combined
// after all goroutines finish.
//
// For inputs smaller than DefaultMinForkSize, all operations fall back to
// sequential execution — the goroutine overhead isn't worth it.
//
// This is a Go port of Eclipse Collections' ParallelIterate utility class.
package parallel

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// DefaultMinForkSize is the minimum slice length before parallelism kicks in.
// Below this threshold, operations run sequentially.
const DefaultMinForkSize = 10_000

// defaultTaskCount returns the default number of goroutines to use.
// Matches Java's formula: (NCPU + 1) * 2, capped at 200.
func defaultTaskCount() int {
	n := (runtime.NumCPU() + 1) * 2
	if n > 200 {
		return 200
	}
	return n
}

// ForEach calls f for every element of data in parallel batches.
// Order of calls within a batch is sequential (ascending index);
// order between batches is not guaranteed.
func ForEach[T any](data []T, f func(T)) {
	ForEachWith(data, f, DefaultMinForkSize, defaultTaskCount())
}

// ForEachWith is like ForEach with explicit minForkSize and taskCount.
func ForEachWith[T any](data []T, f func(T), minForkSize, taskCount int) {
	n := len(data)
	if n == 0 {
		return
	}
	if n < minForkSize || taskCount <= 1 {
		for _, v := range data {
			f(v)
		}
		return
	}

	batches := splitBatches(n, taskCount)
	var wg sync.WaitGroup
	wg.Add(len(batches))
	for _, b := range batches {
		go func(lo, hi int) {
			defer wg.Done()
			for i := lo; i < hi; i++ {
				f(data[i])
			}
		}(b.lo, b.hi)
	}
	wg.Wait()
}

// Select returns a new slice containing only elements that satisfy the
// predicate. The result preserves the relative order of the input.
func Select[T any](data []T, predicate func(T) bool) []T {
	return SelectWith(data, predicate, DefaultMinForkSize, defaultTaskCount())
}

// SelectWith is like Select with explicit minForkSize and taskCount.
func SelectWith[T any](data []T, predicate func(T) bool, minForkSize, taskCount int) []T {
	n := len(data)
	if n == 0 {
		return nil
	}
	if n < minForkSize || taskCount <= 1 {
		return selectSeq(data, predicate)
	}

	batches := splitBatches(n, taskCount)
	parts := make([][]T, len(batches))
	var wg sync.WaitGroup
	wg.Add(len(batches))
	for idx, b := range batches {
		go func(i, lo, hi int) {
			defer wg.Done()
			parts[i] = selectSeq(data[lo:hi], predicate)
		}(idx, b.lo, b.hi)
	}
	wg.Wait()
	return concat(parts)
}

// Reject returns a new slice containing only elements that do NOT satisfy
// the predicate. The result preserves the relative order of the input.
func Reject[T any](data []T, predicate func(T) bool) []T {
	return RejectWith(data, predicate, DefaultMinForkSize, defaultTaskCount())
}

// RejectWith is like Reject with explicit minForkSize and taskCount.
func RejectWith[T any](data []T, predicate func(T) bool, minForkSize, taskCount int) []T {
	return SelectWith(data, func(v T) bool { return !predicate(v) }, minForkSize, taskCount)
}

// Collect applies the transform function to each element in parallel and
// returns the results. The output order matches the input order.
func Collect[T any, R any](data []T, transform func(T) R) []R {
	return CollectWith(data, transform, DefaultMinForkSize, defaultTaskCount())
}

// CollectWith is like Collect with explicit minForkSize and taskCount.
func CollectWith[T any, R any](data []T, transform func(T) R, minForkSize, taskCount int) []R {
	n := len(data)
	if n == 0 {
		return nil
	}
	if n < minForkSize || taskCount <= 1 {
		return collectSeq(data, transform)
	}

	result := make([]R, n)
	batches := splitBatches(n, taskCount)
	var wg sync.WaitGroup
	wg.Add(len(batches))
	for _, b := range batches {
		go func(lo, hi int) {
			defer wg.Done()
			for i := lo; i < hi; i++ {
				result[i] = transform(data[i])
			}
		}(b.lo, b.hi)
	}
	wg.Wait()
	return result
}

// Count returns the number of elements satisfying the predicate,
// counted in parallel batches.
func Count[T any](data []T, predicate func(T) bool) int {
	return CountWith(data, predicate, DefaultMinForkSize, defaultTaskCount())
}

// CountWith is like Count with explicit minForkSize and taskCount.
func CountWith[T any](data []T, predicate func(T) bool, minForkSize, taskCount int) int {
	n := len(data)
	if n == 0 {
		return 0
	}
	if n < minForkSize || taskCount <= 1 {
		return countSeq(data, predicate)
	}

	batches := splitBatches(n, taskCount)
	counts := make([]int, len(batches))
	var wg sync.WaitGroup
	wg.Add(len(batches))
	for idx, b := range batches {
		go func(i, lo, hi int) {
			defer wg.Done()
			counts[i] = countSeq(data[lo:hi], predicate)
		}(idx, b.lo, b.hi)
	}
	wg.Wait()

	total := 0
	for _, c := range counts {
		total += c
	}
	return total
}

// AnySatisfy returns true if any element satisfies the predicate.
// Short-circuits across batches: once one batch finds a match, remaining
// batches are not started (but already-running batches complete).
func AnySatisfy[T any](data []T, predicate func(T) bool) bool {
	return AnySatisfyWith(data, predicate, DefaultMinForkSize, defaultTaskCount())
}

// AnySatisfyWith is like AnySatisfy with explicit minForkSize and taskCount.
func AnySatisfyWith[T any](data []T, predicate func(T) bool, minForkSize, taskCount int) bool {
	n := len(data)
	if n == 0 {
		return false
	}
	if n < minForkSize || taskCount <= 1 {
		for _, v := range data {
			if predicate(v) {
				return true
			}
		}
		return false
	}

	batches := splitBatches(n, taskCount)
	var found atomic.Bool
	var wg sync.WaitGroup
	wg.Add(len(batches))
	for _, b := range batches {
		go func(lo, hi int) {
			defer wg.Done()
			for i := lo; i < hi; i++ {
				if found.Load() {
					return
				}
				if predicate(data[i]) {
					found.Store(true)
					return
				}
			}
		}(b.lo, b.hi)
	}
	wg.Wait()
	return found.Load()
}

// AllSatisfy returns true if all elements satisfy the predicate.
func AllSatisfy[T any](data []T, predicate func(T) bool) bool {
	return !AnySatisfyWith(data, func(v T) bool { return !predicate(v) }, DefaultMinForkSize, defaultTaskCount())
}

// Sum returns the sum of all elements computed in parallel batches.
func Sum[T interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
}](data []T) T {
	return SumWith(data, DefaultMinForkSize, defaultTaskCount())
}

// SumWith is like Sum with explicit minForkSize and taskCount.
func SumWith[T interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
}](data []T, minForkSize, taskCount int) T {
	n := len(data)
	if n == 0 {
		var zero T
		return zero
	}
	if n < minForkSize || taskCount <= 1 {
		var sum T
		for _, v := range data {
			sum += v
		}
		return sum
	}

	batches := splitBatches(n, taskCount)
	sums := make([]T, len(batches))
	var wg sync.WaitGroup
	wg.Add(len(batches))
	for idx, b := range batches {
		go func(i, lo, hi int) {
			defer wg.Done()
			var s T
			for j := lo; j < hi; j++ {
				s += data[j]
			}
			sums[i] = s
		}(idx, b.lo, b.hi)
	}
	wg.Wait()

	var total T
	for _, s := range sums {
		total += s
	}
	return total
}

// ── internal helpers ──────────────────────────────────────────────────

type batch struct{ lo, hi int }

func splitBatches(n, taskCount int) []batch {
	if taskCount > n {
		taskCount = n
	}
	if taskCount < 1 {
		taskCount = 1
	}
	batches := make([]batch, taskCount)
	chunkSize := n / taskCount
	remainder := n % taskCount
	lo := 0
	for i := range batches {
		hi := lo + chunkSize
		if i < remainder {
			hi++
		}
		batches[i] = batch{lo, hi}
		lo = hi
	}
	return batches
}

func selectSeq[T any](data []T, predicate func(T) bool) []T {
	var result []T
	for _, v := range data {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

func collectSeq[T any, R any](data []T, transform func(T) R) []R {
	result := make([]R, len(data))
	for i, v := range data {
		result[i] = transform(v)
	}
	return result
}

func countSeq[T any](data []T, predicate func(T) bool) int {
	n := 0
	for _, v := range data {
		if predicate(v) {
			n++
		}
	}
	return n
}

func concat[T any](parts [][]T) []T {
	total := 0
	for _, p := range parts {
		total += len(p)
	}
	if total == 0 {
		return nil
	}
	result := make([]T, 0, total)
	for _, p := range parts {
		result = append(result, p...)
	}
	return result
}
