
package hashset

import (
	"iter"
)

// ImmutableInt64HashSet is an immutable view of a Int64HashSet.
type ImmutableInt64HashSet struct {
	delegate *Int64HashSet
}

// NewImmutableInt64HashSet creates an immutable set from the given values.
func NewImmutableInt64HashSet(values ...int64) *ImmutableInt64HashSet {
	return &ImmutableInt64HashSet{delegate: Int64HashSetOf(values...)}
}

// ImmutableInt64HashSetFrom creates an immutable copy of a mutable set.
func ImmutableInt64HashSetFrom(s *Int64HashSet) *ImmutableInt64HashSet {
	copy := Int64HashSetOf(s.ToSlice()...)
	return &ImmutableInt64HashSet{delegate: copy}
}

// Contains returns true if the set contains the given value.
func (s *ImmutableInt64HashSet) Contains(value int64) bool {
	return s.delegate.Contains(value)
}

// Size returns the number of elements.
func (s *ImmutableInt64HashSet) Size() int {
	return s.delegate.Size()
}

// IsEmpty returns true if the set contains no elements.
func (s *ImmutableInt64HashSet) IsEmpty() bool {
	return s.delegate.IsEmpty()
}

// All returns an iter.Seq that yields all elements.
func (s *ImmutableInt64HashSet) All() iter.Seq[int64] {
	return s.delegate.All()
}

// ForEach calls the given function for each element.
func (s *ImmutableInt64HashSet) ForEach(f func(int64)) {
	s.delegate.ForEach(f)
}

// Select returns a new immutable set with elements satisfying the predicate.
func (s *ImmutableInt64HashSet) Select(predicate func(int64) bool) *ImmutableInt64HashSet {
	return &ImmutableInt64HashSet{delegate: s.delegate.Select(predicate)}
}

// Reject returns a new immutable set with elements not satisfying the predicate.
func (s *ImmutableInt64HashSet) Reject(predicate func(int64) bool) *ImmutableInt64HashSet {
	return &ImmutableInt64HashSet{delegate: s.delegate.Reject(predicate)}
}

// AnySatisfy returns true if any element satisfies the predicate.
func (s *ImmutableInt64HashSet) AnySatisfy(predicate func(int64) bool) bool {
	return s.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (s *ImmutableInt64HashSet) AllSatisfy(predicate func(int64) bool) bool {
	return s.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (s *ImmutableInt64HashSet) NoneSatisfy(predicate func(int64) bool) bool {
	return s.delegate.NoneSatisfy(predicate)
}

// Union returns a new immutable set with elements from both sets.
func (s *ImmutableInt64HashSet) Union(other *ImmutableInt64HashSet) *ImmutableInt64HashSet {
	return &ImmutableInt64HashSet{delegate: s.delegate.Union(other.delegate)}
}

// Intersect returns a new immutable set with elements in both sets.
func (s *ImmutableInt64HashSet) Intersect(other *ImmutableInt64HashSet) *ImmutableInt64HashSet {
	return &ImmutableInt64HashSet{delegate: s.delegate.Intersect(other.delegate)}
}

// Difference returns a new immutable set with elements in this but not other.
func (s *ImmutableInt64HashSet) Difference(other *ImmutableInt64HashSet) *ImmutableInt64HashSet {
	return &ImmutableInt64HashSet{delegate: s.delegate.Difference(other.delegate)}
}

// ToSlice returns all elements as a slice.
func (s *ImmutableInt64HashSet) ToSlice() []int64 {
	return s.delegate.ToSlice()
}

// String returns a string representation.
func (s *ImmutableInt64HashSet) String() string {
	return s.delegate.String()
}

// Equals returns true if the other immutable set has the same elements.
func (s *ImmutableInt64HashSet) Equals(other *ImmutableInt64HashSet) bool {
	return s.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this set.
func (s *ImmutableInt64HashSet) ToMutable() *Int64HashSet {
	return Int64HashSetOf(s.ToSlice()...)
}
