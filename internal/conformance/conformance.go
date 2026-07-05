// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Package conformance holds the collection conformance laws (todo 14 §4) as
// generic, reusable assertions. Each law is expressed ONCE here and stamped
// across every code-generated family (from internal/codegen) instead of being
// hand-duplicated per family × primitive. The stamped tests import this package
// and call the exported Check* assertions with a concrete collection's methods.
//
// The laws are pure predicates first (check*, returning ok + a diff message) so
// they are unit-testable without a *testing.T; the exported wrappers add
// t.Helper()/t.Error. Element types must be comparable and free of NaN — the
// stamped fixtures never contain NaN, so map-keying and == comparison are exact.
package conformance

import (
	"fmt"
	"iter"
	"slices"
	"testing"
)

// AllMatchesToSlice asserts law 1: iterating All() yields exactly the elements
// of ToSlice(). When ordered is true the two must agree element-for-element (the
// family documents a stable iteration order — insertion, sorted, LIFO, …);
// otherwise they need only match as multisets (same elements with the same
// multiplicities, in any order — the unordered hash families).
func AllMatchesToSlice[T comparable](t *testing.T, all iter.Seq[T], toSlice []T, ordered bool) {
	t.Helper()
	if ok, msg := checkAllMatchesToSlice(all, toSlice, ordered); !ok {
		t.Error(msg)
	}
}

// checkAllMatchesToSlice is the pure law-1 check: it returns whether All() and
// ToSlice() agree (respecting the ordered flag) and, on failure, a diff message.
func checkAllMatchesToSlice[T comparable](all iter.Seq[T], toSlice []T, ordered bool) (bool, string) {
	got := slices.Collect(all)
	if ordered {
		if slices.Equal(got, toSlice) {
			return true, ""
		}
		return false, fmt.Sprintf("law 1 (All ≡ ToSlice, ordered): All()=%v, ToSlice()=%v", got, toSlice)
	}
	if sameMultiset(got, toSlice) {
		return true, ""
	}
	return false, fmt.Sprintf("law 1 (All ≡ ToSlice, unordered): All()=%v and ToSlice()=%v differ as multisets", got, toSlice)
}

// sameMultiset reports whether a and b contain the same elements with the same
// multiplicities, regardless of order.
func sameMultiset[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	counts := make(map[T]int, len(a))
	for _, x := range a {
		counts[x]++
	}
	for _, x := range b {
		counts[x]--
		if counts[x] < 0 {
			return false
		}
	}
	// Lengths are equal and no count went negative, so every count is exactly 0.
	return true
}
