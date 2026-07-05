// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package roaring_test

// Integration test: the whole point of giving roaring Segments(n) is that it now
// satisfies par.Segmenter[uint32], so par.From runs a parallel fan-out over a
// compressed bitmap with NO bespoke conversion — the shared iteration protocol
// paying off across package boundaries (doc 14 §2/§3).

import (
	"context"
	"testing"

	"github.com/mapdb/mapdb-golang/par"
	"github.com/mapdb/mapdb-golang/roaring"
)

func TestParFromRoaringMatchesSequential(t *testing.T) {
	s := roaring.NewRoaringU32()
	// Several thousand values scattered across many chunks and both container
	// kinds, so par actually splits into multiple worker segments.
	for i := uint32(0); i < 5000; i++ {
		s.Add(i*7 + (i%4)<<16)
	}

	// Sequential ground truth over the deduped set.
	var wantEven, wantTotal int
	for v := range s.All() {
		wantTotal++
		if v%2 == 0 {
			wantEven++
		}
	}
	if wantTotal < 100 {
		t.Fatalf("fixture too small (%d) — split would be trivial", wantTotal)
	}

	// par.From(s) COMPILES ONLY because *RoaringU32 satisfies par.Segmenter[uint32]
	// (its Segments(n) method) — the de-islanding payoff.
	view := par.From[uint32](s)
	ctx := context.Background()

	// A parallel Count with a false-everywhere-except-even predicate, and a
	// total count (pred always true) — both deterministic regardless of how the
	// bitmap is partitioned across workers.
	gotEven, err := view.Count(ctx, func(v uint32) bool { return v%2 == 0 })
	if err != nil {
		t.Fatal(err)
	}
	gotTotal, err := view.Count(ctx, func(uint32) bool { return true })
	if err != nil {
		t.Fatal(err)
	}
	if gotEven != wantEven || gotTotal != wantTotal {
		t.Fatalf("par Count (even=%d,total=%d) != sequential (even=%d,total=%d)",
			gotEven, gotTotal, wantEven, wantTotal)
	}
}
