// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package par

import "iter"

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
// Short-circuiting terminals (Any/All/None) stop the puller once they have the
// answer, so they terminate even over unbounded sources. Non-short-circuiting
// reductions (Count/Sum/CountBy/…) consume the whole source, so an unbounded
// source must be bounded upstream (e.g. seq.Take) before the parallel hop.
func FromSeq[T any](s iter.Seq[T], opts ...Option) View[T] {
	return View[T]{
		pump: s,
		size: -1,
		cfg:  newConfig(opts),
	}
}
