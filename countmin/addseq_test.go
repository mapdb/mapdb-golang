// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package countmin

import (
	"slices"
	"testing"
)

// TestAddSeqEqualsRepeatedAddOne pins the absorber law: AddSeq(seq) equals a loop
// of AddOne over the sequence — so a repeated item accumulates its count. Checked
// via Estimate over the added set plus an absent probe.
func TestAddSeqEqualsRepeatedAddOne(t *testing.T) {
	items := []int32{5, 1, 5, 9, 2, 6, 5, 100, -1} // 5 appears three times

	byAdd := NewCountMinWithParams(4, 256)
	for _, v := range items {
		byAdd.AddOne(v)
	}
	bySeq := NewCountMinWithParams(4, 256)
	bySeq.AddSeq(slices.Values(items))

	for _, probe := range []int32{5, 1, 9, 2, 6, 100, -1, 42, 7} {
		if byAdd.Estimate(probe) != bySeq.Estimate(probe) {
			t.Fatalf("Estimate(%d): AddOne-loop=%d, AddSeq=%d", probe, byAdd.Estimate(probe), bySeq.Estimate(probe))
		}
	}
	// The thrice-added 5 must estimate at least 3.
	if bySeq.Estimate(5) < 3 {
		t.Fatalf("Estimate(5) = %d, want >= 3", bySeq.Estimate(5))
	}
}

// TestAddSeqEmpty confirms AddSeq over an empty sequence is a no-op.
func TestAddSeqEmpty(t *testing.T) {
	c := NewCountMinWithParams(4, 256)
	c.AddSeq(slices.Values([]int32{}))
	if c.Estimate(7) != 0 {
		t.Fatal("empty AddSeq should leave all counters zero")
	}
}
