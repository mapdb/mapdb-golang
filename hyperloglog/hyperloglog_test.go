// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package hyperloglog

import (
	"bytes"
	"encoding/hex"
	"math"
	"testing"

	"github.com/mapdb/mapdb-golang/hash"
)

func mustNew(t *testing.T, p uint8) HyperLogLog {
	t.Helper()
	h, err := NewHyperLogLogWithPrecision(p)
	if err != nil {
		t.Fatalf("NewHyperLogLogWithPrecision(%d): %v", p, err)
	}
	return h
}

// expectedSplit recomputes (idx, rho) for an i32 item independently of the
// implementation, so the register-update tests are a real oracle check.
func expectedSplit(item int32, p uint8) (uint32, uint8) {
	x := hash.Hash64(uint64(uint32(item)), 0)
	pp := uint(p)
	idx := uint32(x >> (64 - pp))
	w := (x << pp) | (uint64(1) << (pp - 1))
	return idx, uint8(bitsLeadingZeros64(w) + 1)
}

func bitsLeadingZeros64(x uint64) int {
	n := 0
	for i := 63; i >= 0; i-- {
		if x&(uint64(1)<<uint(i)) != 0 {
			break
		}
		n++
	}
	return n
}

func toHex(b []byte) string { return hex.EncodeToString(b) }

// mustBytes returns the serialized form of a fresh sketch at precision p
// (ToBytes has a pointer receiver, so it needs an addressable value).
func mustBytes(t *testing.T, p uint8) []byte {
	t.Helper()
	h := mustNew(t, p)
	return h.ToBytes()
}

// goldenRatioStep is the 32-bit golden-ratio multiplier (0x9E3779B1) used to
// spread the i32 add-sequence into genuinely distinct values for the estimate
// tolerance tests. Written as a wrapped int32 (the constant overflows int32).
const goldenRatioStep = int32(-1640531527) // = int32(uint32(2654435761))

// ---- Construction & p-range ----------------------------------------------

func TestWithPrecisionAllocatesMRegisters(t *testing.T) {
	for p := MinPrecision; p <= MaxPrecision; p++ {
		h := mustNew(t, p)
		if got := h.RegisterCount(); got != 1<<p {
			t.Fatalf("p=%d register count %d, want %d", p, got, 1<<p)
		}
		for _, r := range h.Registers() {
			if r != 0 {
				t.Fatalf("p=%d fresh register nonzero", p)
			}
		}
		if h.NonzeroRegisters() != 0 || h.MaxRegister() != 0 {
			t.Fatalf("p=%d fresh nonzero/max not 0", p)
		}
	}
}

func TestPOutOfRangeErrorsNeverClamps(t *testing.T) {
	for _, p := range []uint8{0, 1, 3, 19, 255} {
		if _, err := NewHyperLogLogWithPrecision(p); err == nil {
			t.Fatalf("p=%d must error", p)
		}
	}
}

// ---- rho / clz / guard-bit exactness -------------------------------------

func TestSplitGuardBitAllZeroRemainderGivesMaxRho(t *testing.T) {
	// Craft x whose low (64 - p) bits are all zero: x = idx << (64 - p). Then
	// w = (x << p) | guard = guard = 1 << (p-1); clz64 = 64-p; rho = 64-p+1.
	for _, p := range []uint8{4, 7, 14, 18} {
		pp := uint(p)
		idx := uint64(5) & ((uint64(1) << pp) - 1)
		x := idx << (64 - pp)
		gi, rho := split(x, p)
		if uint64(gi) != idx {
			t.Fatalf("p=%d idx %d, want %d", p, gi, idx)
		}
		if rho != 64-p+1 || rho != rhoCeiling(p) {
			t.Fatalf("p=%d all-zero-remainder rho %d, want %d", p, rho, 64-p+1)
		}
	}
}

func TestAddZeroIsAllZeroRemainder(t *testing.T) {
	// Hash64(0,0)=0, so Add(0) drives the crafted all-zero-remainder path: idx
	// 0, rho = 64-p+1.
	for _, p := range []uint8{4, 10, 18} {
		h := mustNew(t, p)
		h.Add(0)
		if h.Registers()[0] != rhoCeiling(p) {
			t.Fatalf("p=%d Add(0) register[0] %d, want %d", p, h.Registers()[0], rhoCeiling(p))
		}
		if h.MaxRegister() != rhoCeiling(p) {
			t.Fatalf("p=%d Add(0) max %d, want %d", p, h.MaxRegister(), rhoCeiling(p))
		}
	}
}

func TestSplitMinRhoIsOne(t *testing.T) {
	// Top remaining bit set -> clz64(w) = 0 -> rho = 1 (the minimum).
	p := uint8(4)
	x := uint64(1) << (64 - uint(p) - 1)
	if _, rho := split(x, p); rho != 1 {
		t.Fatalf("rho %d, want 1", rho)
	}
}

func TestSplitIdxIsTopPBitsLogical(t *testing.T) {
	// High-bit-set x must use a LOGICAL shift for the index.
	x := uint64(0xffffffffffffffff)
	for _, p := range []uint8{4, 10, 18} {
		idx, _ := split(x, p)
		if want := uint32(1)<<p - 1; idx != want {
			t.Fatalf("p=%d top-p-bits idx %d, want %d", p, idx, want)
		}
	}
}

func TestSplitRhoWithinPerPBounds(t *testing.T) {
	for p := MinPrecision; p <= MaxPrecision; p++ {
		for item := int32(0); item < 2000; item++ {
			x := hash.Hash64(uint64(uint32(item)), 0)
			idx, rho := split(x, p)
			if idx >= uint32(1)<<p {
				t.Fatalf("p=%d idx %d out of range", p, idx)
			}
			if rho < 1 || rho > rhoCeiling(p) {
				t.Fatalf("p=%d rho %d out of [1,%d]", p, rho, rhoCeiling(p))
			}
		}
	}
}

// ---- register max-update & idempotence -----------------------------------

func TestAddUpdatesExpectedRegisterToRho(t *testing.T) {
	p := uint8(14)
	h := mustNew(t, p)
	idx, rho := expectedSplit(42, p)
	h.Add(42)
	if h.Registers()[idx] != rho {
		t.Fatalf("register[%d] %d, want %d", idx, h.Registers()[idx], rho)
	}
	if h.NonzeroRegisters() != 1 || h.MaxRegister() != rho {
		t.Fatalf("nonzero/max mismatch")
	}
}

func TestAddIsMaxNotOverwriteAndIdempotent(t *testing.T) {
	p := uint8(4)
	a := mustNew(t, p)
	a.Add(7)
	b := mustNew(t, p)
	b.Add(7)
	b.Add(7)
	b.Add(7)
	if !bytes.Equal(a.Registers(), b.Registers()) {
		t.Fatal("re-add of same item must be idempotent")
	}
}

func TestAddOrderIndependent(t *testing.T) {
	p := uint8(6)
	ab := mustNew(t, p)
	ab.Add(11)
	ab.Add(99999)
	ba := mustNew(t, p)
	ba.Add(99999)
	ba.Add(11)
	if !bytes.Equal(ab.Registers(), ba.Registers()) {
		t.Fatal("add order must not change registers")
	}
}

func TestNegOneZeroExtendDiffersFromSignExtend(t *testing.T) {
	// Add(-1) encodes 0x00000000ffffffff (zero-extend). The would-be
	// sign-extend (0xffffffffffffffff) yields a different x -> different rho/idx.
	p := uint8(4)
	h := mustNew(t, p)
	h.Add(-1)
	zi, zr := split(hash.Hash64(0x00000000ffffffff, 0), p)
	si, sr := split(hash.Hash64(0xffffffffffffffff, 0), p)
	if h.Registers()[zi] != zr {
		t.Fatalf("zero-extend register mismatch")
	}
	if zi == si && zr == sr {
		t.Fatal("sign-extend would route identically -- test is vacuous")
	}
}

// ---- merge: element-wise max, p-mismatch ---------------------------------

func TestMergeIsElementwiseMax(t *testing.T) {
	p := uint8(4)
	a := mustNew(t, p)
	for _, v := range []int32{1, 2, 3} {
		a.Add(v)
	}
	b := mustNew(t, p)
	for _, v := range []int32{3, 4, 5} {
		b.Add(v)
	}
	expected := append([]uint8(nil), a.Registers()...)
	for i, bb := range b.Registers() {
		if bb > expected[i] {
			expected[i] = bb
		}
	}
	if err := a.Merge(&b); err != nil {
		t.Fatalf("merge: %v", err)
	}
	if !bytes.Equal(a.Registers(), expected) {
		t.Fatal("merge != element-wise max")
	}
}

func TestMergeCommutativeAndIdempotent(t *testing.T) {
	p := uint8(5)
	build := func(items ...int32) HyperLogLog {
		h := mustNew(t, p)
		for _, v := range items {
			h.Add(v)
		}
		return h
	}
	ab := build(10, 20, 30)
	bset := build(30, 40, 50)
	if err := ab.Merge(&bset); err != nil {
		t.Fatal(err)
	}
	ba := build(30, 40, 50)
	aset := build(10, 20, 30)
	if err := ba.Merge(&aset); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(ab.Registers(), ba.Registers()) {
		t.Fatal("merge must commute")
	}
	a := build(10, 20, 30)
	aa := build(10, 20, 30)
	if err := aa.Merge(&a); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(aa.Registers(), a.Registers()) {
		t.Fatal("merge must be idempotent")
	}
}

func TestMergePMismatchErrors(t *testing.T) {
	a := mustNew(t, 4)
	b := mustNew(t, 5)
	if err := a.Merge(&b); err == nil {
		t.Fatal("p-mismatch merge must error")
	}
}

// ---- serialization round-trip + rejections -------------------------------

func TestSerializeRoundtripAndHeader(t *testing.T) {
	h := mustNew(t, 4)
	h.Add(1)
	h.Add(7)
	h.Add(-1)
	b := h.ToBytes()
	if len(b) != 5+16 {
		t.Fatalf("len %d, want 21", len(b))
	}
	if !bytes.Equal(b[0:4], []byte("HLL1")) || b[4] != 4 {
		t.Fatal("header mismatch")
	}
	back, err := HyperLogLogFromBytes(b)
	if err != nil {
		t.Fatalf("from_bytes: %v", err)
	}
	if !bytes.Equal(back.Registers(), h.Registers()) || back.Precision() != h.Precision() {
		t.Fatal("roundtrip mismatch")
	}
}

func TestEmptyP4RegisterHexAnchor(t *testing.T) {
	h := mustNew(t, 4)
	// "HLL1" + 0x04 + 16 zero bytes.
	if got := toHex(h.ToBytes()); got != "484c4c310400000000000000000000000000000000" {
		t.Fatalf("anchor hex %s", got)
	}
}

func TestFromBytesRejectsBadMagic(t *testing.T) {
	b := mustBytes(t, 4)
	b[0] = 0x00
	if _, err := HyperLogLogFromBytes(b); err == nil {
		t.Fatal("bad magic must reject")
	}
}

func TestFromBytesRejectsTooShort(t *testing.T) {
	if _, err := HyperLogLogFromBytes([]byte{0x48, 0x4c, 0x4c}); err == nil {
		t.Fatal("too-short must reject")
	}
}

func TestFromBytesRejectsBadP(t *testing.T) {
	b := mustBytes(t, 4)
	b[4] = 3
	if _, err := HyperLogLogFromBytes(b); err == nil {
		t.Fatal("p=3 must reject")
	}
	b[4] = 19
	if _, err := HyperLogLogFromBytes(b); err == nil {
		t.Fatal("p=19 must reject")
	}
}

func TestFromBytesRejectsLengthMismatch(t *testing.T) {
	b := mustBytes(t, 4)
	b = append(b, 0) // one byte too many
	if _, err := HyperLogLogFromBytes(b); err == nil {
		t.Fatal("over-long must reject")
	}
	short := mustBytes(t, 4)[:20]
	if _, err := HyperLogLogFromBytes(short); err == nil {
		t.Fatal("short must reject")
	}
}

func TestFromBytesRejectsRegisterAboveCeiling(t *testing.T) {
	// At p=4 the ceiling is 64-4+1 = 61. A byte of 62 must be rejected.
	b := mustBytes(t, 4)
	b[5] = 62
	if _, err := HyperLogLogFromBytes(b); err == nil {
		t.Fatal("register 62 > ceiling 61 must reject")
	}
	b[5] = 61 // legal max accepted
	if _, err := HyperLogLogFromBytes(b); err != nil {
		t.Fatalf("register 61 must be accepted: %v", err)
	}
}

func TestFromBytesCeilingIsPerP(t *testing.T) {
	// At p=18 the ceiling is 64-18+1 = 47.
	b := mustBytes(t, 18)
	b[5] = 48
	if _, err := HyperLogLogFromBytes(b); err == nil {
		t.Fatal("register 48 > ceiling 47 must reject")
	}
	b[5] = 47
	if _, err := HyperLogLogFromBytes(b); err != nil {
		t.Fatalf("register 47 must be accepted: %v", err)
	}
}

// ---- estimate(): native-only, tolerance-bounded --------------------------

func TestFreshHLLEstimatesZero(t *testing.T) {
	// Z = m, E = alpha*m, E <= 2.5m and V = m > 0 -> m*ln(m/m) = m*ln(1) = 0.
	for _, p := range []uint8{4, 7, 14} {
		h := mustNew(t, p)
		if est := h.Estimate(); est != 0.0 {
			t.Fatalf("p=%d fresh estimate %v, want 0", p, est)
		}
	}
}

func TestEstimateWithinToleranceOnKnownCardinality(t *testing.T) {
	// Documented tolerance: relative error < 5% (covers HLL's ~1.04/sqrt(m)
	// standard error AND cross-libm float drift). p=14 -> m=16384.
	p := uint8(14)
	const n = 10000
	h := mustNew(t, p)
	for i := int32(0); i < n; i++ {
		h.Add(i * goldenRatioStep)
	}
	est := h.Estimate()
	rel := math.Abs(est-n) / n
	if rel >= 0.05 {
		t.Fatalf("estimate %v for n=%d: relative error %v >= 0.05", est, n, rel)
	}
}

func TestEstimateSmallCardinalityLinearCounting(t *testing.T) {
	// A few hundred distinct values: linear counting is active (E <= 2.5m, V > 0).
	p := uint8(14)
	const n = 300
	h := mustNew(t, p)
	for i := int32(0); i < n; i++ {
		h.Add(i * goldenRatioStep)
	}
	est := h.Estimate()
	rel := math.Abs(est-n) / n
	if rel >= 0.05 {
		t.Fatalf("small-card estimate %v for n=%d: rel %v", est, n, rel)
	}
}

func TestAlphaMMatchesPinnedConstants(t *testing.T) {
	if alphaM(4) != 0.673 || alphaM(5) != 0.697 || alphaM(6) != 0.709 {
		t.Fatal("pinned alpha constants mismatch")
	}
	if alphaM(7) != 0.7213/(1.0+1.079/128.0) {
		t.Fatal("alpha closed form mismatch at p=7")
	}
}

func TestEstimateLargeRangeCorrectionIsFinite(t *testing.T) {
	// Drive a high-register state (all registers NEAR the per-p max) via
	// FromBytes so raw E exceeds (1/30)*2^64 while staying below 2^64; assert
	// Estimate() is finite (the 2^64 ceiling keeps ln(1 - E/2^64)'s argument
	// positive; a 2^32 ceiling would return NaN here).
	p := uint8(4)
	nearMax := rhoCeiling(p) - 1
	b := mustBytes(t, p)
	for i := 5; i < len(b); i++ {
		b[i] = nearMax
	}
	h, err := HyperLogLogFromBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	if est := h.Estimate(); math.IsInf(est, 0) || math.IsNaN(est) {
		t.Fatalf("large-range estimate must be finite, got %v", est)
	}
}

// ---- scenario oracle anchors ---------------------------------------------

func TestScenarioOracleAnchors(t *testing.T) {
	// single add(1) at p=4 -> the committed scenario register_hex.
	h := mustNew(t, 4)
	h.Add(1)
	if got := "0x" + toHex(h.ToBytes()); got != "0x484c4c310400000000000000000000000200000000" {
		t.Fatalf("single add(1) p4 hex %s", got)
	}
	if h.NonzeroRegisters() != 1 {
		t.Fatal("single add nonzero != 1")
	}
	// high-rho: Add(0) at p=4 -> rho 61 at idx 0.
	hz := mustNew(t, 4)
	hz.Add(0)
	if got := "0x" + toHex(hz.ToBytes()); got != "0x484c4c31043d000000000000000000000000000000" {
		t.Fatalf("Add(0) p4 hex %s", got)
	}
}
