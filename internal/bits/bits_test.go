package bits

import "testing"

func TestNextPowerOfTwo(t *testing.T) {
	cases := []struct{ in, want int }{
		{-100, 16}, {-1, 16}, {0, 16}, // floor for non-positive
		{1, 1}, {2, 2}, {3, 4}, {5, 8}, {8, 8}, {9, 16},
		{16, 16}, {17, 32}, {1000, 1024}, {1024, 1024}, {1025, 2048},
		{1 << 29, 1 << 29}, {(1 << 29) + 1, 1 << 30}, // stays within int32 for 32-bit portability
	}
	for _, c := range cases {
		if got := NextPowerOfTwo(c.in); got != c.want {
			t.Errorf("NextPowerOfTwo(%d) = %d, want %d", c.in, got, c.want)
		}
	}
}
