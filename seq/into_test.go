// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package seq

import (
	"slices"
	"testing"

	"github.com/mapdb/mapdb-golang/object"
)

// listSink is a minimal order-preserving Adder — proves Into works structurally
// with any Add(T) bool type, no object dependency required.
type listSink struct{ xs []int }

func (s *listSink) Add(v int) bool { s.xs = append(s.xs, v); return true }

func TestIntoStructuralSinkPreservesOrder(t *testing.T) {
	sink := Into(Range(0, 5), &listSink{})
	if !slices.Equal(sink.xs, []int{0, 1, 2, 3, 4}) {
		t.Fatalf("Into(list) = %v, want [0 1 2 3 4]", sink.xs)
	}
}

func TestIntoReturnsSameSinkInstance(t *testing.T) {
	// The returned value must be the very sink passed in (identity), not a copy —
	// so a caller can keep using the variable it already held.
	sink := &listSink{}
	got := Into(Of(3, 1, 2), sink)
	if got != sink {
		t.Fatal("Into returned a different sink instance, want the same pointer")
	}
	if len(got.xs) != 3 {
		t.Fatalf("returned sink len = %d, want 3", len(got.xs))
	}
}

func TestIntoEmptySeqLeavesSinkUntouched(t *testing.T) {
	sink := &listSink{xs: []int{99}} // pre-populated
	got := Into(Range(0, 0), sink)   // empty seq
	if got != sink || len(got.xs) != 1 || got.xs[0] != 99 {
		t.Fatalf("empty Into altered the sink: %v", got.xs)
	}
}

// The set case is the whole reason the Add bool is IGNORED: pouring duplicates
// into a set dedups (Add returns false for repeats) but Into must consume the
// ENTIRE sequence, not stop at the first false.
func TestIntoSetIgnoresInsertionBool(t *testing.T) {
	// 1,1,2,2,3 — five elements, three distinct. A stop-on-false loop would halt
	// after the second 1.
	set := Into(Of(1, 1, 2, 2, 3), object.NewHashSet[int]())
	if set.Len() != 3 {
		t.Fatalf("set.Len() = %d, want 3 (Into stopped early on a duplicate?)", set.Len())
	}
	for _, v := range []int{1, 2, 3} {
		if !set.Contains(v) {
			t.Fatalf("set missing %d", v)
		}
	}
}

// Into into a bag preserves multiplicity (bag Add always true).
func TestIntoBagKeepsMultiplicity(t *testing.T) {
	bag := Into(Of(7, 7, 7, 9), object.NewHashBag[int]())
	if bag.Len() != 4 {
		t.Fatalf("bag.Len() = %d, want 4", bag.Len())
	}
	if bag.OccurrencesOf(7) != 3 {
		t.Fatalf("OccurrencesOf(7) = %d, want 3", bag.OccurrencesOf(7))
	}
}
