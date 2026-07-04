// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package rangev

import "testing"

// Regression for M-2 (todo/fable-golang/01-critical-bugs.md): PutCoalescing used a
// single ascending pass, so `merged` growing during iteration coalesced an abutting
// equal-valued chain rightward but not leftward — a direction-dependent result. The
// fix iterates to a fixpoint; the two mirror-image chains below must now produce the
// same single coalesced entry.
//
// The cross-language oracle (spec 20-range-set-map) only pins single-neighbour and
// bridge-two-separate-entries cases, both of which still pass; a pre-existing chain
// is unspecified there, so the chosen semantics is "no two connected equal-valued
// entries remain".

func TestRangeMapPutCoalescing_ChainRightIsDirectionIndependent(t *testing.T) {
	// Chain on the LEFT of the coalesced range.
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 2), 7)
	m.Put(ClosedOpen(2, 3), 7)
	m.PutCoalescing(ClosedOpen(3, 4), 7)
	assertEntries(t, m, entry(ClosedOpen(1, 4), 7))
}

func TestRangeMapPutCoalescing_ChainLeftIsDirectionIndependent(t *testing.T) {
	// Mirror image: chain on the RIGHT of the coalesced range.
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(2, 3), 7)
	m.Put(ClosedOpen(3, 4), 7)
	m.PutCoalescing(ClosedOpen(1, 2), 7)
	assertEntries(t, m, entry(ClosedOpen(1, 4), 7))
}

func TestRangeMapPutCoalescing_LongChainBothSides(t *testing.T) {
	m := NewInt32Int32RangeMap()
	for i := int32(0); i < 5; i++ {
		m.Put(ClosedOpen(i, i+1), 9)
	}
	for i := int32(6); i < 10; i++ {
		m.Put(ClosedOpen(i, i+1), 9)
	}
	// Bridge the two chains; everything equal-valued and connected collapses.
	m.PutCoalescing(ClosedOpen(5, 6), 9)
	assertEntries(t, m, entry(ClosedOpen(0, 10), 9))
}

// A different value in the middle must still block coalescing across it.
func TestRangeMapPutCoalescing_ChainBlockedByDifferentValue(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 2), 7)
	m.Put(ClosedOpen(2, 3), 8) // different value
	m.PutCoalescing(ClosedOpen(3, 4), 7)
	assertEntries(t, m,
		entry(ClosedOpen(1, 2), 7),
		entry(ClosedOpen(2, 3), 8),
		entry(ClosedOpen(3, 4), 7),
	)
}
