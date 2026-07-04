// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Package par is the parallel layer: bounded-worker fan-out over splittable
// sources. It is the eventual replacement for parallel/ (slice-only, []T-typed,
// no context, no panic containment); see 13-parallel-design.md.
//
// # Execution model — segments, not work-stealing
//
// The source knows how to cut itself into n roughly-equal, independently
// iterable pieces (a [Segmenter]); the executor is a boring bounded worker pool.
// No stealing, no recursive splitting, no task framework. A [View] is a reusable
// handle over such a source — zero goroutines until a terminal op runs.
//
// # Contracts every terminal honors
//
//   - context — ctx is per-operation (call-scoped), not per-view. Every terminal
//     takes ctx first and returns error so cancellation is reportable (ctx.Err()).
//     Infallible, uncancellable callers pass context.Background() and discard the
//     error. Cancellation stops scheduling and pulling; an in-flight non-ctx
//     callback runs to completion (it has no channel to hear cancellation on).
//   - panic containment — the first worker panic wins: siblings are cancelled and
//     the panic is re-raised on the caller's goroutine wrapped in [*PanicError]
//     (original value + stack preserved). Panics are re-raised, not returned as
//     error, because a panic in user code is a bug and converting bugs to errors
//     invites swallowing them.
//   - ordering — a stated property per op, not a mode. Map/Filter restore segment
//     order (segments carry their index; results concatenate). ForEach is
//     unordered. Reduce/Fold require an associative combiner and neutral identity
//     — a documented contract, not a checked one.
//
// # Parallelism sizing
//
// A terminal requests up to Workers segments (default runtime.GOMAXPROCS(0)); a
// source returns k ≤ n, and one goroutine runs per returned segment, so
// concurrency never exceeds Workers. When the source reports its size, small
// inputs fall back to a single segment (sequential) below MinPerWorker — the
// goroutine overhead isn't worth it. MinPerWorker's default is provisional
// pending the crossover benchmarks of 13-parallel-design.md §8.
//
// # Scope (first slice)
//
// Implemented: [FromSlice] and the generic [From] over any [Segmenter]; the
// terminals ForEach/Count/Filter/Reduce and the free functions [Map]/[Fold];
// panic containment and cancellation. Not yet built (later slices): the
// chunk-pump FromSeq/FromMap, short-circuiting Any/Find, the fallible …Err twins,
// GroupBy/TopK/Sum and other reducers, and the generated Par() collection
// on-ramp. Nothing here imports the collection families, so it stays additive.
package par
