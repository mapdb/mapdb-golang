// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package hash

import (
	"reflect"
	"testing"
)

// ---- Self-consistency anchors (spec "Self-consistency anchors") -----------

func TestAnchorHash32Zero(t *testing.T) {
	// hash32(0, 0): seed32=0, h=0, every step 0 -> 0. Universal self-check.
	if got := Hash32(0x00000000, 0); got != 0x00000000 {
		t.Fatalf("Hash32(0,0) = %#08x, want 0x00000000", got)
	}
}

func TestAnchorHash64Zero(t *testing.T) {
	if got := Hash64(0, 0); got != 0 {
		t.Fatalf("Hash64(0,0) = %#016x, want 0", got)
	}
	if got := Hash64Hi(0, 0); got != 0x00000000 {
		t.Fatalf("Hash64Hi(0,0) = %#08x, want 0", got)
	}
	if got := Hash64Lo(0, 0); got != 0x00000000 {
		t.Fatalf("Hash64Lo(0,0) = %#08x, want 0", got)
	}
}

func TestHash64AllOnesLogicalShift(t *testing.T) {
	// h>>33 on all-ones is 0x000000007fffffff (top 33 bits zero): a logical,
	// not arithmetic, shift. The value below is the reference output.
	if got := Hash64(0xffffffffffffffff, 0); got != 0x64b5720b4b825f21 {
		t.Fatalf("Hash64(all ones,0) = %#016x, want 0x64b5720b4b825f21", got)
	}
}

func TestSeedFoldIdentity(t *testing.T) {
	// seed32 = seed ^ (seed>>32). 0x00000000ffffffff -> ffffffff^00000000 =
	// ffffffff. 0xffffffff00000000 -> 00000000 ^ ffffffff = ffffffff. Equal.
	if Hash32(0x12345678, 0x00000000ffffffff) != Hash32(0x12345678, 0xffffffff00000000) {
		t.Fatal("seeds folding to the same seed32 must give identical Hash32")
	}
	// Two seeds that fold to DIFFERENT seed32 produce different hashes.
	if Hash32(0x12345678, 0x0000000000000001) == Hash32(0x12345678, 0x0000000000000002) {
		t.Fatal("seeds folding to different seed32 must differ")
	}
	// The high word genuinely affects the fold.
	if Hash32(0x12345678, 0x0000000100000000) == Hash32(0x12345678, 0x0000000000000000) {
		t.Fatal("seed high word must affect the fold")
	}
}

// ---- Per-type encoder pins (spec "Input -> input-word derivation") --------

func TestI32ReinterpretNotSignExtend(t *testing.T) {
	if EncodeInt32Word32(-1) != 0xffffffff {
		t.Fatal("i32(-1) must reinterpret to 0xffffffff")
	}
	if Hash32Int32(-1, 0) != Hash32(0xffffffff, 0) {
		t.Fatal("Hash32Int32(-1) must equal Hash32(0xffffffff)")
	}
	if EncodeInt32Word32(-2147483648) != 0x80000000 {
		t.Fatal("i32 INT_MIN must reinterpret to 0x80000000")
	}
	if EncodeInt32Word32(2147483647) != 0x7fffffff {
		t.Fatal("i32 INT_MAX must reinterpret to 0x7fffffff")
	}
}

func TestI32ZeroExtendForHash64(t *testing.T) {
	// i32(-1) -> u64 0x00000000ffffffff (ZERO-extend, not 0xffffffffffffffff).
	if EncodeInt32Word64(-1) != 0x00000000ffffffff {
		t.Fatal("i32(-1) must zero-extend to 0x00000000ffffffff")
	}
	if Hash64Int32(-1, 0) != Hash64(0x00000000ffffffff, 0) {
		t.Fatal("Hash64Int32(-1) must equal Hash64(0x00000000ffffffff)")
	}
	// The sign-extend trap made observable: zero-extended != all-ones.
	if Hash64Int32(-1, 0) == Hash64(0xffffffffffffffff, 0) {
		t.Fatal("zero-extended i32(-1) must differ from all-ones (sign-extend trap)")
	}
	if EncodeInt32Word64(-2147483648) != 0x0000000080000000 {
		t.Fatal("i32 INT_MIN must zero-extend to 0x0000000080000000")
	}
}

func TestBytesLEFold32(t *testing.T) {
	// [01 02 03 04] reads LE to lane 0x04030201 then XOR len(4).
	if EncodeBytesWord32([]byte{0x01, 0x02, 0x03, 0x04}) != 0x04030201^4 {
		t.Fatal("4-byte LE fold mismatch")
	}
	if Hash32Bytes([]byte{0x01, 0x02, 0x03, 0x04}, 0) != Hash32(0x04030201^4, 0) {
		t.Fatal("Hash32Bytes must equal Hash32 of the folded word")
	}
}

func TestBytesLEFold64(t *testing.T) {
	// [01..08] reads LE to lane 0x0807060504030201 then XOR len(8).
	if EncodeBytesWord64([]byte{1, 2, 3, 4, 5, 6, 7, 8}) != 0x0807060504030201^8 {
		t.Fatal("8-byte LE fold mismatch")
	}
	if Hash64Bytes([]byte{1, 2, 3, 4, 5, 6, 7, 8}, 0) != Hash64(0x0807060504030201^8, 0) {
		t.Fatal("Hash64Bytes must equal Hash64 of the folded word")
	}
}

func TestBytesTailAndLengthDistinguish(t *testing.T) {
	// Sub-lane tail goes to LOW bytes; length XOR'd in so these all differ.
	h3 := Hash32Bytes([]byte{0x01, 0x02, 0x03}, 0)
	h2 := Hash32Bytes([]byte{0x01, 0x02}, 0)
	h4 := Hash32Bytes([]byte{0x01, 0x02, 0x03, 0x00}, 0)
	if h3 == h2 {
		t.Fatal("[01 02 03] must differ from [01 02]")
	}
	if h3 == h4 {
		t.Fatal("[01 02 03] must differ from [01 02 03 00]")
	}
	// [00] != [00,00] (length XOR distinguishes equal-byte tails).
	if Hash32Bytes([]byte{0x00}, 0) == Hash32Bytes([]byte{0x00, 0x00}, 0) {
		t.Fatal("[00] must differ from [00 00]")
	}
	// Tail in LOW bytes: [01] folds to lane 0x00000001.
	if EncodeBytesWord32([]byte{0x01}) != 0x00000001^1 {
		t.Fatal("[01] must fold to lane 0x00000001 ^ len")
	}
}

// ---- Authoritative numeric test vectors (the committed oracle) ------------
// These rows mirror spec/features/hash-pipeline.md "Test vectors" and the
// frozen Rust reference (feat/hash-pipeline 53d4a50).

var hash32Words = [6]uint32{
	0x00000000, 0x00000001, 0xffffffff, 0x80000000, 0x7fffffff, 0x04030201,
}
var hash32Seeds = [4]uint64{
	0x0000000000000000, 0x0000000000000001, 0x00000000ffffffff, 0xffffffff00000000,
}

func TestHash32Vectors(t *testing.T) {
	expected := [6][4]uint32{
		{0x00000000, 0x514e28b7, 0x81f16f39, 0x81f16f39},
		{0x514e28b7, 0x00000000, 0x7995c304, 0x7995c304},
		{0x81f16f39, 0x7995c304, 0x00000000, 0x00000000},
		{0x6d3c65a0, 0x8b7f7a6a, 0xf9cc0ea8, 0xf9cc0ea8},
		{0xf9cc0ea8, 0x551b50f6, 0x6d3c65a0, 0x6d3c65a0},
		{0xd839eaff, 0x54ec0422, 0xaf02bbbc, 0xaf02bbbc},
	}
	for wi, w := range hash32Words {
		for si, s := range hash32Seeds {
			if got := Hash32(w, s); got != expected[wi][si] {
				t.Errorf("Hash32(%#08x, %#018x) = %#08x, want %#08x", w, s, got, expected[wi][si])
			}
		}
	}
	// The two high/low-only seeds fold to the SAME seed32 (0xffffffff), so
	// their columns are identical -- an intentional, pinned property.
	for _, w := range hash32Words {
		if Hash32(w, 0x00000000ffffffff) != Hash32(w, 0xffffffff00000000) {
			t.Errorf("seed-fold column equality broken for word %#08x", w)
		}
	}
}

var hash64Words = [6]uint64{
	0x0000000000000000, 0x0000000000000001, 0x00000000ffffffff,
	0x0000000080000000, 0xffffffffffffffff, 0x0807060504030201,
}
var hash64Seeds = [4]uint64{
	0x0000000000000000, 0x0000000000000001, 0x00000000ffffffff, 0xffffffff00000000,
}

func TestHash64Vectors(t *testing.T) {
	expected := [6][4]uint64{
		{0x0000000000000000, 0xb456bcfc34c2cb2c, 0xcc71ecda2aa8bcc6, 0xc9213cd20c528300},
		{0xb456bcfc34c2cb2c, 0x0000000000000000, 0x0789620c2ee64a3e, 0x2640647a5ca0376b},
		{0xcc71ecda2aa8bcc6, 0x0789620c2ee64a3e, 0x0000000000000000, 0x64b5720b4b825f21},
		{0xe3beca1f9a7e4886, 0x81b875318ee00b8e, 0x8a662c1a93a26b91, 0xc4ca27146b0a922f},
		{0x64b5720b4b825f21, 0x3a8593886c55a02b, 0xc9213cd20c528300, 0xcc71ecda2aa8bcc6},
		{0x9b57670c60240a13, 0xda66ed8bc89ffb5f, 0xbe7f6184429515e7, 0x916bf52bf4cf0681},
	}
	for wi, w := range hash64Words {
		for si, s := range hash64Seeds {
			if got := Hash64(w, s); got != expected[wi][si] {
				t.Errorf("Hash64(%#018x, %#018x) = %#016x, want %#016x", w, s, got, expected[wi][si])
			}
			// Lane accessors must agree with the combined value.
			full := expected[wi][si]
			if Hash64Hi(w, s) != uint32(full>>32) || Hash64Lo(w, s) != uint32(full) {
				t.Errorf("lane split mismatch for Hash64(%#018x, %#018x)", w, s)
			}
		}
	}
}

// Sanity checks called out in the porting brief.
func TestSanityChecks(t *testing.T) {
	cases := []struct {
		got, want uint64
		name      string
	}{
		{uint64(Hash32(1, 0)), 0x514e28b7, "hash32(1,0)"},
		{Hash64(1, 0), 0xb456bcfc34c2cb2c, "hash64(1,0)"},
		{uint64(Hash32(0x80000000, 0)), 0x6d3c65a0, "hash32(0x80000000,0)"},
		{uint64(Hash32(0, 0)), 0, "hash32(0,0)"},
		{Hash64(0, 0), 0, "hash64(0,0)"},
		{Hash64(0xffffffffffffffff, 0), 0x64b5720b4b825f21, "hash64(all ones,0)"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %#x, want %#x", c.name, c.got, c.want)
		}
	}
	if Hash64Hi(1, 0) != 0xb456bcfc || Hash64Lo(1, 0) != 0x34c2cb2c {
		t.Errorf("hash64(1,0) lanes = (%#08x,%#08x), want (0xb456bcfc,0x34c2cb2c)", Hash64Hi(1, 0), Hash64Lo(1, 0))
	}
}

// ---- positions_from_hashes oracle rows (spec position matrix) -------------

func TestPositionsVectorRows(t *testing.T) {
	cases := []struct {
		h1, h2, m, k uint32
		want         []uint32
	}{
		{0x00000000, 0x00000001, 16, 4, []uint32{0, 1, 2, 3}},
		{0x0000000a, 0x00000003, 16, 4, []uint32{10, 13, 0, 3}},
		{0xffffffff, 0x00000001, 16, 3, []uint32{15, 0, 1}},
		// i*h2 multiply wrap: i=2, 2*0x80000000 = 0x100000000 -> 0.
		{0x80000000, 0x80000000, 7, 3, []uint32{2, 0, 2}},
		// addition wrap + unsigned mod with high bit set.
		{0xfffffffd, 0x00000002, 1000, 5, []uint32{293, 295, 1, 3, 5}},
	}
	for _, c := range cases {
		if got := PositionsFromHashes(c.h1, c.h2, c.m, c.k); !reflect.DeepEqual(got, c.want) {
			t.Errorf("PositionsFromHashes(%#x,%#x,%d,%d) = %v, want %v", c.h1, c.h2, c.m, c.k, got, c.want)
		}
	}
}

func TestPositionsPow2EqualsModulo(t *testing.T) {
	v := Positions([]byte("hello"), 64, 7)
	h1 := Hash32Bytes([]byte("hello"), 0)
	h2 := Hash32Bytes([]byte("hello"), SALT2)
	for i, p := range v {
		combined := h1 + uint32(i)*h2
		if p != combined&63 {
			t.Errorf("position %d: mask result mismatch", i)
		}
		if p != combined%64 {
			t.Errorf("position %d: modulo result mismatch", i)
		}
	}
}

func TestPositionsPublicUsesInternalSeeds(t *testing.T) {
	h1 := Hash32Bytes([]byte("abc"), 0)
	h2 := Hash32Bytes([]byte("abc"), SALT2)
	if !reflect.DeepEqual(Positions([]byte("abc"), 1000, 5), PositionsFromHashes(h1, h2, 1000, 5)) {
		t.Fatal("Positions must use internal seeds 0 and SALT2")
	}
}

// ---- HllSplit sanity (pre-stated) -----------------------------------------

func TestHllSplitBasic(t *testing.T) {
	idx, rho := HllSplit([]byte("x"), 12)
	x := Hash64Bytes([]byte("x"), 0)
	if idx != uint32(x>>(64-12)) {
		t.Errorf("idx = %#x, want %#x", idx, uint32(x>>(64-12)))
	}
	if rho < 1 {
		t.Errorf("rho must be >= 1, got %d", rho)
	}
	if idx >= (1 << 12) {
		t.Errorf("idx must be < 2^12, got %d", idx)
	}
}
