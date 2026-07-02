package bitset

import (
	"fmt"
	"math/bits"
	"strings"
)

// BitSet is a compact bit-packed storage for booleans backed by []uint64.
// O(1) Set/Clear/Flip/Get; Cardinality and bitwise ops are O(n/64).
type BitSet struct {
	words     []uint64
	bitLength int
}

const bitsPerWord = 64

// NewBitSet creates a new empty BitSet.
func NewBitSet() *BitSet { return &BitSet{} }

// NewBitSetOfLength creates a BitSet preallocated for n_bits bits (all zero).
func NewBitSetOfLength(nBits int) *BitSet {
	if nBits < 0 {
		panic("BitSet: length must not be negative")
	}
	nWords := (nBits + bitsPerWord - 1) / bitsPerWord
	return &BitSet{words: make([]uint64, nWords), bitLength: nBits}
}

// BitSetOfIndices builds a BitSet with the given bit indices set, sizing the
// backing words once for the maximum index in a single O(n) pass instead of
// growing on each Set. Indices need not be sorted; a negative index panics. It is
// the BitSet's bulk-load convenience and yields the same BitSet as repeated Set.
func BitSetOfIndices(indices ...int) *BitSet {
	max := -1
	for _, i := range indices {
		if i < 0 {
			panic("BitSet: bit index must not be negative")
		}
		if i > max {
			max = i
		}
	}
	b := NewBitSetOfLength(max + 1)
	for _, i := range indices {
		b.words[i/bitsPerWord] |= 1 << uint(i%bitsPerWord)
	}
	return b
}

// SetAll sets every bit in indices in a single pass, presizing the backing words
// once for the maximum index. Indices need not be sorted; a negative index
// panics. Equivalent to calling Set for each index but without per-index growth.
func (b *BitSet) SetAll(indices ...int) {
	max := -1
	for _, i := range indices {
		if i < 0 {
			panic("BitSet: bit index must not be negative")
		}
		if i > max {
			max = i
		}
	}
	if max >= 0 {
		b.ensure(max)
	}
	for _, i := range indices {
		b.words[i/bitsPerWord] |= 1 << uint(i%bitsPerWord)
	}
}

func (b *BitSet) ensure(bit int) {
	if bit < 0 {
		panic("BitSet: bit index must not be negative")
	}
	needed := bit/bitsPerWord + 1
	if len(b.words) < needed {
		ext := make([]uint64, needed-len(b.words))
		b.words = append(b.words, ext...)
	}
	if bit+1 > b.bitLength {
		b.bitLength = bit + 1
	}
}

// Set sets the bit at index `bit` to 1.
func (b *BitSet) Set(bit int) {
	b.ensure(bit)
	b.words[bit/bitsPerWord] |= 1 << uint(bit%bitsPerWord)
}

// Clear clears the bit at index `bit`. No-op for out-of-range bits.
func (b *BitSet) Clear(bit int) {
	if bit < 0 {
		return
	}
	wi := bit / bitsPerWord
	if wi >= len(b.words) {
		return
	}
	b.words[wi] &^= 1 << uint(bit%bitsPerWord)
}

// Flip toggles the bit at index `bit`.
func (b *BitSet) Flip(bit int) {
	b.ensure(bit)
	b.words[bit/bitsPerWord] ^= 1 << uint(bit%bitsPerWord)
}

// Get returns true if the bit at `bit` is 1. Out-of-range returns false.
func (b *BitSet) Get(bit int) bool {
	if bit < 0 {
		return false
	}
	wi := bit / bitsPerWord
	if wi >= len(b.words) {
		return false
	}
	return (b.words[wi] & (1 << uint(bit%bitsPerWord))) != 0
}

// Cardinality returns the number of set bits.
func (b *BitSet) Cardinality() int {
	if b.bitLength == 0 {
		return 0
	}
	lastIdx := (b.bitLength - 1) / bitsPerWord
	count := 0
	for i, w := range b.words {
		if i < lastIdx {
			count += bits.OnesCount64(w)
		} else if i == lastIdx {
			rem := b.bitLength - i*bitsPerWord
			var mask uint64
			if rem == bitsPerWord {
				mask = ^uint64(0)
			} else {
				mask = (1 << uint(rem)) - 1
			}
			count += bits.OnesCount64(w & mask)
		}
	}
	return count
}

// Length returns the logical bit length.
func (b *BitSet) Length() int { return b.bitLength }

// IsEmpty returns true if no bits are set.
func (b *BitSet) IsEmpty() bool { return b.Cardinality() == 0 }

// ClearAll clears all bits (keeps capacity).
func (b *BitSet) ClearAll() {
	for i := range b.words {
		b.words[i] = 0
	}
}

// Intersects returns true if any bit is set in both receivers.
func (b *BitSet) Intersects(other *BitSet) bool {
	min := len(b.words)
	if len(other.words) < min {
		min = len(other.words)
	}
	for i := 0; i < min; i++ {
		if (b.words[i] & other.words[i]) != 0 {
			return true
		}
	}
	return false
}

// AndInPlace performs b = b AND other.
func (b *BitSet) AndInPlace(other *BitSet) {
	for i := range b.words {
		var ow uint64
		if i < len(other.words) {
			ow = other.words[i]
		}
		b.words[i] &= ow
	}
}

// OrInPlace performs b = b OR other, extending b if needed.
func (b *BitSet) OrInPlace(other *BitSet) {
	if len(other.words) > len(b.words) {
		ext := make([]uint64, len(other.words)-len(b.words))
		b.words = append(b.words, ext...)
	}
	if other.bitLength > b.bitLength {
		b.bitLength = other.bitLength
	}
	for i, ow := range other.words {
		b.words[i] |= ow
	}
}

// XorInPlace performs b = b XOR other, extending b if needed.
func (b *BitSet) XorInPlace(other *BitSet) {
	if len(other.words) > len(b.words) {
		ext := make([]uint64, len(other.words)-len(b.words))
		b.words = append(b.words, ext...)
	}
	if other.bitLength > b.bitLength {
		b.bitLength = other.bitLength
	}
	for i, ow := range other.words {
		b.words[i] ^= ow
	}
}

// AndNotInPlace performs b = b AND NOT other.
func (b *BitSet) AndNotInPlace(other *BitSet) {
	min := len(b.words)
	if len(other.words) < min {
		min = len(other.words)
	}
	for i := 0; i < min; i++ {
		b.words[i] &^= other.words[i]
	}
}

// NextSetBit returns the index of the next set bit at or after `from`, or -1.
func (b *BitSet) NextSetBit(from int) int {
	if from < 0 {
		from = 0
	}
	wi := from / bitsPerWord
	if wi >= len(b.words) {
		return -1
	}
	offset := uint(from % bitsPerWord)
	word := b.words[wi] & (^uint64(0) << offset)
	for {
		if word != 0 {
			return wi*bitsPerWord + bits.TrailingZeros64(word)
		}
		wi++
		if wi >= len(b.words) {
			return -1
		}
		word = b.words[wi]
	}
}

// ToSlice returns the indices of all set bits in ascending order.
func (b *BitSet) ToSlice() []int {
	out := make([]int, 0, b.Cardinality())
	bit := b.NextSetBit(0)
	for bit >= 0 {
		out = append(out, bit)
		bit = b.NextSetBit(bit + 1)
	}
	return out
}

// Equals returns true if both BitSets have the same length and bits.
func (b *BitSet) Equals(other *BitSet) bool {
	if b.bitLength != other.bitLength {
		return false
	}
	n := len(b.words)
	if len(other.words) > n {
		n = len(other.words)
	}
	for i := 0; i < n; i++ {
		var a, c uint64
		if i < len(b.words) {
			a = b.words[i]
		}
		if i < len(other.words) {
			c = other.words[i]
		}
		if a != c {
			return false
		}
	}
	return true
}

// String returns a string representation of the set bits: "{1, 3, 5}".
func (b *BitSet) String() string {
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	bit := b.NextSetBit(0)
	for bit >= 0 {
		if !first {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%d", bit)
		first = false
		bit = b.NextSetBit(bit + 1)
	}
	sb.WriteString("}")
	return sb.String()
}
