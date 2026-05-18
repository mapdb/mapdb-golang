
package treeset

import (
	"fmt"
	"iter"
	"strings"
)

const (
	int16TreeSetNodeRed   = false
	int16TreeSetNodeBlack = true
)

type int16TreeSetNode struct {
	key    int16
	left   *int16TreeSetNode
	right  *int16TreeSetNode
	parent *int16TreeSetNode
	color  bool
}

// Int16TreeSet is a sorted set of int16 values, backed by a red-black tree.
type Int16TreeSet struct {
	root *int16TreeSetNode
	size int
}

// NewInt16TreeSet creates a new empty sorted set.
func NewInt16TreeSet() *Int16TreeSet {
	return &Int16TreeSet{}
}

// Int16TreeSetOf creates a new sorted set from the given values.
func Int16TreeSetOf(values ...int16) *Int16TreeSet {
	s := NewInt16TreeSet()
	for _, v := range values {
		s.Add(v)
	}
	return s
}

// Add inserts a value. Returns true if added (not already present).
func (s *Int16TreeSet) Add(value int16) bool {
	if s.root == nil {
		s.root = &int16TreeSetNode{key: value, color: int16TreeSetNodeBlack}
		s.size++
		return true
	}
	node := s.root
	for {
		if value < node.key {
			if node.left == nil {
				node.left = &int16TreeSetNode{key: value, parent: node, color: int16TreeSetNodeRed}
				s.fixAfterInsert(node.left)
				s.size++
				return true
			}
			node = node.left
		} else if value > node.key {
			if node.right == nil {
				node.right = &int16TreeSetNode{key: value, parent: node, color: int16TreeSetNodeRed}
				s.fixAfterInsert(node.right)
				s.size++
				return true
			}
			node = node.right
		} else {
			return false // already exists
		}
	}
}

// Remove removes a value. Returns true if found and removed.
func (s *Int16TreeSet) Remove(value int16) bool {
	node := s.findNode(value)
	if node == nil {
		return false
	}
	s.deleteNode(node)
	s.size--
	return true
}

// Contains returns true if the set contains the value.
func (s *Int16TreeSet) Contains(value int16) bool {
	return s.findNode(value) != nil
}

// Size returns the number of elements.
func (s *Int16TreeSet) Size() int { return s.size }

// IsEmpty returns true if the set is empty.
func (s *Int16TreeSet) IsEmpty() bool { return s.size == 0 }

// Clear removes all elements.
func (s *Int16TreeSet) Clear() { s.root = nil; s.size = 0 }

// Min returns the smallest element, or zero and false if empty.
func (s *Int16TreeSet) Min() (int16, bool) {
	if s.root == nil {
		return 0, false
	}
	return s.minNode(s.root).key, true
}

// Max returns the largest element, or zero and false if empty.
func (s *Int16TreeSet) Max() (int16, bool) {
	if s.root == nil {
		return 0, false
	}
	return s.maxNode(s.root).key, true
}

// Floor returns the largest element <= value, or zero and false.
func (s *Int16TreeSet) Floor(value int16) (int16, bool) {
	var result *int16TreeSetNode
	node := s.root
	for node != nil {
		if value == node.key {
			return node.key, true
		}
		if value > node.key {
			result = node
			node = node.right
		} else {
			node = node.left
		}
	}
	if result == nil {
		return 0, false
	}
	return result.key, true
}

// Ceiling returns the smallest element >= value, or zero and false.
func (s *Int16TreeSet) Ceiling(value int16) (int16, bool) {
	var result *int16TreeSetNode
	node := s.root
	for node != nil {
		if value == node.key {
			return node.key, true
		}
		if value < node.key {
			result = node
			node = node.left
		} else {
			node = node.right
		}
	}
	if result == nil {
		return 0, false
	}
	return result.key, true
}

// All returns an iter.Seq that yields elements in ascending order.
func (s *Int16TreeSet) All() iter.Seq[int16] {
	return func(yield func(int16) bool) {
		var inorder func(node *int16TreeSetNode) bool
		inorder = func(node *int16TreeSetNode) bool {
			if node == nil {
				return true
			}
			if !inorder(node.left) {
				return false
			}
			if !yield(node.key) {
				return false
			}
			return inorder(node.right)
		}
		inorder(s.root)
	}
}

// RangeValues returns an iter.Seq that yields elements in [from, to).
func (s *Int16TreeSet) RangeValues(from, to int16) iter.Seq[int16] {
	return func(yield func(int16) bool) {
		for v := range s.All() {
			if v < from {
				continue
			}
			if v >= to {
				return
			}
			if !yield(v) {
				return
			}
		}
	}
}

// ForEach calls the function for each element in ascending order.
func (s *Int16TreeSet) ForEach(f func(int16)) {
	for v := range s.All() {
		f(v)
	}
}

// Select returns a new sorted set with elements satisfying the predicate.
func (s *Int16TreeSet) Select(predicate func(int16) bool) *Int16TreeSet {
	result := NewInt16TreeSet()
	for v := range s.All() {
		if predicate(v) {
			result.Add(v)
		}
	}
	return result
}

// Reject returns a new sorted set with elements NOT satisfying the predicate.
func (s *Int16TreeSet) Reject(predicate func(int16) bool) *Int16TreeSet {
	result := NewInt16TreeSet()
	for v := range s.All() {
		if !predicate(v) {
			result.Add(v)
		}
	}
	return result
}

// Detect returns the first element satisfying the predicate, or (zero, false) if none.
func (s *Int16TreeSet) Detect(predicate func(int16) bool) (int16, bool) {
	for v := range s.All() {
		if predicate(v) {
			return v, true
		}
	}
	var zero int16
	return zero, false
}

// AnySatisfy returns true if any element satisfies the predicate.
func (s *Int16TreeSet) AnySatisfy(predicate func(int16) bool) bool {
	for v := range s.All() {
		if predicate(v) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (s *Int16TreeSet) AllSatisfy(predicate func(int16) bool) bool {
	for v := range s.All() {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (s *Int16TreeSet) NoneSatisfy(predicate func(int16) bool) bool {
	for v := range s.All() {
		if predicate(v) {
			return false
		}
	}
	return true
}

// Count returns the number of elements satisfying the predicate.
func (s *Int16TreeSet) Count(predicate func(int16) bool) int {
	c := 0
	for v := range s.All() {
		if predicate(v) {
			c++
		}
	}
	return c
}

// Union returns a new sorted set with elements from both sets.
func (s *Int16TreeSet) Union(other *Int16TreeSet) *Int16TreeSet {
	result := NewInt16TreeSet()
	for v := range s.All() {
		result.Add(v)
	}
	for v := range other.All() {
		result.Add(v)
	}
	return result
}

// Intersect returns a new sorted set with elements in both sets.
func (s *Int16TreeSet) Intersect(other *Int16TreeSet) *Int16TreeSet {
	result := NewInt16TreeSet()
	for v := range s.All() {
		if other.Contains(v) {
			result.Add(v)
		}
	}
	return result
}

// Difference returns a new sorted set with elements in this but not other.
func (s *Int16TreeSet) Difference(other *Int16TreeSet) *Int16TreeSet {
	result := NewInt16TreeSet()
	for v := range s.All() {
		if !other.Contains(v) {
			result.Add(v)
		}
	}
	return result
}

// ToSlice returns elements as a sorted slice.
func (s *Int16TreeSet) ToSlice() []int16 {
	result := make([]int16, 0, s.size)
	for v := range s.All() {
		result = append(result, v)
	}
	return result
}

// With returns the set after adding the value (fluent API).
func (s *Int16TreeSet) With(value int16) *Int16TreeSet { s.Add(value); return s }

// Without returns the set after removing the value (fluent API).
func (s *Int16TreeSet) Without(value int16) *Int16TreeSet { s.Remove(value); return s }

// String returns a string representation in sorted order.
func (s *Int16TreeSet) String() string {
	if s.size == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for v := range s.All() {
		if !first {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v", v)
		first = false
	}
	sb.WriteString("}")
	return sb.String()
}

// --- Red-black tree internals (same as TreeMap) ---

func (s *Int16TreeSet) findNode(key int16) *int16TreeSetNode {
	node := s.root
	for node != nil {
		if key < node.key {
			node = node.left
		} else if key > node.key {
			node = node.right
		} else {
			return node
		}
	}
	return nil
}
func (s *Int16TreeSet) minNode(node *int16TreeSetNode) *int16TreeSetNode {
	for node.left != nil {
		node = node.left
	}
	return node
}
func (s *Int16TreeSet) maxNode(node *int16TreeSetNode) *int16TreeSetNode {
	for node.right != nil {
		node = node.right
	}
	return node
}

func (s *Int16TreeSet) rotateLeft(x *int16TreeSetNode) {
	y := x.right
	x.right = y.left
	if y.left != nil {
		y.left.parent = x
	}
	y.parent = x.parent
	if x.parent == nil {
		s.root = y
	} else if x == x.parent.left {
		x.parent.left = y
	} else {
		x.parent.right = y
	}
	y.left = x
	x.parent = y
}
func (s *Int16TreeSet) rotateRight(x *int16TreeSetNode) {
	y := x.left
	x.left = y.right
	if y.right != nil {
		y.right.parent = x
	}
	y.parent = x.parent
	if x.parent == nil {
		s.root = y
	} else if x == x.parent.right {
		x.parent.right = y
	} else {
		x.parent.left = y
	}
	y.right = x
	x.parent = y
}

func (s *Int16TreeSet) fixAfterInsert(z *int16TreeSetNode) {
	for z.parent != nil && z.parent.color == int16TreeSetNodeRed {
		if z.parent == z.parent.parent.left {
			y := z.parent.parent.right
			if y != nil && y.color == int16TreeSetNodeRed {
				z.parent.color = int16TreeSetNodeBlack
				y.color = int16TreeSetNodeBlack
				z.parent.parent.color = int16TreeSetNodeRed
				z = z.parent.parent
			} else {
				if z == z.parent.right {
					z = z.parent
					s.rotateLeft(z)
				}
				z.parent.color = int16TreeSetNodeBlack
				z.parent.parent.color = int16TreeSetNodeRed
				s.rotateRight(z.parent.parent)
			}
		} else {
			y := z.parent.parent.left
			if y != nil && y.color == int16TreeSetNodeRed {
				z.parent.color = int16TreeSetNodeBlack
				y.color = int16TreeSetNodeBlack
				z.parent.parent.color = int16TreeSetNodeRed
				z = z.parent.parent
			} else {
				if z == z.parent.left {
					z = z.parent
					s.rotateRight(z)
				}
				z.parent.color = int16TreeSetNodeBlack
				z.parent.parent.color = int16TreeSetNodeRed
				s.rotateLeft(z.parent.parent)
			}
		}
	}
	s.root.color = int16TreeSetNodeBlack
}

func (s *Int16TreeSet) deleteNode(z *int16TreeSetNode) {
	if z.left != nil && z.right != nil {
		succ := s.minNode(z.right)
		z.key = succ.key
		z = succ
	}
	var child *int16TreeSetNode
	if z.left != nil {
		child = z.left
	} else {
		child = z.right
	}
	if child != nil {
		child.parent = z.parent
		if z.parent == nil {
			s.root = child
		} else if z == z.parent.left {
			z.parent.left = child
		} else {
			z.parent.right = child
		}
		if z.color == int16TreeSetNodeBlack {
			s.fixAfterDelete(child)
		}
	} else if z.parent == nil {
		s.root = nil
	} else {
		if z.color == int16TreeSetNodeBlack {
			s.fixAfterDelete(z)
		}
		if z.parent != nil {
			if z == z.parent.left {
				z.parent.left = nil
			} else {
				z.parent.right = nil
			}
		}
	}
}

func (s *Int16TreeSet) fixAfterDelete(x *int16TreeSetNode) {
	for x != s.root && x.color == int16TreeSetNodeBlack {
		if x == x.parent.left {
			w := x.parent.right
			if w == nil {
				x = x.parent
				continue
			}
			if w.color == int16TreeSetNodeRed {
				w.color = int16TreeSetNodeBlack
				x.parent.color = int16TreeSetNodeRed
				s.rotateLeft(x.parent)
				w = x.parent.right
			}
			if w == nil {
				x = x.parent
				continue
			}
			lb := w.left == nil || w.left.color == int16TreeSetNodeBlack
			rb := w.right == nil || w.right.color == int16TreeSetNodeBlack
			if lb && rb {
				w.color = int16TreeSetNodeRed
				x = x.parent
			} else {
				if rb {
					if w.left != nil {
						w.left.color = int16TreeSetNodeBlack
					}
					w.color = int16TreeSetNodeRed
					s.rotateRight(w)
					w = x.parent.right
				}
				w.color = x.parent.color
				x.parent.color = int16TreeSetNodeBlack
				if w.right != nil {
					w.right.color = int16TreeSetNodeBlack
				}
				s.rotateLeft(x.parent)
				x = s.root
			}
		} else {
			w := x.parent.left
			if w == nil {
				x = x.parent
				continue
			}
			if w.color == int16TreeSetNodeRed {
				w.color = int16TreeSetNodeBlack
				x.parent.color = int16TreeSetNodeRed
				s.rotateRight(x.parent)
				w = x.parent.left
			}
			if w == nil {
				x = x.parent
				continue
			}
			lb := w.left == nil || w.left.color == int16TreeSetNodeBlack
			rb := w.right == nil || w.right.color == int16TreeSetNodeBlack
			if lb && rb {
				w.color = int16TreeSetNodeRed
				x = x.parent
			} else {
				if lb {
					if w.right != nil {
						w.right.color = int16TreeSetNodeBlack
					}
					w.color = int16TreeSetNodeRed
					s.rotateLeft(w)
					w = x.parent.left
				}
				w.color = x.parent.color
				x.parent.color = int16TreeSetNodeBlack
				if w.left != nil {
					w.left.color = int16TreeSetNodeBlack
				}
				s.rotateRight(x.parent)
				x = s.root
			}
		}
	}
	x.color = int16TreeSetNodeBlack
}
