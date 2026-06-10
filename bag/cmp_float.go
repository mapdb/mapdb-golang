package bag

import "math"

// cmpFloat32 / cmpFloat64 implement the IEEE 754 totalOrder construction
// (bit-identical to Rust's f32::total_cmp / f64::total_cmp), as mandated by
// the collection spec (algorithms.md "Float ordering"). Used by the sorted
// entry list in TreeBag:
//
//	-NaN < -Inf < negative finite < -0.0 < +0.0 < positive finite < +Inf < +NaN
//
// A raw unsigned bit compare for the NaN case is intransitive (a negative
// float's sign bit makes its bit pattern sort above a positive NaN), which
// silently loses entries in the sorted search. The sign-flip-then-signed
// compare below is a true total order.
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
