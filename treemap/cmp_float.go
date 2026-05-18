package treemap

import "math"

// cmpFloat32 returns -1/0/1 for a vs b. NaN is handled via bit-pattern
// tiebreak so the ordering is consistent with the hashmap's bit-identity
// equality. Without this, `a < b` and `a > b` are both false for NaN,
// which silently collapses every NaN key onto the existing root in a
// red-black tree.
func cmpFloat32(a, b float32) int {
	if math.IsNaN(float64(a)) || math.IsNaN(float64(b)) {
		ab, bb := math.Float32bits(a), math.Float32bits(b)
		switch {
		case ab < bb:
			return -1
		case ab > bb:
			return 1
		default:
			return 0
		}
	}
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	// Numerically equal: tiebreak by bit pattern (keeps +0/-0 distinct).
	ab, bb := math.Float32bits(a), math.Float32bits(b)
	switch {
	case ab < bb:
		return -1
	case ab > bb:
		return 1
	default:
		return 0
	}
}

func cmpFloat64(a, b float64) int {
	if math.IsNaN(a) || math.IsNaN(b) {
		ab, bb := math.Float64bits(a), math.Float64bits(b)
		switch {
		case ab < bb:
			return -1
		case ab > bb:
			return 1
		default:
			return 0
		}
	}
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	ab, bb := math.Float64bits(a), math.Float64bits(b)
	switch {
	case ab < bb:
		return -1
	case ab > bb:
		return 1
	default:
		return 0
	}
}
