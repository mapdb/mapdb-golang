package object

import (
	"math"
	"slices"
	"testing"
)

func checkEmptyPutAll[K comparable](t *testing.T, absent, present K) {
	t.Helper()
	for _, zeroValue := range []bool{false, true} {
		t.Run(map[bool]string{false: "constructor", true: "zero-value"}[zeroValue], func(t *testing.T) {
			m := NewHashMultimap[K, int]()
			if zeroValue {
				m = &HashMultimap[K, int]{}
			}
			for _, values := range [][]int{nil, {}, nil} {
				m.PutAll(absent, values...)
			}
			if m.ContainsKey(absent) || m.SizeDistinct() != 0 || m.Len() != 0 || len(slices.Collect(m.Keys())) != 0 || len(slices.Collect(m.Values())) != 0 {
				t.Fatalf("empty PutAll created a key: distinct=%d size=%d", m.SizeDistinct(), m.Len())
			}
			m.PutAll(present, 4, 4, 7)
			m.PutAll(present)
			m.PutAll(absent, []int{}...)
			if !m.ContainsKey(present) || m.ContainsKey(absent) || m.SizeDistinct() != 1 || m.Len() != 3 || !slices.Equal(m.Get(present), []int{4, 4, 7}) {
				t.Fatal("empty PutAll changed existing values or introduced an absent key")
			}
			if got := m.RemoveKey(present); !slices.Equal(got, []int{4, 4, 7}) {
				t.Fatal(got)
			}
			m.PutAll(present)
			if m.ContainsKey(present) || m.SizeDistinct() != 0 || m.Len() != 0 {
				t.Fatal("empty PutAll resurrected removed key")
			}
		})
	}
}

func TestHashMultimapEmptyPutAll(t *testing.T) {
	t.Run("string", func(t *testing.T) { checkEmptyPutAll(t, "absent", "present") })
	t.Run("signed-zero", func(t *testing.T) { checkEmptyPutAll(t, math.Copysign(0, -1), float64(0)) })
	t.Run("nan-payload", func(t *testing.T) {
		checkEmptyPutAll(t, math.Float64frombits(0x7ff8000000000001), math.Float64frombits(0x7ff8000000000002))
	})
}
