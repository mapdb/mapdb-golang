// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package par

import (
	"context"
	"iter"
)

// FromSeq builds a chunk-pump View over a single-shot seq (§6): the on-ramp for
// sources that cannot be split — channels, IO, lazy pipelines. One puller
// goroutine drains the seq into MinPerWorker-sized chunks; a pool of Workers
// consumes them. Because the seq is consumed by pulling in a single pass,
// terminals over a FromSeq view differ from the segment engine (FromSlice/From):
//
//   - UNORDERED: results (Map/Filter/…) arrive in worker-completion order, not
//     source order; Find returns an arbitrary match, not the first.
//   - single pass per terminal: the source is ranged once each time a terminal
//     runs, so re-runnability follows the source (a seq over a slice re-runs; a
//     seq over a channel does not).
//   - Reduce/Fold need a COMMUTATIVE merge (not merely associative), since the
//     per-chunk partials combine in nondeterministic order.
//
// Cancellation is cooperative and bounded by the pull model: the puller can stop
// the source only at a yield boundary. A source that BLOCKS INSIDE ITS OWN BODY
// between yields — a bare `for x := range ch` on an idle channel — is not
// interrupted until it yields again, so neither cancellation nor short-circuit
// can tear it down. For a source that may block, use [FromSeqCtx] so its receive
// selects on the engine's context. Over a plain FromSeq that can block, short-
// circuit terminals (Any/All/None/Find) may not terminate promptly; reserve them
// for always-yielding or splittable sources, or bound the stream first.
//
// Non-short-circuiting reductions (Count/Sum/CountBy/…) consume the whole source,
// so an unbounded source must be bounded upstream (e.g. seq.Take) before the hop.
func FromSeq[T any](s iter.Seq[T], opts ...Option) View[T] {
	return View[T]{
		pump: s,
		size: -1,
		cfg:  newConfig(opts),
	}
}

// FromSeqCtx builds a chunk-pump View over a CONTEXT-AWARE single-shot source: gen
// is handed the engine's internal context each time a terminal runs and must
// return a seq whose blocking operations select on that context's Done channel
// (e.g. `func(ctx) iter.Seq[int] { return seq.FromChannelCtx(ctx, ch).Std() }`).
// This is the escape hatch for blocking sources (channels, IO): when the terminal
// is cancelled OR a short-circuiting terminal has its answer, the engine cancels
// that context, so the source's blocked receive unblocks and the operation tears
// down promptly — the guarantee [FromSeq] cannot make for a source that blocks
// between yields. All the other FromSeq semantics (unordered, single-pass,
// commutative Reduce/Fold) apply unchanged.
func FromSeqCtx[T any](gen func(ctx context.Context) iter.Seq[T], opts ...Option) View[T] {
	return View[T]{
		pumpCtx: gen,
		size:    -1,
		cfg:     newConfig(opts),
	}
}
