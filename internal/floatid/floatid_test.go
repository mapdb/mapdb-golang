package floatid

import (
	"math"
	"reflect"
	"testing"
)

type celsius float64
type single float32

func TestKind(t *testing.T) {
	cases := []struct {
		name string
		got  reflect.Kind
		want reflect.Kind
	}{
		{"float32", Kind[float32](), reflect.Float32},
		{"float64", Kind[float64](), reflect.Float64},
		{"named float64", Kind[celsius](), reflect.Float64},
		{"named float32", Kind[single](), reflect.Float32},
		{"int", Kind[int](), reflect.Invalid},
		{"string", Kind[string](), reflect.Invalid},
		{"struct with float", Kind[struct{ F float64 }](), reflect.Invalid},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s: Kind = %v, want %v", c.name, c.got, c.want)
		}
	}
}

// same is the bit-pattern identity the callers build from Kind and Bits.
func same[T comparable](a, b T) bool {
	if k := Kind[T](); k != reflect.Invalid {
		return Bits(k, a) == Bits(k, b)
	}
	return a == b
}

func TestBitsIdentity(t *testing.T) {
	nan := math.Float64frombits(0x7ff8000000000000)
	nanP := math.Float64frombits(0x7ff8000000000001)
	negZero := math.Copysign(0, -1)
	cases := []struct {
		name string
		got  bool
		want bool
	}{
		{"NaN self", same(nan, nan), true},
		{"NaN payloads", same(nan, nanP), false},
		{"NaN sign", same(nan, -nan), false},
		{"signed zeros", same(0.0, negZero), false},
		{"+Inf", same(math.Inf(1), math.Inf(1)), true},
		{"+Inf/-Inf", same(math.Inf(1), math.Inf(-1)), false},
		{"float32 NaN", same(float32(nan), float32(nan)), true},
		{"float32 zeros", same(float32(0), float32(negZero)), false},
		{"named NaN", same(celsius(nan), celsius(nan)), true},
		{"named zeros", same(celsius(0), celsius(negZero)), false},
		{"int", same(3, 3), true},
		{"struct keeps ==", same(struct{ F float64 }{nan}, struct{ F float64 }{nan}), false},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s: identity = %v, want %v", c.name, c.got, c.want)
		}
	}
}
