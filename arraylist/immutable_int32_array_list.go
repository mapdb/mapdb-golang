
package arraylist

import (
	"iter"
)

// ImmutableInt32ArrayList is an immutable view of a Int32ArrayList.
type ImmutableInt32ArrayList struct {
	delegate *Int32ArrayList
}

// NewImmutableInt32ArrayList creates an immutable list from the given values.
func NewImmutableInt32ArrayList(values ...int32) *ImmutableInt32ArrayList {
	return &ImmutableInt32ArrayList{delegate: Int32ArrayListOf(values...)}
}

// ImmutableInt32ArrayListFrom creates an immutable copy of a mutable list.
func ImmutableInt32ArrayListFrom(l *Int32ArrayList) *ImmutableInt32ArrayList {
	copy := Int32ArrayListOf(l.ToSlice()...)
	return &ImmutableInt32ArrayList{delegate: copy}
}

// Get returns the value at the given index, or an error if the index is out of bounds.
func (l *ImmutableInt32ArrayList) Get(index int) (int32, error) {
	return l.delegate.Get(index)
}

// Size returns the number of elements.
func (l *ImmutableInt32ArrayList) Size() int {
	return l.delegate.Size()
}

// IsEmpty returns true if the list contains no elements.
func (l *ImmutableInt32ArrayList) IsEmpty() bool {
	return l.delegate.IsEmpty()
}

// Contains returns true if the list contains the given value.
func (l *ImmutableInt32ArrayList) Contains(value int32) bool {
	return l.delegate.Contains(value)
}

// IndexOf returns the index of the first occurrence, or -1.
func (l *ImmutableInt32ArrayList) IndexOf(value int32) int {
	return l.delegate.IndexOf(value)
}

// All returns an iter.Seq that yields all elements in order.
func (l *ImmutableInt32ArrayList) All() iter.Seq[int32] {
	return l.delegate.All()
}

// AllWithIndex returns an iter.Seq2 that yields (index, value) pairs.
func (l *ImmutableInt32ArrayList) AllWithIndex() iter.Seq2[int, int32] {
	return l.delegate.AllWithIndex()
}

// ForEach calls the given function for each element.
func (l *ImmutableInt32ArrayList) ForEach(f func(int32)) {
	l.delegate.ForEach(f)
}

// Select returns a new immutable list with elements satisfying the predicate.
func (l *ImmutableInt32ArrayList) Select(predicate func(int32) bool) *ImmutableInt32ArrayList {
	return &ImmutableInt32ArrayList{delegate: l.delegate.Select(predicate)}
}

// Reject returns a new immutable list with elements not satisfying the predicate.
func (l *ImmutableInt32ArrayList) Reject(predicate func(int32) bool) *ImmutableInt32ArrayList {
	return &ImmutableInt32ArrayList{delegate: l.delegate.Reject(predicate)}
}

// Detect returns the first element satisfying the predicate, or zero and false.
func (l *ImmutableInt32ArrayList) Detect(predicate func(int32) bool) (int32, bool) {
	return l.delegate.Detect(predicate)
}

// AnySatisfy returns true if any element satisfies the predicate.
func (l *ImmutableInt32ArrayList) AnySatisfy(predicate func(int32) bool) bool {
	return l.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (l *ImmutableInt32ArrayList) AllSatisfy(predicate func(int32) bool) bool {
	return l.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (l *ImmutableInt32ArrayList) NoneSatisfy(predicate func(int32) bool) bool {
	return l.delegate.NoneSatisfy(predicate)
}

// Count returns the number of elements satisfying the predicate.
func (l *ImmutableInt32ArrayList) Count(predicate func(int32) bool) int {
	return l.delegate.Count(predicate)
}

// Reversed returns a new immutable list in reverse order.
func (l *ImmutableInt32ArrayList) Reversed() *ImmutableInt32ArrayList {
	return &ImmutableInt32ArrayList{delegate: l.delegate.Reversed()}
}

// Distinct returns a new immutable list with duplicates removed.
func (l *ImmutableInt32ArrayList) Distinct() *ImmutableInt32ArrayList {
	return &ImmutableInt32ArrayList{delegate: l.delegate.Distinct()}
}

// ToSlice returns a copy of all elements as a slice.
func (l *ImmutableInt32ArrayList) ToSlice() []int32 {
	return l.delegate.ToSlice()
}

// String returns a string representation.
func (l *ImmutableInt32ArrayList) String() string {
	return l.delegate.String()
}

// Equals returns true if the other immutable list has the same elements in order.
func (l *ImmutableInt32ArrayList) Equals(other *ImmutableInt32ArrayList) bool {
	return l.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this list.
func (l *ImmutableInt32ArrayList) ToMutable() *Int32ArrayList {
	return Int32ArrayListOf(l.ToSlice()...)
}
