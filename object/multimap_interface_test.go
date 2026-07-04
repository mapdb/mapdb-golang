// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package object

import (
	"slices"
	"testing"
)

// Drive both multimap implementations through the MutableMultimap interface — the
// contract a grouping terminal (GroupBy) can return instead of a concrete type.
func TestMutableMultimapInterface(t *testing.T) {
	cases := []struct {
		name string
		mm   MutableMultimap[string, int]
	}{
		{"hash", NewHashMultimap[string, int]()},
		{"tree", NewTreeMultimap[string, int](NaturalComparator[string]())},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mm := tc.mm
			mm.Put("a", 1)
			mm.PutAll("a", 2, 3)
			mm.Put("b", 9)

			if mm.Len() != 4 {
				t.Errorf("Len = %d, want 4 (total pairs)", mm.Len())
			}
			if mm.SizeDistinct() != 2 {
				t.Errorf("SizeDistinct = %d, want 2 (distinct keys)", mm.SizeDistinct())
			}
			if !mm.ContainsKey("a") || mm.ContainsKey("z") {
				t.Error("ContainsKey wrong")
			}
			got := mm.Get("a")
			slices.Sort(got)
			if !slices.Equal(got, []int{1, 2, 3}) {
				t.Errorf("Get(a) = %v, want [1 2 3]", got)
			}
			// Get returns a copy: mutating it must not affect the multimap.
			got[0] = 999
			if g2 := mm.Get("a"); slices.Contains(g2, 999) {
				t.Error("Get did not return a copy — mutation leaked back")
			}

			// Keys distinct, Values all-across-keys.
			var keys []string
			for k := range mm.Keys() {
				keys = append(keys, k)
			}
			slices.Sort(keys)
			if !slices.Equal(keys, []string{"a", "b"}) {
				t.Errorf("Keys = %v, want [a b]", keys)
			}
			var vals []int
			for v := range mm.Values() {
				vals = append(vals, v)
			}
			if len(vals) != 4 {
				t.Errorf("Values count = %d, want 4", len(vals))
			}

			// RemoveKey returns the removed values and drops the key.
			removed := mm.RemoveKey("a")
			slices.Sort(removed)
			if !slices.Equal(removed, []int{1, 2, 3}) {
				t.Errorf("RemoveKey(a) = %v, want [1 2 3]", removed)
			}
			if mm.ContainsKey("a") || mm.Len() != 1 {
				t.Errorf("after RemoveKey(a): ContainsKey=%v Len=%d, want false,1", mm.ContainsKey("a"), mm.Len())
			}
		})
	}
}
