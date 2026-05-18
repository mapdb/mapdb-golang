
package stack

import (
	"fmt"
	"iter"
	"strings"
)

// CharArrayStack is a LIFO (last-in, first-out) stack backed by a uint16 slice.
type CharArrayStack struct {
	items []uint16
}

// NewCharArrayStack creates a new empty CharArrayStack.
func NewCharArrayStack() *CharArrayStack {
	return &CharArrayStack{
		items: make([]uint16, 0, 16),
	}
}

// CharArrayStackOf creates a new CharArrayStack from the given values.
// The last value becomes the top of the stack.
func CharArrayStackOf(values ...uint16) *CharArrayStack {
	s := &CharArrayStack{
		items: make([]uint16, len(values)),
	}
	copy(s.items, values)
	return s
}

// Push adds a value to the top of the stack.
func (s *CharArrayStack) Push(value uint16) {
	s.items = append(s.items, value)
}

// Pop removes and returns the top value, or an error if the stack is empty.
func (s *CharArrayStack) Pop() (uint16, error) {
	if len(s.items) == 0 {
		return 0, fmt.Errorf("CharArrayStack: Pop on empty stack")
	}
	top := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return top, nil
}

// Peek returns the top value without removing it, or an error if the stack is empty.
func (s *CharArrayStack) Peek() (uint16, error) {
	if len(s.items) == 0 {
		return 0, fmt.Errorf("CharArrayStack: Peek on empty stack")
	}
	return s.items[len(s.items)-1], nil
}

// PeekAt returns the element at the given distance from the top (0 = top),
// or an error if the index is out of bounds.
func (s *CharArrayStack) PeekAt(index int) (uint16, error) {
	if index < 0 || index >= len(s.items) {
		return 0, fmt.Errorf("CharArrayStack: PeekAt index out of bounds: %d (size %d)", index, len(s.items))
	}
	return s.items[len(s.items)-1-index], nil
}

// Size returns the number of elements in the stack.
func (s *CharArrayStack) Size() int {
	return len(s.items)
}

// IsEmpty returns true if the stack contains no elements.
func (s *CharArrayStack) IsEmpty() bool {
	return len(s.items) == 0
}

// Clear removes all elements from the stack.
func (s *CharArrayStack) Clear() {
	s.items = s.items[:0]
}

// Contains returns true if the stack contains the given value.
func (s *CharArrayStack) Contains(value uint16) bool {
	for _, v := range s.items {
		if v == value {
			return true
		}
	}
	return false
}

// All returns an iter.Seq that yields elements from top to bottom.
func (s *CharArrayStack) All() iter.Seq[uint16] {
	return func(yield func(uint16) bool) {
		for i := len(s.items) - 1; i >= 0; i-- {
			if !yield(s.items[i]) {
				return
			}
		}
	}
}

// ForEach calls the given function for each element from top to bottom.
func (s *CharArrayStack) ForEach(f func(uint16)) {
	for i := len(s.items) - 1; i >= 0; i-- {
		f(s.items[i])
	}
}

// Select returns a new stack containing only elements that satisfy the predicate.
// Order is preserved (top of result corresponds to top of original that passed).
func (s *CharArrayStack) Select(predicate func(uint16) bool) *CharArrayStack {
	result := NewCharArrayStack()
	for _, v := range s.items {
		if predicate(v) {
			result.Push(v)
		}
	}
	return result
}

// Reject returns a new stack containing only elements that do not satisfy the predicate.
func (s *CharArrayStack) Reject(predicate func(uint16) bool) *CharArrayStack {
	result := NewCharArrayStack()
	for _, v := range s.items {
		if !predicate(v) {
			result.Push(v)
		}
	}
	return result
}

// Detect returns the first element from the top that satisfies the predicate, or zero and false.
func (s *CharArrayStack) Detect(predicate func(uint16) bool) (uint16, bool) {
	for i := len(s.items) - 1; i >= 0; i-- {
		if predicate(s.items[i]) {
			return s.items[i], true
		}
	}
	return 0, false
}

// AnySatisfy returns true if any element satisfies the predicate.
func (s *CharArrayStack) AnySatisfy(predicate func(uint16) bool) bool {
	for _, v := range s.items {
		if predicate(v) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (s *CharArrayStack) AllSatisfy(predicate func(uint16) bool) bool {
	for _, v := range s.items {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (s *CharArrayStack) NoneSatisfy(predicate func(uint16) bool) bool {
	for _, v := range s.items {
		if predicate(v) {
			return false
		}
	}
	return true
}

// Count returns the number of elements that satisfy the predicate.
func (s *CharArrayStack) Count(predicate func(uint16) bool) int {
	count := 0
	for _, v := range s.items {
		if predicate(v) {
			count++
		}
	}
	return count
}

// InjectInto performs a left fold from bottom to top.
func (s *CharArrayStack) InjectInto(initial uint16, f func(uint16, uint16) uint16) uint16 {
	result := initial
	for _, v := range s.items {
		result = f(result, v)
	}
	return result
}

// ToSlice returns all elements as a slice (top element first).
func (s *CharArrayStack) ToSlice() []uint16 {
	result := make([]uint16, len(s.items))
	for i, j := len(s.items)-1, 0; i >= 0; i, j = i-1, j+1 {
		result[j] = s.items[i]
	}
	return result
}

// ToList returns the elements as a slice in stack order (bottom first, for internal use).
func (s *CharArrayStack) toList() []uint16 {
	result := make([]uint16, len(s.items))
	copy(result, s.items)
	return result
}

// With returns the stack after pushing the value (fluent API).
func (s *CharArrayStack) With(value uint16) *CharArrayStack {
	s.Push(value)
	return s
}

// WithAll returns the stack after pushing all values (fluent API).
func (s *CharArrayStack) WithAll(values ...uint16) *CharArrayStack {
	for _, v := range values {
		s.Push(v)
	}
	return s
}

// ToImmutable returns an immutable copy of this stack.
func (s *CharArrayStack) ToImmutable() *ImmutableCharArrayStack {
	return ImmutableCharArrayStackFrom(s)
}

// String returns a string representation of the stack (top element first).
func (s *CharArrayStack) String() string {
	if len(s.items) == 0 {
		return "[]"
	}
	var sb strings.Builder
	sb.WriteString("[")
	for i := len(s.items) - 1; i >= 0; i-- {
		if i < len(s.items)-1 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v", s.items[i])
	}
	sb.WriteString("]")
	return sb.String()
}

// Equals returns true if the other stack has the same elements in the same order.
func (s *CharArrayStack) Equals(other *CharArrayStack) bool {
	if len(s.items) != len(other.items) {
		return false
	}
	for i := range s.items {
		if s.items[i] != other.items[i] {
			return false
		}
	}
	return true
}
