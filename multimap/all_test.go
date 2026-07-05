// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package multimap

import "testing"

// TestAllYieldsEveryPairList confirms All() yields every (key, value) pair —
// including both values of a repeated key — with the count matching Len().
func TestAllYieldsEveryPairList(t *testing.T) {
	m := NewInt32Int32List()
	m.Put(1, 10)
	m.Put(1, 20)
	m.Put(2, 30)

	seen := map[[2]int32]int{}
	n := 0
	for k, v := range m.All() {
		seen[[2]int32{k, v}]++
		n++
	}
	if n != m.Len() {
		t.Fatalf("All() yielded %d pairs, Len()=%d", n, m.Len())
	}
	for _, want := range [][2]int32{{1, 10}, {1, 20}, {2, 30}} {
		if seen[want] != 1 {
			t.Fatalf("pair %v yielded %d times, want 1", want, seen[want])
		}
	}
}

// TestAllEarlyBreak is the load-bearing case: All() must honor an early break.
// This is why All() iterates m.data natively rather than wrapping ForEach — a
// non-bool ForEach callback cannot stop, and calling yield after a break would
// panic ("range function continued iteration after loop body returned"). Both
// List and Set are exercised.
func TestAllEarlyBreak(t *testing.T) {
	list := NewInt32Int32List()
	list.Put(1, 10)
	list.Put(2, 20)
	list.Put(3, 30)
	n := 0
	for range list.All() { // must not panic
		n++
		break
	}
	if n != 1 {
		t.Fatalf("List early-break saw %d, want 1", n)
	}

	set := NewInt32Int32Set()
	set.Put(1, 10)
	set.Put(2, 20)
	set.Put(3, 30)
	n = 0
	for range set.All() { // must not panic
		n++
		break
	}
	if n != 1 {
		t.Fatalf("Set early-break saw %d, want 1", n)
	}
}
