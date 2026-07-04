// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package stack_test

import (
	"context"
	"slices"
	"testing"

	"github.com/mapdb/mapdb-golang/par"
	"github.com/mapdb/mapdb-golang/stack"
)

// Every stack variant is a par.Segmenter[int32] via the generated Segments (the
// load-bearing par.From on-ramp).
var (
	_ par.Segmenter[int32] = (*stack.Int32)(nil)
	_ par.Segmenter[int32] = (*stack.ImmutableInt32)(nil)
	_ par.Segmenter[int32] = (*stack.SynchronizedInt32)(nil)
)

func segmentsMultiset(seg par.Segmenter[int32], n int) []int32 {
	var got []int32
	for _, s := range seg.Segments(n) {
		for v := range s {
			got = append(got, v)
		}
	}
	slices.Sort(got)
	return got
}

// TestSegmentsCoverAllInt32 pins the Segments law for stacks. Note a stack's
// All yields top-to-bottom while Segments follows the backing array — the two
// differ in order but not as a multiset, which is exactly what the (unordered)
// Segmenter contract promises.
func TestSegmentsCoverAllInt32(t *testing.T) {
	for _, size := range []int{0, 1, 7, 8, 100} {
		want := make([]int32, size)
		for i := range want {
			want[i] = int32(i)
		}

		base := stack.Int32Of(want...)
		imm := stack.NewImmutableInt32(want...)
		sync := stack.NewSynchronizedInt32From(stack.Int32Of(want...))

		variants := map[string]par.Segmenter[int32]{"base": base, "immutable": imm, "synchronized": sync}
		for name, seg := range variants {
			for _, n := range []int{1, 2, 7, size + 1} {
				got := segmentsMultiset(seg, n)
				if len(got) == 0 && len(want) == 0 {
					continue
				}
				if !slices.Equal(got, want) {
					t.Fatalf("%s size=%d n=%d: multiset = %v, want %v", name, size, n, got, want)
				}
			}
		}
	}
}

// TestSegmentsFeedParFromInt32 proves a stack plugs into par.From for a real
// parallel reduction.
func TestSegmentsFeedParFromInt32(t *testing.T) {
	s := stack.NewInt32()
	for i := int32(0); i < 10_000; i++ {
		s.Push(i)
	}
	got, err := par.From[int32](s, par.Workers(8)).Count(context.Background(), func(v int32) bool {
		return v%2 == 0
	})
	if err != nil {
		t.Fatalf("par.From(stack).Count: %v", err)
	}
	if got != 5000 {
		t.Fatalf("parallel even-count = %d, want 5000", got)
	}
}
