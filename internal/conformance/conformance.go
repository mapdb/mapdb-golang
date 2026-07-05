// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Package conformance holds the collection conformance laws (todo 14 §4) as
// generic, reusable assertions. Each law is expressed ONCE here and stamped
// across every code-generated family (from internal/codegen) instead of being
// hand-duplicated per family × primitive. The stamped tests import this package
// and call the exported law assertions (AllMatchesToSlice, …) with a concrete
// collection's methods.
//
// The laws are pure predicates first (check*, returning ok + a diff message) so
// they are unit-testable without a *testing.T; the exported wrappers add
// t.Helper()/t.Error. Element types must be comparable and free of NaN — the
// stamped fixtures never contain NaN, so map-keying and == comparison are exact.
package conformance

import (
	"cmp"
	"fmt"
	"iter"
	"maps"
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

// Len2MatchesAll asserts the size-accounting law for a key/value collection:
// Len() equals the number of pairs actually yielded by All(). A stale or
// miscounted size counter (e.g. tombstones counted toward Len) fails here.
func Len2MatchesAll[K, V any](t *testing.T, length int, all iter.Seq2[K, V]) {
	t.Helper()
	if ok, msg := checkLen2MatchesAll(length, all); !ok {
		t.Error(msg)
	}
}

// checkLen2MatchesAll is the pure size-accounting check: Len() vs |All()|.
func checkLen2MatchesAll[K, V any](length int, all iter.Seq2[K, V]) (bool, string) {
	n := 0
	for range all {
		n++
	}
	if n == length {
		return true, ""
	}
	return false, fmt.Sprintf("size law: Len()=%d but All() yielded %d pairs", length, n)
}

// KeysAscending asserts the sorted-map ordering law: All() yields keys in
// strictly ascending order. Applies to the ordered map families (treemap). Keys
// are compared with < so K must be cmp.Ordered and NaN-free (the fixtures are).
func KeysAscending[K cmp.Ordered, V any](t *testing.T, all iter.Seq2[K, V]) {
	t.Helper()
	if ok, msg := checkKeysAscending(all); !ok {
		t.Error(msg)
	}
}

// checkKeysAscending is the pure ordering check: every key strictly greater than
// its predecessor.
func checkKeysAscending[K cmp.Ordered, V any](all iter.Seq2[K, V]) (bool, string) {
	first := true
	var prev K
	for k := range all {
		if !first && !(prev < k) {
			return false, fmt.Sprintf("ordering law: All() keys not strictly ascending: %v then %v", prev, k)
		}
		prev, first = k, false
	}
	return true, ""
}

// SegmentsCoverAll asserts the Segments partition law (todo 14 §4) for a
// structural family. For each split count n in {1, 2, 7, len+1} the segments
// returned by Segments(n) must (a) concatenate to the same multiset as All() —
// a complete, non-overlapping partition, since an overlap would double-count an
// element and a gap would drop one — and (b) each be re-runnable, yielding the
// same multiset on a second iteration. The re-run check is load-bearing: the par
// segment engine iterates a family's segments more than once, so a single-shot
// segment is a real bug, not a stylistic one. Element type is comparable and
// NaN-free (the stamped fixtures never hold NaN), so the multiset comparison is
// exact.
func SegmentsCoverAll[T comparable](t *testing.T, all iter.Seq[T], segments func(n int) []iter.Seq[T]) {
	t.Helper()
	if ok, msg := checkSegmentsCoverAll(all, segments); !ok {
		t.Error(msg)
	}
}

// checkSegmentsCoverAll is the pure Segments-partition predicate: coverage (as a
// multiset) plus per-segment re-runnability, across the four split counts.
func checkSegmentsCoverAll[T comparable](all iter.Seq[T], segments func(n int) []iter.Seq[T]) (bool, string) {
	want := slices.Collect(all)
	for _, n := range []int{1, 2, 7, len(want) + 1} {
		var concat []T
		for i, seg := range segments(n) {
			first := slices.Collect(seg)
			second := slices.Collect(seg) // re-run the same segment
			if !sameMultiset(first, second) {
				return false, fmt.Sprintf("segments law (n=%d): segment %d not re-runnable: %v then %v", n, i, first, second)
			}
			concat = append(concat, first...)
		}
		if !sameMultiset(concat, want) {
			return false, fmt.Sprintf("segments law (n=%d): concat(Segments)=%v does not match All()=%v as a multiset", n, concat, want)
		}
	}
	return true, ""
}

// Segments2CoverAll asserts the Segments2 partition law (todo 14 §4) for a
// key/value family. For each split count n in {1, 2, 7, len+1} the segments from
// Segments2(n) must (a) reproduce exactly the key→value map that All() yields —
// complete coverage with no key split across two segments — and (b) each be
// re-runnable. Map keys are unique, so the law is checked as map equality (which
// also detects a duplicated key: it would be merged from two segments, tripping
// the overlap guard). Keys and values are comparable and NaN-free.
func Segments2CoverAll[K comparable, V comparable](t *testing.T, all iter.Seq2[K, V], segments func(n int) []iter.Seq2[K, V]) {
	t.Helper()
	if ok, msg := checkSegments2CoverAll(all, segments); !ok {
		t.Error(msg)
	}
}

// checkSegments2CoverAll is the pure Segments2-partition predicate: map-equality
// coverage, an explicit no-duplicate-key guard, and per-segment re-runnability,
// across the four split counts. The coverage merge iterates each segment's raw
// pairs (not a collapsed map) so a key emitted twice — whether within one
// segment or across two — trips the guard: a map merge alone would hide a
// same-value duplicate.
func checkSegments2CoverAll[K comparable, V comparable](all iter.Seq2[K, V], segments func(n int) []iter.Seq2[K, V]) (bool, string) {
	want := collectMap(all)
	for _, n := range []int{1, 2, 7, len(want) + 1} {
		got := make(map[K]V, len(want))
		for i, seg := range segments(n) {
			if !maps.Equal(collectMap(seg), collectMap(seg)) { // re-run the same segment
				return false, fmt.Sprintf("segments2 law (n=%d): segment %d not re-runnable", n, i)
			}
			for k, v := range seg {
				if _, dup := got[k]; dup {
					return false, fmt.Sprintf("segments2 law (n=%d): key %v appears more than once (within a segment or across segments)", n, k)
				}
				got[k] = v
			}
		}
		if !maps.Equal(got, want) {
			return false, fmt.Sprintf("segments2 law (n=%d): concat(Segments2)=%v does not reproduce All()=%v", n, got, want)
		}
	}
	return true, ""
}

// collectMap drains an iter.Seq2 into a map (map keys are unique in the families
// this checks, so no pair is lost to key collision).
func collectMap[K comparable, V any](s iter.Seq2[K, V]) map[K]V {
	m := map[K]V{}
	for k, v := range s {
		m[k] = v
	}
	return m
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
