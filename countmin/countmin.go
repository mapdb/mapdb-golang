// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Package countmin is the Count-Min Sketch + Space-Saving probabilistic
// frequency port (see spec/features/count-min.md in mapdb-collection-spec).
//
// Both structures ride the deterministic hash pipeline (package hash). The
// counter matrix / monitored set after a given add-sequence is the
// cross-language oracle: because the d column indices are exactly
// hash.Positions(encodeInt32(item), w, d) -- bit-identical across all five
// ports -- the entire matrix, every estimate, and total are bit-identical too.
// No floating point appears in the deterministic surface (the only float,
// NewCountMinOptimal, is native-test-only and never used by shared scenarios).
package countmin

import (
	"math"

	"github.com/mapdb/mapdb-golang/hash"
)

// eulerE is Euler's number e, used only by the native-only NewCountMinOptimal.
const eulerE = math.E

// CountMin is a Count-Min Sketch over a flat row-major []uint64 matrix of d*w
// counters: a d x w integer counter matrix giving a one-sided over-estimate of
// an element's frequency.
//
// Construct with NewCountMinWithParams (the only constructor the cross-language
// scenarios use) or the native-only NewCountMinOptimal.
type CountMin struct {
	d uint32
	w uint32
	// matrix is the flat row-major matrix: counter matrix[r*w + col] is row r,
	// column col. Length is exactly d*w.
	matrix []uint64
	// total is the running sum of every count argument (the stream length N),
	// saturating.
	total uint64
}

// saturatingAddU64 returns min(a+b, math.MaxUint64) without wrapping. Go has no
// native saturating add, so this is the explicit guard the spec pins.
func saturatingAddU64(a, b uint64) uint64 {
	if a > math.MaxUint64-b {
		return math.MaxUint64
	}
	return a + b
}

// NewCountMinWithParams constructs a d x w sketch with all counters zero. d is
// the depth (rows / hash functions = the k argument to positions); w is the
// width (columns per row = the m argument to positions).
//
// Panics if w == 0 (a zero-column row holds nothing and every modulo would
// divide by zero) -- identical to Bloom's m = 0 ruling. d == 0 is legal and
// degenerate (an empty matrix; Estimate returns math.MaxUint64). Panics if d*w
// overflows the addressable matrix size (a native allocation concern; the
// shared suite never constructs such a sketch).
func NewCountMinWithParams(d, w uint32) *CountMin {
	if w == 0 {
		panic("CountMin width w must be non-zero")
	}
	// d*w in uint64 to detect a length that overflows a Go int (native
	// allocation limit; never hit by the small shared scenarios).
	length := uint64(d) * uint64(w)
	if length > uint64(math.MaxInt) {
		panic("CountMin d*w overflows int (native allocation limit)")
	}
	return &CountMin{
		d:      d,
		w:      w,
		matrix: make([]uint64, int(length)),
	}
}

// NewCountMinOptimal is a native-only convenience constructor sizing the sketch
// from a target additive error epsilon (relative to the total) and failure
// probability delta using the standard Count-Min formulas
// w = ceil(e/epsilon), d = ceil(ln(1/delta)), then delegating to
// NewCountMinWithParams.
//
// Float-quarantined: never used by the cross-language scenarios (the ln/e/ceil
// derivation can drift across libm implementations). It is native-tested
// against the pinned integer table.
//
// Panics unless 0 < epsilon < 1 and 0 < delta < 1; values <= 0, >= 1, NaN, or
// +-Infinity are invalid (they would divide by zero, take ln of a non-positive
// value, or yield a non-finite (d, w)).
func NewCountMinOptimal(epsilon, delta float64) *CountMin {
	if !(epsilon > 0.0 && epsilon < 1.0) {
		panic("CountMin optimal requires 0 < epsilon < 1")
	}
	if !(delta > 0.0 && delta < 1.0) {
		panic("CountMin optimal requires 0 < delta < 1")
	}
	w := math.Ceil(eulerE / epsilon)
	d := math.Ceil(math.Log(1.0 / delta))
	if !(isFiniteGE1(w) && isFiniteGE1(d)) {
		panic("CountMin optimal produced a non-finite (d, w)")
	}
	// Reject out-of-uint32-range widths/depths: float→uint32 conversion of a
	// value above MaxUint32 is implementation-defined per the Go spec (e.g.
	// tiny epsilon → w ≈ 2.7e10). Mirror bloom.Optimal's explicit bound.
	if w > float64(math.MaxUint32) || d > float64(math.MaxUint32) {
		panic("CountMin optimal produced (d, w) exceeding uint32")
	}
	return NewCountMinWithParams(uint32(d), uint32(w))
}

func isFiniteGE1(x float64) bool {
	return !math.IsInf(x, 0) && !math.IsNaN(x) && x >= 1.0
}

// columns returns the d column indices for item, one per row, in derivation
// order: positions(encodeInt32(item), m = w, k = d). cols[r] is the column
// touched in row r.
func (c *CountMin) columns(item int32) []uint32 {
	// Element encoding: i32 -> reinterpret uint32 -> 4 LE bytes -> byte
	// positions path (length fold applied), identical to Bloom.
	u := uint32(item)
	b := []byte{byte(u), byte(u >> 8), byte(u >> 16), byte(u >> 24)}
	return hash.Positions(b, c.w, c.d)
}

// Add increments the d selected counters (one per row) by count, saturating at
// math.MaxUint64. Add(item, count) is not observably five AddOne calls but
// yields the identical counters (increments are commutative). count == 0 is
// legal: a no-op on the counters that still updates total (by 0). Plain CMS --
// increments all d counters (no conservative update).
func (c *CountMin) Add(item int32, count uint64) {
	cols := c.columns(item)
	for r, col := range cols {
		idx := r*int(c.w) + int(col)
		c.matrix[idx] = saturatingAddU64(c.matrix[idx], count)
	}
	c.total = saturatingAddU64(c.total, count)
}

// AddOne is a convenience for Add(item, 1); identical bits.
func (c *CountMin) AddOne(item int32) {
	c.Add(item, 1)
}

// Estimate returns the frequency estimate for item: the MIN over the d rows of
// the selected counter. Never under-estimates (within the uint64 domain). For
// d == 0 the MIN over zero rows is the empty-min identity math.MaxUint64.
func (c *CountMin) Estimate(item int32) uint64 {
	cols := c.columns(item)
	min := uint64(math.MaxUint64)
	for r, col := range cols {
		idx := r*int(c.w) + int(col)
		if c.matrix[idx] < min {
			min = c.matrix[idx]
		}
	}
	return min
}

// Total returns the running sum of every count argument ever added (the stream
// length N), saturating at math.MaxUint64.
func (c *CountMin) Total() uint64 {
	return c.total
}

// Depth returns the depth d (number of rows / hash functions).
func (c *CountMin) Depth() uint32 {
	return c.d
}

// Width returns the width w (number of columns per row).
func (c *CountMin) Width() uint32 {
	return c.w
}

// ToCounters returns the full counter matrix as d*w values, row-major (row 0
// first, column 0 first within a row). Dense (all cells, including zeros).
func (c *CountMin) ToCounters() []uint64 {
	out := make([]uint64, len(c.matrix))
	copy(out, c.matrix)
	return out
}
