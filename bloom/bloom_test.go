package bloom

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func hexOf(bytes []byte) string {
	s := "0x"
	for _, b := range bytes {
		s += fmt.Sprintf("%02x", b)
	}
	return s
}

func mustPanic(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("%s: expected panic, got none", name)
		}
	}()
	fn()
}

// ---- The worked example (spec §"Serialized bit-array form") ----------------

func TestWorkedExampleAdd7(t *testing.T) {
	b := NewBloomWithParams(16, 4)
	if !b.IsEmpty() {
		t.Fatal("fresh filter must be empty")
	}
	b.Add(7)
	if got := b.SetBits(); !reflect.DeepEqual(got, []uint32{0, 2, 7, 9}) {
		t.Fatalf("set_bits = %v, want [0 2 7 9]", got)
	}
	if b.BitCount() != 4 {
		t.Fatalf("bit_count = %d, want 4", b.BitCount())
	}
	if b.IsEmpty() {
		t.Fatal("filter must not be empty after add")
	}
	if got := hexOf(b.ToBytes()); got != "0x8502" {
		t.Fatalf("bytes = %s, want 0x8502", got)
	}
	if !b.MightContain(7) || !b.Contains(7) {
		t.Fatal("contains(7) must be true (no false negative)")
	}
}

// ---- optimal() pinned integer table (spec §"Construction") -----------------

func TestOptimalIntegerTable(t *testing.T) {
	cases := []struct {
		n      uint64
		p      float64
		em, ek uint32
	}{
		{1000, 0.01, 9586, 7},
		{1000, 0.001, 14378, 10},
		{10000, 0.01, 95851, 7},
		{100, 0.1, 480, 3},
		{1, 0.5, 2, 1},
	}
	for _, c := range cases {
		b := Optimal(c.n, c.p)
		if b.MBits() != c.em {
			t.Errorf("Optimal(%d, %g) m = %d, want %d", c.n, c.p, b.MBits(), c.em)
		}
		if b.K() != c.ek {
			t.Errorf("Optimal(%d, %g) k = %d, want %d", c.n, c.p, b.K(), c.ek)
		}
	}
}

func TestOptimalTraps(t *testing.T) {
	mustPanic(t, "n=0", func() { Optimal(0, 0.01) })
	mustPanic(t, "p=0", func() { Optimal(100, 0.0) })
	mustPanic(t, "p<0", func() { Optimal(100, -0.5) })
	mustPanic(t, "p=1", func() { Optimal(100, 1.0) })
	mustPanic(t, "p>1", func() { Optimal(100, 1.5) })
	mustPanic(t, "p=NaN", func() { Optimal(100, math.NaN()) })
	mustPanic(t, "p=+Inf", func() { Optimal(100, math.Inf(1)) })
	mustPanic(t, "p=-Inf", func() { Optimal(100, math.Inf(-1)) })
}

// ---- m = 0 trap; k = 0 vacuous-true ----------------------------------------

func TestMZeroTraps(t *testing.T) {
	mustPanic(t, "m=0", func() { NewBloomWithParams(0, 4) })
}

func TestKZeroVacuousTrue(t *testing.T) {
	b := NewBloomWithParams(16, 0)
	b.Add(5)
	if b.BitCount() != 0 {
		t.Fatalf("k=0 bit_count = %d, want 0", b.BitCount())
	}
	if !b.IsEmpty() {
		t.Fatal("k=0 filter must be empty")
	}
	if got := hexOf(b.ToBytes()); got != "0x0000" {
		t.Fatalf("bytes = %s, want 0x0000", got)
	}
	if !b.MightContain(5) || !b.MightContain(9999) || !b.MightContain(-1) {
		t.Fatal("k=0 might_contain must be vacuously true for everything")
	}
}

// ---- union OR + mismatch trap ----------------------------------------------

func TestUnionIsBitwiseOr(t *testing.T) {
	a := NewBloomWithParams(32, 3)
	a.Add(1)
	a.Add(2)
	c := NewBloomWithParams(32, 3)
	c.Add(3)
	c.Add(4)
	u := a.Union(c)
	if !u.MightContain(1) || !u.MightContain(2) || !u.MightContain(3) || !u.MightContain(4) {
		t.Fatal("union must preserve no-false-negative for both operands")
	}
	for i := range a.words {
		if u.words[i] != a.words[i]|c.words[i] {
			t.Fatalf("word %d: union not bitwise OR", i)
		}
	}
}

func TestUnionMismatchTraps(t *testing.T) {
	mustPanic(t, "m mismatch", func() {
		NewBloomWithParams(16, 4).Union(NewBloomWithParams(32, 4))
	})
	mustPanic(t, "k mismatch", func() {
		NewBloomWithParams(16, 4).Union(NewBloomWithParams(16, 3))
	})
}

// ---- idempotent / order-independent ----------------------------------------

func TestAddIsIdempotent(t *testing.T) {
	once := NewBloomWithParams(64, 5)
	once.Add(7)
	twice := NewBloomWithParams(64, 5)
	twice.Add(7)
	twice.Add(7)
	if !reflect.DeepEqual(once.ToBytes(), twice.ToBytes()) {
		t.Fatal("add is not idempotent (bytes differ)")
	}
	if once.BitCount() != twice.BitCount() {
		t.Fatal("add is not idempotent (bit_count differs)")
	}
}

func TestAddIsOrderIndependent(t *testing.T) {
	ab := NewBloomWithParams(128, 4)
	ab.Add(11)
	ab.Add(22)
	ab.Add(33)
	ba := NewBloomWithParams(128, 4)
	ba.Add(33)
	ba.Add(11)
	ba.Add(22)
	if !reflect.DeepEqual(ab.ToBytes(), ba.ToBytes()) {
		t.Fatal("add is not order-independent")
	}
}

// ---- signed extremes (reinterpret, not sign-extend) ------------------------

func TestSignedExtremes(t *testing.T) {
	b := NewBloomWithParams(256, 4)
	b.Add(-1)
	b.Add(math.MinInt32)
	if !b.MightContain(-1) || !b.MightContain(math.MinInt32) {
		t.Fatal("signed extremes must report present")
	}
	neg := NewBloomWithParams(256, 4)
	neg.Add(-1)
	pos := NewBloomWithParams(256, 4)
	pos.Add(1)
	if reflect.DeepEqual(neg.ToBytes(), pos.ToBytes()) {
		t.Fatal("-1 (reinterpret 0xffffffff) must differ from 1")
	}
}

// ---- no false negative over a set ------------------------------------------

func TestNoFalseNegativeOverASet(t *testing.T) {
	b := NewBloomWithParams(512, 7)
	var elems []int32
	for v := int32(-50); v < 50; v++ {
		elems = append(elems, v)
	}
	elems = append(elems, math.MinInt32, math.MaxInt32, 0)
	for _, e := range elems {
		b.Add(e)
	}
	for _, e := range elems {
		if !b.MightContain(e) {
			t.Fatalf("false negative for %d", e)
		}
	}
}

// ---- tail bits (m not a multiple of 8) -------------------------------------

func TestTailBitsZeroed(t *testing.T) {
	b := NewBloomWithParams(13, 3)
	b.Add(7)
	b.Add(42)
	bytes := b.ToBytes()
	if len(bytes) != 2 {
		t.Fatalf("ceil(13/8) = 2 bytes, got %d", len(bytes))
	}
	for _, p := range b.SetBits() {
		if p >= 13 {
			t.Fatalf("set bit %d must be < 13", p)
		}
	}
	if got := hexOf(bytes); got != "0x9804" {
		t.Fatalf("bytes = %s, want 0x9804", got)
	}
}

// ---- host-endianness independence of to_bytes ------------------------------

func TestToBytesLSBFirstIndependentOfWordWidth(t *testing.T) {
	b := NewBloomWithParams(16, 1)
	b.setBit(8)
	if got := b.ToBytes(); !reflect.DeepEqual(got, []byte{0x00, 0x01}) {
		t.Fatalf("bit 8 -> %v, want [0x00 0x01]", got)
	}
	c := NewBloomWithParams(16, 1)
	c.setBit(0)
	if got := c.ToBytes(); !reflect.DeepEqual(got, []byte{0x01, 0x00}) {
		t.Fatalf("bit 0 -> %v, want [0x01 0x00]", got)
	}
	d := NewBloomWithParams(16, 1)
	d.setBit(7)
	if got := d.ToBytes(); !reflect.DeepEqual(got, []byte{0x80, 0x00}) {
		t.Fatalf("bit 7 -> %v, want [0x80 0x00]", got)
	}
	e := NewBloomWithParams(200, 1)
	e.setBit(130) // byte 130/8 = 16, bit 130%8 = 2 -> 0x04
	bytes := e.ToBytes()
	if len(bytes) != 25 {
		t.Fatalf("ceil(200/8) = 25 bytes, got %d", len(bytes))
	}
	if bytes[16] != 0x04 {
		t.Fatalf("byte[16] = %#02x, want 0x04", bytes[16])
	}
}

func TestEmptyFilterSerializesToZeroOfFullLength(t *testing.T) {
	b := NewBloomWithParams(16, 4)
	if got := b.ToBytes(); !reflect.DeepEqual(got, []byte{0x00, 0x00}) {
		t.Fatalf("empty -> %v, want [0x00 0x00]", got)
	}
	if b.BitCount() != 0 || !b.IsEmpty() {
		t.Fatal("empty filter must have bit_count 0 and IsEmpty true")
	}
	if b.MightContain(7) {
		t.Fatal("empty k>=1 filter reports absent for everything")
	}
}

// ---- scenario sanity checks (mirror the cross-language oracles) ------------

func TestScenarioSanityValues(t *testing.T) {
	// bloom_multi_add: m=64,k=3, add {10,20,30}.
	mm := NewBloomWithParams(64, 3)
	mm.Add(10)
	mm.Add(20)
	mm.Add(30)
	if got := hexOf(mm.ToBytes()); got != "0x0020002200000298" {
		t.Fatalf("multi_add bytes = %s, want 0x0020002200000298", got)
	}
	if mm.BitCount() != 7 {
		t.Fatalf("multi_add bit_count = %d, want 7", mm.BitCount())
	}

	// bloom_collision_small_m: m=5,k=3, add 0 -> 0x18, bit_count 2.
	cs := NewBloomWithParams(5, 3)
	cs.Add(0)
	if got := hexOf(cs.ToBytes()); got != "0x18" {
		t.Fatalf("collision bytes = %s, want 0x18", got)
	}
	if cs.BitCount() != 2 {
		t.Fatalf("collision bit_count = %d, want 2", cs.BitCount())
	}

	// bloom_false_positive: m=8,k=3, add {1,2,3} -> 0xf5, contains_9 true, contains_4 false.
	fp := NewBloomWithParams(8, 3)
	fp.Add(1)
	fp.Add(2)
	fp.Add(3)
	if got := hexOf(fp.ToBytes()); got != "0xf5" {
		t.Fatalf("false_positive bytes = %s, want 0xf5", got)
	}
	if !fp.MightContain(9) {
		t.Fatal("false_positive contains_9 must be true (deterministic false positive)")
	}
	if fp.MightContain(4) {
		t.Fatal("false_positive contains_4 must be false (genuine absent)")
	}

	// bloom_union: two (m=32,k=3) -> 0xd0614504, union bit_count 10.
	ua := NewBloomWithParams(32, 3)
	ua.Add(1)
	ua.Add(2)
	ub := NewBloomWithParams(32, 3)
	ub.Add(100)
	ub.Add(200)
	u := ua.Union(ub)
	if got := hexOf(u.ToBytes()); got != "0xd0614504" {
		t.Fatalf("union bytes = %s, want 0xd0614504", got)
	}
	if u.BitCount() != 10 {
		t.Fatalf("union bit_count = %d, want 10", u.BitCount())
	}
}
