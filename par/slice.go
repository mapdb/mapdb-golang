// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package par

import (
	"iter"

	"github.com/mapdb/mapdb-golang/internal/segment"
)

// FromSlice builds a View over xs. The slice is split into balanced contiguous
// index ranges (O(1) per segment); the slice is not copied, so mutating it while
// a terminal runs is undefined behavior — the same live-view rule as any
// Segmenter. The slice's length is known, so small inputs fall back to
// sequential execution below MinPerWorker.
//
// It splits via the shared internal/segment.Split, so FromSlice(xs) and
// From(listOver(xs)) — whose generated Segments calls the same helper — cut a
// source into the same ranges.
func FromSlice[T any](xs []T, opts ...Option) View[T] {
	return View[T]{
		segment: func(n int) []iter.Seq[T] { return segment.Split(xs, n) },
		size:    len(xs),
		cfg:     newConfig(opts),
	}
}
