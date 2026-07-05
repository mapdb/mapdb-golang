// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package object_test

// Hand-written conformance stamp (todo 14 §4). The generic object collections
// are not part of the per-primitive codegen; object.ArrayList[T] is the object
// family that exposes Segments, so it is checked here against the same
// internal/conformance predicates the generated families use. Instantiated at
// both int and string so the law is exercised across a numeric and a non-numeric
// element type. An ArrayList preserves insertion order, so law 1 is ordered.

import (
	"testing"

	"github.com/mapdb/mapdb-golang/internal/conformance"
	"github.com/mapdb/mapdb-golang/object"
)

func TestConformanceAllMatchesToSliceArrayListInt(t *testing.T) {
	l := object.NewArrayListFrom(3, 1, 4, 1, 5, 9, 2)
	conformance.AllMatchesToSlice(t, l.All(), l.ToSlice(), true)
}

func TestConformanceSegmentsArrayListInt(t *testing.T) {
	l := object.NewArrayListFrom(3, 1, 4, 1, 5, 9, 2)
	conformance.SegmentsCoverAll(t, l.All(), l.Segments)
}

func TestConformanceAllMatchesToSliceArrayListString(t *testing.T) {
	l := object.NewArrayListFrom("c", "a", "d", "a", "e", "i", "b")
	conformance.AllMatchesToSlice(t, l.All(), l.ToSlice(), true)
}

func TestConformanceSegmentsArrayListString(t *testing.T) {
	l := object.NewArrayListFrom("c", "a", "d", "a", "e", "i", "b")
	conformance.SegmentsCoverAll(t, l.All(), l.Segments)
}
