
package treeset

import (
	"fmt"
	"iter"
	"strings"
)

const (
	float32TreeSetNodeRed   = false
	float32TreeSetNodeBlack = true
)

type float32TreeSetNode struct {
	key    float32
	left   *float32TreeSetNode
	right  *float32TreeSetNode
	parent *float32TreeSetNode
	color  bool
}

// Float32TreeSet is a sorted set of float32 values, backed by a red-black tree.
type Float32TreeSet struct {
	root *float32TreeSetNode
	size int
}

// NewFloat32TreeSet creates a new empty sorted set.
func NewFloat32TreeSet() *Float32TreeSet {
	return &Float32TreeSet{}
}

// Float32TreeSetOf creates a new sorted set from the given values.
func Float32TreeSetOf(values ...float32) *Float32TreeSet {
	s := NewFloat32TreeSet()
	for _, v := range values {
		s.Add(v)
	}
	return s
}

// Add inserts a value. Returns true if added (not already present).
func (s *Float32TreeSet) Add(value float32) bool {
	if s.root == nil {
		s.root = &float32TreeSetNode{key: value, color: float32TreeSetNodeBlack}
		s.size++
		return true
	}
	node := s.root
	for {
		if value < node.key {
			if node.left == nil {
				node.left = &float32TreeSetNode{key: value, parent: node, color: float32TreeSetNodeRed}
				s.fixAfterInsert(node.left)
				s.size++
				return true
			}
			node = node.left
		} else if value > node.key {
			if node.right == nil {
				node.right = &float32TreeSetNode{key: value, parent: node, color: float32TreeSetNodeRed}
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
func (s *Float32TreeSet) Remove(value float32) bool {
	node := s.findNode(value)
	if node == nil {
		return false
	}
	s.deleteNode(node)
	s.size--
	return true
}

// Contains returns true if the set contains the value.
func (s *Float32TreeSet) Contains(value float32) bool {
	return s.findNode(value) != nil
}

// Size returns the number of elements.
func (s *Float32TreeSet) Size() int { return s.size }

// IsEmpty returns true if the set is empty.
func (s *Float32TreeSet) IsEmpty() bool { return s.size == 0 }

// Clear removes all elements.
func (s *Float32TreeSet) Clear() { s.root = nil; s.size = 0 }

// Min returns the smallest element, or zero and false if empty.
func (s *Float32TreeSet) Min() (float32, bool) {
	if s.root == nil {
		return 0.0, false
	}
	return s.minNode(s.root).key, true
}

// Max returns the largest element, or zero and false if empty.
func (s *Float32TreeSet) Max() (float32, bool) {
	if s.root == nil {
		return 0.0, false
	}
	return s.maxNode(s.root).key, true
}

// Floor returns the largest element <= value, or zero and false.
func (s *Float32TreeSet) Floor(value float32) (float32, bool) {
	var result *float32TreeSetNode
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
		return 0.0, false
	}
	return result.key, true
}

// Ceiling returns the smallest element >= value, or zero and false.
func (s *Float32TreeSet) Ceiling(value float32) (float32, bool) {
	var result *float32TreeSetNode
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
		return 0.0, false
	}
	return result.key, true
}

// All returns an iter.Seq that yields elements in ascending order.
func (s *Float32TreeSet) All() iter.Seq[float32] {
	return func(yield func(float32) bool) {
		var inorder func(node *float32TreeSetNode) bool
		inorder = func(node *float32TreeSetNode) bool {
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
func (s *Float32TreeSet) RangeValues(from, to float32) iter.Seq[float32] {
	return func(yield func(float32) bool) {
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
func (s *Float32TreeSet) ForEach(f func(float32)) {
	for v := range s.All() {
		f(v)
	}
}

// Select returns a new sorted set with elements satisfying the predicate.
func (s *Float32TreeSet) Select(predicate func(float32) bool) *Float32TreeSet {
	result := NewFloat32TreeSet()
	for v := range s.All() {
		if predicate(v) {
			result.Add(v)
		}
	}
	return result
}

// Reject returns a new sorted set with elements NOT satisfying the predicate.
func (s *Float32TreeSet) Reject(predicate func(float32) bool) *Float32TreeSet {
	result := NewFloat32TreeSet()
	for v := range s.All() {
		if !predicate(v) {
			result.Add(v)
		}
	}
	return result
}

// Detect returns the first element satisfying the predicate, or (zero, false) if none.
func (s *Float32TreeSet) Detect(predicate func(float32) bool) (float32, bool) {
	for v := range s.All() {
		if predicate(v) {
			return v, true
		}
	}
	var zero float32
	return zero, false
}

// AnySatisfy returns true if any element satisfies the predicate.
func (s *Float32TreeSet) AnySatisfy(predicate func(float32) bool) bool {
	for v := range s.All() {
		if predicate(v) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (s *Float32TreeSet) AllSatisfy(predicate func(float32) bool) bool {
	for v := range s.All() {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (s *Float32TreeSet) NoneSatisfy(predicate func(float32) bool) bool {
	for v := range s.All() {
		if predicate(v) {
			return false
		}
	}
	return true
}

// Count returns the number of elements satisfying the predicate.
func (s *Float32TreeSet) Count(predicate func(float32) bool) int {
	c := 0
	for v := range s.All() {
		if predicate(v) {
			c++
		}
	}
	return c
}

// Union returns a new sorted set with elements from both sets.
func (s *Float32TreeSet) Union(other *Float32TreeSet) *Float32TreeSet {
	result := NewFloat32TreeSet()
	for v := range s.All() {
		result.Add(v)
	}
	for v := range other.All() {
		result.Add(v)
	}
	return result
}

// Intersect returns a new sorted set with elements in both sets.
func (s *Float32TreeSet) Intersect(other *Float32TreeSet) *Float32TreeSet {
	result := NewFloat32TreeSet()
	for v := range s.All() {
		if other.Contains(v) {
			result.Add(v)
		}
	}
	return result
}

// Difference returns a new sorted set with elements in this but not other.
func (s *Float32TreeSet) Difference(other *Float32TreeSet) *Float32TreeSet {
	result := NewFloat32TreeSet()
	for v := range s.All() {
		if !other.Contains(v) {
			result.Add(v)
		}
	}
	return result
}

// ToSlice returns elements as a sorted slice.
func (s *Float32TreeSet) ToSlice() []float32 {
	result := make([]float32, 0, s.size)
	for v := range s.All() {
		result = append(result, v)
	}
	return result
}

// With returns the set after adding the value (fluent API).
func (s *Float32TreeSet) With(value float32) *Float32TreeSet { s.Add(value); return s }

// Without returns the set after removing the value (fluent API).
func (s *Float32TreeSet) Without(value float32) *Float32TreeSet { s.Remove(value); return s }

// String returns a string representation in sorted order.
func (s *Float32TreeSet) String() string {
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

func (s *Float32TreeSet) findNode(key float32) *float32TreeSetNode {
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
func (s *Float32TreeSet) minNode(node *float32TreeSetNode) *float32TreeSetNode {
	for node.left != nil {
		node = node.left
	}
	return node
}
func (s *Float32TreeSet) maxNode(node *float32TreeSetNode) *float32TreeSetNode {
	for node.right != nil {
		node = node.right
	}
	return node
}

func (s *Float32TreeSet) rotateLeft(x *float32TreeSetNode) {
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
func (s *Float32TreeSet) rotateRight(x *float32TreeSetNode) {
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

func (s *Float32TreeSet) fixAfterInsert(z *float32TreeSetNode) {
	for z.parent != nil && z.parent.color == float32TreeSetNodeRed {
		if z.parent == z.parent.parent.left {
			y := z.parent.parent.right
			if y != nil && y.color == float32TreeSetNodeRed {
				z.parent.color = float32TreeSetNodeBlack
				y.color = float32TreeSetNodeBlack
				z.parent.parent.color = float32TreeSetNodeRed
				z = z.parent.parent
			} else {
				if z == z.parent.right {
					z = z.parent
					s.rotateLeft(z)
				}
				z.parent.color = float32TreeSetNodeBlack
				z.parent.parent.color = float32TreeSetNodeRed
				s.rotateRight(z.parent.parent)
			}
		} else {
			y := z.parent.parent.left
			if y != nil && y.color == float32TreeSetNodeRed {
				z.parent.color = float32TreeSetNodeBlack
				y.color = float32TreeSetNodeBlack
				z.parent.parent.color = float32TreeSetNodeRed
				z = z.parent.parent
			} else {
				if z == z.parent.left {
					z = z.parent
					s.rotateRight(z)
				}
				z.parent.color = float32TreeSetNodeBlack
				z.parent.parent.color = float32TreeSetNodeRed
				s.rotateLeft(z.parent.parent)
			}
		}
	}
	s.root.color = float32TreeSetNodeBlack
}

func (s *Float32TreeSet) deleteNode(z *float32TreeSetNode) {
	if z.left != nil && z.right != nil {
		succ := s.minNode(z.right)
		z.key = succ.key
		z = succ
	}
	var child *float32TreeSetNode
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
		if z.color == float32TreeSetNodeBlack {
			s.fixAfterDelete(child)
		}
	} else if z.parent == nil {
		s.root = nil
	} else {
		if z.color == float32TreeSetNodeBlack {
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

func (s *Float32TreeSet) fixAfterDelete(x *float32TreeSetNode) {
	for x != s.root && x.color == float32TreeSetNodeBlack {
		if x == x.parent.left {
			w := x.parent.right
			if w == nil {
				x = x.parent
				continue
			}
			if w.color == float32TreeSetNodeRed {
				w.color = float32TreeSetNodeBlack
				x.parent.color = float32TreeSetNodeRed
				s.rotateLeft(x.parent)
				w = x.parent.right
			}
			if w == nil {
				x = x.parent
				continue
			}
			lb := w.left == nil || w.left.color == float32TreeSetNodeBlack
			rb := w.right == nil || w.right.color == float32TreeSetNodeBlack
			if lb && rb {
				w.color = float32TreeSetNodeRed
				x = x.parent
			} else {
				if rb {
					if w.left != nil {
						w.left.color = float32TreeSetNodeBlack
					}
					w.color = float32TreeSetNodeRed
					s.rotateRight(w)
					w = x.parent.right
				}
				w.color = x.parent.color
				x.parent.color = float32TreeSetNodeBlack
				if w.right != nil {
					w.right.color = float32TreeSetNodeBlack
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
			if w.color == float32TreeSetNodeRed {
				w.color = float32TreeSetNodeBlack
				x.parent.color = float32TreeSetNodeRed
				s.rotateRight(x.parent)
				w = x.parent.left
			}
			if w == nil {
				x = x.parent
				continue
			}
			lb := w.left == nil || w.left.color == float32TreeSetNodeBlack
			rb := w.right == nil || w.right.color == float32TreeSetNodeBlack
			if lb && rb {
				w.color = float32TreeSetNodeRed
				x = x.parent
			} else {
				if lb {
					if w.right != nil {
						w.right.color = float32TreeSetNodeBlack
					}
					w.color = float32TreeSetNodeRed
					s.rotateLeft(w)
					w = x.parent.left
				}
				w.color = x.parent.color
				x.parent.color = float32TreeSetNodeBlack
				if w.left != nil {
					w.left.color = float32TreeSetNodeBlack
				}
				s.rotateRight(x.parent)
				x = s.root
			}
		}
	}
	x.color = float32TreeSetNodeBlack
}
