// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package bitset_test

// Hand-written conformance stamp (todo 14 §4). bitset is not part of the
// per-primitive codegen, so its law tests are written directly here, reusing the
// same internal/conformance predicates the generated families call. A bit set
// iterates set bits low→high, so law 1 is checked in ordered mode.

import (
	"testing"

	"github.com/mapdb/mapdb-golang/bitset"
	"github.com/mapdb/mapdb-golang/internal/conformance"
)

// conformanceBits is the shared fixture: a duplicate (1) collapses (bit already
// set), leaving the distinct ascending set {1,2,3,4,5,9} spread across words.
var conformanceBits = []int{3, 1, 4, 1, 5, 9, 2}

func buildConformanceBitSet() *bitset.BitSet {
	b := bitset.NewBitSet()
	for _, bit := range conformanceBits {
		b.Set(bit)
	}
	return b
}

func buildConformanceSyncBitSet() *bitset.SynchronizedBitSet {
	b := bitset.NewSynchronizedBitSet()
	for _, bit := range conformanceBits {
		b.Set(bit)
	}
	return b
}

// TestConformanceAllMatchesToSliceBitSet pins law 1: All() ≡ ToSlice(), in the
// bit set's ascending set-bit order.
func TestConformanceAllMatchesToSliceBitSet(t *testing.T) {
	b := buildConformanceBitSet()
	conformance.AllMatchesToSlice(t, b.All(), b.ToSlice(), true)
}

// TestConformanceSegmentsBitSet pins the Segments partition law: word-range
// segments cover every set bit exactly once and are re-runnable.
func TestConformanceSegmentsBitSet(t *testing.T) {
	b := buildConformanceBitSet()
	conformance.SegmentsCoverAll(t, b.All(), b.Segments)
}

func TestConformanceAllMatchesToSliceSyncBitSet(t *testing.T) {
	b := buildConformanceSyncBitSet()
	conformance.AllMatchesToSlice(t, b.All(), b.ToSlice(), true)
}

func TestConformanceSegmentsSyncBitSet(t *testing.T) {
	b := buildConformanceSyncBitSet()
	conformance.SegmentsCoverAll(t, b.All(), b.Segments)
}
