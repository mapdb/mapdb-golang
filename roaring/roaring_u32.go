// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Package roaring implements RoaringU32, a sparse, compressed 32-bit integer
// set (a Roaring-style bitmap). See spec/features/roaring-u32.md.
//
// The universe (2^32 values) is split into 2^16 chunks keyed by the high 16
// bits of a value. Each non-empty chunk is stored as a container: an ARRAY
// (sorted distinct uint16 slice) for cardinality 1..=4096, or a BITMAP
// (1024 x uint64) for cardinality 4097..=65536. The container type is a pure
// function of the chunk's current cardinality (history-independent), which
// makes the serialized form canonical.
//
// Ordering is UNSIGNED uint32 ascending throughout (iteration, Min/Max,
// serialized chunk order). An int32 element is bit-reinterpreted to uint32 (not
// sign-extended), so int32 -1 is 0xFFFFFFFF and sorts last.
package roaring

import (
	"errors"
	"fmt"
	"math/bits"
	"sort"
)

// arrayMax is the cardinality at and below which a chunk is an ARRAY; above
// which it is a BITMAP. 4096 is the classic Roaring break-even
// (4096 * 2 bytes == 8192, the bitmap size). c <= 4096 => ARRAY, c > 4096 =>
// BITMAP.
const arrayMax = 4096

// bitmapWords is the fixed BITMAP container size: 1024 uint64 words (2^16 bits).
const bitmapWords = 1024

// magic is the serialized header magic 0x32523055 (LE bytes 55 30 52 32).
const magic = 0x32523055

// version is the serialized format version.
const version = 1

const (
	tagArray  = 0x01
	tagBitmap = 0x02
)

// containerKind distinguishes the canonical container representation.
type containerKind uint8

const (
	kindArray containerKind = iota
	kindBitmap
)

// container is a per-chunk container. Exactly one of array/words is meaningful,
// selected by kind. The kind is always the canonical type for the contained
// cardinality (ARRAY for 1..=4096, BITMAP for 4097..=65536).
type container struct {
	kind containerKind
	// array holds sorted, distinct low-16-bit keys (length == cardinality).
	array []uint16
	// words is the dense bitmap: bit (w*64+b) is low key w*64+b.
	words []uint64
	// count is the cached BITMAP popcount (== cardinality). Unused for ARRAY.
	count uint32
}

// cardinality reports the number of present low keys (1..=65536).
func (c *container) cardinality() uint32 {
	if c.kind == kindArray {
		return uint32(len(c.array))
	}
	return c.count
}

func (c *container) contains(low uint16) bool {
	if c.kind == kindArray {
		i := sort.Search(len(c.array), func(i int) bool { return c.array[i] >= low })
		return i < len(c.array) && c.array[i] == low
	}
	w, b := int(low)>>6, int(low)&63
	return c.words[w]&(1<<uint(b)) != 0
}

// add inserts low. Returns whether the container changed. The caller converts
// ARRAY -> BITMAP if the cardinality rises above arrayMax.
func (c *container) add(low uint16) bool {
	if c.kind == kindArray {
		i := sort.Search(len(c.array), func(i int) bool { return c.array[i] >= low })
		if i < len(c.array) && c.array[i] == low {
			return false
		}
		c.array = append(c.array, 0)
		copy(c.array[i+1:], c.array[i:])
		c.array[i] = low
		return true
	}
	w, b := int(low)>>6, int(low)&63
	bit := uint64(1) << uint(b)
	if c.words[w]&bit == 0 {
		c.words[w] |= bit
		c.count++
		return true
	}
	return false
}

// remove deletes low. Returns whether the container changed. The caller converts
// BITMAP -> ARRAY if the cardinality drops to arrayMax or below, and drops the
// whole chunk if the cardinality hits zero.
func (c *container) remove(low uint16) bool {
	if c.kind == kindArray {
		i := sort.Search(len(c.array), func(i int) bool { return c.array[i] >= low })
		if i < len(c.array) && c.array[i] == low {
			c.array = append(c.array[:i], c.array[i+1:]...)
			return true
		}
		return false
	}
	w, b := int(low)>>6, int(low)&63
	bit := uint64(1) << uint(b)
	if c.words[w]&bit != 0 {
		c.words[w] &^= bit
		c.count--
		return true
	}
	return false
}

// lows returns all present low keys in unsigned ascending order.
func (c *container) lows() []uint16 {
	if c.kind == kindArray {
		out := make([]uint16, len(c.array))
		copy(out, c.array)
		return out
	}
	out := make([]uint16, 0, c.count)
	for w, word := range c.words {
		for word != 0 {
			b := bits.TrailingZeros64(word)
			out = append(out, uint16(w*64+b))
			word &= word - 1
		}
	}
	return out
}

// minLow returns the minimum present low key (unsigned).
func (c *container) minLow() uint16 {
	if c.kind == kindArray {
		return c.array[0]
	}
	for w, word := range c.words {
		if word != 0 {
			return uint16(w*64 + bits.TrailingZeros64(word))
		}
	}
	panic("non-empty bitmap has a set bit")
}

// maxLow returns the maximum present low key (unsigned).
func (c *container) maxLow() uint16 {
	if c.kind == kindArray {
		return c.array[len(c.array)-1]
	}
	for w := len(c.words) - 1; w >= 0; w-- {
		if c.words[w] != 0 {
			return uint16(w*64 + (63 - bits.LeadingZeros64(c.words[w])))
		}
	}
	panic("non-empty bitmap has a set bit")
}

// bitmapFromLows builds a BITMAP container from a sorted low-key list.
func bitmapFromLows(lows []uint16) *container {
	words := make([]uint64, bitmapWords)
	for _, low := range lows {
		w, b := int(low)>>6, int(low)&63
		words[w] |= uint64(1) << uint(b)
	}
	return &container{kind: kindBitmap, words: words, count: uint32(len(lows))}
}

// arrayFromLows wraps a sorted, distinct, non-empty low-key list as an ARRAY.
func arrayFromLows(lows []uint16) *container {
	return &container{kind: kindArray, array: lows}
}

// canonicalFromLows normalizes a low-key list (assumed sorted, distinct,
// non-empty) into the canonical container for its cardinality.
func canonicalFromLows(lows []uint16) *container {
	if len(lows) <= arrayMax {
		return arrayFromLows(lows)
	}
	return bitmapFromLows(lows)
}

// chunk is a (high key, container) pair. The owning set keeps chunks in
// strictly-ascending high-key order with no empty containers.
type chunk struct {
	high uint16
	c    *container
}

// RoaringU32 is a sparse, compressed set of uint32 values (Roaring-style
// bitmap).
//
// int32 elements are bit-reinterpreted to uint32 (-1 -> 0xFFFFFFFF) before
// being split into a 16-bit high key (chunk) and 16-bit low key. Ordering is
// unsigned uint32 ascending throughout.
type RoaringU32 struct {
	// chunks holds non-empty chunks in unsigned high-key ascending order;
	// invariant: strictly ascending, no empty containers.
	chunks []chunk
}

func split(value uint32) (uint16, uint16) {
	return uint16(value >> 16), uint16(value & 0xFFFF)
}

func join(high, low uint16) uint32 {
	return uint32(high)<<16 | uint32(low)
}

// NewRoaringU32 returns an empty set.
func NewRoaringU32() *RoaringU32 {
	return &RoaringU32{}
}

// find locates the chunk index for high. found reports presence; when absent,
// i is the insertion point preserving ascending order.
func (s *RoaringU32) find(high uint16) (i int, found bool) {
	i = sort.Search(len(s.chunks), func(i int) bool { return s.chunks[i].high >= high })
	if i < len(s.chunks) && s.chunks[i].high == high {
		return i, true
	}
	return i, false
}

// Add inserts value. Returns whether the set changed (was newly added).
func (s *RoaringU32) Add(value uint32) bool {
	high, low := split(value)
	i, found := s.find(high)
	if !found {
		s.chunks = append(s.chunks, chunk{})
		copy(s.chunks[i+1:], s.chunks[i:])
		s.chunks[i] = chunk{high: high, c: arrayFromLows([]uint16{low})}
		return true
	}
	changed := s.chunks[i].c.add(low)
	if changed && s.chunks[i].c.cardinality() > arrayMax {
		// ARRAY -> BITMAP up-conversion at cardinality 4097.
		if s.chunks[i].c.kind == kindArray {
			s.chunks[i].c = bitmapFromLows(s.chunks[i].c.array)
		}
	}
	return changed
}

// Remove deletes value. Returns whether the set changed (was present).
func (s *RoaringU32) Remove(value uint32) bool {
	high, low := split(value)
	i, found := s.find(high)
	if !found {
		return false
	}
	if !s.chunks[i].c.remove(low) {
		return false
	}
	card := s.chunks[i].c.cardinality()
	switch {
	case card == 0:
		// Empty-chunk normalization: drop the chunk entirely.
		s.chunks = append(s.chunks[:i], s.chunks[i+1:]...)
	case card <= arrayMax:
		// BITMAP -> ARRAY down-conversion at cardinality 4096.
		if s.chunks[i].c.kind == kindBitmap {
			s.chunks[i].c = arrayFromLows(s.chunks[i].c.lows())
		}
	}
	return true
}

// Contains reports whether value is present.
func (s *RoaringU32) Contains(value uint32) bool {
	high, low := split(value)
	if i, found := s.find(high); found {
		return s.chunks[i].c.contains(low)
	}
	return false
}

// Cardinality returns the logical cardinality (up to 2^32).
func (s *RoaringU32) Cardinality() uint64 {
	var total uint64
	for i := range s.chunks {
		total += uint64(s.chunks[i].c.cardinality())
	}
	return total
}

// IsEmpty reports whether the set is empty.
func (s *RoaringU32) IsEmpty() bool {
	return len(s.chunks) == 0
}

// Clear removes all values (canonical empty set).
func (s *RoaringU32) Clear() {
	s.chunks = nil
}

// ChunkCount returns the number of non-empty chunks (the serialized
// CHUNK_COUNT).
func (s *RoaringU32) ChunkCount() int {
	return len(s.chunks)
}

// Min returns the unsigned minimum present value. ok is false if empty.
func (s *RoaringU32) Min() (value uint32, ok bool) {
	if len(s.chunks) == 0 {
		return 0, false
	}
	first := &s.chunks[0]
	return join(first.high, first.c.minLow()), true
}

// Max returns the unsigned maximum present value. ok is false if empty.
func (s *RoaringU32) Max() (value uint32, ok bool) {
	if len(s.chunks) == 0 {
		return 0, false
	}
	last := &s.chunks[len(s.chunks)-1]
	return join(last.high, last.c.maxLow()), true
}

// ToSortedSlice returns all values in unsigned uint32 ascending order.
func (s *RoaringU32) ToSortedSlice() []uint32 {
	out := make([]uint32, 0, s.Cardinality())
	for i := range s.chunks {
		high := s.chunks[i].high
		for _, low := range s.chunks[i].c.lows() {
			out = append(out, join(high, low))
		}
	}
	return out
}

// Iterate calls fn for each value in unsigned uint32 ascending order, stopping
// early if fn returns false.
func (s *RoaringU32) Iterate(fn func(uint32) bool) {
	for i := range s.chunks {
		high := s.chunks[i].high
		for _, low := range s.chunks[i].c.lows() {
			if !fn(join(high, low)) {
				return
			}
		}
	}
}

// ContainerTypes returns per-chunk container-type tags in chunk order, each
// "array" or "bitmap".
func (s *RoaringU32) ContainerTypes() []string {
	out := make([]string, len(s.chunks))
	for i := range s.chunks {
		if s.chunks[i].c.kind == kindArray {
			out[i] = "array"
		} else {
			out[i] = "bitmap"
		}
	}
	return out
}

// ---- Set algebra (container-granularity, scalar) -------------------------

// combine is the generic chunk-merge driver. merge combines two containers
// sharing a high key into a low-key list; keepA/keepB decide whether an
// only-in-A / only-in-B chunk contributes (a rebuilt copy).
func (s *RoaringU32) combine(
	other *RoaringU32,
	keepA, keepB bool,
	merge func(a, b *container) []uint16,
) *RoaringU32 {
	var chunks []chunk
	i, j := 0, 0
	for i < len(s.chunks) && j < len(other.chunks) {
		ha, hb := s.chunks[i].high, other.chunks[j].high
		switch {
		case ha < hb:
			if keepA {
				chunks = append(chunks, chunk{ha, canonicalFromLows(s.chunks[i].c.lows())})
			}
			i++
		case ha > hb:
			if keepB {
				chunks = append(chunks, chunk{hb, canonicalFromLows(other.chunks[j].c.lows())})
			}
			j++
		default:
			lows := merge(s.chunks[i].c, other.chunks[j].c)
			if len(lows) != 0 {
				chunks = append(chunks, chunk{ha, canonicalFromLows(lows)})
			}
			i++
			j++
		}
	}
	if keepA {
		for ; i < len(s.chunks); i++ {
			chunks = append(chunks, chunk{s.chunks[i].high, canonicalFromLows(s.chunks[i].c.lows())})
		}
	}
	if keepB {
		for ; j < len(other.chunks); j++ {
			chunks = append(chunks, chunk{other.chunks[j].high, canonicalFromLows(other.chunks[j].c.lows())})
		}
	}
	return &RoaringU32{chunks: chunks}
}

// Or returns the union (v in A or v in B) as a new, independent set.
func (s *RoaringU32) Or(other *RoaringU32) *RoaringU32 {
	return s.combine(other, true, true, func(a, b *container) []uint16 {
		return sortedUnion(a.lows(), b.lows())
	})
}

// And returns the intersection (v in A and v in B) as a new, independent set.
func (s *RoaringU32) And(other *RoaringU32) *RoaringU32 {
	return s.combine(other, false, false, func(a, b *container) []uint16 {
		return sortedIntersect(a.lows(), b.lows())
	})
}

// AndNot returns the difference (v in A and v not in B; asymmetric A \ B) as a
// new, independent set.
func (s *RoaringU32) AndNot(other *RoaringU32) *RoaringU32 {
	return s.combine(other, true, false, func(a, b *container) []uint16 {
		return sortedAndNot(a.lows(), b.lows())
	})
}

// Xor returns the symmetric difference (exactly one of A, B) as a new,
// independent set.
func (s *RoaringU32) Xor(other *RoaringU32) *RoaringU32 {
	return s.combine(other, true, true, func(a, b *container) []uint16 {
		return sortedXor(a.lows(), b.lows())
	})
}

// OrInPlace sets s to s | other.
func (s *RoaringU32) OrInPlace(other *RoaringU32) { *s = *s.Or(other) }

// AndInPlace sets s to s & other.
func (s *RoaringU32) AndInPlace(other *RoaringU32) { *s = *s.And(other) }

// AndNotInPlace sets s to s \ other.
func (s *RoaringU32) AndNotInPlace(other *RoaringU32) { *s = *s.AndNot(other) }

// XorInPlace sets s to s ^ other.
func (s *RoaringU32) XorInPlace(other *RoaringU32) { *s = *s.Xor(other) }

// ---- Serialization (little-endian, canonical) ----------------------------

// Serialize returns the canonical little-endian v1 byte image.
func (s *RoaringU32) Serialize() []byte {
	out := make([]byte, 0, 12)
	out = appendU32(out, magic)
	out = appendU16(out, version)
	out = appendU16(out, 0) // RESERVED
	out = appendU32(out, uint32(len(s.chunks)))
	for i := range s.chunks {
		c := s.chunks[i].c
		out = appendU16(out, s.chunks[i].high)
		card := c.cardinality()
		if c.kind == kindArray {
			out = append(out, tagArray, 0) // tag, PAD
			out = appendU16(out, uint16(card-1))
			for _, low := range c.array {
				out = appendU16(out, low)
			}
		} else {
			out = append(out, tagBitmap, 0) // tag, PAD
			out = appendU16(out, uint16(card-1))
			for _, word := range c.words {
				out = appendU64(out, word)
			}
		}
	}
	return out
}

// Deserialize parses a canonical v1 byte image. It returns an error for any
// non-canonical / corrupt / foreign image (see spec reader-MUST-reject rules).
func Deserialize(bytesIn []byte) (*RoaringU32, error) {
	r := &reader{bytes: bytesIn}
	m, err := r.u32()
	if err != nil {
		return nil, err
	}
	if m != magic {
		return nil, fmt.Errorf("bad MAGIC: %#010x", m)
	}
	ver, err := r.u16()
	if err != nil {
		return nil, err
	}
	if ver != version {
		return nil, fmt.Errorf("unsupported VERSION: %d", ver)
	}
	reserved, err := r.u16()
	if err != nil {
		return nil, err
	}
	if reserved != 0 {
		return nil, fmt.Errorf("non-zero RESERVED: %#06x", reserved)
	}
	chunkCount, err := r.u32()
	if err != nil {
		return nil, err
	}
	if chunkCount > 65536 {
		return nil, fmt.Errorf("CHUNK_COUNT > 65536: %d", chunkCount)
	}
	chunks := make([]chunk, 0, chunkCount)
	havePrevHigh := false
	var prevHigh uint16
	for k := uint32(0); k < chunkCount; k++ {
		high, err := r.u16()
		if err != nil {
			return nil, err
		}
		if havePrevHigh && high <= prevHigh {
			return nil, fmt.Errorf("non-ascending or duplicate high key: %#06x after %#06x", high, prevHigh)
		}
		havePrevHigh = true
		prevHigh = high
		tag, err := r.u8()
		if err != nil {
			return nil, err
		}
		pad, err := r.u8()
		if err != nil {
			return nil, err
		}
		if pad != 0 {
			return nil, fmt.Errorf("non-zero PAD: %#04x", pad)
		}
		cardMinus1, err := r.u16()
		if err != nil {
			return nil, err
		}
		card := uint32(cardMinus1) + 1 // CARDINALITY_MINUS_1 + 1
		switch tag {
		case tagArray:
			if card > arrayMax {
				return nil, fmt.Errorf("non-canonical ARRAY cardinality %d (> %d)", card, arrayMax)
			}
			lows := make([]uint16, 0, card)
			havePrev := false
			var prev uint16
			for n := uint32(0); n < card; n++ {
				low, err := r.u16()
				if err != nil {
					return nil, err
				}
				if havePrev && low <= prev {
					return nil, fmt.Errorf("non-ascending or duplicate ARRAY low key: %#06x after %#06x", low, prev)
				}
				havePrev = true
				prev = low
				lows = append(lows, low)
			}
			chunks = append(chunks, chunk{high, arrayFromLows(lows)})
		case tagBitmap:
			if card <= arrayMax {
				return nil, fmt.Errorf("non-canonical BITMAP cardinality %d (<= %d)", card, arrayMax)
			}
			words := make([]uint64, bitmapWords)
			var popcount uint32
			for w := range words {
				word, err := r.u64()
				if err != nil {
					return nil, err
				}
				popcount += uint32(bits.OnesCount64(word))
				words[w] = word
			}
			if popcount != card {
				return nil, fmt.Errorf("BITMAP popcount %d != stored cardinality %d", popcount, card)
			}
			chunks = append(chunks, chunk{high, &container{kind: kindBitmap, words: words, count: card}})
		default:
			return nil, fmt.Errorf("unknown CONTAINER_TYPE tag: %#04x", tag)
		}
	}
	if !r.atEnd() {
		return nil, fmt.Errorf("%d trailing bytes after chunk records", r.remaining())
	}
	return &RoaringU32{chunks: chunks}, nil
}

func appendU16(b []byte, v uint16) []byte {
	return append(b, byte(v), byte(v>>8))
}

func appendU32(b []byte, v uint32) []byte {
	return append(b, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

func appendU64(b []byte, v uint64) []byte {
	return append(b,
		byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
		byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56))
}

// reader is a bounds-checked little-endian byte reader.
type reader struct {
	bytes []byte
	pos   int
}

var errTruncated = errors.New("truncated input")

func (r *reader) take(n int) ([]byte, error) {
	if r.pos+n > len(r.bytes) {
		return nil, fmt.Errorf("%w: need %d bytes at offset %d, have %d", errTruncated, n, r.pos, len(r.bytes)-r.pos)
	}
	s := r.bytes[r.pos : r.pos+n]
	r.pos += n
	return s, nil
}

func (r *reader) u8() (uint8, error) {
	s, err := r.take(1)
	if err != nil {
		return 0, err
	}
	return s[0], nil
}

func (r *reader) u16() (uint16, error) {
	s, err := r.take(2)
	if err != nil {
		return 0, err
	}
	return uint16(s[0]) | uint16(s[1])<<8, nil
}

func (r *reader) u32() (uint32, error) {
	s, err := r.take(4)
	if err != nil {
		return 0, err
	}
	return uint32(s[0]) | uint32(s[1])<<8 | uint32(s[2])<<16 | uint32(s[3])<<24, nil
}

func (r *reader) u64() (uint64, error) {
	s, err := r.take(8)
	if err != nil {
		return 0, err
	}
	return uint64(s[0]) | uint64(s[1])<<8 | uint64(s[2])<<16 | uint64(s[3])<<24 |
		uint64(s[4])<<32 | uint64(s[5])<<40 | uint64(s[6])<<48 | uint64(s[7])<<56, nil
}

func (r *reader) atEnd() bool { return r.pos == len(r.bytes) }

func (r *reader) remaining() int { return len(r.bytes) - r.pos }

// ---- Scalar sorted-list set algebra (low-key lists) ----------------------

func sortedUnion(a, b []uint16) []uint16 {
	out := make([]uint16, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] < b[j]:
			out = append(out, a[i])
			i++
		case a[i] > b[j]:
			out = append(out, b[j])
			j++
		default:
			out = append(out, a[i])
			i++
			j++
		}
	}
	out = append(out, a[i:]...)
	out = append(out, b[j:]...)
	return out
}

func sortedIntersect(a, b []uint16) []uint16 {
	var out []uint16
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] < b[j]:
			i++
		case a[i] > b[j]:
			j++
		default:
			out = append(out, a[i])
			i++
			j++
		}
	}
	return out
}

func sortedAndNot(a, b []uint16) []uint16 {
	var out []uint16
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] < b[j]:
			out = append(out, a[i])
			i++
		case a[i] > b[j]:
			j++
		default:
			i++
			j++
		}
	}
	out = append(out, a[i:]...)
	return out
}

func sortedXor(a, b []uint16) []uint16 {
	var out []uint16
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] < b[j]:
			out = append(out, a[i])
			i++
		case a[i] > b[j]:
			out = append(out, b[j])
			j++
		default:
			i++
			j++
		}
	}
	out = append(out, a[i:]...)
	out = append(out, b[j:]...)
	return out
}
