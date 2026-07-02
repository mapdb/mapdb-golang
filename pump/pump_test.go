package pump

import (
	"math"
	"math/bits"
	"testing"
)

// TestHashCapacityFor_ZeroRehashFormula verifies the presize formula satisfies
// the strict growth rule cap*3 >= 4*n + 1, which is the condition under which the
// generated hash families never rehash mid-load (their needsResize predicate is
// (size+1)*4 >= cap*3, so inserting the n-th element must not trigger it).
func TestHashCapacityFor_ZeroRehashFormula(t *testing.T) {
	for n := 0; n <= 4096; n++ {
		c := HashCapacityFor(n)
		if n == 0 {
			if c != 0 {
				t.Fatalf("HashCapacityFor(0) = %d, want 0", c)
			}
			continue
		}
		// power of two
		if c&(c-1) != 0 {
			t.Fatalf("HashCapacityFor(%d) = %d is not a power of two", n, c)
		}
		// strict growth rule: cap*3 >= 4n + 1
		if c*3 < 4*n+1 {
			t.Fatalf("HashCapacityFor(%d) = %d violates cap*3 >= 4n+1 (%d < %d)", n, c, c*3, 4*n+1)
		}
		// minimality: cap/2 must NOT satisfy the rule (we presize tightly)
		if half := c / 2; half >= 1 && half*3 >= 4*n+1 {
			t.Fatalf("HashCapacityFor(%d) = %d not minimal: %d would also fit", n, c, half)
		}
	}
}

// TestHashCapacityFor_PowersOf3x2k spot-checks the n = 3*2^k boundary where a
// naive ceil(n/0.75) presize would land exactly on a power of two and still
// rehash. The correct formula must give one more doubling there.
func TestHashCapacityFor_PowersOf3x2k(t *testing.T) {
	for _, n := range []int{3, 6, 12, 24, 48, 96} {
		c := HashCapacityFor(n)
		if c*3 < 4*n+1 {
			t.Fatalf("n=%d: cap=%d rehashes (cap*3=%d < 4n+1=%d)", n, c, c*3, 4*n+1)
		}
	}
}

// TestHashCapacityFor_LargeBoundaryNoOverflow exercises n large enough that the
// old int-arithmetic computation of 4*(n/3) wrapped past MaxInt and returned a
// negative / tiny capacity. The capacity must stay a positive power of two that
// satisfies the strict growth rule, with the intermediate computed in wider
// arithmetic.
func TestHashCapacityFor_LargeBoundaryNoOverflow(t *testing.T) {
	// Largest n for which a power-of-two capacity still fits in a positive int:
	// required = floor(4n/3)+1 must be <= 1<<(UintSize-2).
	maxPow2 := 1 << (bits.UintSize - 2)
	// pick an n just below the limit where 4*(n/3) would overflow int but the
	// result is still representable.
	n := (maxPow2/4)*3 - 8 // comfortably below the panic threshold, well above MaxInt/4
	if n <= math.MaxInt32 {
		// 32-bit platform: still ensure a value above MaxInt/4 is handled.
		n = (math.MaxInt32 / 4) * 3
	}
	c := HashCapacityFor(n)
	if c <= 0 {
		t.Fatalf("HashCapacityFor(%d) = %d overflowed to non-positive", n, c)
	}
	if c&(c-1) != 0 {
		t.Fatalf("HashCapacityFor(%d) = %d is not a power of two", n, c)
	}
	// strict growth rule cap*3 >= 4n+1, checked in uint64 to avoid the very
	// overflow we are guarding against in the assertion itself.
	if uint64(c)*3 < uint64(n)*4+1 {
		t.Fatalf("HashCapacityFor(%d) = %d violates cap*3 >= 4n+1", n, c)
	}
}

// TestHashCapacityFor_TooLargePanics verifies n with no representable capacity
// panics rather than silently wrapping.
func TestHashCapacityFor_TooLargePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("HashCapacityFor(MaxInt) should panic, did not")
		}
	}()
	HashCapacityFor(math.MaxInt)
}

// TestRedBlackRedLevel checks the JDK buildFromSorted red-level math for the
// power-of-two-minus-one perfect trees (no red level => returns depth) and a few
// known sizes.
func TestRedBlackRedLevel(t *testing.T) {
	cases := map[int]int{0: 0, 1: 1, 2: 1, 3: 2, 4: 2, 7: 3, 8: 3, 15: 4}
	for n, want := range cases {
		if got := RedBlackRedLevel(n); got != want {
			t.Errorf("RedBlackRedLevel(%d) = %d, want %d", n, got, want)
		}
	}
}
