// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package bloom

import (
	"slices"
	"testing"
)

// TestAddSeqEqualsRepeatedAdd pins the absorber law: AddSeq(seq) is exactly a
// loop of Add over the sequence. A filter filled via AddSeq answers MightContain
// identically to one filled element-by-element, over both the added set and a
// disjoint probe set (so equality isn't hidden by everything reading "present").
func TestAddSeqEqualsRepeatedAdd(t *testing.T) {
	items := []int32{5, 1, 5, 9, 2, 6, 100, -1}

	byAdd := NewBloomWithParams(256, 4)
	for _, v := range items {
		byAdd.Add(v)
	}
	bySeq := NewBloomWithParams(256, 4)
	bySeq.AddSeq(slices.Values(items))

	for probe := int32(-5); probe <= 200; probe++ {
		if byAdd.MightContain(probe) != bySeq.MightContain(probe) {
			t.Fatalf("MightContain(%d): Add-loop=%v, AddSeq=%v", probe, byAdd.MightContain(probe), bySeq.MightContain(probe))
		}
	}
	// Every added element must be present (no false negative).
	for _, v := range items {
		if !bySeq.MightContain(v) {
			t.Fatalf("AddSeq: added %d not present", v)
		}
	}
}

// TestAddSeqEmpty confirms AddSeq over an empty sequence is a no-op.
func TestAddSeqEmpty(t *testing.T) {
	b := NewBloomWithParams(256, 4)
	b.AddSeq(slices.Values([]int32{}))
	if b.MightContain(7) {
		t.Fatal("empty AddSeq should leave the filter empty")
	}
}
