package seq

import (
	"iter"
	"slices"
	"strconv"
	"testing"
)

func TestMapAndChainResumes(t *testing.T) {
	// Map changes type (int→string), then a method resumes the chain.
	got := Map(Of(1, 2, 3, 4), func(n int) string { return strconv.Itoa(n * n) }).
		Filter(func(s string) bool { return len(s) == 1 }).
		ToSlice()
	if want := []string{"1", "4", "9"}; !slices.Equal(got, want) {
		t.Errorf("Map+Filter = %v, want %v", got, want)
	}
}

func TestMapWhere(t *testing.T) {
	got := MapWhere(Of(1, 2, 3, 4),
		func(n int) bool { return n%2 == 0 },
		func(n int) int { return n * 10 }).ToSlice()
	if want := []int{20, 40}; !slices.Equal(got, want) {
		t.Errorf("MapWhere = %v, want %v", got, want)
	}
}

func TestFlatMap(t *testing.T) {
	got := FlatMap(Of(1, 2, 3), func(n int) iter.Seq[int] {
		return Range(0, n).Std()
	}).ToSlice()
	if want := []int{0, 0, 1, 0, 1, 2}; !slices.Equal(got, want) {
		t.Errorf("FlatMap = %v, want %v", got, want)
	}
}

func TestDistinct(t *testing.T) {
	got := Distinct(Of(1, 2, 2, 3, 1, 4, 3)).ToSlice()
	if want := []int{1, 2, 3, 4}; !slices.Equal(got, want) {
		t.Errorf("Distinct = %v, want %v", got, want)
	}
	// DistinctBy on absolute value, first-seen order.
	got = DistinctBy(Of(1, -1, 2, -2, 3), func(n int) int {
		if n < 0 {
			return -n
		}
		return n
	}).ToSlice()
	if want := []int{1, 2, 3}; !slices.Equal(got, want) {
		t.Errorf("DistinctBy = %v, want %v", got, want)
	}
}

func TestDistinctIsLazy(t *testing.T) {
	// Distinct must short-circuit: over the naturals, First stops immediately.
	nats := Iterate(0, func(n int) int { return n + 1 })
	if v, ok := Distinct(nats).First(); !ok || v != 0 {
		t.Errorf("Distinct(nats).First = %d,%v", v, ok)
	}
}

func TestFoldAndScan(t *testing.T) {
	// Fold to a different type: concatenate ints into a string.
	s := Fold(Of(1, 2, 3), "", func(acc string, n int) string {
		return acc + strconv.Itoa(n)
	})
	if s != "123" {
		t.Errorf("Fold = %q, want \"123\"", s)
	}
	// Scan yields the initial then each running sum.
	got := Scan(Of(1, 2, 3), 0, func(a, b int) int { return a + b }).ToSlice()
	if want := []int{0, 1, 3, 6}; !slices.Equal(got, want) {
		t.Errorf("Scan = %v, want %v", got, want)
	}
}

func TestPartition(t *testing.T) {
	evens, odds := Partition(Of(1, 2, 3, 4, 5), func(n int) bool { return n%2 == 0 })
	if got := evens.ToSlice(); !slices.Equal(got, []int{2, 4}) {
		t.Errorf("Partition matching = %v, want [2 4]", got)
	}
	if got := odds.ToSlice(); !slices.Equal(got, []int{1, 3, 5}) {
		t.Errorf("Partition rest = %v, want [1 3 5]", got)
	}
}

func TestSumMinMax(t *testing.T) {
	s := Of(3, 1, 4, 1, 5)
	if sum := Sum(s); sum != 14 {
		t.Errorf("Sum = %d", sum)
	}
	if v, ok := Min(s); !ok || v != 1 {
		t.Errorf("Min = %d,%v", v, ok)
	}
	if v, ok := Max(s); !ok || v != 5 {
		t.Errorf("Max = %d,%v", v, ok)
	}
	if _, ok := Min(Of[int]()); ok {
		t.Error("Min of empty should be false")
	}
	if sum := Sum(Of[float64]()); sum != 0 {
		t.Errorf("Sum of empty = %v, want 0", sum)
	}
}
