package stream

import (
	"slices"
	"testing"
)

// countingSeq yields 0..n-1, incrementing *pulled once per element actually
// requested from the source. It lets a test observe over-consumption.
func countingSeq(n int, pulled *int) func(yield func(int) bool) {
	return func(yield func(int) bool) {
		for i := 0; i < n; i++ {
			*pulled++
			if !yield(i) {
				return
			}
		}
	}
}

// Take(n<=0) must yield nothing AND pull nothing from the source (it used to
// pull one element before returning — observable for a stateful/pull source).
func TestTakeNonPositiveDoesNotConsume(t *testing.T) {
	for _, n := range []int{0, -1, -100} {
		pulled := 0
		got := ToSlice(Take(countingSeq(5, &pulled), n))
		if len(got) != 0 {
			t.Errorf("Take(%d) = %v, want empty", n, got)
		}
		if pulled != 0 {
			t.Errorf("Take(%d) pulled %d elements from source, want 0", n, pulled)
		}
	}
}

// Take(n) must pull exactly n elements from the source, not n+1.
func TestTakeExactPullCount(t *testing.T) {
	pulled := 0
	got := ToSlice(Take(countingSeq(100, &pulled), 3))
	if !slices.Equal(got, []int{0, 1, 2}) {
		t.Errorf("Take(3) = %v, want [0 1 2]", got)
	}
	if pulled != 3 {
		t.Errorf("Take(3) pulled %d elements, want exactly 3", pulled)
	}
}

// Chunk(n<=0) is a programmer error and must panic eagerly (n<0 would panic in
// make with a negative capacity; n==0 would build one unbounded chunk).
func TestChunkNonPositivePanics(t *testing.T) {
	for _, n := range []int{0, -1} {
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Chunk(%d) did not panic", n)
				}
			}()
			Chunk(seqOf(1, 2, 3), n)
		}()
	}
}
