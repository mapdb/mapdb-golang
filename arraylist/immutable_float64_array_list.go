
package arraylist

import (
	"iter"
)

// ImmutableFloat64ArrayList is an immutable view of a Float64ArrayList.
type ImmutableFloat64ArrayList struct {
	delegate *Float64ArrayList
}

// NewImmutableFloat64ArrayList creates an immutable list from the given values.
func NewImmutableFloat64ArrayList(values ...float64) *ImmutableFloat64ArrayList {
	return &ImmutableFloat64ArrayList{delegate: Float64ArrayListOf(values...)}
}

// ImmutableFloat64ArrayListFrom creates an immutable copy of a mutable list.
func ImmutableFloat64ArrayListFrom(l *Float64ArrayList) *ImmutableFloat64ArrayList {
	copy := Float64ArrayListOf(l.ToSlice()...)
	return &ImmutableFloat64ArrayList{delegate: copy}
}

// Get returns the value at the given index, or an error if the index is out of bounds.
func (l *ImmutableFloat64ArrayList) Get(index int) (float64, error) {
	return l.delegate.Get(index)
}

// Size returns the number of elements.
func (l *ImmutableFloat64ArrayList) Size() int {
	return l.delegate.Size()
}

// IsEmpty returns true if the list contains no elements.
func (l *ImmutableFloat64ArrayList) IsEmpty() bool {
	return l.delegate.IsEmpty()
}

// Contains returns true if the list contains the given value.
func (l *ImmutableFloat64ArrayList) Contains(value float64) bool {
	return l.delegate.Contains(value)
}

// IndexOf returns the index of the first occurrence, or -1.
func (l *ImmutableFloat64ArrayList) IndexOf(value float64) int {
	return l.delegate.IndexOf(value)
}

// All returns an iter.Seq that yields all elements in order.
func (l *ImmutableFloat64ArrayList) All() iter.Seq[float64] {
	return l.delegate.All()
}

// AllWithIndex returns an iter.Seq2 that yields (index, value) pairs.
func (l *ImmutableFloat64ArrayList) AllWithIndex() iter.Seq2[int, float64] {
	return l.delegate.AllWithIndex()
}

// ForEach calls the given function for each element.
func (l *ImmutableFloat64ArrayList) ForEach(f func(float64)) {
	l.delegate.ForEach(f)
}

// Select returns a new immutable list with elements satisfying the predicate.
func (l *ImmutableFloat64ArrayList) Select(predicate func(float64) bool) *ImmutableFloat64ArrayList {
	return &ImmutableFloat64ArrayList{delegate: l.delegate.Select(predicate)}
}

// Reject returns a new immutable list with elements not satisfying the predicate.
func (l *ImmutableFloat64ArrayList) Reject(predicate func(float64) bool) *ImmutableFloat64ArrayList {
	return &ImmutableFloat64ArrayList{delegate: l.delegate.Reject(predicate)}
}

// Detect returns the first element satisfying the predicate, or zero and false.
func (l *ImmutableFloat64ArrayList) Detect(predicate func(float64) bool) (float64, bool) {
	return l.delegate.Detect(predicate)
}

// AnySatisfy returns true if any element satisfies the predicate.
func (l *ImmutableFloat64ArrayList) AnySatisfy(predicate func(float64) bool) bool {
	return l.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (l *ImmutableFloat64ArrayList) AllSatisfy(predicate func(float64) bool) bool {
	return l.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (l *ImmutableFloat64ArrayList) NoneSatisfy(predicate func(float64) bool) bool {
	return l.delegate.NoneSatisfy(predicate)
}

// Count returns the number of elements satisfying the predicate.
func (l *ImmutableFloat64ArrayList) Count(predicate func(float64) bool) int {
	return l.delegate.Count(predicate)
}

// Reversed returns a new immutable list in reverse order.
func (l *ImmutableFloat64ArrayList) Reversed() *ImmutableFloat64ArrayList {
	return &ImmutableFloat64ArrayList{delegate: l.delegate.Reversed()}
}

// Distinct returns a new immutable list with duplicates removed.
func (l *ImmutableFloat64ArrayList) Distinct() *ImmutableFloat64ArrayList {
	return &ImmutableFloat64ArrayList{delegate: l.delegate.Distinct()}
}

// ToSlice returns a copy of all elements as a slice.
func (l *ImmutableFloat64ArrayList) ToSlice() []float64 {
	return l.delegate.ToSlice()
}

// String returns a string representation.
func (l *ImmutableFloat64ArrayList) String() string {
	return l.delegate.String()
}

// Equals returns true if the other immutable list has the same elements in order.
func (l *ImmutableFloat64ArrayList) Equals(other *ImmutableFloat64ArrayList) bool {
	return l.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this list.
func (l *ImmutableFloat64ArrayList) ToMutable() *Float64ArrayList {
	return Float64ArrayListOf(l.ToSlice()...)
}
