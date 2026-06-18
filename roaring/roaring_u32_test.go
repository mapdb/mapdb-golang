// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package roaring

import (
	"bytes"
	"reflect"
	"testing"
)

func build(vals ...uint32) *RoaringU32 {
	s := NewRoaringU32()
	for _, v := range vals {
		s.Add(v)
	}
	return s
}

func TestEmptySet(t *testing.T) {
	s := NewRoaringU32()
	if !s.IsEmpty() {
		t.Fatal("new set must be empty")
	}
	if s.Cardinality() != 0 {
		t.Fatalf("cardinality = %d, want 0", s.Cardinality())
	}
	if _, ok := s.Min(); ok {
		t.Fatal("Min on empty must be absent")
	}
	if _, ok := s.Max(); ok {
		t.Fatal("Max on empty must be absent")
	}
	if s.ChunkCount() != 0 {
		t.Fatalf("chunk_count = %d, want 0", s.ChunkCount())
	}
	if len(s.ToSortedSlice()) != 0 {
		t.Fatal("empty set must iterate empty")
	}
	want := []byte{0x55, 0x30, 0x52, 0x32, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if got := s.Serialize(); !bytes.Equal(got, want) {
		t.Fatalf("empty serialize = %x, want %x", got, want)
	}
}

func TestSingleElement(t *testing.T) {
	s := build(42)
	if s.Cardinality() != 1 || s.ChunkCount() != 1 {
		t.Fatalf("card=%d chunks=%d", s.Cardinality(), s.ChunkCount())
	}
	if got := s.ContainerTypes(); !reflect.DeepEqual(got, []string{"array"}) {
		t.Fatalf("container types = %v", got)
	}
	if v, ok := s.Min(); !ok || v != 42 {
		t.Fatalf("min = %d,%v", v, ok)
	}
	if v, ok := s.Max(); !ok || v != 42 {
		t.Fatalf("max = %d,%v", v, ok)
	}
	b := s.Serialize()
	// header(12) + high(2) + tag/pad(2) + card-1(2) + one u16 low.
	if len(b) != 12+6+2 {
		t.Fatalf("len = %d", len(b))
	}
	wantTail := []byte{0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x2a, 0x00}
	if !bytes.Equal(b[12:], wantTail) {
		t.Fatalf("tail = %x, want %x", b[12:], wantTail)
	}
}

func TestIdempotentAddRemove(t *testing.T) {
	s := build(5)
	if s.Add(5) {
		t.Fatal("re-add must return false")
	}
	if !s.Add(6) {
		t.Fatal("new add must return true")
	}
	if !s.Remove(6) {
		t.Fatal("remove present must return true")
	}
	if s.Remove(6) {
		t.Fatal("re-remove must return false")
	}
	if s.Remove(99) {
		t.Fatal("remove absent must return false")
	}
}

func TestUnsignedOrderWithSignedExtremes(t *testing.T) {
	// i32 {INT_MIN, -1, 0, INT_MAX} reinterpreted to u32.
	intMin, negOne, intMax := int32(-2147483648), int32(-1), int32(2147483647)
	s := build(uint32(intMin), uint32(negOne), 0, uint32(intMax))
	want := []uint32{0x00000000, 0x7FFFFFFF, 0x80000000, 0xFFFFFFFF}
	if got := s.ToSortedSlice(); !reflect.DeepEqual(got, want) {
		t.Fatalf("sorted = %x, want %x", got, want)
	}
	if v, _ := s.Min(); v != 0 {
		t.Fatalf("min = %#x, want 0", v)
	}
	if v, _ := s.Max(); v != 0xFFFFFFFF {
		t.Fatalf("max = %#x, want 0xFFFFFFFF", v)
	}
	// Chunk highs unsigned ascending: 0x0000, 0x7FFF, 0x8000, 0xFFFF.
	b := s.Serialize()
	wantHighs := []uint16{0x0000, 0x7FFF, 0x8000, 0xFFFF}
	for i, wh := range wantHighs {
		off := 12 + i*8
		got := uint16(b[off]) | uint16(b[off+1])<<8
		if got != wh {
			t.Fatalf("chunk %d high = %#x, want %#x", i, got, wh)
		}
	}
}

func TestThresholdArray4096Bitmap4097(t *testing.T) {
	s := NewRoaringU32()
	for v := uint32(0); v < 4096; v++ {
		s.Add(v)
	}
	if s.Cardinality() != 4096 {
		t.Fatalf("card = %d", s.Cardinality())
	}
	if got := s.ContainerTypes(); !reflect.DeepEqual(got, []string{"array"}) {
		t.Fatalf("4096 must be array, got %v", got)
	}
	s.Add(4096)
	if s.Cardinality() != 4097 {
		t.Fatalf("card = %d", s.Cardinality())
	}
	if got := s.ContainerTypes(); !reflect.DeepEqual(got, []string{"bitmap"}) {
		t.Fatalf("4097 must be bitmap, got %v", got)
	}
}

func TestArrayToBitmapAndBackSameBytes(t *testing.T) {
	grown := NewRoaringU32()
	for v := uint32(0); v <= 4096; v++ {
		grown.Add(v)
	}
	if got := grown.ContainerTypes(); !reflect.DeepEqual(got, []string{"bitmap"}) {
		t.Fatalf("grown must be bitmap, got %v", got)
	}
	grown.Remove(4096)
	if got := grown.ContainerTypes(); !reflect.DeepEqual(got, []string{"array"}) {
		t.Fatalf("down-converted must be array, got %v", got)
	}
	never := NewRoaringU32()
	for v := uint32(0); v < 4096; v++ {
		never.Add(v)
	}
	if !bytes.Equal(grown.Serialize(), never.Serialize()) {
		t.Fatal("history-independence: grown-then-shrunk must equal never-grown bytes")
	}
}

func TestContainerTypePureFunctionOfCardinality(t *testing.T) {
	// Reach cardinality 4096 by two different paths; expect identical bytes.
	a := NewRoaringU32()
	for v := uint32(0); v < 5000; v++ {
		a.Add(v)
	}
	for v := uint32(4096); v < 5000; v++ {
		a.Remove(v)
	}
	b := NewRoaringU32()
	for v := int(4095); v >= 0; v-- {
		b.Add(uint32(v))
	}
	if got := a.ContainerTypes(); !reflect.DeepEqual(got, []string{"array"}) {
		t.Fatalf("a must be array, got %v", got)
	}
	if !bytes.Equal(a.Serialize(), b.Serialize()) {
		t.Fatal("type-by-cardinality: ascending and descending insertion paths must match")
	}
}

func TestFullChunk(t *testing.T) {
	s := NewRoaringU32()
	for v := uint32(0); v <= 65535; v++ {
		s.Add(v)
	}
	if s.Cardinality() != 65536 {
		t.Fatalf("card = %d", s.Cardinality())
	}
	if got := s.ContainerTypes(); !reflect.DeepEqual(got, []string{"bitmap"}) {
		t.Fatalf("full chunk must be bitmap, got %v", got)
	}
	b := s.Serialize()
	// CARDINALITY_MINUS_1 == 0xFFFF at offset 12+4.
	if b[16] != 0xFF || b[17] != 0xFF {
		t.Fatalf("CARDINALITY_MINUS_1 bytes = %x %x, want ff ff", b[16], b[17])
	}
	if len(b) != 12+6+8192 {
		t.Fatalf("len = %d", len(b))
	}
	back, err := Deserialize(b)
	if err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if !bytes.Equal(back.Serialize(), b) {
		t.Fatal("round-trip mismatch")
	}
	if back.Cardinality() != 65536 {
		t.Fatalf("round-trip card = %d", back.Cardinality())
	}
}

func TestDropEmptyChunk(t *testing.T) {
	s := build(100000, 5)
	if s.ChunkCount() != 2 {
		t.Fatalf("chunks = %d", s.ChunkCount())
	}
	s.Remove(100000)
	if s.ChunkCount() != 1 {
		t.Fatalf("after drop chunks = %d", s.ChunkCount())
	}
	if !bytes.Equal(s.Serialize(), build(5).Serialize()) {
		t.Fatal("dropped-chunk bytes must equal a set that only held {5}")
	}
}

func TestClear(t *testing.T) {
	s := build(1, 70000, 200000)
	s.Clear()
	if !s.IsEmpty() || s.ChunkCount() != 0 {
		t.Fatal("clear must yield canonical empty set")
	}
	if !bytes.Equal(s.Serialize(), NewRoaringU32().Serialize()) {
		t.Fatal("cleared bytes must equal a fresh empty set")
	}
}

func TestSetAlgebraBasic(t *testing.T) {
	a := build(1, 2, 3, 70000)
	b := build(2, 3, 4, 140000)
	check := func(name string, got *RoaringU32, want []uint32) {
		if g := got.ToSortedSlice(); !reflect.DeepEqual(g, want) {
			t.Fatalf("%s = %v, want %v", name, g, want)
		}
	}
	check("or", a.Or(b), []uint32{1, 2, 3, 4, 70000, 140000})
	check("and", a.And(b), []uint32{2, 3})
	check("and_not", a.AndNot(b), []uint32{1, 70000})
	check("xor", a.Xor(b), []uint32{1, 4, 70000, 140000})
	// operands unchanged
	check("a", a, []uint32{1, 2, 3, 70000})
	check("b", b, []uint32{2, 3, 4, 140000})
}

func TestXorBitmapNormalizesToArray(t *testing.T) {
	a := NewRoaringU32()
	b := NewRoaringU32()
	for v := uint32(0); v < 5000; v++ {
		a.Add(v)
		b.Add(v)
	}
	for v := uint32(5000); v < 5030; v++ {
		a.Add(v)
	}
	if got := a.ContainerTypes(); !reflect.DeepEqual(got, []string{"bitmap"}) {
		t.Fatalf("a must be bitmap, got %v", got)
	}
	x := a.Xor(b)
	if x.Cardinality() != 30 {
		t.Fatalf("xor card = %d", x.Cardinality())
	}
	if got := x.ContainerTypes(); !reflect.DeepEqual(got, []string{"array"}) {
		t.Fatalf("xor result must be array, got %v", got)
	}
}

func TestOrArrayNormalizesToBitmap(t *testing.T) {
	a := NewRoaringU32()
	for v := uint32(0); v < 3000; v++ {
		a.Add(v)
	}
	b := NewRoaringU32()
	for v := uint32(2000); v < 6000; v++ {
		b.Add(v)
	}
	if got := a.ContainerTypes(); !reflect.DeepEqual(got, []string{"array"}) {
		t.Fatalf("a must be array, got %v", got)
	}
	u := a.Or(b)
	if u.Cardinality() != 6000 {
		t.Fatalf("union card = %d", u.Cardinality())
	}
	if got := u.ContainerTypes(); !reflect.DeepEqual(got, []string{"bitmap"}) {
		t.Fatalf("union result must be bitmap, got %v", got)
	}
}

func TestAndNotEmptiesChunk(t *testing.T) {
	a := build(1, 2, 70000, 70001)
	other := build(70000, 70001)
	d := a.AndNot(other)
	if d.ChunkCount() != 1 {
		t.Fatalf("chunks = %d", d.ChunkCount())
	}
	if got := d.ToSortedSlice(); !reflect.DeepEqual(got, []uint32{1, 2}) {
		t.Fatalf("result = %v", got)
	}
	if !bytes.Equal(d.Serialize(), build(1, 2).Serialize()) {
		t.Fatal("emptied-chunk result must equal {1,2}")
	}
}

func TestResultIndependentOfOperands(t *testing.T) {
	a := build(1, 2, 3)
	b := build(3, 4, 5)
	u := a.Or(b)
	u.Add(999)
	if a.Contains(999) || b.Contains(999) {
		t.Fatal("mutating result must not affect operands")
	}
	// only-A chunk copied, not aliased
	a2 := build(1)
	b2 := NewRoaringU32()
	u2 := a2.Or(b2)
	a2.Add(2)
	if u2.Contains(2) {
		t.Fatal("only-A chunk must be copied, not aliased")
	}
}

func TestInPlaceOps(t *testing.T) {
	a := build(1, 2, 3)
	b := build(2, 3, 4)
	a.OrInPlace(b)
	if got := a.ToSortedSlice(); !reflect.DeepEqual(got, []uint32{1, 2, 3, 4}) {
		t.Fatalf("or_in_place = %v", got)
	}
	c := build(1, 2, 3)
	c.AndInPlace(build(2, 3, 4))
	if got := c.ToSortedSlice(); !reflect.DeepEqual(got, []uint32{2, 3}) {
		t.Fatalf("and_in_place = %v", got)
	}
}

func TestSerializeDeserializeRoundTrip(t *testing.T) {
	// Deterministic LCG; spans many chunks and a mix of array/bitmap.
	s := NewRoaringU32()
	var x uint64 = 0x12345678
	for i := 0; i < 20000; i++ {
		x = x*6364136223846793005 + 1442695040888963407
		s.Add(uint32(x >> 16))
	}
	b := s.Serialize()
	back, err := Deserialize(b)
	if err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if !bytes.Equal(back.Serialize(), b) {
		t.Fatal("re-serialize mismatch")
	}
	if !reflect.DeepEqual(back.ToSortedSlice(), s.ToSortedSlice()) {
		t.Fatal("iteration mismatch after round-trip")
	}
}

func TestAddRangeRemoveRange(t *testing.T) {
	s := NewRoaringU32()
	for v := uint32(0); v <= 4095; v++ {
		s.Add(v)
	}
	if s.Cardinality() != 4096 {
		t.Fatalf("card = %d", s.Cardinality())
	}
	for v := uint32(100); v <= 200; v++ {
		s.Remove(v)
	}
	if s.Cardinality() != 4096-101 {
		t.Fatalf("card = %d, want %d", s.Cardinality(), 4096-101)
	}
}

func TestCardinalityWidth(t *testing.T) {
	s := NewRoaringU32()
	for v := uint32(0); v <= 65535; v++ {
		s.Add(v)
	}
	var c uint64 = s.Cardinality()
	if c != 65536 {
		t.Fatalf("card = %d", c)
	}
}

// ---- deserialize rejection tests ----

func validBytes() []byte { return build(1, 70000).Serialize() }

func TestRejectBadMagic(t *testing.T) {
	b := validBytes()
	b[0] = 0x00
	if _, err := Deserialize(b); err == nil {
		t.Fatal("must reject bad magic")
	}
}

func TestRejectBadVersion(t *testing.T) {
	b := validBytes()
	b[4] = 0x02
	if _, err := Deserialize(b); err == nil {
		t.Fatal("must reject bad version")
	}
}

func TestRejectNonZeroReserved(t *testing.T) {
	b := validBytes()
	b[6] = 0x01
	if _, err := Deserialize(b); err == nil {
		t.Fatal("must reject non-zero reserved")
	}
}

func TestRejectNonZeroPad(t *testing.T) {
	b := validBytes()
	b[15] = 0x01 // first chunk PAD at offset 12+2+1
	if _, err := Deserialize(b); err == nil {
		t.Fatal("must reject non-zero pad")
	}
}

func TestRejectUnknownTag(t *testing.T) {
	b := validBytes()
	b[14] = 0x03 // first chunk tag at offset 12+2
	if _, err := Deserialize(b); err == nil {
		t.Fatal("must reject unknown tag")
	}
}

func TestRejectTrailingBytes(t *testing.T) {
	b := append(validBytes(), 0x00)
	if _, err := Deserialize(b); err == nil {
		t.Fatal("must reject trailing bytes")
	}
}

func TestRejectTruncated(t *testing.T) {
	b := validBytes()
	if _, err := Deserialize(b[:len(b)-1]); err == nil {
		t.Fatal("must reject truncated payload")
	}
	if _, err := Deserialize(b[:5]); err == nil {
		t.Fatal("must reject truncated header")
	}
}

func TestRejectChunkCountTooLarge(t *testing.T) {
	b := validBytes()
	// CHUNK_COUNT at offset 8..12 -> 70000 (LE)
	b[8], b[9], b[10], b[11] = 0x70, 0x11, 0x01, 0x00
	if _, err := Deserialize(b); err == nil {
		t.Fatal("must reject CHUNK_COUNT > 65536")
	}
}

func header(chunkCount uint32) []byte {
	b := []byte{0x55, 0x30, 0x52, 0x32, 0x01, 0x00, 0x00, 0x00}
	b = append(b, byte(chunkCount), byte(chunkCount>>8), byte(chunkCount>>16), byte(chunkCount>>24))
	return b
}

func leU16(v uint16) []byte { return []byte{byte(v), byte(v >> 8)} }

func TestRejectNonCanonicalArrayCardinality(t *testing.T) {
	// ARRAY with cardinality 4097 (> arrayMax): illegal.
	b := header(1)
	b = append(b, leU16(0)...) // high
	b = append(b, tagArray, 0)
	b = append(b, leU16(4096)...) // card-1 = 4096 => card 4097
	for low := 0; low < 4097; low++ {
		b = append(b, leU16(uint16(low))...)
	}
	if _, err := Deserialize(b); err == nil {
		t.Fatal("must reject non-canonical ARRAY cardinality")
	}
}

func TestRejectNonCanonicalBitmapCardinality(t *testing.T) {
	// BITMAP with cardinality 1 (<= arrayMax): illegal.
	b := header(1)
	b = append(b, leU16(0)...)
	b = append(b, tagBitmap, 0)
	b = append(b, leU16(0)...) // card 1
	words := make([]uint64, bitmapWords)
	words[0] = 1
	for _, w := range words {
		b = appendU64(b, w)
	}
	if _, err := Deserialize(b); err == nil {
		t.Fatal("must reject non-canonical BITMAP cardinality")
	}
}

func TestRejectNonAscendingArrayLows(t *testing.T) {
	b := header(1)
	b = append(b, leU16(0)...)
	b = append(b, tagArray, 0)
	b = append(b, leU16(1)...) // card 2
	b = append(b, leU16(5)...)
	b = append(b, leU16(5)...) // duplicate
	if _, err := Deserialize(b); err == nil {
		t.Fatal("must reject non-ascending/duplicate ARRAY lows")
	}
}

func TestRejectBitmapPopcountMismatch(t *testing.T) {
	b := header(1)
	b = append(b, leU16(0)...)
	b = append(b, tagBitmap, 0)
	b = append(b, leU16(4096)...) // claims card 4097
	for _, w := range make([]uint64, bitmapWords) {
		b = appendU64(b, w) // popcount 0
	}
	if _, err := Deserialize(b); err == nil {
		t.Fatal("must reject BITMAP popcount mismatch")
	}
}

func TestRejectNonAscendingChunkHighs(t *testing.T) {
	b := header(2)
	// chunk 1: high 5
	b = append(b, leU16(5)...)
	b = append(b, tagArray, 0)
	b = append(b, leU16(0)...)
	b = append(b, leU16(0)...)
	// chunk 2: high 5 again (non-ascending)
	b = append(b, leU16(5)...)
	b = append(b, tagArray, 0)
	b = append(b, leU16(0)...)
	b = append(b, leU16(0)...)
	if _, err := Deserialize(b); err == nil {
		t.Fatal("must reject non-ascending chunk highs")
	}
}

func TestIterateEarlyStop(t *testing.T) {
	s := build(1, 2, 3, 70000)
	var seen []uint32
	s.Iterate(func(v uint32) bool {
		seen = append(seen, v)
		return v != 2 // stop after 2
	})
	if !reflect.DeepEqual(seen, []uint32{1, 2}) {
		t.Fatalf("early-stop iterate = %v", seen)
	}
}

func TestBitmapBitOrder(t *testing.T) {
	// Force a BITMAP, then set sparse low keys spanning distant words.
	s := NewRoaringU32()
	for v := uint32(0); v <= 4096; v++ {
		s.Add(v)
	}
	for _, v := range []uint32{5000, 9000, 60000, 65535} {
		s.Add(v)
	}
	if got := s.ContainerTypes(); !reflect.DeepEqual(got, []string{"bitmap"}) {
		t.Fatalf("must be bitmap, got %v", got)
	}
	for _, v := range []uint32{4096, 5000, 9000, 60000, 65535, 0} {
		if !s.Contains(v) {
			t.Fatalf("must contain %d", v)
		}
	}
	if s.Contains(4097) {
		t.Fatal("must not contain 4097")
	}
	b := s.Serialize()
	back, err := Deserialize(b)
	if err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if !bytes.Equal(back.Serialize(), b) {
		t.Fatal("bit-order round-trip mismatch")
	}
}
