package seq

import (
	"slices"
	"testing"
)

func TestZip(t *testing.T) {
	var got [][2]int
	for a, b := range Zip(Of(1, 2, 3), Of(10, 20, 30, 40)) {
		got = append(got, [2]int{a, b})
	}
	want := [][2]int{{1, 10}, {2, 20}, {3, 30}}
	if !slices.Equal(got, want) {
		t.Errorf("Zip = %v, want %v", got, want)
	}
	// Zip stops when the shorter (a) runs dry even though b is infinite.
	nats := Iterate(0, func(n int) int { return n + 1 })
	if n := Zip(Of("x", "y"), nats).Keys().Count(); n != 2 {
		t.Errorf("Zip against infinite b = %d, want 2", n)
	}
}

func TestZipReRunnable(t *testing.T) {
	z := Zip(Of(1, 2), Of("a", "b"))
	for r := 0; r < 2; r++ {
		if got := z.Keys().ToSlice(); !slices.Equal(got, []int{1, 2}) {
			t.Errorf("run %d: %v", r, got)
		}
	}
}

func TestPairwise(t *testing.T) {
	var got [][2]int
	for a, b := range Pairwise(Of(1, 2, 3, 4)) {
		got = append(got, [2]int{a, b})
	}
	want := [][2]int{{1, 2}, {2, 3}, {3, 4}}
	if !slices.Equal(got, want) {
		t.Errorf("Pairwise = %v, want %v", got, want)
	}
	if n := Pairwise(Of(1)).Keys().Count(); n != 0 {
		t.Errorf("Pairwise of one element = %d, want 0", n)
	}
}

func TestWindow(t *testing.T) {
	var got [][]int
	for w := range Window(Of(1, 2, 3, 4), 2) {
		got = append(got, w)
	}
	if len(got) != 3 || !slices.Equal(got[0], []int{1, 2}) ||
		!slices.Equal(got[1], []int{2, 3}) || !slices.Equal(got[2], []int{3, 4}) {
		t.Errorf("Window(2) = %v", got)
	}
	// shorter than the window yields nothing
	var none int
	for range Window(Of(1, 2), 3) {
		none++
	}
	if none != 0 {
		t.Errorf("Window larger than source yielded %d", none)
	}
}

func TestWindowOutputsDoNotAlias(t *testing.T) {
	var got [][]int
	for w := range Window(Of(1, 2, 3, 4), 2) {
		got = append(got, w)
	}
	// Mutating an earlier retained window must not disturb a later one.
	got[0][0] = 999
	if !slices.Equal(got[1], []int{2, 3}) {
		t.Errorf("windows alias the internal buffer: got[1]=%v", got[1])
	}
}

func TestZipEarlyBreakStops(t *testing.T) {
	// Zip against an infinite b, taken to 1: the Pull must stop cleanly and the
	// range must short-circuit rather than spin.
	nats := Iterate(0, func(n int) int { return n + 1 })
	n := 0
	for range Zip(Of("a", "b", "c"), nats).Keys().Take(1).Std() {
		n++
	}
	if n != 1 {
		t.Errorf("Zip early break yielded %d, want 1", n)
	}
}

func TestWindowNonPositivePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Window(0) did not panic")
		}
	}()
	Window(Of(1), 0)
}

func TestCountBySumByAggregateBy(t *testing.T) {
	words := Of("apple", "ant", "bee", "bear", "cat")
	byFirst := func(s string) byte { return s[0] }

	cb := CountBy(words, byFirst)
	if cb['a'] != 2 || cb['b'] != 2 || cb['c'] != 1 {
		t.Errorf("CountBy = %v", cb)
	}

	sb := SumBy(words, byFirst, func(s string) int { return len(s) })
	if sb['a'] != 8 || sb['b'] != 7 || sb['c'] != 3 { // apple+ant=8, bee+bear=7, cat=3
		t.Errorf("SumBy = %v", sb)
	}

	agg := AggregateBy(words, byFirst,
		func() []string { return nil },
		func(acc []string, s string) []string { return append(acc, s) })
	if !slices.Equal(agg['a'], []string{"apple", "ant"}) {
		t.Errorf("AggregateBy['a'] = %v", agg['a'])
	}
}

func TestAverage(t *testing.T) {
	if avg, ok := Average(Of(1, 2, 3, 4)); !ok || avg != 2.5 {
		t.Errorf("Average = %v,%v", avg, ok)
	}
	if _, ok := Average(Of[int]()); ok {
		t.Error("Average of empty should be false")
	}
}

func TestGroupByAndInto(t *testing.T) {
	words := Of("apple", "ant", "bee")
	mm := GroupBy(words, func(s string) byte { return s[0] })
	if got := mm.Get('a'); !slices.Equal(got, []string{"apple", "ant"}) {
		t.Errorf("GroupBy['a'] = %v", got)
	}
	if mm.Len() != 3 {
		t.Errorf("GroupBy total = %d", mm.Len())
	}
}

func TestTopK(t *testing.T) {
	cmp := func(a, b int) int { return a - b }
	got := TopK(Of(3, 1, 4, 1, 5, 9, 2, 6), 3, cmp)
	if !slices.Equal(got, []int{9, 6, 5}) {
		t.Errorf("TopK(3) = %v, want [9 6 5]", got)
	}
	// k larger than the sequence returns all, still descending
	got = TopK(Of(2, 1, 3), 10, cmp)
	if !slices.Equal(got, []int{3, 2, 1}) {
		t.Errorf("TopK overshoot = %v", got)
	}
	if TopK(Of(1, 2), 0, cmp) != nil {
		t.Error("TopK(0) should be nil")
	}
	// works lazily over an infinite source when bounded upstream
	nats := Iterate(1, func(n int) int { return n + 1 })
	got = TopK(nats.Take(1000), 2, cmp)
	if !slices.Equal(got, []int{1000, 999}) {
		t.Errorf("TopK over bounded-infinite = %v", got)
	}
}
