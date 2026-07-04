// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package hyperloglog

import (
	"bytes"
	"testing"
)

// FuzzHyperLogLogFromBytes drives the hand-written wire parser with arbitrary
// input. The contract under test:
//   - HyperLogLogFromBytes must never panic on any []byte (only return an error).
//   - Whatever it accepts must round-trip: ToBytes reproduces the exact accepted
//     bytes, the precision is in range, the register array length is 2^p, and
//     every register is within the per-p rho ceiling (the states the parser
//     carefully rejects must never slip through and then re-serialize differently).
func FuzzHyperLogLogFromBytes(f *testing.F) {
	// Seed with valid serialized sketches across a couple of precisions.
	for _, p := range []uint8{4, 6, 10} {
		h, err := NewHyperLogLogWithPrecision(p)
		if err != nil {
			f.Fatal(err)
		}
		for i := int32(0); i < 100; i++ {
			h.Add(i)
		}
		f.Add(h.ToBytes())
	}
	f.Add([]byte{}) // too short
	f.Add([]byte("HLL1\x06"))

	f.Fuzz(func(t *testing.T, data []byte) {
		h, err := HyperLogLogFromBytes(data)
		if err != nil {
			return // rejection is fine; it must just not panic
		}
		if !bytes.Equal(h.ToBytes(), data) {
			t.Fatalf("accepted bytes do not round-trip through ToBytes (len %d)", len(data))
		}
		p := h.Precision()
		if p < MinPrecision || p > MaxPrecision {
			t.Fatalf("accepted out-of-range precision %d", p)
		}
		if got, want := h.RegisterCount(), 1<<p; got != want {
			t.Fatalf("register count %d != 2^p (%d)", got, want)
		}
		ceiling := rhoCeiling(p)
		for i, r := range h.Registers() {
			if r > ceiling {
				t.Fatalf("register[%d]=%d exceeds ceiling %d", i, r, ceiling)
			}
		}
		// Estimate must stay finite for any accepted state (spec mandate).
		if e := h.Estimate(); e != e { // NaN check without math import
			t.Fatal("Estimate returned NaN for an accepted state")
		}
	})
}
