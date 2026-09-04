package treeset

import (
	"sort"
	"testing"

	"github.com/mapdb/mapdb-golang/pump"
)

// assertSetSizeInvariant verifies the subtree-size invariant at every node
// (size == 1 + size(left) + size(right)) and that the root total equals Len().
func assertSetSizeInvariant(t *testing.T, s *Int32) {
	t.Helper()
	var check func(n *int32Node) int
	check = func(n *int32Node) int {
		if n == nil {
			return 0
		}
		l := check(n.left)
		r := check(n.right)
		if n.size != 1+l+r {
			t.Fatalf("subtree-size invariant violated at key %d: size=%d want=%d", n.key, n.size, 1+l+r)
		}
		return n.size
	}
	if total := check(s.root); total != s.Len() {
		t.Fatalf("root size %d mismatches Len() %d", total, s.Len())
	}
}

func TestInt32_RankSelectBasic(t *testing.T) {
	s := Int32Of(10, 20, 30, 40, 50)
	for _, c := range []struct {
		k, want int32
	}{{10, 0}, {30, 2}, {50, 4}, {5, 0}, {25, 2}, {55, 5}} {
		if got := s.Rank(c.k); got != int(c.want) {
			t.Errorf("Rank(%d) = %d, want %d", c.k, got, c.want)
		}
	}
	for _, c := range []struct {
		i    int
		want int32
	}{{0, 10}, {2, 30}, {4, 50}} {
		if v, ok := s.Select(c.i); !ok || v != c.want {
			t.Errorf("Select(%d) = (%d, %v), want (%d, true)", c.i, v, ok, c.want)
		}
	}
	if v, ok := s.Select(5); ok {
		t.Errorf("Select(5) = (%d, true), want absence", v)
	}
}

func TestInt32_SelectNegativeIndexAbsence(t *testing.T) {
	s := Int32Of(10, 20, 30)
	for _, i := range []int{-1, -2, -1000} {
		if v, ok := s.Select(i); ok {
			t.Errorf("Select(%d) = (%d, true), want absence", i, v)
		}
	}
	empty := NewInt32()
	if _, ok := empty.Select(-1); ok {
		t.Error("Select(-1) on empty should be absent")
	}
	if _, ok := empty.Select(0); ok {
		t.Error("Select(0) on empty should be absent")
	}
}

func TestInt32_RankSelectEmptySingle(t *testing.T) {
	empty := NewInt32()
	if r := empty.Rank(5); r != 0 {
		t.Errorf("empty Rank(5) = %d, want 0", r)
	}
	if _, ok := empty.Select(0); ok {
		t.Error("empty Select(0) should be absent")
	}

	s := Int32Of(7)
	for _, c := range []struct {
		k, want int32
	}{{6, 0}, {7, 0}, {8, 1}} {
		if got := s.Rank(c.k); got != int(c.want) {
			t.Errorf("single Rank(%d) = %d, want %d", c.k, got, c.want)
		}
	}
	if v, ok := s.Select(0); !ok || v != 7 {
		t.Errorf("single Select(0) = (%d, %v), want (7, true)", v, ok)
	}
	if _, ok := s.Select(1); ok {
		t.Error("single Select(1) should be absent")
	}
}

func TestInt32_RankSelectSigned(t *testing.T) {
	const minI32 = int32(-2147483648)
	const maxI32 = int32(2147483647)
	s := Int32Of(minI32, -1, 0, 1, maxI32)
	if r := s.Rank(minI32); r != 0 {
		t.Errorf("Rank(MIN) = %d, want 0", r)
	}
	if r := s.Rank(0); r != 2 {
		t.Errorf("Rank(0) = %d, want 2", r)
	}
	if r := s.Rank(maxI32); r != 4 {
		t.Errorf("Rank(MAX) = %d, want 4", r)
	}
	if v, ok := s.Select(0); !ok || v != minI32 {
		t.Errorf("Select(0) = (%d, %v), want (MIN, true)", v, ok)
	}
	if v, ok := s.Select(4); !ok || v != maxI32 {
		t.Errorf("Select(4) = (%d, %v), want (MAX, true)", v, ok)
	}
	if _, ok := s.Select(5); ok {
		t.Error("Select(5) should be absent")
	}
}

func TestInt32_RankSelectAfterRemoveRoundTrip(t *testing.T) {
	s := Int32Of(10, 20, 30, 40, 50)
	if !s.Remove(30) {
		t.Fatal("Remove(30) failed")
	}
	if r := s.Rank(40); r != 2 {
		t.Errorf("Rank(40) = %d, want 2", r)
	}
	if r := s.Rank(35); r != 2 {
		t.Errorf("Rank(35) = %d, want 2", r)
	}
	if v, ok := s.Select(2); !ok || v != 40 {
		t.Errorf("Select(2) = (%d, %v), want (40, true)", v, ok)
	}
	if _, ok := s.Select(4); ok {
		t.Error("Select(4) should be absent")
	}
	assertSetSizeInvariant(t, s)
	// round-trip identity over the live set
	var elems []int32
	for v := range s.All() {
		elems = append(elems, v)
	}
	for _, x := range elems {
		if got, ok := s.Select(s.Rank(x)); !ok || got != x {
			t.Errorf("Select(Rank(%d)) = (%d, %v), want (%d, true)", x, got, ok, x)
		}
	}
	for i := 0; i < s.Len(); i++ {
		x, ok := s.Select(i)
		if !ok {
			t.Fatalf("Select(%d) absent within range", i)
		}
		if r := s.Rank(x); r != i {
			t.Errorf("Rank(Select(%d)=%d) = %d, want %d", i, x, r, i)
		}
	}
}

func xorshift(state *uint64) uint64 {
	x := *state
	x ^= x << 13
	x ^= x >> 7
	x ^= x << 17
	*state = x
	return x
}

func TestInt32_SizeInvariantRandomizedInsertRemove(t *testing.T) {
	s := NewInt32()
	present := map[int32]bool{}
	state := uint64(0x9E3779B97F4A7C15)
	for i := 0; i < 4000; i++ {
		key := int32(xorshift(&state) % 200)
		if xorshift(&state)&1 == 0 {
			s.Add(key)
			present[key] = true
		} else {
			s.Remove(key)
			delete(present, key)
		}
		assertSetSizeInvariant(t, s)
		if s.Len() != len(present) {
			t.Fatalf("Len() %d mismatches oracle %d at step %d", s.Len(), len(present), i)
		}
	}
	sorted := make([]int32, 0, len(present))
	for k := range present {
		sorted = append(sorted, k)
	}
	sort.Slice(sorted, func(a, b int) bool { return sorted[a] < sorted[b] })
	for i, k := range sorted {
		if r := s.Rank(k); r != i {
			t.Fatalf("Rank(%d) = %d, want %d", k, r, i)
		}
		if got, ok := s.Select(i); !ok || got != k {
			t.Fatalf("Select(%d) = (%d, %v), want (%d, true)", i, got, ok, k)
		}
	}
	if _, ok := s.Select(len(sorted)); ok {
		t.Error("Select(size) should be absent")
	}
}

// TestInt32_RankSelectAfterFromSorted is the treeset twin of the treemap
// bulk-load regression guard: the bottom-up builder must establish the
// subtree-size augmentation, since it bypasses the insert/rotation paths that
// normally maintain it.
func TestInt32_RankSelectAfterFromSorted(t *testing.T) {
	values := []int32{-10, 0, 5, 20, 21}
	s, err := NewInt32FromSorted(values, pump.ErrorOnDuplicate)
	if err != nil {
		t.Fatalf("NewInt32FromSorted: %v", err)
	}
	assertSetSizeInvariant(t, s)
	for i, v := range values {
		if got := s.Rank(v); got != i {
			t.Errorf("Rank(%d) = %d, want %d", v, got, i)
		}
		gotV, ok := s.Select(i)
		if !ok || gotV != v {
			t.Errorf("Select(%d) = (%d, %v), want (%d, true)", i, gotV, ok, v)
		}
	}
	if got := s.Rank(22); got != len(values) {
		t.Errorf("Rank(22) = %d, want %d", got, len(values))
	}

	big := make([]int32, 257)
	for i := range big {
		big[i] = int32(i * 3)
	}
	bs, err := NewInt32FromSorted(big, pump.ErrorOnDuplicate)
	if err != nil {
		t.Fatalf("NewInt32FromSorted(big): %v", err)
	}
	assertSetSizeInvariant(t, bs)
	for i, v := range big {
		if got := bs.Rank(v); got != i {
			t.Fatalf("big Rank(%d) = %d, want %d", v, got, i)
		}
	}
}
