// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.

package object

import (
	"iter"
	"slices"
	"testing"
)

// HeadMap/TailMap/SubMap are call-time snapshots (navigable-map spec):
// inserting or deleting a key inside the range after obtaining the view must
// not change what the view yields.
func TestTreeMap_RangeViewsAreSnapshots(t *testing.T) {
	cases := []struct {
		name     string
		view     func(m *TreeMap[int, string]) iter.Seq2[int, string]
		want     []int
		in, del  int
		wantVals []string
	}{
		{"HeadMap", func(m *TreeMap[int, string]) iter.Seq2[int, string] { return m.HeadMap(30) }, []int{10, 20}, 15, 20, []string{"a", "b"}},
		{"TailMap", func(m *TreeMap[int, string]) iter.Seq2[int, string] { return m.TailMap(30) }, []int{30, 40, 50}, 45, 40, []string{"c", "d", "e"}},
		{"SubMap", func(m *TreeMap[int, string]) iter.Seq2[int, string] { return m.SubMap(20, 40) }, []int{20, 30}, 25, 30, []string{"b", "c"}},
	}
	for _, tc := range cases {
		for _, mut := range []string{"insert", "delete"} {
			t.Run(tc.name+"/"+mut, func(t *testing.T) {
				m := buildTestTreeMap()
				view := tc.view(m)
				if mut == "insert" {
					m.Put(tc.in, "new")
				} else if _, ok := m.Remove(tc.del); !ok {
					t.Fatalf("Remove(%d) found nothing", tc.del)
				}
				for pass := 0; pass < 2; pass++ {
					var ks []int
					var vs []string
					for k, v := range view {
						ks = append(ks, k)
						vs = append(vs, v)
					}
					if !slices.Equal(ks, tc.want) || !slices.Equal(vs, tc.wantVals) {
						t.Fatalf("pass %d = %v %v, want %v %v", pass, ks, vs, tc.want, tc.wantVals)
					}
				}
			})
		}
	}
}

// A reverse-comparator map snapshots in its own order.
func TestTreeMap_RangeViewSnapshotReverseComparator(t *testing.T) {
	m := NewTreeMap[int, string](func(a, b int) int { return b - a })
	for _, k := range []int{1, 2, 3, 4, 5} {
		m.Put(k, "")
	}
	head := m.HeadMap(3) // keys before 3 in reverse order: 5, 4
	m.Remove(4)
	m.Put(6, "")
	var ks []int
	for k := range head {
		ks = append(ks, k)
	}
	if !slices.Equal(ks, []int{5, 4}) {
		t.Fatalf("HeadMap = %v, want [5 4]", ks)
	}
}
