package treemap

import (
	"sort"
	"testing"
)

// assertSizeInvariant verifies the subtree-size invariant holds at every node
// (size == 1 + size(left) + size(right)) and that the root total equals Len().
func assertSizeInvariant(t *testing.T, m *Int32Int32) {
	t.Helper()
	var check func(n *int32Int32TreeNode) int
	check = func(n *int32Int32TreeNode) int {
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
	if total := check(m.root); total != m.Len() {
		t.Fatalf("root size %d mismatches Len() %d", total, m.Len())
	}
}

func mapOf(keys ...int32) *Int32Int32 {
	m := NewInt32Int32()
	for _, k := range keys {
		m.Put(k, k*10)
	}
	return m
}

func TestInt32Int32_RankPresentAndAbsent(t *testing.T) {
	m := mapOf(10, 20, 30, 40, 50)
	// present keys -> their 0-based index
	cases := []struct {
		k, want int32
	}{
		{10, 0}, {30, 2}, {50, 4},
		// absent keys -> lower-bound index
		{5, 0}, {25, 2}, {55, 5},
	}
	for _, c := range cases {
		if got := m.Rank(c.k); got != int(c.want) {
			t.Errorf("Rank(%d) = %d, want %d", c.k, got, c.want)
		}
	}
}

func TestInt32Int32_SelectKeyAndEntry(t *testing.T) {
	m := mapOf(10, 20, 30, 40, 50)
	for _, c := range []struct {
		i    int
		want int32
	}{{0, 10}, {2, 30}, {4, 50}} {
		k, ok := m.SelectKey(c.i)
		if !ok || k != c.want {
			t.Errorf("SelectKey(%d) = (%d, %v), want (%d, true)", c.i, k, ok, c.want)
		}
		// entry form carries value = key*10
		ek, ev, eok := m.SelectEntry(c.i)
		if !eok || ek != c.want || ev != c.want*10 {
			t.Errorf("SelectEntry(%d) = (%d, %d, %v), want (%d, %d, true)", c.i, ek, ev, eok, c.want, c.want*10)
		}
	}
	// i == size and beyond -> absence, no trap
	if k, ok := m.SelectKey(5); ok {
		t.Errorf("SelectKey(5) = (%d, true), want absence", k)
	}
	if _, _, ok := m.SelectEntry(5); ok {
		t.Error("SelectEntry(5) should be absent")
	}
	if k, ok := m.SelectKey(999); ok {
		t.Errorf("SelectKey(999) = (%d, true), want absence", k)
	}
}

func TestInt32Int32_SelectNegativeIndexAbsence(t *testing.T) {
	m := mapOf(10, 20, 30)
	// Go int can be negative: must return absence and MUST NOT trap.
	for _, i := range []int{-1, -2, -1000} {
		if k, ok := m.SelectKey(i); ok {
			t.Errorf("SelectKey(%d) = (%d, true), want absence", i, k)
		}
		if _, _, ok := m.SelectEntry(i); ok {
			t.Errorf("SelectEntry(%d) should be absent", i)
		}
	}
	// Empty map: every index is absence.
	empty := NewInt32Int32()
	if _, ok := empty.SelectKey(-1); ok {
		t.Error("SelectKey(-1) on empty should be absent")
	}
	if _, ok := empty.SelectKey(0); ok {
		t.Error("SelectKey(0) on empty should be absent")
	}
}

func TestInt32Int32_RankSelectEmptySingle(t *testing.T) {
	empty := NewInt32Int32()
	if r := empty.Rank(5); r != 0 {
		t.Errorf("empty Rank(5) = %d, want 0", r)
	}
	if _, ok := empty.SelectKey(0); ok {
		t.Error("empty SelectKey(0) should be absent")
	}

	single := mapOf(7)
	for _, c := range []struct {
		k, want int32
	}{{6, 0}, {7, 0}, {8, 1}} {
		if got := single.Rank(c.k); got != int(c.want) {
			t.Errorf("single Rank(%d) = %d, want %d", c.k, got, c.want)
		}
	}
	if k, ok := single.SelectKey(0); !ok || k != 7 {
		t.Errorf("single SelectKey(0) = (%d, %v), want (7, true)", k, ok)
	}
	if k, v, ok := single.SelectEntry(0); !ok || k != 7 || v != 70 {
		t.Errorf("single SelectEntry(0) = (%d, %d, %v), want (7, 70, true)", k, v, ok)
	}
	if _, ok := single.SelectKey(1); ok {
		t.Error("single SelectKey(1) should be absent")
	}
}

func TestInt32Int32_RankSelectSignedExtremes(t *testing.T) {
	const minI32 = int32(-2147483648)
	const maxI32 = int32(2147483647)
	m := mapOf(minI32, -1, 0, 1, maxI32)
	if r := m.Rank(minI32); r != 0 {
		t.Errorf("Rank(MIN) = %d, want 0", r)
	}
	if r := m.Rank(0); r != 2 {
		t.Errorf("Rank(0) = %d, want 2", r)
	}
	if r := m.Rank(maxI32); r != 4 {
		t.Errorf("Rank(MAX) = %d, want 4", r)
	}
	if k, ok := m.SelectKey(0); !ok || k != minI32 {
		t.Errorf("SelectKey(0) = (%d, %v), want (MIN, true)", k, ok)
	}
	if k, ok := m.SelectKey(4); !ok || k != maxI32 {
		t.Errorf("SelectKey(4) = (%d, %v), want (MAX, true)", k, ok)
	}
	if _, ok := m.SelectKey(5); ok {
		t.Error("SelectKey(5) should be absent")
	}
}

func TestInt32Int32_RankSelectAfterRemove(t *testing.T) {
	m := mapOf(10, 20, 30, 40, 50)
	if _, ok := m.Remove(30); !ok {
		t.Fatal("Remove(30) failed")
	}
	// {10,20,40,50}; stale subtree sizes after a remove/transplant would corrupt these.
	if r := m.Rank(40); r != 2 {
		t.Errorf("Rank(40) = %d, want 2", r)
	}
	if r := m.Rank(35); r != 2 {
		t.Errorf("Rank(35) = %d, want 2", r)
	}
	if k, ok := m.SelectKey(2); !ok || k != 40 {
		t.Errorf("SelectKey(2) = (%d, %v), want (40, true)", k, ok)
	}
	if _, ok := m.SelectKey(4); ok {
		t.Error("SelectKey(4) should be absent")
	}
	assertSizeInvariant(t, m)
}

func TestInt32Int32_RoundTripSelectRank(t *testing.T) {
	m := mapOf(10, 20, 30, 40, 50, -7, 0, 99)
	// select(rank(k)) == k for every present key
	var keys []int32
	for k := range m.Keys() {
		keys = append(keys, k)
	}
	for _, k := range keys {
		if got, ok := m.SelectKey(m.Rank(k)); !ok || got != k {
			t.Errorf("SelectKey(Rank(%d)) = (%d, %v), want (%d, true)", k, got, ok, k)
		}
	}
	// rank(select(i)) == i for every 0 <= i < size
	for i := 0; i < m.Len(); i++ {
		k, ok := m.SelectKey(i)
		if !ok {
			t.Fatalf("SelectKey(%d) absent within range", i)
		}
		if r := m.Rank(k); r != i {
			t.Errorf("Rank(SelectKey(%d)=%d) = %d, want %d", i, k, r, i)
		}
	}
	// select(size) is absence
	if _, ok := m.SelectKey(m.Len()); ok {
		t.Error("SelectKey(Len()) should be absent")
	}
}

// xorshift is a deterministic PRNG so the randomized invariant test never relies
// on external randomness (and matches the Rust/other-port reproducibility note).
func xorshift(state *uint64) uint64 {
	x := *state
	x ^= x << 13
	x ^= x >> 7
	x ^= x << 17
	*state = x
	return x
}

func TestInt32Int32_SizeInvariantRandomizedInsertRemove(t *testing.T) {
	m := NewInt32Int32()
	present := map[int32]bool{}
	state := uint64(0x9E3779B97F4A7C15)
	for i := 0; i < 4000; i++ {
		key := int32(xorshift(&state) % 200)
		if xorshift(&state)&1 == 0 {
			m.Put(key, key*10)
			present[key] = true
		} else {
			m.Remove(key)
			delete(present, key)
		}
		assertSizeInvariant(t, m)
		if m.Len() != len(present) {
			t.Fatalf("Len() %d mismatches oracle %d at step %d", m.Len(), len(present), i)
		}
	}
	// After the churn, rank/select must agree with the oracle ordering.
	sorted := make([]int32, 0, len(present))
	for k := range present {
		sorted = append(sorted, k)
	}
	sort.Slice(sorted, func(a, b int) bool { return sorted[a] < sorted[b] })
	for i, k := range sorted {
		if r := m.Rank(k); r != i {
			t.Fatalf("Rank(%d) = %d, want %d", k, r, i)
		}
		if got, ok := m.SelectKey(i); !ok || got != k {
			t.Fatalf("SelectKey(%d) = (%d, %v), want (%d, true)", i, got, ok, k)
		}
	}
	if _, ok := m.SelectKey(len(sorted)); ok {
		t.Error("SelectKey(size) should be absent")
	}
}
