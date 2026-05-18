
package arraylist

import (
	"iter"
)

// ImmutableCharArrayList is an immutable view of a CharArrayList.
type ImmutableCharArrayList struct {
	delegate *CharArrayList
}

// NewImmutableCharArrayList creates an immutable list from the given values.
func NewImmutableCharArrayList(values ...uint16) *ImmutableCharArrayList {
	return &ImmutableCharArrayList{delegate: CharArrayListOf(values...)}
}

// ImmutableCharArrayListFrom creates an immutable copy of a mutable list.
func ImmutableCharArrayListFrom(l *CharArrayList) *ImmutableCharArrayList {
	copy := CharArrayListOf(l.ToSlice()...)
	return &ImmutableCharArrayList{delegate: copy}
}

// Get returns the value at the given index, or an error if the index is out of bounds.
func (l *ImmutableCharArrayList) Get(index int) (uint16, error) {
	return l.delegate.Get(index)
}

// Size returns the number of elements.
func (l *ImmutableCharArrayList) Size() int {
	return l.delegate.Size()
}

// IsEmpty returns true if the list contains no elements.
func (l *ImmutableCharArrayList) IsEmpty() bool {
	return l.delegate.IsEmpty()
}

// Contains returns true if the list contains the given value.
func (l *ImmutableCharArrayList) Contains(value uint16) bool {
	return l.delegate.Contains(value)
}

// IndexOf returns the index of the first occurrence, or -1.
func (l *ImmutableCharArrayList) IndexOf(value uint16) int {
	return l.delegate.IndexOf(value)
}

// All returns an iter.Seq that yields all elements in order.
func (l *ImmutableCharArrayList) All() iter.Seq[uint16] {
	return l.delegate.All()
}

// AllWithIndex returns an iter.Seq2 that yields (index, value) pairs.
func (l *ImmutableCharArrayList) AllWithIndex() iter.Seq2[int, uint16] {
	return l.delegate.AllWithIndex()
}

// ForEach calls the given function for each element.
func (l *ImmutableCharArrayList) ForEach(f func(uint16)) {
	l.delegate.ForEach(f)
}

// Select returns a new immutable list with elements satisfying the predicate.
func (l *ImmutableCharArrayList) Select(predicate func(uint16) bool) *ImmutableCharArrayList {
	return &ImmutableCharArrayList{delegate: l.delegate.Select(predicate)}
}

// Reject returns a new immutable list with elements not satisfying the predicate.
func (l *ImmutableCharArrayList) Reject(predicate func(uint16) bool) *ImmutableCharArrayList {
	return &ImmutableCharArrayList{delegate: l.delegate.Reject(predicate)}
}

// Detect returns the first element satisfying the predicate, or zero and false.
func (l *ImmutableCharArrayList) Detect(predicate func(uint16) bool) (uint16, bool) {
	return l.delegate.Detect(predicate)
}

// AnySatisfy returns true if any element satisfies the predicate.
func (l *ImmutableCharArrayList) AnySatisfy(predicate func(uint16) bool) bool {
	return l.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (l *ImmutableCharArrayList) AllSatisfy(predicate func(uint16) bool) bool {
	return l.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (l *ImmutableCharArrayList) NoneSatisfy(predicate func(uint16) bool) bool {
	return l.delegate.NoneSatisfy(predicate)
}

// Count returns the number of elements satisfying the predicate.
func (l *ImmutableCharArrayList) Count(predicate func(uint16) bool) int {
	return l.delegate.Count(predicate)
}

// Reversed returns a new immutable list in reverse order.
func (l *ImmutableCharArrayList) Reversed() *ImmutableCharArrayList {
	return &ImmutableCharArrayList{delegate: l.delegate.Reversed()}
}

// Distinct returns a new immutable list with duplicates removed.
func (l *ImmutableCharArrayList) Distinct() *ImmutableCharArrayList {
	return &ImmutableCharArrayList{delegate: l.delegate.Distinct()}
}

// ToSlice returns a copy of all elements as a slice.
func (l *ImmutableCharArrayList) ToSlice() []uint16 {
	return l.delegate.ToSlice()
}

// String returns a string representation.
func (l *ImmutableCharArrayList) String() string {
	return l.delegate.String()
}

// Equals returns true if the other immutable list has the same elements in order.
func (l *ImmutableCharArrayList) Equals(other *ImmutableCharArrayList) bool {
	return l.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this list.
func (l *ImmutableCharArrayList) ToMutable() *CharArrayList {
	return CharArrayListOf(l.ToSlice()...)
}
