// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Package hash is the deterministic, byte-exact, cross-language hash pipeline
// (see spec/features/hash-pipeline.md in mapdb-collection-spec).
//
// This is a port of the frozen reference (mapdb-rust src/hash.rs) whose entire
// contract is bit-exactness across all five language ports: every (input, seed)
// produces the identical Hash32 / Hash64 / Positions bits in Rust, Go,
// TypeScript, Zig and Java. It is a separate, additive module -- it does NOT
// touch the collections' bucket hash (algorithms.md "Hash function"), which
// keeps its native-hash carve-out. This module has no carve-out.
//
//   - Hash32 -- MurmurHash3 32-bit finalizer (fmix32) over the input word XOR'd
//     with a 32-bit fold of the 64-bit seed.
//   - Hash64 -- MurmurHash3 64-bit finalizer (fmix64, constants
//     0xff51afd7ed558ccd / 0xc4ceb9fe1a85ec53, shifts 33/33/33 -- NOT the
//     SplitMix64 generator's final mix) over the input word XOR'd with the seed.
//   - Positions -- Kirsch-Mitzenmacher double hashing (h1 + i*h2 mod m), all
//     32-bit wrapping, unsigned modulo.
//   - HllSplit -- pre-stated (register_index, leading_zero_run) split for the
//     later HyperLogLog feature.
//
// All multiplies are two's-complement wrapping at their declared width; all
// right shifts are logical (unsigned). Go's uint32/uint64 arithmetic wraps
// natively and ">>" on unsigned values is a logical shift, so both are free.
package hash

import "math/bits"

// MurmurHash3 fmix32 finalizer constants (the published values).
const (
	fmix32C1 uint32 = 0x85ebca6b
	fmix32C2 uint32 = 0xc2b2ae35
)

// MurmurHash3 fmix64 finalizer constants (the published values). NOTE: these
// are the MurmurHash3 fmix64 constants with three 33-bit shifts, NOT the
// SplitMix64 *generator's* final mix (0xbf58476d1ce4e5b9 / 0x94d049bb133111eb,
// shifts 30/27/31) -- a different function with different bits.
const (
	fmix64C1 uint64 = 0xff51afd7ed558ccd
	fmix64C2 uint64 = 0xc4ceb9fe1a85ec53
)

// SALT2 is the fixed 32-bit salt for the second base hash of the double-hashing
// position scheme (the 32-bit golden-ratio prime). Distinct from the 64-bit
// collection Fibonacci constant 0x9E3779B97F4A7C15.
const SALT2 uint64 = 0x9e3779b1

// Hash32 is the 32-bit named hash: the MurmurHash3 fmix32 finalizer applied to
// one 32-bit lane derived from word and a 32-bit fold of the 64-bit seed.
//
// The seed is folded with seed ^ (seed >> 32) so two seeds differing only in
// their high 32 bits still produce different hashes. Seed 0 is an ordinary seed
// (XOR'd in; no special case).
func Hash32(word uint32, seed uint64) uint32 {
	// Fold the full 64-bit seed into one 32-bit lane (low XOR high).
	seed32 := uint32(seed ^ (seed >> 32))
	h := word ^ seed32
	h ^= h >> 16
	h *= fmix32C1
	h ^= h >> 13
	h *= fmix32C2
	h ^= h >> 16
	return h
}

// Hash64 is the 64-bit named hash: the MurmurHash3 fmix64 finalizer applied to
// word ^ seed. The seed is mixed in first as a 64-bit integer (no endianness,
// no special case for seed 0).
func Hash64(word uint64, seed uint64) uint64 {
	h := word ^ seed
	h ^= h >> 33
	h *= fmix64C1
	h ^= h >> 33
	h *= fmix64C2
	h ^= h >> 33
	return h
}

// Hash64Hi is the high 32-bit lane of Hash64 (pins the TypeScript hi/lo lane
// split).
func Hash64Hi(word uint64, seed uint64) uint32 {
	return uint32(Hash64(word, seed) >> 32)
}

// Hash64Lo is the low 32-bit lane of Hash64.
func Hash64Lo(word uint64, seed uint64) uint32 {
	return uint32(Hash64(word, seed))
}

// ---- Per-type input-word encoders ----------------------------------------

// EncodeInt32Word32 encodes an i32 element to the Hash32 input word: a
// two's-complement bit reinterpret to uint32 (NOT a sign-extend).
func EncodeInt32Word32(value int32) uint32 {
	return uint32(value)
}

// EncodeInt32Word64 encodes an i32 element to the Hash64 input word:
// reinterpret to uint32 then ZERO-extend to uint64 (so the high 32 bits are
// always 0; the seed supplies the high-word entropy). NOT a sign-extend.
func EncodeInt32Word64(value int32) uint64 {
	return uint64(uint32(value))
}

// EncodeBytesWord32 folds a raw byte slice into the Hash32 input word: read 4
// bytes at a time as little-endian uint32 lanes, XOR-combine, zero-pad a
// sub-lane tail to the low bytes, then XOR in len(bytes) mod 2^32.
func EncodeBytesWord32(b []byte) uint32 {
	var word uint32
	n := len(b)
	full := n - n%4
	for i := 0; i < full; i += 4 {
		word ^= uint32(b[i]) | uint32(b[i+1])<<8 | uint32(b[i+2])<<16 | uint32(b[i+3])<<24
	}
	if full < n {
		// Tail goes in the LOW bytes of its lane; remaining high bytes are 0.
		var lane uint32
		for j := 0; full+j < n; j++ {
			lane |= uint32(b[full+j]) << (8 * uint(j))
		}
		word ^= lane
	}
	// Length reduced mod 2^32 before the XOR.
	return word ^ uint32(n)
}

// EncodeBytesWord64 folds a raw byte slice into the Hash64 input word: read 8
// bytes at a time as little-endian uint64 lanes, XOR-combine, zero-pad a
// sub-lane tail to the low bytes, then XOR in len(bytes) mod 2^64.
func EncodeBytesWord64(b []byte) uint64 {
	var word uint64
	n := len(b)
	full := n - n%8
	for i := 0; i < full; i += 8 {
		word ^= uint64(b[i]) | uint64(b[i+1])<<8 | uint64(b[i+2])<<16 | uint64(b[i+3])<<24 |
			uint64(b[i+4])<<32 | uint64(b[i+5])<<40 | uint64(b[i+6])<<48 | uint64(b[i+7])<<56
	}
	if full < n {
		var lane uint64
		for j := 0; full+j < n; j++ {
			lane |= uint64(b[full+j]) << (8 * uint(j))
		}
		word ^= lane
	}
	return word ^ uint64(n)
}

// Hash32Int32 returns Hash32 of an i32 element (reinterpret encoding).
func Hash32Int32(value int32, seed uint64) uint32 {
	return Hash32(EncodeInt32Word32(value), seed)
}

// Hash32Bytes returns Hash32 of a raw byte slice (little-endian fold encoding).
func Hash32Bytes(b []byte, seed uint64) uint32 {
	return Hash32(EncodeBytesWord32(b), seed)
}

// Hash64Int32 returns Hash64 of an i32 element (reinterpret + zero-extend
// encoding).
func Hash64Int32(value int32, seed uint64) uint64 {
	return Hash64(EncodeInt32Word64(value), seed)
}

// Hash64Bytes returns Hash64 of a raw byte slice (little-endian fold encoding).
func Hash64Bytes(b []byte, seed uint64) uint64 {
	return Hash64(EncodeBytesWord64(b), seed)
}

// ---- Derived positions (Kirsch-Mitzenmacher double hashing) --------------

// PositionsFromHashes derives k array positions over a table of size m from two
// base hashes h1/h2, combined linearly: p_i = (h1 + i*h2) mod m, all 32-bit
// wrapping, unsigned modulo. Returned in derivation order p_0 .. p_{k-1}.
//
// This is the inner function the test-vector oracle is stated on; it is
// independent of the Hash32 layer and the byte encoding.
func PositionsFromHashes(h1, h2, m, k uint32) []uint32 {
	out := make([]uint32, 0, k)
	for i := uint32(0); i < k; i++ {
		combined := h1 + i*h2 // 32-bit wrapping add + multiply
		out = append(out, combined%m)
	}
	return out
}

// Positions derives k array positions for input over a table of size m using
// Kirsch-Mitzenmacher double hashing. h1 = Hash32(input, 0),
// h2 = Hash32(input, SALT2); then PositionsFromHashes.
func Positions(input []byte, m, k uint32) []uint32 {
	h1 := Hash32Bytes(input, 0)
	h2 := Hash32Bytes(input, SALT2)
	return PositionsFromHashes(h1, h2, m, k)
}

// ---- HyperLogLog split (pre-stated for the HLL feature) ------------------

// HllSplit is the pre-stated HyperLogLog split: from a single 64-bit hash,
// derive a (register_index, leading_zero_run) pair. p = log2(number of
// registers), 4 <= p <= 18. Only Hash64(input, 0) is locked here; HLL itself
// is a separate later feature.
//
//   - idx = the top p bits of the hash (the register index).
//   - rho = clz64(w) + 1, the 1-based leading-zero run of the remaining bits
//     shifted up with a guard bit set at position p - 1.
func HllSplit(input []byte, p uint32) (idx uint32, rho uint32) {
	x := Hash64Bytes(input, 0)
	idx = uint32(x >> (64 - p))
	w := (x << p) | (uint64(1) << (p - 1))
	rho = uint32(bits.LeadingZeros64(w)) + 1
	return idx, rho
}
