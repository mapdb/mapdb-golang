// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Package fenwick is a Fenwick tree / Binary Indexed Tree (prefix & range sums).
//
// A fixed-size index structure with O(log n) point-update and O(log n)
// prefix/range sum over signed int32 element values accumulated in a wrapping
// int64 accumulator. See spec/features/fenwick.md (mapdb-collection-spec) for
// the pinned design; this is a port of the frozen reference
// (mapdb-rust src/fenwick.rs).
//
// Pinned invariants realized here:
//   - Indexing: the public API is 0-based (0 .. n-1); the BIT is classically
//     1-based internally (internal = public + 1). The 1-based index is never
//     observable. The backing array is length n+1 with slot 0 unused.
//   - Ranges: PrefixSum(i) is the INCLUSIVE prefix [0..=i]; RangeSum(lo, hi) is
//     the INCLUSIVE closed range [lo..=hi]; Total() == PrefixSum(n-1) (and 0 for
//     the empty tree).
//   - Accumulator: each slot and every sum is a wrapping two's-complement int64.
//     Go's native int64 arithmetic wraps, so the wrap needs no explicit clamp.
//     The per-element value widens to int64 and does NOT re-wrap at int32, so Get
//     returns int64.
//   - Out-of-range: mutators (Update/Set), Get, and PrefixSum panic on an
//     out-of-domain index (both i < 0 and i >= n, since Go indexes with a signed
//     int). RangeSum validates BOTH endpoints first (out-of-domain endpoint
//     panics), THEN returns 0 for an empty lo > hi range.
package fenwick

import (
	"fmt"
	"iter"
)

// FenwickTree is a Fenwick tree (Binary Indexed Tree) over int32 element values
// with a wrapping int64 accumulator. Fixed size; no resize.
//
// The backing array tree has length n+1: slot 0 is the unused BIT terminator and
// tree[1 .. n] are the 1-based partial sums.
type FenwickTree struct {
	// tree holds 1-based partial sums; tree[0] is unused. Length is n+1.
	tree []int64
	// n is the public size (number of valid 0-based indices).
	n int
}

// NewFenwickTreeWithSize constructs an all-zero tree of size n. n must be
// non-negative; NewFenwickTreeWithSize(0) is a valid empty tree (Total() == 0,
// IsEmpty() == true).
//
// Panics if n < 0 (the same trap posture as the out-of-range mutators; Go's
// constructor takes a signed int).
func NewFenwickTreeWithSize(n int) *FenwickTree {
	if n < 0 {
		panic(fmt.Sprintf("FenwickTree size %d must be non-negative", n))
	}
	return &FenwickTree{
		tree: make([]int64, n+1),
		n:    n,
	}
}

// NewFenwickTreeFromValues builds from an initial int32 array; the tree has
// Size() == len(values) and Get(i) == values[i]. Uses the O(n) in-place build,
// which produces the identical tree as NewFenwickTreeWithSize(len) then
// Update(i, values[i]) for each i (build determinism).
//
// NewFenwickTreeFromValues(nil) / an empty slice builds a valid empty tree.
func NewFenwickTreeFromValues(values []int32) *FenwickTree {
	n := len(values)
	tree := make([]int64, n+1)
	// Seed each 1-based slot with the (widened) element value.
	for i, v := range values {
		tree[i+1] = int64(v)
	}
	// O(n) in-place build: push each slot's running sum to its parent.
	// Over the 1-based array: parent = i + (i & -i).
	for i := 1; i <= n; i++ {
		parent := i + lowbit(i)
		if parent <= n {
			tree[parent] += tree[i]
		}
	}
	return &FenwickTree{tree: tree, n: n}
}

// Len returns the number of valid 0-based indices.
func (f *FenwickTree) Len() int {
	return f.n
}

// Size is an alias for Len (the spec's structural key).
func (f *FenwickTree) Size() int {
	return f.n
}

// IsEmpty reports whether the tree is empty (n == 0).
func (f *FenwickTree) IsEmpty() bool {
	return f.n == 0
}

// Update adds delta (int32, widened to int64) to the value at 0-based index i.
//
// Panics if i is out of the fixed 0 .. n-1 domain (i < 0 or i >= n).
func (f *FenwickTree) Update(i int, delta int32) {
	f.addInternal(i, int64(delta))
}

// Set point-assigns: makes the value at i equal value (int32).
//
// Implemented as a Fenwick difference-add computed in wrapping int64
// (delta = int64(value) - Get(i)), NOT routed through the int32 Update
// signature -- so the internal delta stays exact even when the current slot
// value already exceeds int32.
//
// Panics if i is out of range (i < 0 or i >= n).
func (f *FenwickTree) Set(i int, value int32) {
	delta := int64(value) - f.Get(i)
	f.addInternal(i, delta)
}

// Get returns the single logical value currently at 0-based index i, as int64.
// Equivalent to RangeSum(i, i).
//
// Panics if i is out of range (i < 0 or i >= n).
func (f *FenwickTree) Get(i int) int64 {
	if i < 0 || i >= f.n {
		panic(fmt.Sprintf("FenwickTree.Get index %d out of range 0..%d", i, f.n))
	}
	// Get(i) == PrefixSum(i) - PrefixSum(i-1); PrefixSum(-1) := 0.
	if i == 0 {
		return f.prefixSumInternal(0)
	}
	return f.prefixSumInternal(i) - f.prefixSumInternal(i-1)
}

// Values returns an iter.Seq that yields each element value f.Get(0..n-1) in
// index order (law 1). It walks prefix sums incrementally, so the whole sweep is
// O(n log n) — the same as calling Get per index, but rangeable and bridgeable
// into the seq/par layers (seq.From(tree.Values())). Iteration is lazy and honors
// an early break.
func (f *FenwickTree) Values() iter.Seq[int64] {
	return func(yield func(int64) bool) {
		prev := int64(0) // PrefixSum(-1) := 0
		for i := 0; i < f.n; i++ {
			cur := f.prefixSumInternal(i)
			if !yield(cur - prev) {
				return
			}
			prev = cur
		}
	}
}

// PrefixSum returns the inclusive prefix sum Σ values[0..=i], as wrapping int64.
//
// Panics if i is out of range (i < 0 or i >= n).
func (f *FenwickTree) PrefixSum(i int) int64 {
	if i < 0 || i >= f.n {
		panic(fmt.Sprintf("FenwickTree.PrefixSum index %d out of range 0..%d", i, f.n))
	}
	return f.prefixSumInternal(i)
}

// RangeSum returns the inclusive range sum Σ values[lo..=hi], as wrapping int64.
//
// Validates BOTH endpoints first: lo and hi must be valid public indices
// (0 <= lo < n and 0 <= hi < n). Only after both are valid, if lo > hi the range
// is empty and returns 0.
//
// Panics if lo or hi is out of range (an out-of-domain endpoint). On the empty
// tree every call panics (no valid endpoint exists). An empty closed range
// (lo > hi, both endpoints valid) returns 0 and does NOT panic.
func (f *FenwickTree) RangeSum(lo, hi int) int64 {
	if lo < 0 || lo >= f.n {
		panic(fmt.Sprintf("FenwickTree.RangeSum lo %d out of range 0..%d", lo, f.n))
	}
	if hi < 0 || hi >= f.n {
		panic(fmt.Sprintf("FenwickTree.RangeSum hi %d out of range 0..%d", hi, f.n))
	}
	// Both endpoints valid; an empty closed range (lo > hi) is a defined 0.
	if lo > hi {
		return 0
	}
	// RangeSum = PrefixSum(hi) - PrefixSum(lo-1); PrefixSum(-1) := 0.
	upper := f.prefixSumInternal(hi)
	var lower int64
	if lo > 0 {
		lower = f.prefixSumInternal(lo - 1)
	}
	return upper - lower
}

// Total returns the grand total Σ of all values, == PrefixSum(n-1) for n >= 1,
// and 0 for the empty tree.
func (f *FenwickTree) Total() int64 {
	if f.n == 0 {
		return 0
	}
	return f.prefixSumInternal(f.n - 1)
}

// CanonicalTree returns the canonical 1-based BIT projection: a length-n int64
// array where element j-1 (0-based in the returned slice) is the partial sum the
// tree stores for the 1-based index j -- i.e. tree[1 .. n]. This is the
// layout-independent secondary determinism oracle.
func (f *FenwickTree) CanonicalTree() []int64 {
	out := make([]int64, f.n)
	copy(out, f.tree[1:f.n+1])
	return out
}

// ---- internals (1-based BIT navigation) -----------------------------------

// addInternal adds a wrapping-int64 delta at 0-based index i via the low-bit
// walk. Panics if i is out of range (shared by Update/Set/Get's mutator path).
func (f *FenwickTree) addInternal(i int, delta int64) {
	if i < 0 || i >= f.n {
		panic(fmt.Sprintf("FenwickTree mutator index %d out of range 0..%d", i, f.n))
	}
	j := i + 1 // public -> 1-based BIT
	for j <= f.n {
		f.tree[j] += delta
		j += lowbit(j)
	}
}

// prefixSumInternal computes the inclusive prefix sum for 0-based index i (the
// caller guarantees 0 <= i < n).
func (f *FenwickTree) prefixSumInternal(i int) int64 {
	var acc int64
	j := i + 1 // public -> 1-based BIT
	for j > 0 {
		acc += f.tree[j]
		j -= lowbit(j)
	}
	return acc
}

// lowbit returns the low bit (j & -j) of a 1-based index (j >= 1).
func lowbit(j int) int {
	return j & -j
}
