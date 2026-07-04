package seq

import (
	"slices"
	"testing"
)

func TestSortedReversed(t *testing.T) {
	cmp := func(a, b int) int { return a - b }
	src := Of(3, 1, 4, 1, 5)
	if got := src.Sorted(cmp).ToSlice(); !slices.Equal(got, []int{1, 1, 3, 4, 5}) {
		t.Errorf("Sorted = %v", got)
	}
	if got := src.Reversed().ToSlice(); !slices.Equal(got, []int{5, 1, 4, 1, 3}) {
		t.Errorf("Reversed = %v", got)
	}
	// eager result is re-runnable and does not re-sort/re-reverse the original
	sorted := src.Sorted(cmp)
	for r := 0; r < 2; r++ {
		if got := sorted.ToSlice(); !slices.Equal(got, []int{1, 1, 3, 4, 5}) {
			t.Errorf("Sorted re-run %d = %v", r, got)
		}
	}
}

func TestOnEmpty(t *testing.T) {
	fb := Of(-1, -2)
	if got := Of(1, 2).OnEmpty(fb).ToSlice(); !slices.Equal(got, []int{1, 2}) {
		t.Errorf("OnEmpty non-empty = %v", got)
	}
	if got := Of[int]().OnEmpty(fb).ToSlice(); !slices.Equal(got, []int{-1, -2}) {
		t.Errorf("OnEmpty empty = %v", got)
	}
}

func TestCacheReRunsSingleShot(t *testing.T) {
	// A single-shot source: a Pull-backed seq that yields 1,2,3 exactly once.
	runs := 0
	oneShot := func() Seq[int] {
		yielded := false
		return func(yield func(int) bool) {
			if yielded {
				return // already consumed — single-shot
			}
			yielded = true
			runs++
			for _, v := range []int{1, 2, 3} {
				if !yield(v) {
					return
				}
			}
		}
	}()
	cached := oneShot.Cache()
	first := cached.ToSlice()
	second := cached.ToSlice()
	if !slices.Equal(first, []int{1, 2, 3}) || !slices.Equal(second, []int{1, 2, 3}) {
		t.Errorf("Cache: first=%v second=%v", first, second)
	}
	if runs != 1 {
		t.Errorf("Cache ran the source %d times, want 1", runs)
	}
}

func TestCacheEarlyBreakDoesNotPoison(t *testing.T) {
	src := Of(1, 2, 3, 4, 5).Cache()
	// break early on the first pass — nothing should be cached
	got := src.Take(2).ToSlice()
	if !slices.Equal(got, []int{1, 2}) {
		t.Errorf("Cache early = %v", got)
	}
	// full pass afterwards still sees everything (source here is re-runnable)
	if got := src.ToSlice(); !slices.Equal(got, []int{1, 2, 3, 4, 5}) {
		t.Errorf("Cache full after early break = %v", got)
	}
}

func TestMergeSorted(t *testing.T) {
	cmp := func(a, b int) int { return a - b }
	got := MergeSorted(cmp, Of(1, 4, 7), Of(2, 3, 8), Of(5, 6)).ToSlice()
	if !slices.Equal(got, []int{1, 2, 3, 4, 5, 6, 7, 8}) {
		t.Errorf("MergeSorted = %v", got)
	}
	// re-runnable
	m := MergeSorted(cmp, Of(1, 3), Of(2, 4))
	for r := 0; r < 2; r++ {
		if got := m.ToSlice(); !slices.Equal(got, []int{1, 2, 3, 4}) {
			t.Errorf("MergeSorted re-run %d = %v", r, got)
		}
	}
	// short-circuits: only pull what's needed from each input
	if got := MergeSorted(cmp, Range(0, 1000000), Range(0, 1000000)).Take(4).ToSlice(); !slices.Equal(got, []int{0, 0, 1, 1}) {
		t.Errorf("MergeSorted.Take = %v", got)
	}
}

func TestMergeSortedDistinct(t *testing.T) {
	cmp := func(a, b int) int { return a - b }
	got := MergeSortedDistinct(cmp, Of(1, 3, 5), Of(2, 3, 6), Of(1, 6)).ToSlice()
	if !slices.Equal(got, []int{1, 2, 3, 5, 6}) {
		t.Errorf("MergeSortedDistinct = %v", got)
	}
}

func TestMustBeSorted(t *testing.T) {
	cmp := func(a, b int) int { return a - b }
	if got := MustBeSorted(Of(1, 2, 2, 3), cmp).ToSlice(); !slices.Equal(got, []int{1, 2, 2, 3}) {
		t.Errorf("MustBeSorted passthrough = %v", got)
	}
	defer func() {
		if recover() == nil {
			t.Error("MustBeSorted did not panic on unsorted input")
		}
	}()
	MustBeSorted(Of(1, 3, 2), cmp).ToSlice()
}
