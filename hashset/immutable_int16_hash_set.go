
package hashset

import (
	"iter"
)

// ImmutableInt16HashSet is an immutable view of a Int16HashSet.
type ImmutableInt16HashSet struct {
	delegate *Int16HashSet
}

// NewImmutableInt16HashSet creates an immutable set from the given values.
func NewImmutableInt16HashSet(values ...int16) *ImmutableInt16HashSet {
	return &ImmutableInt16HashSet{delegate: Int16HashSetOf(values...)}
}

// ImmutableInt16HashSetFrom creates an immutable copy of a mutable set.
func ImmutableInt16HashSetFrom(s *Int16HashSet) *ImmutableInt16HashSet {
	copy := Int16HashSetOf(s.ToSlice()...)
	return &ImmutableInt16HashSet{delegate: copy}
}

// Contains returns true if the set contains the given value.
func (s *ImmutableInt16HashSet) Contains(value int16) bool {
	return s.delegate.Contains(value)
}

// Size returns the number of elements.
func (s *ImmutableInt16HashSet) Size() int {
	return s.delegate.Size()
}

// IsEmpty returns true if the set contains no elements.
func (s *ImmutableInt16HashSet) IsEmpty() bool {
	return s.delegate.IsEmpty()
}

// All returns an iter.Seq that yields all elements.
func (s *ImmutableInt16HashSet) All() iter.Seq[int16] {
	return s.delegate.All()
}

// ForEach calls the given function for each element.
func (s *ImmutableInt16HashSet) ForEach(f func(int16)) {
	s.delegate.ForEach(f)
}

// Select returns a new immutable set with elements satisfying the predicate.
func (s *ImmutableInt16HashSet) Select(predicate func(int16) bool) *ImmutableInt16HashSet {
	return &ImmutableInt16HashSet{delegate: s.delegate.Select(predicate)}
}

// Reject returns a new immutable set with elements not satisfying the predicate.
func (s *ImmutableInt16HashSet) Reject(predicate func(int16) bool) *ImmutableInt16HashSet {
	return &ImmutableInt16HashSet{delegate: s.delegate.Reject(predicate)}
}

// AnySatisfy returns true if any element satisfies the predicate.
func (s *ImmutableInt16HashSet) AnySatisfy(predicate func(int16) bool) bool {
	return s.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (s *ImmutableInt16HashSet) AllSatisfy(predicate func(int16) bool) bool {
	return s.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (s *ImmutableInt16HashSet) NoneSatisfy(predicate func(int16) bool) bool {
	return s.delegate.NoneSatisfy(predicate)
}

// Union returns a new immutable set with elements from both sets.
func (s *ImmutableInt16HashSet) Union(other *ImmutableInt16HashSet) *ImmutableInt16HashSet {
	return &ImmutableInt16HashSet{delegate: s.delegate.Union(other.delegate)}
}

// Intersect returns a new immutable set with elements in both sets.
func (s *ImmutableInt16HashSet) Intersect(other *ImmutableInt16HashSet) *ImmutableInt16HashSet {
	return &ImmutableInt16HashSet{delegate: s.delegate.Intersect(other.delegate)}
}

// Difference returns a new immutable set with elements in this but not other.
func (s *ImmutableInt16HashSet) Difference(other *ImmutableInt16HashSet) *ImmutableInt16HashSet {
	return &ImmutableInt16HashSet{delegate: s.delegate.Difference(other.delegate)}
}

// ToSlice returns all elements as a slice.
func (s *ImmutableInt16HashSet) ToSlice() []int16 {
	return s.delegate.ToSlice()
}

// String returns a string representation.
func (s *ImmutableInt16HashSet) String() string {
	return s.delegate.String()
}

// Equals returns true if the other immutable set has the same elements.
func (s *ImmutableInt16HashSet) Equals(other *ImmutableInt16HashSet) bool {
	return s.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this set.
func (s *ImmutableInt16HashSet) ToMutable() *Int16HashSet {
	return Int16HashSetOf(s.ToSlice()...)
}
