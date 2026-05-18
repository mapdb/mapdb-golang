
package stack

import (
	"fmt"
	"iter"
)

// ImmutableCharArrayStack is an immutable LIFO stack of uint16 values.
type ImmutableCharArrayStack struct {
	delegate *CharArrayStack
}

// NewImmutableCharArrayStack creates an immutable stack from the given values.
// The last value becomes the top of the stack.
func NewImmutableCharArrayStack(values ...uint16) *ImmutableCharArrayStack {
	return &ImmutableCharArrayStack{delegate: CharArrayStackOf(values...)}
}

// ImmutableCharArrayStackFrom creates an immutable copy of a mutable stack.
func ImmutableCharArrayStackFrom(s *CharArrayStack) *ImmutableCharArrayStack {
	copy := &CharArrayStack{items: make([]uint16, len(s.items))}
	for i := range s.items {
		copy.items[i] = s.items[i]
	}
	return &ImmutableCharArrayStack{delegate: copy}
}

// Peek returns the top value without removing it, or an error if the stack is empty.
func (s *ImmutableCharArrayStack) Peek() (uint16, error) {
	return s.delegate.Peek()
}

// PeekAt returns the element at the given distance from the top,
// or an error if the index is out of bounds.
func (s *ImmutableCharArrayStack) PeekAt(index int) (uint16, error) {
	return s.delegate.PeekAt(index)
}

// Size returns the number of elements.
func (s *ImmutableCharArrayStack) Size() int {
	return s.delegate.Size()
}

// IsEmpty returns true if the stack contains no elements.
func (s *ImmutableCharArrayStack) IsEmpty() bool {
	return s.delegate.IsEmpty()
}

// Contains returns true if the stack contains the given value.
func (s *ImmutableCharArrayStack) Contains(value uint16) bool {
	return s.delegate.Contains(value)
}

// All returns an iter.Seq that yields elements from top to bottom.
func (s *ImmutableCharArrayStack) All() iter.Seq[uint16] {
	return s.delegate.All()
}

// ForEach calls the given function for each element from top to bottom.
func (s *ImmutableCharArrayStack) ForEach(f func(uint16)) {
	s.delegate.ForEach(f)
}

// Select returns a new immutable stack with elements satisfying the predicate.
func (s *ImmutableCharArrayStack) Select(predicate func(uint16) bool) *ImmutableCharArrayStack {
	return &ImmutableCharArrayStack{delegate: s.delegate.Select(predicate)}
}

// Reject returns a new immutable stack with elements not satisfying the predicate.
func (s *ImmutableCharArrayStack) Reject(predicate func(uint16) bool) *ImmutableCharArrayStack {
	return &ImmutableCharArrayStack{delegate: s.delegate.Reject(predicate)}
}

// Detect returns the first element from top satisfying the predicate, or zero and false.
func (s *ImmutableCharArrayStack) Detect(predicate func(uint16) bool) (uint16, bool) {
	return s.delegate.Detect(predicate)
}

// AnySatisfy returns true if any element satisfies the predicate.
func (s *ImmutableCharArrayStack) AnySatisfy(predicate func(uint16) bool) bool {
	return s.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (s *ImmutableCharArrayStack) AllSatisfy(predicate func(uint16) bool) bool {
	return s.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (s *ImmutableCharArrayStack) NoneSatisfy(predicate func(uint16) bool) bool {
	return s.delegate.NoneSatisfy(predicate)
}

// Count returns the number of elements satisfying the predicate.
func (s *ImmutableCharArrayStack) Count(predicate func(uint16) bool) int {
	return s.delegate.Count(predicate)
}

// ToSlice returns all elements as a slice (top first).
func (s *ImmutableCharArrayStack) ToSlice() []uint16 {
	return s.delegate.ToSlice()
}

// String returns a string representation.
func (s *ImmutableCharArrayStack) String() string {
	return s.delegate.String()
}

// Equals returns true if the other immutable stack has the same elements.
func (s *ImmutableCharArrayStack) Equals(other *ImmutableCharArrayStack) bool {
	return s.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this stack.
func (s *ImmutableCharArrayStack) ToMutable() *CharArrayStack {
	copy := &CharArrayStack{items: make([]uint16, len(s.delegate.items))}
	for i := range s.delegate.items {
		copy.items[i] = s.delegate.items[i]
	}
	return copy
}

// Push returns a NEW immutable stack with the value pushed on top.
// The original stack is not modified.
func (s *ImmutableCharArrayStack) Push(value uint16) *ImmutableCharArrayStack {
	newItems := make([]uint16, len(s.delegate.items)+1)
	copy(newItems, s.delegate.items)
	newItems[len(s.delegate.items)] = value
	return &ImmutableCharArrayStack{delegate: &CharArrayStack{items: newItems}}
}

// Pop returns a NEW immutable stack with the top element removed, and the removed value.
// The original stack is not modified. Returns an error if the stack is empty.
func (s *ImmutableCharArrayStack) Pop() (*ImmutableCharArrayStack, uint16, error) {
	if s.delegate.IsEmpty() {
		return nil, 0, fmt.Errorf("ImmutableCharArrayStack: Pop on empty stack")
	}
	top := s.delegate.items[len(s.delegate.items)-1]
	newItems := make([]uint16, len(s.delegate.items)-1)
	copy(newItems, s.delegate.items[:len(s.delegate.items)-1])
	return &ImmutableCharArrayStack{delegate: &CharArrayStack{items: newItems}}, top, nil
}
