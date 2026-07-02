package treemap

import (
	"errors"
	"math"
	"math/rand"
	"testing"

	"github.com/mapdb/mapdb-golang/pump"
)

// validateRBInt32Int32 asserts the red-black invariants on a built tree:
// root is black, no red node has a red child, and every root-to-leaf path has the
// same number of black nodes. Returns the black-height. A violation fails the test.
func validateRBInt32Int32(t *testing.T, m *Int32Int32) {
	t.Helper()
	if m.root == nil {
		return
	}
	if m.root.color != int32Int32TreeNodeBlack {
		t.Fatalf("root is not black")
	}
	var check func(n *int32Int32TreeNode) int
	check = func(n *int32Int32TreeNode) int {
		if n == nil {
			return 1 // nil leaves are black
		}
		if n.color == int32Int32TreeNodeRed {
			if (n.left != nil && n.left.color == int32Int32TreeNodeRed) ||
				(n.right != nil && n.right.color == int32Int32TreeNodeRed) {
				t.Fatalf("red node with red child at key %d", n.key)
			}
		}
		if n.left != nil && n.left.parent != n {
			t.Fatalf("bad parent pointer (left) at key %d", n.key)
		}
		if n.right != nil && n.right.parent != n {
			t.Fatalf("bad parent pointer (right) at key %d", n.key)
		}
		lb := check(n.left)
		rb := check(n.right)
		if lb != rb {
			t.Fatalf("black-height mismatch at key %d: %d vs %d", n.key, lb, rb)
		}
		if n.color == int32Int32TreeNodeBlack {
			return lb + 1
		}
		return lb
	}
	check(m.root)
}

func keysOfInt32Int32(m *Int32Int32) []int32 {
	var ks []int32
	for k := range m.Keys() {
		ks = append(ks, k)
	}
	return ks
}

func TestInt32Int32FromSorted_EqualsIncremental(t *testing.T) {
	for _, n := range []int{0, 1, 2, 3, 5, 7, 8, 15, 16, 17, 100, 257} {
		keys := make([]int32, n)
		vals := make([]int32, n)
		for i := 0; i < n; i++ {
			keys[i] = int32(i * 2)
			vals[i] = int32(i * 10)
		}
		built, err := New32x32(t, keys, vals)
		if err != nil {
			t.Fatalf("n=%d: %v", n, err)
		}
		if built.Len() != n {
			t.Fatalf("n=%d: Len=%d", n, built.Len())
		}
		validateRBInt32Int32(t, built)

		// height bound 2*log2(n+1)
		if h := heightInt32Int32(built.root); n > 0 && float64(h) > 2*math.Log2(float64(n+1))+1 {
			t.Fatalf("n=%d: height %d exceeds bound", n, h)
		}

		incr := NewInt32Int32()
		for i := 0; i < n; i++ {
			incr.Put(keys[i], vals[i])
		}
		bk, ik := keysOfInt32Int32(built), keysOfInt32Int32(incr)
		if len(bk) != len(ik) {
			t.Fatalf("n=%d: key count differs", n)
		}
		for i := range bk {
			if bk[i] != ik[i] {
				t.Fatalf("n=%d: order differs at %d", n, i)
			}
			bv, _ := built.Get(bk[i])
			iv, _ := incr.Get(ik[i])
			if bv != iv {
				t.Fatalf("n=%d: value differs at key %d", n, bk[i])
			}
		}
	}
}

func New32x32(t *testing.T, keys, vals []int32) (*Int32Int32, error) {
	return NewInt32Int32FromSorted(keys, vals, pump.ErrorOnDuplicate)
}

func heightInt32Int32(n *int32Int32TreeNode) int {
	if n == nil {
		return 0
	}
	l, r := heightInt32Int32(n.left), heightInt32Int32(n.right)
	if l > r {
		return l + 1
	}
	return r + 1
}

func TestInt32Int32FromSorted_PostBuildMutation(t *testing.T) {
	n := 200
	keys := make([]int32, n)
	vals := make([]int32, n)
	for i := range keys {
		keys[i] = int32(i)
		vals[i] = int32(i)
	}
	m, err := NewInt32Int32FromSorted(keys, vals, pump.ErrorOnDuplicate)
	if err != nil {
		t.Fatal(err)
	}
	rng := rand.New(rand.NewSource(1))
	model := map[int32]int32{}
	for _, k := range keys {
		model[k] = k
	}
	for i := 0; i < 5000; i++ {
		k := int32(rng.Intn(2 * n))
		if rng.Intn(2) == 0 {
			v := int32(rng.Int())
			m.Put(k, v)
			model[k] = v
		} else {
			m.Remove(k)
			delete(model, k)
		}
		validateRBInt32Int32(t, m)
	}
	if m.Len() != len(model) {
		t.Fatalf("size %d != model %d", m.Len(), len(model))
	}
	for k, v := range model {
		got, ok := m.Get(k)
		if !ok || got != v {
			t.Fatalf("key %d: got (%d,%v) want %d", k, got, ok, v)
		}
	}
}

func TestInt32Int32FromSorted_Errors(t *testing.T) {
	// out of order at first/middle/last
	cases := [][]int32{
		{2, 1, 3, 4},
		{1, 2, 5, 4, 6},
		{1, 2, 3, 2},
	}
	for i, keys := range cases {
		vals := make([]int32, len(keys))
		_, err := NewInt32Int32FromSorted(keys, vals, pump.ErrorOnDuplicate)
		if !errors.Is(err, pump.ErrNotSorted) {
			t.Fatalf("case %d: expected ErrNotSorted, got %v", i, err)
		}
	}
	// duplicate -> ErrDuplicateKey
	_, err := NewInt32Int32FromSorted([]int32{1, 2, 2, 3}, []int32{0, 0, 0, 0}, pump.ErrorOnDuplicate)
	if !errors.Is(err, pump.ErrDuplicateKey) {
		t.Fatalf("expected ErrDuplicateKey, got %v", err)
	}
	// IgnoreDuplicates -> first wins, no error
	m, err := NewInt32Int32FromSorted([]int32{1, 2, 2, 3}, []int32{10, 20, 99, 30}, pump.IgnoreDuplicates)
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := m.Get(2); v != 20 {
		t.Fatalf("IgnoreDuplicates kept %d, want first 20", v)
	}
	if m.Len() != 3 {
		t.Fatalf("len %d want 3", m.Len())
	}
}

func TestInt32Int32FromSorted_LengthMismatchPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on length mismatch")
		}
	}()
	NewInt32Int32FromSorted([]int32{1, 2}, []int32{1}, pump.ErrorOnDuplicate)
}

func TestInt32Int32Sink(t *testing.T) {
	s := NewInt32Int32Sink(pump.ErrorOnDuplicate)
	for i := int32(0); i < 10; i++ {
		s.Put(i, i*i)
	}
	m, err := s.Build()
	if err != nil {
		t.Fatal(err)
	}
	if m.Len() != 10 {
		t.Fatalf("len %d", m.Len())
	}
	validateRBInt32Int32(t, m)
	// double build panics
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic on second Build")
			}
		}()
		s.Build()
	}()
	// put after build panics
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic on Put after Build")
			}
		}()
		s.Put(99, 0)
	}()
}

func TestInt32Int32Sink_OrderErrorPoisons(t *testing.T) {
	s := NewInt32Int32Sink(pump.ErrorOnDuplicate)
	s.Put(5, 0)
	s.Put(3, 0) // out of order, reported at Build
	if _, err := s.Build(); !errors.Is(err, pump.ErrNotSorted) {
		t.Fatalf("expected ErrNotSorted, got %v", err)
	}
}

// TestFloat64Int32FromSorted_FloatEdges exercises the total-order pump path with
// NaN, +/-0, +/-Inf.
func TestFloat64Int32FromSorted_FloatEdges(t *testing.T) {
	negZero := math.Copysign(0, -1)
	keys := []float64{
		math.Inf(-1), -1.5, negZero, 0.0, 1.5, math.Inf(1), math.NaN(),
	}
	vals := make([]int32, len(keys))
	for i := range vals {
		vals[i] = int32(i)
	}
	m, err := NewFloat64Int32FromSorted(keys, vals, pump.ErrorOnDuplicate)
	if err != nil {
		t.Fatalf("float pump failed: %v", err)
	}
	if m.Len() != len(keys) {
		t.Fatalf("len %d want %d", m.Len(), len(keys))
	}
	// in-order traversal must equal the sorted input order (total order)
	var got []float64
	for k := range m.Keys() {
		got = append(got, k)
	}
	for i := range keys {
		if math.Float64bits(got[i]) != math.Float64bits(keys[i]) {
			t.Fatalf("order differs at %d: got bits %x want %x", i,
				math.Float64bits(got[i]), math.Float64bits(keys[i]))
		}
	}
	// -0 and +0 are distinct keys
	if _, ok := m.Get(negZero); !ok {
		t.Fatal("-0 not found")
	}
	if _, ok := m.Get(0.0); !ok {
		t.Fatal("+0 not found")
	}
	// out-of-order float (NaN before Inf) errors
	bad := []float64{math.NaN(), math.Inf(1)}
	if _, err := NewFloat64Int32FromSorted(bad, []int32{0, 0}, pump.ErrorOnDuplicate); !errors.Is(err, pump.ErrNotSorted) {
		t.Fatalf("expected ErrNotSorted for NaN-before-Inf, got %v", err)
	}
}
