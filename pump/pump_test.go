package pump

import "testing"

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
