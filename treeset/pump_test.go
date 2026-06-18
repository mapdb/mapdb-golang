package treeset

import (
	"errors"
	"math"
	"math/rand"
	"testing"

	"github.com/mapdb/mapdb-golang/pump"
)

func validateRBInt32(t *testing.T, s *Int32) {
	t.Helper()
	if s.root == nil {
		return
	}
	if s.root.color != int32NodeBlack {
		t.Fatal("root not black")
	}
	var check func(n *int32Node) int
	check = func(n *int32Node) int {
		if n == nil {
			return 1
		}
		if n.color == int32NodeRed {
			if (n.left != nil && n.left.color == int32NodeRed) ||
				(n.right != nil && n.right.color == int32NodeRed) {
				t.Fatalf("red-red at %d", n.key)
			}
		}
		if n.left != nil && n.left.parent != n {
			t.Fatalf("bad parent at %d", n.key)
		}
		if n.right != nil && n.right.parent != n {
			t.Fatalf("bad parent at %d", n.key)
		}
		lb, rb := check(n.left), check(n.right)
		if lb != rb {
			t.Fatalf("black-height mismatch at %d", n.key)
		}
		if n.color == int32NodeBlack {
			return lb + 1
		}
		return lb
	}
	check(s.root)
}

func TestInt32FromSorted_EqualsIncremental(t *testing.T) {
	for _, n := range []int{0, 1, 2, 3, 7, 8, 16, 100} {
		vals := make([]int32, n)
		for i := range vals {
			vals[i] = int32(i)
		}
		built, err := NewInt32FromSorted(vals, pump.ErrorOnDuplicate)
		if err != nil {
			t.Fatalf("n=%d: %v", n, err)
		}
		validateRBInt32(t, built)
		incr := NewInt32()
		for _, v := range vals {
			incr.Add(v)
		}
		var b, i []int32
		for v := range built.All() {
			b = append(b, v)
		}
		for v := range incr.All() {
			i = append(i, v)
		}
		if len(b) != len(i) {
			t.Fatalf("n=%d count differs", n)
		}
		for k := range b {
			if b[k] != i[k] {
				t.Fatalf("n=%d order differs", n)
			}
		}
	}
}

func TestInt32FromSorted_PostBuildMutation(t *testing.T) {
	n := 150
	vals := make([]int32, n)
	for i := range vals {
		vals[i] = int32(i)
	}
	s, err := NewInt32FromSorted(vals, pump.ErrorOnDuplicate)
	if err != nil {
		t.Fatal(err)
	}
	rng := rand.New(rand.NewSource(7))
	model := map[int32]bool{}
	for _, v := range vals {
		model[v] = true
	}
	for i := 0; i < 4000; i++ {
		v := int32(rng.Intn(2 * n))
		if rng.Intn(2) == 0 {
			s.Add(v)
			model[v] = true
		} else {
			s.Remove(v)
			delete(model, v)
		}
		validateRBInt32(t, s)
	}
	if s.Len() != len(model) {
		t.Fatalf("size %d != %d", s.Len(), len(model))
	}
}

func TestInt32FromSorted_Errors(t *testing.T) {
	if _, err := NewInt32FromSorted([]int32{3, 1}, pump.ErrorOnDuplicate); !errors.Is(err, pump.ErrNotSorted) {
		t.Fatalf("got %v", err)
	}
	if _, err := NewInt32FromSorted([]int32{1, 2, 2}, pump.ErrorOnDuplicate); !errors.Is(err, pump.ErrDuplicateKey) {
		t.Fatalf("got %v", err)
	}
	s, err := NewInt32FromSorted([]int32{1, 2, 2, 3}, pump.IgnoreDuplicates)
	if err != nil || s.Len() != 3 {
		t.Fatalf("ignore dup: len=%d err=%v", s.Len(), err)
	}
}

func TestInt32Sink(t *testing.T) {
	s := NewInt32Sink(pump.ErrorOnDuplicate)
	for i := int32(0); i < 5; i++ {
		s.Add(i)
	}
	set, err := s.Build()
	if err != nil || set.Len() != 5 {
		t.Fatalf("len=%d err=%v", set.Len(), err)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on second build")
		}
	}()
	s.Build()
}

func TestFloat64FromSorted_FloatEdges(t *testing.T) {
	negZero := math.Copysign(0, -1)
	vals := []float64{math.Inf(-1), -1, negZero, 0, 1, math.Inf(1), math.NaN()}
	s, err := NewFloat64FromSorted(vals, pump.ErrorOnDuplicate)
	if err != nil {
		t.Fatal(err)
	}
	var got []float64
	for v := range s.All() {
		got = append(got, v)
	}
	for i := range vals {
		if math.Float64bits(got[i]) != math.Float64bits(vals[i]) {
			t.Fatalf("order differs at %d", i)
		}
	}
}
