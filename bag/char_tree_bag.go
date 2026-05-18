
package bag

import (
	"fmt"
	"iter"
	"sort"
	"strings"
)

// CharTreeBagEntry holds a value and its occurrence count in a CharTreeBag.
type CharTreeBagEntry struct {
	value uint16
	count int
}

// CharTreeBag is a sorted bag (multiset) that counts occurrences of uint16 values.
// Backed by a sorted slice of entries with binary search for O(log n) lookup.
type CharTreeBag struct {
	entries []CharTreeBagEntry
	size    int // total count including duplicates
}

// NewCharTreeBag creates a new empty CharTreeBag.
func NewCharTreeBag() *CharTreeBag {
	return &CharTreeBag{
		entries: nil,
		size:    0,
	}
}

// CharTreeBagOf creates a new CharTreeBag from the given values.
func CharTreeBagOf(values ...uint16) *CharTreeBag {
	b := NewCharTreeBag()
	for _, v := range values {
		b.Add(v)
	}
	return b
}

// search returns the index where value is or would be inserted.
// Returns (index, found).
func (b *CharTreeBag) search(value uint16) (int, bool) {
	lo, hi := 0, len(b.entries)
	for lo < hi {
		mid := lo + (hi-lo)/2
		if b.entries[mid].value < value {
			lo = mid + 1
		} else if b.entries[mid].value > value {
			hi = mid
		} else {
			return mid, true
		}
	}
	return lo, false
}

// Add adds one occurrence of the value.
func (b *CharTreeBag) Add(value uint16) {
	idx, found := b.search(value)
	if found {
		b.entries[idx].count++
		b.size++
		return
	}
	// Insert at idx to keep sorted order
	b.entries = append(b.entries, CharTreeBagEntry{})
	copy(b.entries[idx+1:], b.entries[idx:])
	b.entries[idx] = CharTreeBagEntry{value: value, count: 1}
	b.size++
}

// AddOccurrences adds the given number of occurrences of the value.
// Returns the new count for this value. Panics if occurrences is negative.
func (b *CharTreeBag) AddOccurrences(value uint16, occurrences int) int {
	if occurrences < 0 {
		panic(fmt.Sprintf("CharTreeBag: cannot add negative occurrences: %d", occurrences))
	}
	if occurrences == 0 {
		idx, found := b.search(value)
		if found {
			return b.entries[idx].count
		}
		return 0
	}
	idx, found := b.search(value)
	if found {
		b.entries[idx].count += occurrences
		b.size += occurrences
		return b.entries[idx].count
	}
	b.entries = append(b.entries, CharTreeBagEntry{})
	copy(b.entries[idx+1:], b.entries[idx:])
	b.entries[idx] = CharTreeBagEntry{value: value, count: occurrences}
	b.size += occurrences
	return occurrences
}

// Remove removes one occurrence of the value. Returns true if the value was present.
func (b *CharTreeBag) Remove(value uint16) bool {
	idx, found := b.search(value)
	if !found {
		return false
	}
	if b.entries[idx].count == 1 {
		b.entries = append(b.entries[:idx], b.entries[idx+1:]...)
	} else {
		b.entries[idx].count--
	}
	b.size--
	return true
}

// RemoveOccurrences removes the given number of occurrences. Returns true if any were removed.
func (b *CharTreeBag) RemoveOccurrences(value uint16, occurrences int) bool {
	if occurrences <= 0 {
		return false
	}
	idx, found := b.search(value)
	if !found {
		return false
	}
	if occurrences >= b.entries[idx].count {
		b.size -= b.entries[idx].count
		b.entries = append(b.entries[:idx], b.entries[idx+1:]...)
	} else {
		b.entries[idx].count -= occurrences
		b.size -= occurrences
	}
	return true
}

// RemoveAll removes all occurrences of the value. Returns the previous count.
func (b *CharTreeBag) RemoveAll(value uint16) int {
	idx, found := b.search(value)
	if !found {
		return 0
	}
	count := b.entries[idx].count
	b.entries = append(b.entries[:idx], b.entries[idx+1:]...)
	b.size -= count
	return count
}

// OccurrencesOf returns the number of occurrences of the value.
func (b *CharTreeBag) OccurrencesOf(value uint16) int {
	idx, found := b.search(value)
	if !found {
		return 0
	}
	return b.entries[idx].count
}

// Contains returns true if the bag contains at least one occurrence of the value.
func (b *CharTreeBag) Contains(value uint16) bool {
	_, found := b.search(value)
	return found
}

// Size returns the total number of elements including duplicates.
func (b *CharTreeBag) Size() int {
	return b.size
}

// SizeDistinct returns the number of distinct elements.
func (b *CharTreeBag) SizeDistinct() int {
	return len(b.entries)
}

// IsEmpty returns true if the bag contains no elements.
func (b *CharTreeBag) IsEmpty() bool {
	return b.size == 0
}

// Clear removes all elements from the bag.
func (b *CharTreeBag) Clear() {
	b.entries = nil
	b.size = 0
}

// Min returns the smallest element, or zero value and false if empty.
func (b *CharTreeBag) Min() (uint16, bool) {
	if len(b.entries) == 0 {
		return 0, false
	}
	return b.entries[0].value, true
}

// Max returns the largest element, or zero value and false if empty.
func (b *CharTreeBag) Max() (uint16, bool) {
	if len(b.entries) == 0 {
		return 0, false
	}
	return b.entries[len(b.entries)-1].value, true
}

// All returns an iter.Seq that yields each element once per occurrence, in sorted order.
func (b *CharTreeBag) All() iter.Seq[uint16] {
	return func(yield func(uint16) bool) {
		for _, entry := range b.entries {
			for i := 0; i < entry.count; i++ {
				if !yield(entry.value) {
					return
				}
			}
		}
	}
}

// AllDistinct returns an iter.Seq that yields each distinct element once, in sorted order.
func (b *CharTreeBag) AllDistinct() iter.Seq[uint16] {
	return func(yield func(uint16) bool) {
		for _, entry := range b.entries {
			if !yield(entry.value) {
				return
			}
		}
	}
}

// AllWithOccurrences returns an iter.Seq2 that yields (value, count) pairs in sorted order.
func (b *CharTreeBag) AllWithOccurrences() iter.Seq2[uint16, int] {
	return func(yield func(uint16, int) bool) {
		for _, entry := range b.entries {
			if !yield(entry.value, entry.count) {
				return
			}
		}
	}
}

// ForEach calls the given function for each element (once per occurrence), in sorted order.
func (b *CharTreeBag) ForEach(f func(uint16)) {
	for _, entry := range b.entries {
		for i := 0; i < entry.count; i++ {
			f(entry.value)
		}
	}
}

// ForEachWithOccurrences calls the given function with each distinct element and its count, in sorted order.
func (b *CharTreeBag) ForEachWithOccurrences(f func(uint16, int)) {
	for _, entry := range b.entries {
		f(entry.value, entry.count)
	}
}

// Select returns a new tree bag containing only elements that satisfy the predicate.
func (b *CharTreeBag) Select(predicate func(uint16) bool) *CharTreeBag {
	result := NewCharTreeBag()
	for _, entry := range b.entries {
		if predicate(entry.value) {
			result.AddOccurrences(entry.value, entry.count)
		}
	}
	return result
}

// Reject returns a new tree bag containing only elements that do not satisfy the predicate.
func (b *CharTreeBag) Reject(predicate func(uint16) bool) *CharTreeBag {
	result := NewCharTreeBag()
	for _, entry := range b.entries {
		if !predicate(entry.value) {
			result.AddOccurrences(entry.value, entry.count)
		}
	}
	return result
}

// Detect returns the first distinct element (in sorted order) that satisfies the predicate, or zero value and false.
func (b *CharTreeBag) Detect(predicate func(uint16) bool) (uint16, bool) {
	for _, entry := range b.entries {
		if predicate(entry.value) {
			return entry.value, true
		}
	}
	return 0, false
}

// AnySatisfy returns true if any distinct element satisfies the predicate.
func (b *CharTreeBag) AnySatisfy(predicate func(uint16) bool) bool {
	for _, entry := range b.entries {
		if predicate(entry.value) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all distinct elements satisfy the predicate.
func (b *CharTreeBag) AllSatisfy(predicate func(uint16) bool) bool {
	for _, entry := range b.entries {
		if !predicate(entry.value) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no distinct element satisfies the predicate.
func (b *CharTreeBag) NoneSatisfy(predicate func(uint16) bool) bool {
	for _, entry := range b.entries {
		if predicate(entry.value) {
			return false
		}
	}
	return true
}

// TopOccurrences returns the n elements with the highest occurrence counts.
func (b *CharTreeBag) TopOccurrences(n int) []struct {
	Value uint16
	Count int
} {
	// Copy entries and sort by count descending
	sorted := make([]CharTreeBagEntry, len(b.entries))
	copy(sorted, b.entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})
	if n > len(sorted) {
		n = len(sorted)
	}
	result := make([]struct {
		Value uint16
		Count int
	}, n)
	for i := 0; i < n; i++ {
		result[i].Value = sorted[i].value
		result[i].Count = sorted[i].count
	}
	return result
}

// ToSlice returns all elements as a slice (elements repeated per occurrence count), in sorted order.
func (b *CharTreeBag) ToSlice() []uint16 {
	result := make([]uint16, 0, b.size)
	for _, entry := range b.entries {
		for i := 0; i < entry.count; i++ {
			result = append(result, entry.value)
		}
	}
	return result
}

// ToSortedSlice returns all distinct elements as a sorted slice.
func (b *CharTreeBag) ToSortedSlice() []uint16 {
	result := make([]uint16, 0, len(b.entries))
	for _, entry := range b.entries {
		result = append(result, entry.value)
	}
	return result
}

// With returns the bag after adding one occurrence of the value (fluent API).
func (b *CharTreeBag) With(value uint16) *CharTreeBag {
	b.Add(value)
	return b
}

// Without returns the bag after removing all occurrences of the value (fluent API).
func (b *CharTreeBag) Without(value uint16) *CharTreeBag {
	b.RemoveAll(value)
	return b
}

// WithAll returns the bag after adding all values (fluent API).
func (b *CharTreeBag) WithAll(values ...uint16) *CharTreeBag {
	for _, v := range values {
		b.Add(v)
	}
	return b
}

// WithoutAll removes all occurrences of the given values (fluent API).
func (b *CharTreeBag) WithoutAll(values ...uint16) *CharTreeBag {
	for _, v := range values {
		b.RemoveAll(v)
	}
	return b
}

// String returns a string representation of the bag in sorted order.
func (b *CharTreeBag) String() string {
	if b.size == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for _, entry := range b.entries {
		if !first {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v\u00d7%d", entry.value, entry.count)
		first = false
	}
	sb.WriteString("}")
	return sb.String()
}

// Equals returns true if the other tree bag has the same elements with the same counts.
func (b *CharTreeBag) Equals(other *CharTreeBag) bool {
	if b.size != other.size || len(b.entries) != len(other.entries) {
		return false
	}
	for i, entry := range b.entries {
		if entry.value != other.entries[i].value || entry.count != other.entries[i].count {
			return false
		}
	}
	return true
}
