package treeset

import "math"

// cmpFloat32 / cmpFloat64 return -1/0/1 with NaN handled via bit-pattern
// tiebreak. Naive `<` / `>` return false for NaN comparisons and would
// collapse every NaN entry onto the tree root.
func cmpFloat32(a, b float32) int {
	if math.IsNaN(float64(a)) || math.IsNaN(float64(b)) {
		return cmpU32(math.Float32bits(a), math.Float32bits(b))
	}
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return cmpU32(math.Float32bits(a), math.Float32bits(b))
}

func cmpFloat64(a, b float64) int {
	if math.IsNaN(a) || math.IsNaN(b) {
		return cmpU64(math.Float64bits(a), math.Float64bits(b))
	}
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return cmpU64(math.Float64bits(a), math.Float64bits(b))
}

func cmpU32(a, b uint32) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func cmpU64(a, b uint64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
