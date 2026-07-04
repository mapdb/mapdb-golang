// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package roaring

import (
	"bytes"
	"testing"
)

// FuzzDeserialize drives the hand-written byte parser with arbitrary input. The
// contract under test:
//   - Deserialize must never panic on any []byte (only return an error).
//   - Whatever it accepts must be in canonical form: re-serializing the result
//     reproduces the exact accepted bytes (the Deserialize canonicality check is
//     what makes serialized bytes a reliable equality oracle), and ToSortedSlice
//     is strictly ascending with no duplicates.
func FuzzDeserialize(f *testing.F) {
	// Seed with a few real serialized forms so the fuzzer starts from valid
	// structure and mutates outward.
	seeds := [][]uint32{
		{},
		{0},
		{1, 2, 3},
		{0, 65535, 65536, 131071, 4294967295},
	}
	for _, s := range seeds {
		r := NewRoaringU32()
		for _, v := range s {
			r.Add(v)
		}
		f.Add(r.Serialize())
	}
	// A dense run to force a bitmap container.
	dense := NewRoaringU32()
	for v := uint32(0); v <= 5000; v++ {
		dense.Add(v)
	}
	f.Add(dense.Serialize())

	f.Fuzz(func(t *testing.T, data []byte) {
		r, err := Deserialize(data)
		if err != nil {
			return // rejection is fine; it must just not panic
		}
		// Accepted => canonical round-trip.
		reser := r.Serialize()
		if !bytes.Equal(reser, data) {
			t.Fatalf("accepted non-canonical bytes: len(in)=%d len(reser)=%d", len(data), len(reser))
		}
		// Elements strictly ascending, no duplicates.
		prev := int64(-1)
		for _, v := range r.ToSortedSlice() {
			if int64(v) <= prev {
				t.Fatalf("ToSortedSlice not strictly ascending at %d (prev %d)", v, prev)
			}
			prev = int64(v)
		}
	})
}
