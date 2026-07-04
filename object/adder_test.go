// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package object

import "testing"

// Every mutable non-stack collection now exposes the unified Add(value) bool and
// so satisfies Adder[T] — the sink a bulk loader targets. The []int (strategy)
// and TreeSet[[]int] instantiations are load-bearing: they type-check only because
// the hierarchy is T-any.
var (
	_ Adder[int]    = (*ArrayList[int])(nil)
	_ Adder[int]    = (*HashSet[int])(nil)
	_ Adder[int]    = (*HashBag[int])(nil)
	_ Adder[int]    = (*TreeSet[int])(nil)
	_ Adder[[]int]  = (*TreeSet[[]int])(nil)
	_ Adder[[]int]  = (*HashSetWithStrategy[[]int])(nil)
	_ Adder[string] = (*LinkedHashSet[string])(nil)
)

// TestAdderReturnContract pins the per-category bool: lists and bags always accept
// (true); a set reports whether the value was newly inserted.
func TestAdderReturnContract(t *testing.T) {
	var list Adder[int] = NewArrayList[int]()
	if !list.Add(5) || !list.Add(5) {
		t.Error("list.Add must always return true (accepts duplicates)")
	}

	var bag Adder[int] = NewHashBag[int]()
	if !bag.Add(5) || !bag.Add(5) {
		t.Error("bag.Add must always return true (accepts duplicate occurrences)")
	}

	var set Adder[int] = NewHashSet[int]()
	if !set.Add(5) {
		t.Error("set.Add(new) must return true (newly inserted)")
	}
	if set.Add(5) {
		t.Error("set.Add(duplicate) must return false (already present)")
	}

	// A comparator set reports insertion the same way through the interface.
	var ts Adder[int] = NewTreeSet[int](NaturalComparator[int]())
	if !ts.Add(1) || ts.Add(1) {
		t.Error("TreeSet.Add: want true then false for a repeated value")
	}
}
