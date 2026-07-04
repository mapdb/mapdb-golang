// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package par

import (
	"context"
	"fmt"
	"runtime"
	"testing"
)

// These benchmarks are the 13-parallel-design.md §8 deliverable: they let the
// MinPerWorker default be justified by a measured crossover and produce the
// worker-scaling curve. Run with:
//
//	go test -run=x -bench=. ./par/...
//
// Findings (32-core linux/amd64 host, trivial callbacks — reproduce on your own
// hardware; crossover shifts with callback cost):
//
//   - CountCrossover: sequential (1 worker) wins up to n≈1024; break-even at
//     n≈4096; parallel (32 workers) wins from n≈16384 (2.4×) up to ~5× at 1M.
//     ⇒ ~750–1024 elements/worker is the trivial-callback crossover, which is why
//     defaultMinPerWorker = 1024. An expensive callback moves this far lower, so
//     MinPerWorker is a per-op knob.
//   - SumScaling (10M int64): 1→2→4→8→32 workers ≈ 1.0×→1.25×→2.0×→2.5×→4.9×.
//     Sublinear because summing int64s is memory-bandwidth-bound, not compute-
//     bound; compute-heavy callbacks scale closer to linear.
//   - PumpVsSegment (1M ints): the chunk-pump is ~8× slower than the segment
//     engine on a slice (channel + copy + coordination). Prefer FromSlice/From
//     when the source is splittable; FromSeq is the fallback for single-shot
//     sources, not a slice path.
//
// CountCrossover sweeps element count × {sequential, parallel} for a trivial
// callback: the size where parallel (GOMAXPROCS workers) first beats sequential
// (1 worker) is the crossover that justifies MinPerWorker.
func BenchmarkCountCrossover(b *testing.B) {
	ctx := context.Background()
	for _, n := range []int{64, 256, 1024, 4096, 16_384, 65_536, 262_144, 1_048_576} {
		xs := iotaSlice(n)
		for _, w := range []int{1, runtime.GOMAXPROCS(0)} {
			// MinPerWorker(1) disables the fallback so we measure the engine itself.
			v := FromSlice(xs, Workers(w), MinPerWorker(1))
			b.Run(fmt.Sprintf("n=%d/workers=%d", n, w), func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					if _, err := v.Count(ctx, func(x int) bool { return x&1 == 0 }); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

// SumScaling is the worker-scaling curve: Sum over 10M int64 across 1..GOMAXPROCS.
func BenchmarkSumScaling(b *testing.B) {
	ctx := context.Background()
	xs := make([]int64, 10_000_000)
	for i := range xs {
		xs[i] = int64(i)
	}
	maxW := runtime.GOMAXPROCS(0)
	for _, w := range dedupWorkers(1, 2, 4, 8, maxW) {
		v := FromSlice(xs, Workers(w), MinPerWorker(1))
		b.Run(fmt.Sprintf("workers=%d", w), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := Sum(ctx, v); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// PumpVsSegment contrasts the two execution models on the same 1M ints with a
// trivial callback: the chunk-pump pays channel + copy overhead the segment
// engine does not.
func BenchmarkPumpVsSegment(b *testing.B) {
	ctx := context.Background()
	xs := iotaSlice(1_048_576)
	seg := FromSlice(xs, MinPerWorker(1))
	pump := FromSeq(seqOfSlice(xs), MinPerWorker(4096))
	b.Run("segment", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := seg.Count(ctx, func(x int) bool { return x&1 == 0 }); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("chunk-pump", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := pump.Count(ctx, func(x int) bool { return x&1 == 0 }); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func dedupWorkers(ws ...int) []int {
	seen := map[int]bool{}
	var out []int
	for _, w := range ws {
		if w >= 1 && !seen[w] {
			seen[w] = true
			out = append(out, w)
		}
	}
	return out
}
