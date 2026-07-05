// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Package hyperloglog is the HyperLogLog distinct-count cardinality sketch
// (see spec/features/hyperloglog.md in mapdb-collection-spec). Port of the
// frozen reference (mapdb-rust src/hyperloglog.rs).
//
// # Float-quarantine ruling (the heart of this feature)
//
// HyperLogLog has two observable surfaces:
//
//  1. The INTEGER register array (m = 2^p uint8s, each the max rho seen). This
//     is exact integer state, a pure-integer function of (p, add-sequence)
//     through the hash pipeline, and is the CROSS-LANGUAGE oracle -- all five
//     ports MUST produce the byte-identical array. The shared JSON scenarios
//     assert ONLY this (via register_hex, NonzeroRegisters, MaxRegister,
//     register_at_N).
//
//  2. The float64 Estimate. It is a function of ln / 2^x / division / summation
//     and CANNOT be required to agree bit-for-bit across five libm
//     implementations. It is specified precisely here and tested NATIVELY
//     against a documented tolerance -- it is NEVER in the shared oracle. There
//     is no estimate assertion key.
//
// Add, Merge, and the register array use ZERO floating point (only Hash64,
// shifts, max, byte packing); the float appears only inside Estimate(), a
// read-only projection that never writes a register.
package hyperloglog

import (
	"fmt"
	"iter"
	"math"
	"math/bits"

	"github.com/mapdb/mapdb-golang/hash"
)

// MinPrecision is the minimum legal precision (m = 16).
const MinPrecision uint8 = 4

// MaxPrecision is the maximum legal precision (m = 262144); the v1 ceiling
// matching hll_split.
const MaxPrecision uint8 = 18

// magic is the 4-byte ASCII magic that version-tags the serialized form
// ("HLL1").
var magic = [4]byte{'H', 'L', 'L', '1'}

// HyperLogLog is a HyperLogLog distinct-count sketch.
//
// Built by NewHyperLogLogWithPrecision; updated by Add / Merge; the register
// array (the oracle) is read via Registers / NonzeroRegisters / MaxRegister;
// the quarantined float answer is Estimate; serialized via ToBytes /
// HyperLogLogFromBytes.
type HyperLogLog struct {
	p uint8
	// registers holds m = 2^p registers, each the max rho seen for that index
	// (0 = empty).
	registers []uint8
}

// NewHyperLogLogWithPrecision constructs an empty sketch with m = 2^p zeroed
// registers.
//
// p must be in [4, 18]; otherwise an error is returned (never a silent clamp --
// a clamp would let two ports build differently-sized arrays from the same
// nominal p).
func NewHyperLogLogWithPrecision(p uint8) (*HyperLogLog, error) {
	if p < MinPrecision || p > MaxPrecision {
		return nil, badPrecisionError(p)
	}
	m := 1 << p
	return &HyperLogLog{p: p, registers: make([]uint8, m)}, nil
}

// Precision returns the precision p (log2(m)).
func (h *HyperLogLog) Precision() uint8 { return h.p }

// RegisterCount returns the register count m = 2^p.
func (h *HyperLogLog) RegisterCount() int { return len(h.registers) }

// rhoCeiling is the per-p maximum possible rho (and the FromBytes byte
// ceiling): 64 - p + 1.
func rhoCeiling(p uint8) uint8 { return 64 - p + 1 }

// split derives (register_index, rho) from a 64-bit hash per the hash
// pipeline's pre-stated hll_split (top p bits -> index; remaining bits + guard
// bit -> clz64 + 1). Pure integer; the guard bit guarantees w != 0 so
// LeadingZeros64 is never called on 0.
func split(x uint64, p uint8) (uint32, uint8) {
	pp := uint(p)
	idx := uint32(x >> (64 - pp))
	// GUARD BIT: OR in 1 << (p - 1). If the remaining 64 - p bits are all zero,
	// w = 1 << (p - 1), clz64(w) = 64 - p, so rho = 64 - p + 1 (its max) and
	// clz64 is never invoked on 0.
	w := (x << pp) | (uint64(1) << (pp - 1))
	rho := uint8(bits.LeadingZeros64(w) + 1)
	return idx, rho
}

// Add adds an int32 element. The item is encoded with the hash pipeline's int32
// rule -- reinterpret to uint32, ZERO-extend to uint64 (NOT sign-extend) --
// then Hash64(word, seed = 0), then hll_split, then
// register[idx] = max(register[idx], rho). Pure integer, zero floating point.
func (h *HyperLogLog) Add(item int32) {
	// int32 -> uint32 reinterpret -> zero-extend to uint64 (high 32 bits always 0).
	inputWord := uint64(uint32(item))
	x := hash.Hash64(inputWord, 0)
	idx, rho := split(x, h.p)
	if rho > h.registers[idx] {
		h.registers[idx] = rho
	}
}

// AddSeq adds every item the sequence yields — the absorber analogue of the
// collection AddSeq protocol, so a sketch can be fed directly from any
// iter.Seq[int32] source without an intermediate slice, e.g. h.AddSeq(set.All()).
func (h *HyperLogLog) AddSeq(seq iter.Seq[int32]) {
	for item := range seq {
		h.Add(item)
	}
}

// Registers returns a copy of the register array (the cross-language oracle
// bytes). A copy — not the internal slice — so callers cannot mutate the sketch
// into states HyperLogLogFromBytes rejects (matches every sibling accessor:
// countmin.ToCounters, fenwick.CanonicalTree, bloom.ToBytes, roaring.ToSortedSlice).
func (h *HyperLogLog) Registers() []uint8 {
	out := make([]uint8, len(h.registers))
	copy(out, h.registers)
	return out
}

// NonzeroRegisters returns the count of registers > 0 (= m - V, where V is the
// zero count).
func (h *HyperLogLog) NonzeroRegisters() uint32 {
	var n uint32
	for _, r := range h.registers {
		if r > 0 {
			n++
		}
	}
	return n
}

// MaxRegister returns the maximum register value (largest rho seen); 0 for a
// fresh sketch.
func (h *HyperLogLog) MaxRegister() uint8 {
	var m uint8
	for _, r := range h.registers {
		if r > m {
			m = r
		}
	}
	return m
}

// Merge merges other into h by element-wise register max (the union's register
// j is the max over both input sets). Requires identical p (else an error).
// Commutative, associative, idempotent. Pure integer, zero floating point.
func (h *HyperLogLog) Merge(other *HyperLogLog) error {
	if h.p != other.p {
		return precisionMismatchError(h.p, other.p)
	}
	for i, b := range other.registers {
		if b > h.registers[i] {
			h.registers[i] = b
		}
	}
	return nil
}

// Estimate returns the distinct cardinality (the QUARANTINED float64,
// native-only and tolerance-tested -- never in the shared oracle).
//
// Original HyperLogLog estimator (Flajolet-Fusy-Gandouet-Meunier 2007) with the
// small-range linear-counting correction and the 2^64 large-range correction
// (this HLL consumes a 64-bit Hash64, so the hash space is 2^64, NOT the 2007
// paper's 2^32).
func (h *HyperLogLog) Estimate() float64 {
	m := float64(len(h.registers))
	alpha := alphaM(h.p)

	// z = sum 2^(-register[j]); register[j] == 0 contributes 2^0 = 1.
	// 1 << register[j] (<= 1 << 61) fits a uint64; compute the shift in integer.
	var z float64
	var v int // zero-register count
	for _, r := range h.registers {
		if r == 0 {
			v++
		}
		z += 1.0 / float64(uint64(1)<<r)
	}
	e := alpha * m * m / z

	// Small-range (linear counting): E small AND there are empties (V > 0).
	if e <= 2.5*m && v > 0 {
		return m * math.Log(m/float64(v))
	}

	// Large-range correction near the HASH-SPACE ceiling (2^64, NOT 2^32).
	const two64 = 18446744073709551616.0 // 2^64, exactly representable.
	if e > (1.0/30.0)*two64 {
		// Guard the log argument: for all reachable states E < 2^64 so
		// (1 - E/2^64) > 0. A fully-saturated deserialized state (every
		// register at the per-p ceiling, constructible via UnmarshalBinary but
		// not via Add) can push raw E >= 2^64, making (1 - E/2^64) <= 0 and
		// Log(<= 0) = NaN. Skip the log correction there and return the raw
		// (large, finite) E so Estimate() stays finite as the spec mandates.
		if e < two64 {
			return -two64 * math.Log(1.0-e/two64)
		}
	}

	return e
}

// alphaM returns the HLL bias constant alpha_m: pinned piecewise literals for
// small m, closed form for m >= 128.
func alphaM(p uint8) float64 {
	switch p {
	case 4:
		return 0.673 // m = 16
	case 5:
		return 0.697 // m = 32
	case 6:
		return 0.709 // m = 64
	default:
		return 0.7213 / (1.0 + 1.079/float64(uint64(1)<<p)) // m >= 128
	}
}

// ToBytes serializes to the v1 wire form: 5-byte header ("HLL1" + p) followed by
// one uint8 per register in index order. Total length 5 + 2^p.
func (h *HyperLogLog) ToBytes() []byte {
	out := make([]byte, 0, 5+len(h.registers))
	out = append(out, magic[:]...)
	out = append(out, h.p)
	out = append(out, h.registers...)
	return out
}

// HyperLogLogFromBytes deserializes from the v1 wire form. Rejects (single MUST
// rule so no two ports disagree on validity): too short, bad magic, p out of
// range, length != 5 + 2^p, or any register byte > 64 - p + 1.
func HyperLogLogFromBytes(b []byte) (*HyperLogLog, error) {
	if len(b) < 5 {
		return nil, tooShortError(len(b))
	}
	var m4 [4]byte
	copy(m4[:], b[0:4])
	if m4 != magic {
		return nil, badMagicError(m4)
	}
	p := b[4]
	if p < MinPrecision || p > MaxPrecision {
		return nil, badPrecisionError(p)
	}
	m := 1 << p
	expected := 5 + m
	if len(b) != expected {
		return nil, lengthMismatchError(expected, len(b))
	}
	ceiling := rhoCeiling(p)
	regs := b[5:]
	for i, r := range regs {
		if r > ceiling {
			return nil, registerOutOfRangeError(i, r, ceiling)
		}
	}
	out := make([]uint8, m)
	copy(out, regs)
	return &HyperLogLog{p: p, registers: out}, nil
}

// ---- Errors --------------------------------------------------------------

func badPrecisionError(p uint8) error {
	return fmt.Errorf("precision %d out of range %d..=%d", p, MinPrecision, MaxPrecision)
}

func precisionMismatchError(left, right uint8) error {
	return fmt.Errorf("merge precision mismatch: %d != %d", left, right)
}

func tooShortError(n int) error {
	return fmt.Errorf("serialized HLL too short: %d bytes (need >= 5)", n)
}

func badMagicError(m [4]byte) error {
	return fmt.Errorf("bad HLL magic: %02x (expected \"HLL1\")", m)
}

func lengthMismatchError(expected, got int) error {
	return fmt.Errorf("HLL length mismatch: expected %d, got %d", expected, got)
}

func registerOutOfRangeError(index int, value, max uint8) error {
	return fmt.Errorf("register[%d] = %d exceeds per-p ceiling %d", index, value, max)
}
