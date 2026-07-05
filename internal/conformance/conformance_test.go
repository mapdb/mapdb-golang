// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package conformance

import (
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

// TestAllMatchesToSliceThroughT drives the exported *testing.T wrapper on a
// passing case to confirm it does not spuriously fail (the failing path is a
// t.Error, which cannot be asserted here without a testing.TB mock — the pure
// predicate above owns the negative coverage).
func TestAllMatchesToSliceThroughT(t *testing.T) {
	vals := []int{5, 3, 5, 1}
	AllMatchesToSlice(t, seqOf(vals), slices.Clone(vals), true)
	AllMatchesToSlice(t, seqOf(vals), []int{1, 3, 5, 5}, false)
}
