package bag

import "math"

// cmpFloat32 / cmpFloat64 return -1/0/1 with NaN handled via bit-pattern
// tiebreak. Used by the sorted entry list in TreeBag.
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
