// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Package bloom is a deterministic Bloom filter (approximate set membership)
// riding the cross-language hash pipeline (see spec/features/bloom.md in
// mapdb-collection-spec).
//
// It is the first end-user collection of the probabilistic wave. It rides
// directly on the deterministic hash pipeline (package hash): it uses that
// module's Positions(input, m, k) (Kirsch-Mitzenmacher double hashing) to pick
// k bit indices in an m-bit array. Because Positions is bit-identical across all
// five ports, the bit array after a given add-sequence is bit-identical too --
// that bit array is the cross-language oracle.
//
// Element encoding (the critical trap): an i32 element v is reinterpreted to
// uint32, encoded to 4 little-endian bytes, and fed to the hash pipeline's
// BYTE-INPUT Positions path -- the exact path the 12-hash-pipeline/positions_*
// scenarios drive, which folds in the byte length. This is NOT the scalar
// Hash32Int32 path: for v=7 the byte-path input word is 0x07^4 = 0x03. Worked
// example: NewBloomWithParams(16, 4) then Add(7) lights bits {0, 2, 7, 9} ->
// ToBytes() = [0x85, 0x02] -> "0x8502", BitCount() == 4.
//
// Guarantees:
//   - No false negative. Add(v) then MightContain(v) is always true.
//   - Idempotent / order-independent. The bit array depends only on the SET of
//     added elements.
//   - Deterministic. Identical (m, k) + add-sequence ==> identical bits on all
//     five ports.
package bloom

import (
	"iter"
	"math"
	"math/bits"

	"github.com/mapdb/mapdb-golang/hash"
)

// ln2 is the natural log of 2 in f64 (the same constant the reference uses for
// the optimal() float derivation; native-test-only).
const ln2 = math.Ln2

// Bloom is a Bloom filter over i32 elements with m bits and k hash functions,
// both fixed at construction. The bit array is stored as a []uint64 word array;
// the internal word width is not observable -- only ToBytes (LSB-first,
// ascending bytes) is the cross-language form.
type Bloom struct {
	// mBits is the number of bits in the array (m, 1 ..= 2^32-1).
	mBits uint32
	// k is the number of hash functions / positions set per element.
	k uint32
	// words is the bit array, ceil(m/64) words; bit i lives in word i/64 at bit
	// position i%64 (LSB-first within a word).
	words []uint64
}

// NewBloomWithParams is the canonical, fully-deterministic constructor: explicit
// bit count mBits and hash count k. The filter starts empty (all bits 0).
//
// mBits == 0 is invalid and panics (a 0-bit array can hold nothing and every
// Positions modulo would be by zero). k == 0 is degenerate but legal (see
// MightContain).
func NewBloomWithParams(mBits, k uint32) *Bloom {
	if mBits == 0 {
		panic("bloom: NewBloomWithParams: mBits must be >= 1")
	}
	nWords := (uint64(mBits) + 63) / 64
	return &Bloom{
		mBits: mBits,
		k:     k,
		words: make([]uint64, nWords),
	}
}

// Optimal is a convenience constructor sizing the filter from an expected
// element count n and a target false-positive probability p, using the standard
// Bloom formulas:
//
//	m = ceil( -n * ln(p) / (ln 2)^2 )
//	k = max( 1, round( (m / n) * ln 2 ) )     # round-half-away-from-zero
//
// then delegates to NewBloomWithParams. This is NATIVE-TEST-ONLY: it never
// appears in the shared cross-language scenarios (the float derivation could
// drift by a ULP across libms -- quarantined to native tests against the pinned
// integer table in spec/features/bloom.md).
//
// Requires n >= 1 and 0 < p < 1. n == 0, p <= 0, p >= 1, NaN, and +/-Infinity
// are invalid and panic (they would divide by zero, take ln of a non-positive
// value, or yield a non-finite m).
func Optimal(n uint64, p float64) *Bloom {
	if n < 1 {
		panic("bloom: Optimal: n must be >= 1")
	}
	if math.IsNaN(p) || math.IsInf(p, 0) || p <= 0.0 || p >= 1.0 {
		panic("bloom: Optimal: p must be finite and in (0, 1)")
	}
	nf := float64(n)
	mF := math.Ceil(-nf * math.Log(p) / (ln2 * ln2))
	if math.IsNaN(mF) || math.IsInf(mF, 0) || mF < 1.0 || mF > float64(math.MaxUint32) {
		panic("bloom: Optimal: derived m out of range")
	}
	m := uint32(mF)
	// round-half-away-from-zero (Go's math.Round), clamped to >= 1.
	kF := math.Round((float64(m) / nf) * ln2)
	k := uint32(1)
	if kF > 1.0 {
		k = uint32(kF)
	}
	return NewBloomWithParams(m, k)
}

// MBits returns the bit count m.
func (b *Bloom) MBits() uint32 { return b.mBits }

// K returns the hash count k.
func (b *Bloom) K() uint32 { return b.k }

// Add adds an i32 element: sets the k bits for v (idempotent). With k == 0 this
// sets no bits.
func (b *Bloom) Add(v int32) {
	enc := encodeInt32(v)
	for _, p := range hash.Positions(enc, b.mBits, b.k) {
		b.setBit(p)
	}
}

// AddSeq adds every value the sequence yields — the absorber analogue of the
// collection AddSeq protocol, so a Bloom filter can be filled directly from any
// iter.Seq[int32] source without an intermediate slice, e.g.
// b.AddSeq(set.All()) or b.AddSeq(seq.Range(0, n)).
func (b *Bloom) AddSeq(seq iter.Seq[int32]) {
	for v := range seq {
		b.Add(v)
	}
}

// MightContain is the canonical membership test. Returns false ==> definitely
// absent; true ==> possibly present (may be a false positive). NEVER returns
// false for an element that was added (no false negative).
//
// With k == 0 the AND over zero positions is vacuously true, so this returns
// true for every element (an all-false-positive filter).
func (b *Bloom) MightContain(v int32) bool {
	enc := encodeInt32(v)
	for _, p := range hash.Positions(enc, b.mBits, b.k) {
		if !b.getBit(p) {
			return false
		}
	}
	return true
}

// Contains is the idiomatic alias for MightContain (what the JSON suite's
// contains_<v> keys probe). The result is APPROXIMATE membership.
func (b *Bloom) Contains(v int32) bool { return b.MightContain(v) }

// IsEmpty reports whether no bit is set (equivalently: nothing has been added,
// or only k == 0 adds). Equal to BitCount() == 0.
func (b *Bloom) IsEmpty() bool {
	for _, w := range b.words {
		if w != 0 {
			return false
		}
	}
	return true
}

// BitCount returns the number of set bits (popcount of the whole bit array). The
// zeroed tail bits never contribute (no Positions index reaches them).
func (b *Bloom) BitCount() uint32 {
	var c uint32
	for _, w := range b.words {
		c += uint32(bits.OnesCount64(w))
	}
	return c
}

// Union returns the bitwise OR of two filters with identical (m, k), as a new
// filter. The result's membership is the union of the two filters' membership
// (no false negatives lost).
//
// Mismatched (m, k) panics (a filter built with different parameters has an
// incompatible bit array; ORing them is meaningless).
func (b *Bloom) Union(other *Bloom) *Bloom {
	if b.mBits != other.mBits || b.k != other.k {
		panic("bloom: Union: parameter mismatch")
	}
	words := make([]uint64, len(b.words))
	for i := range b.words {
		words[i] = b.words[i] | other.words[i]
	}
	return &Bloom{mBits: b.mBits, k: b.k, words: words}
}

// ToBytes is the serialized bit array (spec/features/bloom.md §"Serialized
// bit-array form"): length exactly ceil(m/8) bytes; LSB-first bit order within
// each byte (bit i ==> byte[i/8] |= 1 << (i%8)); ascending byte order;
// little-endian on every host; unused tail bits 0.
func (b *Bloom) ToBytes() []byte {
	nBytes := int((uint64(b.mBits) + 7) / 8)
	out := make([]byte, nBytes)
	for wi, w := range b.words {
		// Each word holds bits [wi*64 .. wi*64 + 64). Emit its 8 bytes
		// little-endian so bit (wi*64 + bi*8 + j) lands at out[...] & (1<<j).
		for bi := 0; bi < 8; bi++ {
			outIdx := wi*8 + bi
			if outIdx >= nBytes {
				// outIdx >= nBytes can only be a fully-zero tail byte (no
				// Positions index reaches >= m), so dropping it is exact.
				break
			}
			out[outIdx] = byte(w >> (8 * uint(bi)))
		}
	}
	return out
}

// SetBits returns the sorted-ascending indices of the set bits -- a
// human-legible alternate oracle to ToBytes (drives the set_bits assertion).
func (b *Bloom) SetBits() []uint32 {
	out := make([]uint32, 0, b.BitCount())
	for wi, w := range b.words {
		for w != 0 {
			j := uint32(bits.TrailingZeros64(w))
			out = append(out, uint32(wi)*64+j)
			w &= w - 1 // clear lowest set bit
		}
	}
	return out
}

// encodeInt32 reinterprets an i32 to u32 (two's-complement bit reinterpret, NOT
// sign-extend) and returns the 4 little-endian bytes the hash pipeline's
// byte-input Positions path consumes.
func encodeInt32(v int32) []byte {
	w := uint32(v)
	return []byte{byte(w), byte(w >> 8), byte(w >> 16), byte(w >> 24)}
}

func (b *Bloom) setBit(i uint32) {
	b.words[i/64] |= 1 << (i % 64)
}

func (b *Bloom) getBit(i uint32) bool {
	return (b.words[i/64]>>(i%64))&1 == 1
}
