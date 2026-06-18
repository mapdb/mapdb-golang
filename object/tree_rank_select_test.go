package object

import (
	"sort"
	"testing"
)

// assertTmSizeInvariant verifies the subtree-size invariant at every node
// (size == 1 + size(left) + size(right)) and that the root total equals Len().
func assertTmSizeInvariant[K any, V any](t *testing.T, m *TreeMap[K, V]) {
	t.Helper()
	var check func(n *tmNode[K, V]) int
	check = func(n *tmNode[K, V]) int {
		if n == nil {
			return 0
		}
		l := check(n.left)
		r := check(n.right)
		if n.size != 1+l+r {
			t.Fatalf("subtree-size invariant violated at key %v: size=%d want=%d", n.key, n.size, 1+l+r)
		}
		return n.size
	}
	if total := check(m.root); total != m.Len() {
		t.Fatalf("root size %d mismatches Len() %d", total, m.Len())
	}
}

func objMapOf(keys ...int) *TreeMap[int, int] {
	m := NewTreeMap[int, int](NaturalComparator[int]())
	for _, k := range keys {
		m.Put(k, k*10)
	}
	return m
}

func TestObjectTreeMap_RankSelect(t *testing.T) {
	m := objMapOf(10, 20, 30, 40, 50)
	for _, c := range []struct{ k, want int }{
		{10, 0}, {30, 2}, {50, 4}, {5, 0}, {25, 2}, {55, 5},
	} {
		if got := m.Rank(c.k); got != c.want {
			t.Errorf("Rank(%d) = %d, want %d", c.k, got, c.want)
		}
	}
	for _, c := range []struct{ i, want int }{{0, 10}, {2, 30}, {4, 50}} {
		if k, ok := m.SelectKey(c.i); !ok || k != c.want {
			t.Errorf("SelectKey(%d) = (%d, %v), want (%d, true)", c.i, k, ok, c.want)
		}
		if k, v, ok := m.SelectEntry(c.i); !ok || k != c.want || v != c.want*10 {
			t.Errorf("SelectEntry(%d) = (%d, %d, %v)", c.i, k, v, ok)
		}
	}
	if _, ok := m.SelectKey(5); ok {
		t.Error("SelectKey(5) should be absent")
	}
	// negative index: absence, no trap
	for _, i := range []int{-1, -100} {
		if _, ok := m.SelectKey(i); ok {
			t.Errorf("SelectKey(%d) should be absent", i)
		}
		if _, _, ok := m.SelectEntry(i); ok {
			t.Errorf("SelectEntry(%d) should be absent", i)
		}
	}
}

func TestObjectTreeMap_RankSelectAfterRemove(t *testing.T) {
	m := objMapOf(10, 20, 30, 40, 50)
	if _, ok := m.Remove(30); !ok {
		t.Fatal("Remove(30) failed")
	}
	if r := m.Rank(40); r != 2 {
		t.Errorf("Rank(40) = %d, want 2", r)
	}
	if k, ok := m.SelectKey(2); !ok || k != 40 {
		t.Errorf("SelectKey(2) = (%d, %v), want (40, true)", k, ok)
	}
	if _, ok := m.SelectKey(4); ok {
		t.Error("SelectKey(4) should be absent")
	}
	assertTmSizeInvariant(t, m)
}

func objXorshift(state *uint64) uint64 {
	x := *state
	x ^= x << 13
	x ^= x >> 7
	x ^= x << 17
	*state = x
	return x
}

func TestObjectTree_SizeInvariantRandomized(t *testing.T) {
	m := NewTreeMap[int, int](NaturalComparator[int]())
	set := NewTreeSet[int](NaturalComparator[int]())
	present := map[int]bool{}
	state := uint64(0x9E3779B97F4A7C15)
	for i := 0; i < 4000; i++ {
		key := int(objXorshift(&state) % 200)
		if objXorshift(&state)&1 == 0 {
			m.Put(key, key*10)
			set.Add(key)
			present[key] = true
		} else {
			m.Remove(key)
			set.Remove(key)
			delete(present, key)
		}
		assertTmSizeInvariant(t, m)
		assertTmSizeInvariant(t, &set.tree)
		if m.Len() != len(present) || set.Len() != len(present) {
			t.Fatalf("len mismatch at step %d: map=%d set=%d oracle=%d", i, m.Len(), set.Len(), len(present))
		}
	}
	sorted := make([]int, 0, len(present))
	for k := range present {
		sorted = append(sorted, k)
	}
	sort.Ints(sorted)
	for i, k := range sorted {
		if r := m.Rank(k); r != i {
			t.Fatalf("map Rank(%d) = %d, want %d", k, r, i)
		}
		if got, ok := m.SelectKey(i); !ok || got != k {
			t.Fatalf("map SelectKey(%d) = (%d, %v), want (%d, true)", i, got, ok, k)
		}
		if got, ok := set.Select(i); !ok || got != k {
			t.Fatalf("set Select(%d) = (%d, %v), want (%d, true)", i, got, ok, k)
		}
	}
	if _, ok := m.SelectKey(len(sorted)); ok {
		t.Error("map SelectKey(size) should be absent")
	}
	if _, ok := set.Select(-1); ok {
		t.Error("set Select(-1) should be absent")
	}
}
