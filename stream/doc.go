// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Package stream provides lazy, composable sequence operations over the standard
// library's iter.Seq[V] / iter.Seq2[K, V] iterators.
//
// Go's type system does not allow generic methods on generic types, so
// operations that change the element type (Map, FlatMap, GroupBy, Partition,
// Enumerate, Chunk, …) cannot live as methods on the collection types. They live
// here as free functions taking and returning iter.Seq values, so pipelines
// compose by nesting:
//
//	total := stream.Reduce(
//		stream.Map(
//			stream.Filter(src, even),
//			square,
//		),
//		0, add,
//	)
//
// # Laziness
//
// Transform and filter functions (Map, Filter, Take, Drop, Chunk, Chain,
// Distinct, …) are lazy: they return a new iter.Seq and pull from the source
// only as the result is iterated. Terminal functions (ToSlice, Reduce, Count,
// ForEach, First, Any/All/None, Min/Max/Sum, GroupByToMap, …) drive iteration to
// completion (or until an early stop).
//
// # Edge-case contracts
//
//   - Take(n) with n <= 0 yields nothing and pulls nothing from the source; a
//     positive n pulls exactly n elements (never n+1).
//   - Chunk(n) requires n > 0 and panics otherwise (a non-positive chunk size is
//     a programmer error).
//   - Early termination propagates: when a consumer stops (a yield returns
//     false), the upstream sources stop too.
package stream
