// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Package collection defines the composable interface vocabulary implemented by
// the per-primitive collection packages (arraylist, hashset, treeset, …).
//
// Following Go's io.Reader / io.Writer / io.ReadWriter pattern, the API is built
// from small, single-concern interfaces (Sized, Iterable, Searchable,
// Convertible) per primitive element type, which the larger collection
// interfaces (Collection, MutableCollection, List, Set, Bag, Stack and their
// Mutable forms) compose. A consumer that only needs the element count depends
// on Int32Sized rather than a concrete list or set type.
//
// One set of interfaces exists per supported primitive (Int32*, Int64*,
// Float64*, …) because Go has no generic method specialization across primitives.
// The files are hand-maintained to mirror the shapes in object/interfaces.go.
//
// FAMILY_MATRIX.md in this directory is generated from the codegen manifest
// (internal/codegen/manifest.go) and lists every code-generated collection
// family with its storage, ordering, type coverage, and variants.
//
//go:generate go run ../internal/codegen matrix
package collection
