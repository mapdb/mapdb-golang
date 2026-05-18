
package hashset

import (
	"iter"
)

// ImmutableBoolHashSet is an immutable view of a BoolHashSet.
type ImmutableBoolHashSet struct {
	delegate *BoolHashSet
}

// NewImmutableBoolHashSet creates an immutable set from the given values.
func NewImmutableBoolHashSet(values ...bool) *ImmutableBoolHashSet {
	return &ImmutableBoolHashSet{delegate: BoolHashSetOf(values...)}
}

// ImmutableBoolHashSetFrom creates an immutable copy of a mutable set.
func ImmutableBoolHashSetFrom(s *BoolHashSet) *ImmutableBoolHashSet {
	copy := BoolHashSetOf(s.ToSlice()...)
	return &ImmutableBoolHashSet{delegate: copy}
}

// Contains returns true if the set contains the given value.
func (s *ImmutableBoolHashSet) Contains(value bool) bool {
	return s.delegate.Contains(value)
}

// Size returns the number of elements.
func (s *ImmutableBoolHashSet) Size() int {
	return s.delegate.Size()
}

// IsEmpty returns true if the set contains no elements.
func (s *ImmutableBoolHashSet) IsEmpty() bool {
	return s.delegate.IsEmpty()
}

// All returns an iter.Seq that yields all elements.
func (s *ImmutableBoolHashSet) All() iter.Seq[bool] {
	return s.delegate.All()
}

// ForEach calls the given function for each element.
func (s *ImmutableBoolHashSet) ForEach(f func(bool)) {
	s.delegate.ForEach(f)
}

// Select returns a new immutable set with elements satisfying the predicate.
func (s *ImmutableBoolHashSet) Select(predicate func(bool) bool) *ImmutableBoolHashSet {
	return &ImmutableBoolHashSet{delegate: s.delegate.Select(predicate)}
}

// Reject returns a new immutable set with elements not satisfying the predicate.
func (s *ImmutableBoolHashSet) Reject(predicate func(bool) bool) *ImmutableBoolHashSet {
	return &ImmutableBoolHashSet{delegate: s.delegate.Reject(predicate)}
}

// AnySatisfy returns true if any element satisfies the predicate.
func (s *ImmutableBoolHashSet) AnySatisfy(predicate func(bool) bool) bool {
	return s.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (s *ImmutableBoolHashSet) AllSatisfy(predicate func(bool) bool) bool {
	return s.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (s *ImmutableBoolHashSet) NoneSatisfy(predicate func(bool) bool) bool {
	return s.delegate.NoneSatisfy(predicate)
}

// Union returns a new immutable set with elements from both sets.
func (s *ImmutableBoolHashSet) Union(other *ImmutableBoolHashSet) *ImmutableBoolHashSet {
	return &ImmutableBoolHashSet{delegate: s.delegate.Union(other.delegate)}
}

// Intersect returns a new immutable set with elements in both sets.
func (s *ImmutableBoolHashSet) Intersect(other *ImmutableBoolHashSet) *ImmutableBoolHashSet {
	return &ImmutableBoolHashSet{delegate: s.delegate.Intersect(other.delegate)}
}

// Difference returns a new immutable set with elements in this but not other.
func (s *ImmutableBoolHashSet) Difference(other *ImmutableBoolHashSet) *ImmutableBoolHashSet {
	return &ImmutableBoolHashSet{delegate: s.delegate.Difference(other.delegate)}
}

// ToSlice returns all elements as a slice.
func (s *ImmutableBoolHashSet) ToSlice() []bool {
	return s.delegate.ToSlice()
}

// String returns a string representation.
func (s *ImmutableBoolHashSet) String() string {
	return s.delegate.String()
}

// Equals returns true if the other immutable set has the same elements.
func (s *ImmutableBoolHashSet) Equals(other *ImmutableBoolHashSet) bool {
	return s.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this set.
func (s *ImmutableBoolHashSet) ToMutable() *BoolHashSet {
	return BoolHashSetOf(s.ToSlice()...)
}
