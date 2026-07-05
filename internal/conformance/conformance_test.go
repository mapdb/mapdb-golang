// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package conformance

import (
	"iter"
	"slices"
	"testing"
)

// seqOf turns a slice into an iter.Seq for exercising the pure law check.
func seqOf[T any](xs []T) func(yield func(T) bool) {
	return func(yield func(T) bool) {
		for _, x := range xs {
			if !yield(x) {
				return
			}
		}
	}
}

// TestCheckAllMatchesToSlice pins the pure law-1 predicate: it must PASS on
// agreeing inputs and FAIL on every way they can disagree — the whole point of a
// conformance law is that it rejects a broken family, so a check that never
// fails is worthless. It is verified in both the ordered and unordered modes.
func TestCheckAllMatchesToSlice(t *testing.T) {
	cases := []struct {
		name    string
		all     []int
		toSlice []int
		ordered bool
		wantOK  bool
	}{
		// ordered mode: identical sequences pass; any order/content change fails.
		{"ordered identical", []int{3, 1, 4, 1, 5}, []int{3, 1, 4, 1, 5}, true, true},
		{"ordered empty", nil, nil, true, true},
		{"ordered reordered fails", []int{1, 2, 3}, []int{3, 2, 1}, true, false},
		{"ordered missing element fails", []int{1, 2, 3}, []int{1, 2}, true, false},
		{"ordered extra element fails", []int{1, 2}, []int{1, 2, 3}, true, false},
		{"ordered wrong multiplicity fails", []int{1, 1, 2}, []int{1, 2, 2}, true, false},
		// unordered mode: any permutation passes; multiset differences fail.
		{"unordered permutation passes", []int{3, 1, 4, 1, 5}, []int{1, 1, 3, 4, 5}, false, true},
		{"unordered empty", nil, nil, false, true},
		{"unordered wrong multiplicity fails", []int{1, 1, 2}, []int{1, 2, 2}, false, false},
		{"unordered missing element fails", []int{1, 2, 3}, []int{1, 2, 4}, false, false},
		{"unordered length mismatch fails", []int{1, 2}, []int{1, 2, 2}, false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ok, msg := checkAllMatchesToSlice(seqOf(c.all), c.toSlice, c.ordered)
			if ok != c.wantOK {
				t.Fatalf("checkAllMatchesToSlice(%v, %v, ordered=%v) ok=%v (%q), want ok=%v",
					c.all, c.toSlice, c.ordered, ok, msg, c.wantOK)
			}
			if !ok && msg == "" {
				t.Fatalf("a failing check must carry a non-empty diff message")
			}
		})
	}
}

// TestSameMultiset checks the multiset helper directly, including the negative
// paths (a bag that ignored multiplicity or order would slip through law 1).
func TestSameMultiset(t *testing.T) {
	cases := []struct {
		a, b []int
		want bool
	}{
		{nil, nil, true},
		{[]int{1, 2, 3}, []int{3, 2, 1}, true},
		{[]int{1, 1, 2}, []int{1, 2, 1}, true},
		{[]int{1, 1, 2}, []int{1, 2, 2}, false},
		{[]int{1, 2, 3}, []int{1, 2}, false},
		{[]int{1, 2}, []int{1, 2, 3}, false},
		{[]int{1, 2, 3}, []int{1, 2, 4}, false},
	}
	for _, c := range cases {
		if got := sameMultiset(c.a, c.b); got != c.want {
			t.Errorf("sameMultiset(%v, %v) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

// seq2Of turns parallel key/value slices into an iter.Seq2 for the map laws.
func seq2Of[K, V any](ks []K, vs []V) func(yield func(K, V) bool) {
	return func(yield func(K, V) bool) {
		for i := range ks {
			if !yield(ks[i], vs[i]) {
				return
			}
		}
	}
}

// TestCheckLen2MatchesAll pins the size-accounting predicate: it passes only when
// the claimed length equals the number of pairs All() yields.
func TestCheckLen2MatchesAll(t *testing.T) {
	ks, vs := []int{3, 1, 4}, []int{30, 10, 40}
	cases := []struct {
		length int
		wantOK bool
	}{
		{3, true},  // correct count
		{2, false}, // undercount (a stale/low size counter)
		{4, false}, // overcount (e.g. tombstones counted toward Len)
		{0, false}, // zero against a non-empty All()
	}
	for _, c := range cases {
		ok, msg := checkLen2MatchesAll(c.length, seq2Of(ks, vs))
		if ok != c.wantOK {
			t.Errorf("checkLen2MatchesAll(length=%d) ok=%v (%q), want %v", c.length, ok, msg, c.wantOK)
		}
		if !ok && msg == "" {
			t.Errorf("length=%d: failing check must carry a message", c.length)
		}
	}
	// Empty map: Len 0 matches an empty All().
	if ok, _ := checkLen2MatchesAll(0, seq2Of([]int{}, []int{})); !ok {
		t.Errorf("empty map: Len()=0 should match empty All()")
	}
}

// TestCheckKeysAscending pins the ordering predicate: strictly ascending keys
// pass; any equal or descending adjacent pair fails.
func TestCheckKeysAscending(t *testing.T) {
	cases := []struct {
		keys   []int
		wantOK bool
	}{
		{[]int{1, 2, 3, 9}, true},
		{nil, true},                // empty is vacuously ascending
		{[]int{5}, true},           // singleton
		{[]int{1, 3, 2}, false},    // descending pair
		{[]int{1, 2, 2, 3}, false}, // duplicate (not STRICTLY ascending)
		{[]int{3, 2, 1}, false},
	}
	for _, c := range cases {
		vs := make([]int, len(c.keys))
		ok, msg := checkKeysAscending(seq2Of(c.keys, vs))
		if ok != c.wantOK {
			t.Errorf("checkKeysAscending(%v) ok=%v (%q), want %v", c.keys, ok, msg, c.wantOK)
		}
	}
}

// TestAllMatchesToSliceThroughT drives the exported *testing.T wrapper on a
// passing case to confirm it does not spuriously fail (the failing path is a
// t.Error, which cannot be asserted here without a testing.TB mock — the pure
// predicate above owns the negative coverage).
func TestAllMatchesToSliceThroughT(t *testing.T) {
	vals := []int{5, 3, 5, 1}
	AllMatchesToSlice(t, seqOf(vals), slices.Clone(vals), true)
	AllMatchesToSlice(t, seqOf(vals), []int{1, 3, 5, 5}, false)
}

// segmentsOf is the well-behaved reference partition: it splits xs into n
// contiguous, stateless (re-runnable) chunks — the shape a correct Segments(n)
// must produce. n is clamped into [1, len(xs)] the way real implementations do.
func segmentsOf[T any](xs []T) func(n int) []iter.Seq[T] {
	return func(n int) []iter.Seq[T] {
		if n < 1 {
			n = 1
		}
		if n > len(xs) {
			n = len(xs)
		}
		if n == 0 {
			return nil
		}
		size := (len(xs) + n - 1) / n
		var segs []iter.Seq[T]
		for i := 0; i < len(xs); i += size {
			end := min(i+size, len(xs))
			segs = append(segs, seqOf(xs[i:end]))
		}
		return segs
	}
}

// TestCheckSegmentsCoverAll pins the pure Segments-partition predicate: it PASSES
// on a correct partition (and on the empty collection) and FAILS on each way a
// partition can be wrong — a dropped element (gap), a double-counted element
// (overlap), and a single-shot segment (not re-runnable). A law that never
// rejects a broken family is worthless, so the negative paths are the point.
func TestCheckSegmentsCoverAll(t *testing.T) {
	xs := []int{3, 1, 4, 1, 5, 9, 2} // dup 1 ⇒ multiset-sensitive coverage

	if ok, msg := checkSegmentsCoverAll(seqOf(xs), segmentsOf(xs)); !ok {
		t.Errorf("correct partition should pass, got %q", msg)
	}
	if ok, msg := checkSegmentsCoverAll(seqOf([]int(nil)), segmentsOf([]int(nil))); !ok {
		t.Errorf("empty collection should pass vacuously, got %q", msg)
	}

	// gap: a partition missing the last element under-covers.
	gap := func(int) []iter.Seq[int] { return []iter.Seq[int]{seqOf(xs[:len(xs)-1])} }
	if ok, _ := checkSegmentsCoverAll(seqOf(xs), gap); ok {
		t.Error("partition dropping an element should fail coverage")
	}

	// overlap: an element yielded by two segments over-counts.
	overlap := func(int) []iter.Seq[int] { return []iter.Seq[int]{seqOf(xs), seqOf(xs[:1])} }
	if ok, _ := checkSegmentsCoverAll(seqOf(xs), overlap); ok {
		t.Error("partition double-counting an element should fail coverage")
	}

	// non-re-runnable: a single-shot segment yields nothing on its second pass.
	singleShot := func(int) []iter.Seq[int] {
		used := false
		seg := func(yield func(int) bool) {
			if used {
				return
			}
			used = true
			for _, x := range xs {
				if !yield(x) {
					return
				}
			}
		}
		return []iter.Seq[int]{seg}
	}
	if ok, _ := checkSegmentsCoverAll(seqOf(xs), singleShot); ok {
		t.Error("single-shot (non-re-runnable) segment should fail")
	}
}

// segments2Of is the well-behaved reference pair-partition: contiguous,
// re-runnable chunks of parallel key/value slices.
func segments2Of[K, V any](keys []K, vals []V) func(n int) []iter.Seq2[K, V] {
	return func(n int) []iter.Seq2[K, V] {
		if n < 1 {
			n = 1
		}
		if n > len(keys) {
			n = len(keys)
		}
		if n == 0 {
			return nil
		}
		size := (len(keys) + n - 1) / n
		var segs []iter.Seq2[K, V]
		for i := 0; i < len(keys); i += size {
			end := min(i+size, len(keys))
			segs = append(segs, seq2Of(keys[i:end], vals[i:end]))
		}
		return segs
	}
}

// TestCheckSegments2CoverAll pins the pure Segments2-partition predicate: PASS on
// a correct partition (and the empty map) and FAIL on a dropped pair (gap), a key
// present in two segments (overlap), and a single-shot segment (not re-runnable).
func TestCheckSegments2CoverAll(t *testing.T) {
	keys := []int{3, 1, 4, 5, 9, 2, 6} // distinct keys (a map has no dup keys)
	vals := []int{0, 1, 2, 3, 4, 5, 6}

	if ok, msg := checkSegments2CoverAll(seq2Of(keys, vals), segments2Of(keys, vals)); !ok {
		t.Errorf("correct partition should pass, got %q", msg)
	}
	if ok, msg := checkSegments2CoverAll(seq2Of([]int(nil), []int(nil)), segments2Of([]int(nil), []int(nil))); !ok {
		t.Errorf("empty map should pass vacuously, got %q", msg)
	}

	// gap: dropping a pair under-covers.
	gap := func(int) []iter.Seq2[int, int] {
		return []iter.Seq2[int, int]{seq2Of(keys[:len(keys)-1], vals[:len(vals)-1])}
	}
	if ok, _ := checkSegments2CoverAll(seq2Of(keys, vals), gap); ok {
		t.Error("dropping a pair should fail coverage")
	}

	// overlap: a key emitted by two segments trips the no-duplicate-key guard.
	overlap := func(int) []iter.Seq2[int, int] {
		return []iter.Seq2[int, int]{seq2Of(keys, vals), seq2Of(keys[:1], vals[:1])}
	}
	if ok, _ := checkSegments2CoverAll(seq2Of(keys, vals), overlap); ok {
		t.Error("a key in two segments should fail")
	}

	// within-segment duplicate: the SAME key emitted twice inside one segment,
	// with the SAME value, must still fail — a map merge alone would hide it.
	dupKeys := append(append([]int{}, keys...), keys[0]) // 3,1,4,5,9,2,6,3
	dupVals := append(append([]int{}, vals...), vals[0])
	withinDup := func(int) []iter.Seq2[int, int] {
		return []iter.Seq2[int, int]{seq2Of(dupKeys, dupVals)}
	}
	if ok, _ := checkSegments2CoverAll(seq2Of(keys, vals), withinDup); ok {
		t.Error("a key duplicated within one segment should fail")
	}

	// non-re-runnable: single-shot segment is empty on its second pass.
	singleShot := func(int) []iter.Seq2[int, int] {
		used := false
		seg := func(yield func(int, int) bool) {
			if used {
				return
			}
			used = true
			for i := range keys {
				if !yield(keys[i], vals[i]) {
					return
				}
			}
		}
		return []iter.Seq2[int, int]{seg}
	}
	if ok, _ := checkSegments2CoverAll(seq2Of(keys, vals), singleShot); ok {
		t.Error("single-shot (non-re-runnable) segment should fail")
	}
}
