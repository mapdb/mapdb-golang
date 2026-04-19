package stream

import (
	"iter"
	"slices"
	"testing"
)

func seqOf[V any](values ...V) iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, v := range values {
			if !yield(v) {
				return
			}
		}
	}
}

func TestFilter(t *testing.T) {
	evens := ToSlice(Filter(seqOf(1, 2, 3, 4, 5), func(v int) bool { return v%2 == 0 }))
	if !slices.Equal(evens, []int{2, 4}) {
		t.Errorf("Filter evens = %v, want [2, 4]", evens)
	}
}

func TestMap(t *testing.T) {
	doubled := ToSlice(Map(seqOf(1, 2, 3), func(v int) int { return v * 2 }))
	if !slices.Equal(doubled, []int{2, 4, 6}) {
		t.Errorf("Map doubled = %v, want [2, 4, 6]", doubled)
	}
}

func TestFlatMap(t *testing.T) {
	result := ToSlice(FlatMap(seqOf(1, 2, 3), func(v int) iter.Seq[int] {
		return seqOf(v, v*10)
	}))
	if !slices.Equal(result, []int{1, 10, 2, 20, 3, 30}) {
		t.Errorf("FlatMap = %v, want [1, 10, 2, 20, 3, 30]", result)
	}
}

func TestTakeDrop(t *testing.T) {
	taken := ToSlice(Take(seqOf(1, 2, 3, 4, 5), 3))
	if !slices.Equal(taken, []int{1, 2, 3}) {
		t.Errorf("Take(3) = %v, want [1, 2, 3]", taken)
	}

	dropped := ToSlice(Drop(seqOf(1, 2, 3, 4, 5), 2))
	if !slices.Equal(dropped, []int{3, 4, 5}) {
		t.Errorf("Drop(2) = %v, want [3, 4, 5]", dropped)
	}
}

func TestReduce(t *testing.T) {
	sum := Reduce(seqOf(1, 2, 3, 4, 5), 0, func(a, b int) int { return a + b })
	if sum != 15 {
		t.Errorf("Reduce sum = %d, want 15", sum)
	}
}

func TestCount(t *testing.T) {
	n := Count(seqOf(1, 2, 3))
	if n != 3 {
		t.Errorf("Count = %d, want 3", n)
	}
}

func TestGroupBy(t *testing.T) {
	groups := GroupBy(seqOf(1, 2, 3, 4, 5, 6), func(v int) string {
		if v%2 == 0 {
			return "even"
		}
		return "odd"
	})
	if len(groups["even"]) != 3 || len(groups["odd"]) != 3 {
		t.Errorf("GroupBy = %v", groups)
	}
}

func TestAnyAllNone(t *testing.T) {
	seq := seqOf(1, 2, 3, 4, 5)
	isEven := func(v int) bool { return v%2 == 0 }

	if !Any(seq, isEven) {
		t.Error("Any(even) should be true")
	}
	if All(seq, isEven) {
		t.Error("All(even) should be false")
	}
	if None(seq, isEven) {
		t.Error("None(even) should be false")
	}
}

func TestContains(t *testing.T) {
	if !Contains(seqOf(1, 2, 3), 2) {
		t.Error("Contains(2) should be true")
	}
	if Contains(seqOf(1, 2, 3), 99) {
		t.Error("Contains(99) should be false")
	}
}

func TestFirstLast(t *testing.T) {
	if v, ok := First(seqOf(10, 20, 30)); !ok || v != 10 {
		t.Errorf("First = (%d, %v), want (10, true)", v, ok)
	}
	if v, ok := Last(seqOf(10, 20, 30)); !ok || v != 30 {
		t.Errorf("Last = (%d, %v), want (30, true)", v, ok)
	}
}

func TestDistinct(t *testing.T) {
	result := ToSlice(Distinct(seqOf(1, 2, 2, 3, 1, 3)))
	if !slices.Equal(result, []int{1, 2, 3}) {
		t.Errorf("Distinct = %v, want [1, 2, 3]", result)
	}
}

func TestChain(t *testing.T) {
	result := ToSlice(Chain(seqOf(1, 2), seqOf(3, 4), seqOf(5)))
	if !slices.Equal(result, []int{1, 2, 3, 4, 5}) {
		t.Errorf("Chain = %v, want [1, 2, 3, 4, 5]", result)
	}
}

func TestLazyPipeline(t *testing.T) {
	// Compose: filter evens, double them, take first 3
	result := ToSlice(
		Take(
			Map(
				Filter(seqOf(1, 2, 3, 4, 5, 6, 7, 8, 9, 10),
					func(v int) bool { return v%2 == 0 }),
				func(v int) int { return v * 2 }),
			3))
	if !slices.Equal(result, []int{4, 8, 12}) {
		t.Errorf("Lazy pipeline = %v, want [4, 8, 12]", result)
	}
}

func TestSumMinMax(t *testing.T) {
	s := Sum(seqOf(1, 2, 3, 4, 5))
	if s != 15 {
		t.Errorf("Sum = %d, want 15", s)
	}
	min, ok := Min(seqOf(3, 1, 4, 1, 5))
	if !ok || min != 1 {
		t.Errorf("Min = (%d, %v), want (1, true)", min, ok)
	}
	max, ok := Max(seqOf(3, 1, 4, 1, 5))
	if !ok || max != 5 {
		t.Errorf("Max = (%d, %v), want (5, true)", max, ok)
	}
}
