package seq

import (
	"slices"
	"testing"
)

func TestFilterFilterNot(t *testing.T) {
	even := func(n int) bool { return n%2 == 0 }
	got := slices.Collect(Of(1, 2, 3, 4, 5, 6).Filter(even).Std())
	if want := []int{2, 4, 6}; !slices.Equal(got, want) {
		t.Errorf("Filter = %v, want %v", got, want)
	}
	got = slices.Collect(Of(1, 2, 3, 4, 5, 6).FilterNot(even).Std())
	if want := []int{1, 3, 5}; !slices.Equal(got, want) {
		t.Errorf("FilterNot = %v, want %v", got, want)
	}
}

func TestChainingIsReRunnable(t *testing.T) {
	s := Of(1, 2, 3, 4).Filter(func(n int) bool { return n > 1 }).Drop(1)
	for i := 0; i < 2; i++ {
		if got := s.ToSlice(); !slices.Equal(got, []int{3, 4}) {
			t.Errorf("run %d: %v, want [3 4]", i, got)
		}
	}
}

// TestTakeZeroDoesNotPull is the confirmed stream.Take bug: Take(0) must yield
// nothing AND pull nothing, so it terminates over an infinite source.
func TestTakeZeroDoesNotPull(t *testing.T) {
	pulls := 0
	src := Generate(func() int { pulls++; return 1 })
	got := src.Take(0).ToSlice()
	if len(got) != 0 {
		t.Errorf("Take(0) = %v, want empty", got)
	}
	if pulls != 0 {
		t.Errorf("Take(0) pulled %d elements, want 0", pulls)
	}
}

func TestTakeExactPullCount(t *testing.T) {
	pulls := 0
	src := Generate(func() int { pulls++; return pulls })
	got := src.Take(3).ToSlice()
	if !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("Take(3) = %v, want [1 2 3]", got)
	}
	if pulls != 3 {
		t.Errorf("Take(3) pulled %d elements, want exactly 3", pulls)
	}
}

func TestTakeWhileDropWhile(t *testing.T) {
	lt3 := func(n int) bool { return n < 3 }
	if got := Of(1, 2, 3, 1).TakeWhile(lt3).ToSlice(); !slices.Equal(got, []int{1, 2}) {
		t.Errorf("TakeWhile = %v, want [1 2]", got)
	}
	if got := Of(1, 2, 3, 1).DropWhile(lt3).ToSlice(); !slices.Equal(got, []int{3, 1}) {
		t.Errorf("DropWhile = %v, want [3 1]", got)
	}
}

func TestPeekObservesWithoutConsuming(t *testing.T) {
	var seen []int
	got := Of(1, 2, 3).Peek(func(n int) { seen = append(seen, n) }).ToSlice()
	if !slices.Equal(got, []int{1, 2, 3}) || !slices.Equal(seen, []int{1, 2, 3}) {
		t.Errorf("Peek: got=%v seen=%v, want both [1 2 3]", got, seen)
	}
}

func TestConcat(t *testing.T) {
	got := Of(1, 2).Concat(Of(3), Of(4, 5)).ToSlice()
	if !slices.Equal(got, []int{1, 2, 3, 4, 5}) {
		t.Errorf("Concat = %v, want [1 2 3 4 5]", got)
	}
}

func TestCycleWithTake(t *testing.T) {
	got := Of(1, 2).Cycle().Take(5).ToSlice()
	if !slices.Equal(got, []int{1, 2, 1, 2, 1}) {
		t.Errorf("Cycle.Take(5) = %v, want [1 2 1 2 1]", got)
	}
	// Cycle over an empty source must terminate, not spin forever.
	if got := Of[int]().Cycle().Take(3).ToSlice(); len(got) != 0 {
		t.Errorf("empty Cycle = %v, want empty", got)
	}
}

func TestChunk(t *testing.T) {
	var chunks [][]int
	for c := range Of(1, 2, 3, 4, 5).Chunk(2) {
		chunks = append(chunks, c)
	}
	if len(chunks) != 3 || !slices.Equal(chunks[0], []int{1, 2}) ||
		!slices.Equal(chunks[1], []int{3, 4}) || !slices.Equal(chunks[2], []int{5}) {
		t.Errorf("Chunk(2) = %v, want [[1 2] [3 4] [5]]", chunks)
	}
}

func TestChunkNonPositivePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Chunk(0) did not panic")
		}
	}()
	Of(1).Chunk(0)
}

func TestSources(t *testing.T) {
	if got := Range(2, 6).ToSlice(); !slices.Equal(got, []int{2, 3, 4, 5}) {
		t.Errorf("Range(2,6) = %v", got)
	}
	if got := Range(6, 2).ToSlice(); len(got) != 0 {
		t.Errorf("Range(6,2) = %v, want empty", got)
	}
	if got := RangeStep(0, 10, 3).ToSlice(); !slices.Equal(got, []int{0, 3, 6, 9}) {
		t.Errorf("RangeStep(0,10,3) = %v", got)
	}
	if got := RangeStep(10, 0, -3).ToSlice(); !slices.Equal(got, []int{10, 7, 4, 1}) {
		t.Errorf("RangeStep(10,0,-3) = %v", got)
	}
	if got := Iterate(1, func(n int) int { return n * 2 }).Take(4).ToSlice(); !slices.Equal(got, []int{1, 2, 4, 8}) {
		t.Errorf("Iterate = %v", got)
	}
	if got := RepeatN("x", 3).ToSlice(); !slices.Equal(got, []string{"x", "x", "x"}) {
		t.Errorf("RepeatN = %v", got)
	}
}

func TestRangeStepZeroPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("RangeStep step 0 did not panic")
		}
	}()
	RangeStep(0, 10, 0)
}

func TestTerminals(t *testing.T) {
	s := Of(3, 1, 4, 1, 5)
	if n := s.Count(); n != 5 {
		t.Errorf("Count = %d", n)
	}
	if n := s.CountFunc(func(n int) bool { return n == 1 }); n != 2 {
		t.Errorf("CountFunc = %d", n)
	}
	if v, ok := s.First(); !ok || v != 3 {
		t.Errorf("First = %d,%v", v, ok)
	}
	if v, ok := s.Last(); !ok || v != 5 {
		t.Errorf("Last = %d,%v", v, ok)
	}
	if v, ok := s.Find(func(n int) bool { return n > 3 }); !ok || v != 4 {
		t.Errorf("Find = %d,%v", v, ok)
	}
	if sum := s.Reduce(0, func(a, b int) int { return a + b }); sum != 14 {
		t.Errorf("Reduce sum = %d", sum)
	}
	cmp := func(a, b int) int { return a - b }
	if v, ok := s.MinFunc(cmp); !ok || v != 1 {
		t.Errorf("MinFunc = %d,%v", v, ok)
	}
	if v, ok := s.MaxFunc(cmp); !ok || v != 5 {
		t.Errorf("MaxFunc = %d,%v", v, ok)
	}
}

func TestPredicateShortCircuitOnInfinite(t *testing.T) {
	nats := Iterate(0, func(n int) int { return n + 1 })
	if !nats.Any(func(n int) bool { return n == 100 }) {
		t.Error("Any should find 100 in the naturals")
	}
	if nats.All(func(n int) bool { return n < 5 }) {
		t.Error("All should short-circuit false")
	}
	if v, ok := nats.Find(func(n int) bool { return n == 7 }); !ok || v != 7 {
		t.Errorf("Find on infinite = %d,%v", v, ok)
	}
}

func TestEmptyTerminals(t *testing.T) {
	e := Of[int]()
	if _, ok := e.First(); ok {
		t.Error("First on empty should be false")
	}
	if !e.All(func(int) bool { return false }) {
		t.Error("All on empty should be vacuously true")
	}
	if !e.None(func(int) bool { return true }) {
		t.Error("None on empty should be true")
	}
	if e.Any(func(int) bool { return true }) {
		t.Error("Any on empty should be false")
	}
}
