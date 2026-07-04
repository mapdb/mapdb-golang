// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Package bitset provides a growable set of non-negative bit indices backed by a
// packed []uint64 word array — the Go analogue of java.util.BitSet.
//
// BitSet stores membership as one bit per index, so a dense set of small indices
// is far more compact than a map[int]struct{}. It supports the usual per-bit
// operations (Set, Clear, Flip, Get), in-place bulk set algebra (AndInPlace,
// AndNotInPlace, OrInPlace, XorInPlace), Intersects/Equals, Cardinality
// (popcount), and iteration over set bits via NextSetBit or ToSlice.
//
// Indices are non-negative; the backing word array grows on demand as higher
// indices are set. A SynchronizedBitSet wrapper exposes the same surface under an
// internal lock for concurrent use.
package bitset
