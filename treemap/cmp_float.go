package treemap

import "math"

// cmpFloat32 / cmpFloat64 implement the IEEE 754 totalOrder construction
// (bit-identical to Rust's f32::total_cmp / f64::total_cmp), as mandated by
// the collection spec (algorithms.md "Float ordering"):
//
//	-NaN < -Inf < negative finite < -0.0 < +0.0 < positive finite < +Inf < +NaN
//
// A naive `<` returns false for any NaN comparison (collapsing every NaN key
// onto the existing root in a red-black tree). A raw unsigned bit compare is
// also wrong: it is intransitive because a negative float's sign bit makes its
// bit pattern sort above a positive NaN, which silently loses keys in the
// tree's binary search. The sign-flip-then-signed-compare trick below is a
// true total order, and keeps +0/-0 and distinct NaN payloads distinct.
func cmpFloat32(a, b float32) int {
	ai := int32(math.Float32bits(a))
	bi := int32(math.Float32bits(b))
	ai ^= int32(uint32(ai>>31) >> 1)
	bi ^= int32(uint32(bi>>31) >> 1)
	switch {
	case ai < bi:
		return -1
	case ai > bi:
		return 1
	default:
		return 0
	}
}

func cmpFloat64(a, b float64) int {
	ai := int64(math.Float64bits(a))
	bi := int64(math.Float64bits(b))
	ai ^= int64(uint64(ai>>63) >> 1)
	bi ^= int64(uint64(bi>>63) >> 1)
	switch {
	case ai < bi:
		return -1
	case ai > bi:
		return 1
	default:
		return 0
	}
}
