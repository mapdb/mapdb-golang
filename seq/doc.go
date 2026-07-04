// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Package seq is the lazy layer: chainable sequences over the standard library's
// iter.Seq iterators. Seq[T] is a defined function type
// (type Seq[T] iter.Seq[T]), so it ranges directly — for v := range s works —
// and carries a fluent method set, while type-changing or constrained operations
// (Map, Distinct, Sum, …) are free functions that return Seq so a pipeline keeps
// flowing.
//
// # Interop
//
// Seq[T] and iter.Seq[T] are distinct defined types with no implicit conversion.
// The two worlds meet through zero-cost adapters: seq.From(s) adopts a stdlib
// iterator, s.Std() releases one. Collection methods (All, FromSeq) speak stdlib
// iter.Seq so collections need no dependency on this package; seq free functions
// and methods all speak Seq so chains need no conversions.
//
// # The laziness contract
//
// Every operation documents three things:
//
//   - lazy or eager — lazy ops do no work until the returned Seq is ranged; eager
//     ops (Partition, and terminals like ToSlice/Sum) consume their input once
//     when called.
//   - re-runnable or single-shot — a Seq derived from a collection or an
//     in-memory source may be ranged repeatedly; a Seq over a one-shot source
//     (a channel, a Pull) is single-shot. Eager ops yield re-runnable results.
//   - allocation class — O(1) for streaming stages (Filter, Map, Take, …),
//     O(distinct) for Distinct (which is still lazy and short-circuiting — it
//     just retains a set of the values seen), O(n) for the materializing ops
//     (Partition, ToSlice).
//
// Infinite sources (Iterate, Generate, Repeat) are first-class: compose them with
// Take / TakeWhile / First, which short-circuit. Because the streaming stages are
// O(1) memory and stop pulling as soon as the consumer stops, a filtered-mapped
// scan of a billion-element source holds a constant footprint.
//
// # Ordering vocabulary
//
// Terms reused module-wide: encounter-ordered (lists, deques), sorted (tree
// views), heap-ordered (priorityqueue — not sorted), unordered (hash structures,
// order unspecified and may vary per run). A Seq preserves the ordering class of
// its source; any op that imposes or destroys an order documents it.
//
// This package supersedes stream/ (which remains as thin aliases for one release).
package seq
