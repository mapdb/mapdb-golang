// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package hyperloglog

import (
	"slices"
	"testing"
)

// TestAddSeqEqualsRepeatedAdd pins the absorber law: AddSeq(seq) equals a loop of
// Add over the sequence — checked bit-for-bit on the register array (the
// cross-language oracle), so any divergence would surface.
func TestAddSeqEqualsRepeatedAdd(t *testing.T) {
	items := []int32{5, 1, 5, 9, 2, 6, 100, -1, 65536, 7}

	byAdd, err := NewHyperLogLogWithPrecision(10)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range items {
		byAdd.Add(v)
	}
	bySeq, err := NewHyperLogLogWithPrecision(10)
	if err != nil {
		t.Fatal(err)
	}
	bySeq.AddSeq(slices.Values(items))

	if !slices.Equal(byAdd.Registers(), bySeq.Registers()) {
		t.Fatal("AddSeq registers differ from repeated Add")
	}
}

// TestAddSeqEmpty confirms AddSeq over an empty sequence is a no-op.
func TestAddSeqEmpty(t *testing.T) {
	h, err := NewHyperLogLogWithPrecision(10)
	if err != nil {
		t.Fatal(err)
	}
	before := slices.Clone(h.Registers())
	h.AddSeq(slices.Values([]int32{}))
	if !slices.Equal(before, h.Registers()) {
		t.Fatal("empty AddSeq mutated the sketch")
	}
}
