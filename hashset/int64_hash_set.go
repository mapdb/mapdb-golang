
package hashset

import (
	"fmt"
	"iter"
	"strings"
)

const (
	int64HashSetDefaultCapacity = 16
	// Load factor 3/4 = 0.75, using integer math to avoid float conversion per insert.
)

type int64HashSetEntry struct {
	key      int64
	occupied bool
}

// Int64HashSet is an open-addressing hash set for int64 values.
type Int64HashSet struct {
	entries []int64HashSetEntry
	size    int
}

// NewInt64HashSet creates a new empty Int64HashSet.
func NewInt64HashSet() *Int64HashSet {
	return NewInt64HashSetWithCapacity(int64HashSetDefaultCapacity)
}

// NewInt64HashSetWithCapacity creates a new empty Int64HashSet with the given initial capacity.
func NewInt64HashSetWithCapacity(capacity int) *Int64HashSet {
	cap := nextPowerOfTwoInt64HashSet(capacity)
	return &Int64HashSet{
		entries: make([]int64HashSetEntry, cap),
		size:    0,
	}
}

// Int64HashSetOf creates a new Int64HashSet from the given values.
func Int64HashSetOf(values ...int64) *Int64HashSet {
	s := NewInt64HashSetWithCapacity(len(values) * 2)
	for _, v := range values {
		s.Add(v)
	}
	return s
}

// Add inserts a value into the set. Returns true if the value was added (not already present).
func (s *Int64HashSet) Add(value int64) bool {
	if s.needsResize() {
		s.resize()
	}
	cap := len(s.entries)
	mask := cap - 1
	idx := int(s.hash(value)) & mask

	for {
		if !s.entries[idx].occupied {
			s.entries[idx].key = value
			s.entries[idx].occupied = true
			s.size++
			return true
		}
		if s.entries[idx].key == value {
			return false
		}
		idx = (idx + 1) & mask
	}
}

// AddAll inserts all values into the set.
func (s *Int64HashSet) AddAll(values ...int64) {
	for _, v := range values {
		s.Add(v)
	}
}

// Remove removes a value from the set. Returns true if the value was found and removed.
func (s *Int64HashSet) Remove(value int64) bool {
	cap := len(s.entries)
	if cap == 0 {
		return false
	}
	mask := cap - 1
	idx := int(s.hash(value)) & mask

	for {
		if !s.entries[idx].occupied {
			return false
		}
		if s.entries[idx].key == value {
			s.entries[idx] = int64HashSetEntry{}
			s.size--
			s.rehashFrom(idx, mask)
			return true
		}
		idx = (idx + 1) & mask
	}
}

// Contains returns true if the set contains the given value.
func (s *Int64HashSet) Contains(value int64) bool {
	cap := len(s.entries)
	if cap == 0 {
		return false
	}
	mask := cap - 1
	idx := int(s.hash(value)) & mask

	for {
		if !s.entries[idx].occupied {
			return false
		}
		if s.entries[idx].key == value {
			return true
		}
		idx = (idx + 1) & mask
	}
}

// Size returns the number of elements in the set.
func (s *Int64HashSet) Size() int {
	return s.size
}

// IsEmpty returns true if the set contains no elements.
func (s *Int64HashSet) IsEmpty() bool {
	return s.size == 0
}

// Clear removes all elements from the set.
func (s *Int64HashSet) Clear() {
	for i := range s.entries {
		s.entries[i] = int64HashSetEntry{}
	}
	s.size = 0
}

// All returns an iter.Seq that yields all elements.
func (s *Int64HashSet) All() iter.Seq[int64] {
	return func(yield func(int64) bool) {
		for i := range s.entries {
			if s.entries[i].occupied {
				if !yield(s.entries[i].key) {
					return
				}
			}
		}
	}
}

// ForEach calls the given function for each element.
func (s *Int64HashSet) ForEach(f func(int64)) {
	for i := range s.entries {
		if s.entries[i].occupied {
			f(s.entries[i].key)
		}
	}
}

// Select returns a new set containing only elements that satisfy the predicate.
func (s *Int64HashSet) Select(predicate func(int64) bool) *Int64HashSet {
	result := NewInt64HashSet()
	for i := range s.entries {
		if s.entries[i].occupied && predicate(s.entries[i].key) {
			result.Add(s.entries[i].key)
		}
	}
	return result
}

// Reject returns a new set containing only elements that do not satisfy the predicate.
func (s *Int64HashSet) Reject(predicate func(int64) bool) *Int64HashSet {
	result := NewInt64HashSet()
	for i := range s.entries {
		if s.entries[i].occupied && !predicate(s.entries[i].key) {
			result.Add(s.entries[i].key)
		}
	}
	return result
}

// Detect returns the first element that satisfies the predicate, or zero value and false.
func (s *Int64HashSet) Detect(predicate func(int64) bool) (int64, bool) {
	for i := range s.entries {
		if s.entries[i].occupied && predicate(s.entries[i].key) {
			return s.entries[i].key, true
		}
	}
	return 0, false
}

// AnySatisfy returns true if any element satisfies the predicate.
func (s *Int64HashSet) AnySatisfy(predicate func(int64) bool) bool {
	for i := range s.entries {
		if s.entries[i].occupied && predicate(s.entries[i].key) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (s *Int64HashSet) AllSatisfy(predicate func(int64) bool) bool {
	for i := range s.entries {
		if s.entries[i].occupied && !predicate(s.entries[i].key) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (s *Int64HashSet) NoneSatisfy(predicate func(int64) bool) bool {
	for i := range s.entries {
		if s.entries[i].occupied && predicate(s.entries[i].key) {
			return false
		}
	}
	return true
}

// Union returns a new set containing all elements from both sets.
func (s *Int64HashSet) Union(other *Int64HashSet) *Int64HashSet {
	result := NewInt64HashSetWithCapacity((s.size + other.size) * 2)
	for i := range s.entries {
		if s.entries[i].occupied {
			result.Add(s.entries[i].key)
		}
	}
	for i := range other.entries {
		if other.entries[i].occupied {
			result.Add(other.entries[i].key)
		}
	}
	return result
}

// Intersect returns a new set containing only elements present in both sets.
func (s *Int64HashSet) Intersect(other *Int64HashSet) *Int64HashSet {
	result := NewInt64HashSet()
	smaller, larger := s, other
	if s.size > other.size {
		smaller, larger = other, s
	}
	for i := range smaller.entries {
		if smaller.entries[i].occupied && larger.Contains(smaller.entries[i].key) {
			result.Add(smaller.entries[i].key)
		}
	}
	return result
}

// Difference returns a new set containing elements in this set but not in the other.
func (s *Int64HashSet) Difference(other *Int64HashSet) *Int64HashSet {
	result := NewInt64HashSet()
	for i := range s.entries {
		if s.entries[i].occupied && !other.Contains(s.entries[i].key) {
			result.Add(s.entries[i].key)
		}
	}
	return result
}

// SymmetricDifference returns a new set containing elements in either set but not both.
func (s *Int64HashSet) SymmetricDifference(other *Int64HashSet) *Int64HashSet {
	result := NewInt64HashSet()
	for i := range s.entries {
		if s.entries[i].occupied && !other.Contains(s.entries[i].key) {
			result.Add(s.entries[i].key)
		}
	}
	for i := range other.entries {
		if other.entries[i].occupied && !s.Contains(other.entries[i].key) {
			result.Add(other.entries[i].key)
		}
	}
	return result
}

// ToSlice returns all elements as a slice.
func (s *Int64HashSet) ToSlice() []int64 {
	result := make([]int64, 0, s.size)
	for i := range s.entries {
		if s.entries[i].occupied {
			result = append(result, s.entries[i].key)
		}
	}
	return result
}

// With returns the set after adding the value (fluent API).
func (s *Int64HashSet) With(value int64) *Int64HashSet {
	s.Add(value)
	return s
}

// Without returns the set after removing the value (fluent API).
func (s *Int64HashSet) Without(value int64) *Int64HashSet {
	s.Remove(value)
	return s
}

// WithAll returns the set after adding all values (fluent API).
func (s *Int64HashSet) WithAll(values ...int64) *Int64HashSet {
	s.AddAll(values...)
	return s
}

// WithoutAll returns the set after removing all given values (fluent API).
func (s *Int64HashSet) WithoutAll(values ...int64) *Int64HashSet {
	for _, v := range values {
		s.Remove(v)
	}
	return s
}

// ToImmutable returns an immutable copy of this set.
func (s *Int64HashSet) ToImmutable() *ImmutableInt64HashSet {
	return ImmutableInt64HashSetFrom(s)
}

// String returns a string representation of the set.
func (s *Int64HashSet) String() string {
	if s.size == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for i := range s.entries {
		if s.entries[i].occupied {
			if !first {
				sb.WriteString(", ")
			}
			fmt.Fprintf(&sb, "%v", s.entries[i].key)
			first = false
		}
	}
	sb.WriteString("}")
	return sb.String()
}

// Equals returns true if the other set has the same elements.
func (s *Int64HashSet) Equals(other *Int64HashSet) bool {
	if s.size != other.size {
		return false
	}
	for i := range s.entries {
		if s.entries[i].occupied && !other.Contains(s.entries[i].key) {
			return false
		}
	}
	return true
}

func (s *Int64HashSet) hash(value int64) uint64 {
	return func() uint64 { h := uint64(value) * 0x9E3779B97F4A7C15; return h ^ (h >> 32) }()
}

func (s *Int64HashSet) needsResize() bool {
	return (s.size+1)*4 > len(s.entries)*3 // 0.75 load factor, integer math
}

func (s *Int64HashSet) resize() {
	oldEntries := s.entries
	newCap := len(oldEntries) * 2
	if newCap == 0 {
		newCap = int64HashSetDefaultCapacity
	}
	s.entries = make([]int64HashSetEntry, newCap)
	s.size = 0

	for i := range oldEntries {
		if oldEntries[i].occupied {
			s.Add(oldEntries[i].key)
		}
	}
}

func (s *Int64HashSet) rehashFrom(deleted int, mask int) {
	c := len(s.entries)
	idx := (deleted + 1) & mask
	for s.entries[idx].occupied {
		ideal := int(s.hash(s.entries[idx].key)) & mask
		distCurrent := (idx - ideal + c) & mask
		distGap := (deleted - ideal + c) & mask
		if distCurrent > distGap {
			s.entries[deleted] = s.entries[idx]
			s.entries[idx] = int64HashSetEntry{}
			deleted = idx
		}
		idx = (idx + 1) & mask
		if idx == deleted {
			break
		}
	}
}

func nextPowerOfTwoInt64HashSet(n int) int {
	if n <= 0 {
		return 16
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	return n
}
