package treeset

import (
	"math"
	"testing"
)

// NaN sign/payload ordering is verified natively here (NOT in the shared
// cross-language validation suite): in TypeScript all NaN bit patterns are a
// single ECMAScript language-level NaN, so an f32 NaN's SIGN and PAYLOAD are
// not cross-language-observable. cmpFloat32 (and the Float32 built on
// it) is the production totalOrder comparator phase 3 fixed.

func b32(bits uint32) float32 { return math.Float32frombits(bits) }

// In the float TREE comparator, -NaN (0xffc00000) sorts BELOW -Infinity, and
// distinct positive NaN payloads order ascending (0x7fc00000 < 0x7fc00001).
func TestCmpFloat32_NaNSignAndPayloadOrdering(t *testing.T) {
	negNaN := b32(0xffc00000)
	negInf := float32(math.Inf(-1))
	if cmpFloat32(negNaN, negInf) >= 0 {
		t.Errorf("cmpFloat32(-NaN, -Inf) = %d; want < 0 (-NaN below -Inf)", cmpFloat32(negNaN, negInf))
	}
	if cmpFloat32(negNaN, -math.MaxFloat32) >= 0 {
		t.Errorf("cmpFloat32(-NaN, -MaxFloat32) >= 0; want -NaN at the very bottom")
	}

	p0 := b32(0x7fc00000)
	p1 := b32(0x7fc00001)
	posInf := float32(math.Inf(1))
	if cmpFloat32(p0, p1) >= 0 {
		t.Errorf("cmpFloat32(0x7fc00000, 0x7fc00001) = %d; want < 0 (ascending payload)", cmpFloat32(p0, p1))
	}
	if cmpFloat32(posInf, p0) >= 0 {
		t.Errorf("cmpFloat32(+Inf, +NaN) >= 0; want +NaN above +Inf")
	}
}

// The same total order, observed via the production Float32 in-order
// traversal: -NaN < -Inf < -finite < -0.0 < +0.0 < +finite < +Inf < +NaN, with
// distinct positive NaN payloads as distinct elements ordered ascending.
func TestFloat32_NaNSignAndPayloadTotalOrder(t *testing.T) {
	s := NewFloat32()
	for _, v := range []float32{
		0.0,
		b32(0x7fc00000), // +NaN
		2.0,
		float32(math.Inf(-1)),
		b32(0xffc00000), // -NaN
		-3.0,
		float32(math.Copysign(0, -1)), // -0.0
		float32(math.Inf(1)),
		b32(0x7fc00001), // +NaN, larger payload (distinct element)
	} {
		s.Add(v)
	}
	if s.Len() != 9 {
		t.Fatalf("Size = %d; want 9 (each distinct bit pattern distinct, incl. ±0 and both +NaN payloads)", s.Len())
	}
	got := s.ToSlice()
	want := []uint32{
		0xffc00000, // -NaN (bottom)
		0xff800000, // -Inf
		0xc0400000, // -3.0
		0x80000000, // -0.0
		0x00000000, // +0.0
		0x40000000, // 2.0
		0x7f800000, // +Inf
		0x7fc00000, // +NaN
		0x7fc00001, // +NaN, larger payload (top)
	}
	if len(got) != len(want) {
		t.Fatalf("ToSlice len = %d; want %d", len(got), len(want))
	}
	for i := range want {
		if math.Float32bits(got[i]) != want[i] {
			t.Errorf("position %d: got bits 0x%08x; want 0x%08x", i, math.Float32bits(got[i]), want[i])
		}
	}
}
