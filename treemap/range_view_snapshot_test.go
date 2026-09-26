package treemap

import (
	"iter"
	"slices"
	"testing"
)

// Range views (RangeKeys/SubMap/HeadMap/TailMap) are call-time snapshots per
// the navigable-map spec: mutating the map after obtaining a view -- inserting
// or deleting a key inside the range -- must not change what the view yields.

func collect2[K comparable, V any](seq iter.Seq2[K, V]) ([]K, []V) {
	var ks []K
	var vs []V
	for k, v := range seq {
		ks = append(ks, k)
		vs = append(vs, v)
	}
	return ks, vs
}

func TestInt32Int32_RangeViewsAreSnapshots(t *testing.T) {
	cases := []struct {
		name string
		view func(m *Int32Int32) iter.Seq2[int32, int32]
		want []int32
		in   int32 // a new key inside the range
		del  int32 // an existing key inside the range
	}{
		{"RangeKeys", func(m *Int32Int32) iter.Seq2[int32, int32] { return m.RangeKeys(20, 40) }, []int32{20, 30}, 25, 20},
		{"SubMap", func(m *Int32Int32) iter.Seq2[int32, int32] { return m.SubMap(20, 40) }, []int32{20, 30}, 35, 30},
		{"HeadMap", func(m *Int32Int32) iter.Seq2[int32, int32] { return m.HeadMap(30) }, []int32{10, 20}, 5, 10},
		{"TailMap", func(m *Int32Int32) iter.Seq2[int32, int32] { return m.TailMap(30) }, []int32{30, 40, 50}, 45, 50},
	}
	for _, tc := range cases {
		for _, mut := range []string{"insert", "delete", "overwrite"} {
			t.Run(tc.name+"/"+mut, func(t *testing.T) {
				m := NewInt32Int32()
				for _, k := range []int32{10, 20, 30, 40, 50} {
					m.Put(k, k*10)
				}
				view := tc.view(m)
				switch mut {
				case "insert":
					m.Put(tc.in, -1)
				case "delete":
					if _, ok := m.Remove(tc.del); !ok {
						t.Fatalf("Remove(%d) found nothing", tc.del)
					}
				case "overwrite":
					m.Put(tc.del, -1)
				}
				for pass := 0; pass < 2; pass++ { // replayable, same snapshot
					ks, vs := collect2(view)
					if !slices.Equal(ks, tc.want) {
						t.Fatalf("pass %d keys = %v, want %v", pass, ks, tc.want)
					}
					for i, k := range ks {
						if vs[i] != k*10 {
							t.Fatalf("pass %d value[%d] = %d, want %d", pass, k, vs[i], k*10)
						}
					}
				}
			})
		}
	}
}

// Deleting every key of the view while ranging over it is safe and visits the
// whole snapshot.
func TestInt32Int32_RangeViewRemoveWhileIterating(t *testing.T) {
	m := NewInt32Int32()
	for k := int32(1); k <= 10; k++ {
		m.Put(k, k)
	}
	var seen []int32
	for k := range m.SubMap(3, 8) {
		seen = append(seen, k)
		m.Remove(k)
	}
	if want := []int32{3, 4, 5, 6, 7}; !slices.Equal(seen, want) {
		t.Fatalf("seen = %v, want %v", seen, want)
	}
	if m.Len() != 5 {
		t.Fatalf("Len = %d, want 5", m.Len())
	}
}

func TestFloat64Int32_RangeViewsAreSnapshots(t *testing.T) {
	m := NewFloat64Int32()
	for i, k := range []float64{-1.5, 0.5, 1.5, 2.5} {
		m.Put(k, int32(i))
	}
	head := m.HeadMap(1.5)
	tail := m.TailMap(0.5)
	sub := m.SubMap(0.0, 2.0)
	m.Put(1.0, 99)
	m.Remove(0.5)
	if ks, _ := collect2(head); !slices.Equal(ks, []float64{-1.5, 0.5}) {
		t.Errorf("HeadMap = %v", ks)
	}
	if ks, _ := collect2(tail); !slices.Equal(ks, []float64{0.5, 1.5, 2.5}) {
		t.Errorf("TailMap = %v", ks)
	}
	if ks, _ := collect2(sub); !slices.Equal(ks, []float64{0.5, 1.5}) {
		t.Errorf("SubMap = %v", ks)
	}
}

func TestCharInt8_RangeViewsAreSnapshots(t *testing.T) {
	m := NewCharInt8()
	for _, k := range []uint16{'a', 'c', 'e', 'g'} {
		m.Put(k, int8(k-'a'))
	}
	rk := m.RangeKeys('b', 'f')
	m.Put('d', 3)
	m.Remove('c')
	if ks, _ := collect2(rk); !slices.Equal(ks, []uint16{'c', 'e'}) {
		t.Errorf("RangeKeys = %v", ks)
	}
}
