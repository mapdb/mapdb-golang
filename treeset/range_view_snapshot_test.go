package treeset

import (
	"slices"
	"testing"
)

// RangeValues is a call-time snapshot (navigable-map spec): mutating the set
// afterwards -- inserting or deleting an element inside the range -- must not
// change what the returned iterator yields.

func TestInt32_RangeValuesIsSnapshot(t *testing.T) {
	for _, mut := range []string{"insert", "delete"} {
		t.Run(mut, func(t *testing.T) {
			s := NewInt32()
			for _, v := range []int32{10, 20, 30, 40, 50} {
				s.Add(v)
			}
			view := s.RangeValues(20, 40)
			if mut == "insert" {
				s.Add(25)
			} else if !s.Remove(30) {
				t.Fatal("Remove(30) found nothing")
			}
			for pass := 0; pass < 2; pass++ {
				if got := slices.Collect(view); !slices.Equal(got, []int32{20, 30}) {
					t.Fatalf("pass %d RangeValues = %v, want [20 30]", pass, got)
				}
			}
		})
	}
}

func TestInt32_RangeValuesRemoveWhileIterating(t *testing.T) {
	s := NewInt32()
	for v := int32(1); v <= 10; v++ {
		s.Add(v)
	}
	var seen []int32
	for v := range s.RangeValues(3, 8) {
		seen = append(seen, v)
		s.Remove(v)
	}
	if want := []int32{3, 4, 5, 6, 7}; !slices.Equal(seen, want) {
		t.Fatalf("seen = %v, want %v", seen, want)
	}
}

func TestFloat32_RangeValuesIsSnapshot(t *testing.T) {
	s := NewFloat32()
	for _, v := range []float32{-1.5, 0.5, 1.5, 2.5} {
		s.Add(v)
	}
	view := s.RangeValues(0, 2)
	s.Add(1.0)
	s.Remove(0.5)
	if got := slices.Collect(view); !slices.Equal(got, []float32{0.5, 1.5}) {
		t.Fatalf("RangeValues = %v, want [0.5 1.5]", got)
	}
}
