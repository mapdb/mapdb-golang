// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package fenwick

import (
	"slices"
	"testing"
)

// TestValuesYieldsElementsInOrder pins Values(): it reproduces each element in
// index order, agrees with Get, has length Len, and reflects a later Set.
func TestValuesYieldsElementsInOrder(t *testing.T) {
	src := []int32{5, -3, 0, 7, 2, -1, 100}
	f := NewFenwickTreeFromValues(src)

	got := slices.Collect(f.Values())

	want := make([]int64, len(src))
	for i := range src {
		want[i] = int64(src[i])
	}
	if !slices.Equal(got, want) {
		t.Fatalf("Values() = %v, want %v", got, want)
	}
	if len(got) != f.Len() {
		t.Fatalf("Values() length = %d, Len() = %d", len(got), f.Len())
	}
	// Agrees with Get at every index.
	for i := range src {
		if got[i] != f.Get(i) {
			t.Fatalf("Values()[%d]=%d != Get(%d)=%d", i, got[i], i, f.Get(i))
		}
	}

	// A mutation is reflected (Values is not a stale snapshot).
	f.Set(3, 42)
	got2 := slices.Collect(f.Values())
	if got2[3] != 42 {
		t.Fatalf("after Set(3,42), Values()[3]=%d, want 42", got2[3])
	}
}

// TestValuesEmpty confirms Values() over an empty tree yields nothing.
func TestValuesEmpty(t *testing.T) {
	if got := slices.Collect(NewFenwickTreeWithSize(0).Values()); len(got) != 0 {
		t.Fatalf("Values() over empty tree = %v, want none", got)
	}
}

// TestValuesEarlyBreak confirms the lazy iterator honors an early break.
func TestValuesEarlyBreak(t *testing.T) {
	f := NewFenwickTreeFromValues([]int32{10, 20, 30, 40})
	var seen []int64
	for v := range f.Values() {
		seen = append(seen, v)
		if v == 20 {
			break
		}
	}
	if !slices.Equal(seen, []int64{10, 20}) {
		t.Fatalf("early-break Values() = %v, want [10 20]", seen)
	}
}
