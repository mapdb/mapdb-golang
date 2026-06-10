
package bag

import (
	"fmt"
	"iter"
	"math"
	"strings"
)

// float32BagEntry stores the original float32 value alongside its count.
// The backing map is keyed by the IEEE-754 bit pattern (math.Float32bits) so
// that NaN keys are findable and +0.0 / -0.0 remain distinct; Go map equality
// on raw float32 would make every NaN a fresh unreachable entry and collapse
// signed zero.
type float32BagEntry struct {
	value float32
	count int
}

// Float32HashBag is a bag (multiset) that counts occurrences of float32 values.
// Backed by a map from value bit pattern to (value, count).
type Float32HashBag struct {
	counts map[uint32]float32BagEntry
	size   int // total count including duplicates
}

// NewFloat32HashBag creates a new empty Float32HashBag.
func NewFloat32HashBag() *Float32HashBag {
	return &Float32HashBag{
		counts: make(map[uint32]float32BagEntry),
		size:   0,
	}
}

// Float32HashBagOf creates a new Float32HashBag from the given values.
func Float32HashBagOf(values ...float32) *Float32HashBag {
	b := NewFloat32HashBag()
	for _, v := range values {
		b.Add(v)
	}
	return b
}

// Add adds one occurrence of the value.
func (b *Float32HashBag) Add(value float32) {
	k := math.Float32bits(value)
	e := b.counts[k]
	e.value = value
	e.count++
	b.counts[k] = e
	b.size++
}

// AddOccurrences adds the given number of occurrences of the value.
// Returns the new count for this value. Panics if occurrences is negative.
func (b *Float32HashBag) AddOccurrences(value float32, occurrences int) int {
	if occurrences < 0 {
		panic(fmt.Sprintf("Float32HashBag: cannot add negative occurrences: %d", occurrences))
	}
	k := math.Float32bits(value)
	if occurrences == 0 {
		return b.counts[k].count
	}
	e := b.counts[k]
	e.value = value
	e.count += occurrences
	b.counts[k] = e
	b.size += occurrences
	return e.count
}

// Remove removes one occurrence of the value. Returns true if the value was present.
func (b *Float32HashBag) Remove(value float32) bool {
	k := math.Float32bits(value)
	e, ok := b.counts[k]
	if !ok || e.count <= 0 {
		return false
	}
	if e.count == 1 {
		delete(b.counts, k)
	} else {
		e.count--
		b.counts[k] = e
	}
	b.size--
	return true
}

// RemoveOccurrences removes the given number of occurrences. Returns true if any were removed.
func (b *Float32HashBag) RemoveOccurrences(value float32, occurrences int) bool {
	if occurrences <= 0 {
		return false
	}
	k := math.Float32bits(value)
	e, ok := b.counts[k]
	if !ok || e.count <= 0 {
		return false
	}
	if occurrences >= e.count {
		delete(b.counts, k)
		b.size -= e.count
	} else {
		e.count -= occurrences
		b.counts[k] = e
		b.size -= occurrences
	}
	return true
}

// RemoveAll removes all occurrences of the value. Returns the previous count.
func (b *Float32HashBag) RemoveAll(value float32) int {
	k := math.Float32bits(value)
	e, ok := b.counts[k]
	if !ok {
		return 0
	}
	delete(b.counts, k)
	b.size -= e.count
	return e.count
}

// OccurrencesOf returns the number of occurrences of the value.
func (b *Float32HashBag) OccurrencesOf(value float32) int {
	return b.counts[math.Float32bits(value)].count
}

// Contains returns true if the bag contains at least one occurrence of the value.
func (b *Float32HashBag) Contains(value float32) bool {
	return b.counts[math.Float32bits(value)].count > 0
}

// Size returns the total number of elements including duplicates.
func (b *Float32HashBag) Size() int {
	return b.size
}

// SizeDistinct returns the number of distinct elements.
func (b *Float32HashBag) SizeDistinct() int {
	return len(b.counts)
}

// IsEmpty returns true if the bag contains no elements.
func (b *Float32HashBag) IsEmpty() bool {
	return b.size == 0
}

// Clear removes all elements from the bag.
func (b *Float32HashBag) Clear() {
	b.counts = make(map[uint32]float32BagEntry)
	b.size = 0
}

// All returns an iter.Seq that yields each element once per occurrence.
func (b *Float32HashBag) All() iter.Seq[float32] {
	return func(yield func(float32) bool) {
		for _, e := range b.counts {
			for i := 0; i < e.count; i++ {
				if !yield(e.value) {
					return
				}
			}
		}
	}
}

// AllDistinct returns an iter.Seq that yields each distinct element once.
func (b *Float32HashBag) AllDistinct() iter.Seq[float32] {
	return func(yield func(float32) bool) {
		for _, e := range b.counts {
			if !yield(e.value) {
				return
			}
		}
	}
}

// AllWithOccurrences returns an iter.Seq2 that yields (value, count) pairs.
func (b *Float32HashBag) AllWithOccurrences() iter.Seq2[float32, int] {
	return func(yield func(float32, int) bool) {
		for _, e := range b.counts {
			if !yield(e.value, e.count) {
				return
			}
		}
	}
}

// ForEach calls the given function for each element (once per occurrence).
func (b *Float32HashBag) ForEach(f func(float32)) {
	for _, e := range b.counts {
		for i := 0; i < e.count; i++ {
			f(e.value)
		}
	}
}

// ForEachWithOccurrences calls the given function with each distinct element and its count.
func (b *Float32HashBag) ForEachWithOccurrences(f func(float32, int)) {
	for _, e := range b.counts {
		f(e.value, e.count)
	}
}

// Select returns a new bag containing only elements that satisfy the predicate.
func (b *Float32HashBag) Select(predicate func(float32) bool) *Float32HashBag {
	result := NewFloat32HashBag()
	for _, e := range b.counts {
		if predicate(e.value) {
			result.AddOccurrences(e.value, e.count)
		}
	}
	return result
}

// Reject returns a new bag containing only elements that do not satisfy the predicate.
func (b *Float32HashBag) Reject(predicate func(float32) bool) *Float32HashBag {
	result := NewFloat32HashBag()
	for _, e := range b.counts {
		if !predicate(e.value) {
			result.AddOccurrences(e.value, e.count)
		}
	}
	return result
}

// Detect returns the first distinct element that satisfies the predicate, or zero value and false.
func (b *Float32HashBag) Detect(predicate func(float32) bool) (float32, bool) {
	for _, e := range b.counts {
		if predicate(e.value) {
			return e.value, true
		}
	}
	return 0.0, false
}

// AnySatisfy returns true if any element satisfies the predicate.
func (b *Float32HashBag) AnySatisfy(predicate func(float32) bool) bool {
	for _, e := range b.counts {
		if predicate(e.value) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all distinct elements satisfy the predicate.
func (b *Float32HashBag) AllSatisfy(predicate func(float32) bool) bool {
	for _, e := range b.counts {
		if !predicate(e.value) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (b *Float32HashBag) NoneSatisfy(predicate func(float32) bool) bool {
	for _, e := range b.counts {
		if predicate(e.value) {
			return false
		}
	}
	return true
}

// TopOccurrences returns the n elements with the highest occurrence counts.
func (b *Float32HashBag) TopOccurrences(n int) []struct {
	Value float32
	Count int
} {
	type pair struct {
		Value float32
		Count int
	}
	pairs := make([]pair, 0, len(b.counts))
	for _, e := range b.counts {
		pairs = append(pairs, pair{e.value, e.count})
	}
	// Simple selection sort for top-n (good enough for typical use)
	for i := 0; i < n && i < len(pairs); i++ {
		maxIdx := i
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].Count > pairs[maxIdx].Count {
				maxIdx = j
			}
		}
		pairs[i], pairs[maxIdx] = pairs[maxIdx], pairs[i]
	}
	if n > len(pairs) {
		n = len(pairs)
	}
	result := make([]struct {
		Value float32
		Count int
	}, n)
	for i := 0; i < n; i++ {
		result[i].Value = pairs[i].Value
		result[i].Count = pairs[i].Count
	}
	return result
}

// ToSlice returns all elements as a slice (elements repeated per occurrence count).
func (b *Float32HashBag) ToSlice() []float32 {
	result := make([]float32, 0, b.size)
	for _, e := range b.counts {
		for i := 0; i < e.count; i++ {
			result = append(result, e.value)
		}
	}
	return result
}

// With returns the bag after adding one occurrence of the value (fluent API).
func (b *Float32HashBag) With(value float32) *Float32HashBag {
	b.Add(value)
	return b
}

// Without returns the bag after removing all occurrences of the value (fluent API).
func (b *Float32HashBag) Without(value float32) *Float32HashBag {
	b.RemoveAll(value)
	return b
}

// String returns a string representation of the bag.
func (b *Float32HashBag) String() string {
	if b.size == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for _, e := range b.counts {
		if !first {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v×%d", e.value, e.count)
		first = false
	}
	sb.WriteString("}")
	return sb.String()
}

// WithAll returns the bag after adding all values (fluent API).
func (b *Float32HashBag) WithAll(values ...float32) *Float32HashBag {
	for _, v := range values {
		b.Add(v)
	}
	return b
}

// WithoutAll removes all occurrences of the given values.
func (b *Float32HashBag) WithoutAll(values ...float32) *Float32HashBag {
	for _, v := range values {
		b.RemoveAll(v)
	}
	return b
}

// ToImmutable returns an immutable copy of this bag.
func (b *Float32HashBag) ToImmutable() *ImmutableFloat32HashBag {
	return ImmutableFloat32HashBagFrom(b)
}

// Equals returns true if the other bag has the same elements with the same counts.
func (b *Float32HashBag) Equals(other *Float32HashBag) bool {
	if b.size != other.size || len(b.counts) != len(other.counts) {
		return false
	}
	for k, e := range b.counts {
		if other.counts[k].count != e.count {
			return false
		}
	}
	return true
}
