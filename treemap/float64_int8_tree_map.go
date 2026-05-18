
package treemap

import (
	"fmt"
	"iter"
	"strings"
)

const (
	float64Int8TreeNodeRed   = false
	float64Int8TreeNodeBlack = true
)

type float64Int8TreeNode struct {
	key    float64
	value  int8
	left   *float64Int8TreeNode
	right  *float64Int8TreeNode
	parent *float64Int8TreeNode
	color  bool
}

// Float64Int8TreeMap is a sorted map with float64 keys and int8 values, backed by a red-black tree.
// Keys are maintained in ascending order.
type Float64Int8TreeMap struct {
	root *float64Int8TreeNode
	size int
}

// NewFloat64Int8TreeMap creates a new empty sorted map.
func NewFloat64Int8TreeMap() *Float64Int8TreeMap {
	return &Float64Int8TreeMap{}
}

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *Float64Int8TreeMap) Put(key float64, value int8) (int8, bool) {
	if m.root == nil {
		m.root = &float64Int8TreeNode{key: key, value: value, color: float64Int8TreeNodeBlack}
		m.size++
		return 0, false
	}
	node := m.root
	for {
		if cmpFloat64(key, node.key) < 0 {
			if node.left == nil {
				node.left = &float64Int8TreeNode{key: key, value: value, parent: node, color: float64Int8TreeNodeRed}
				m.fixAfterInsert(node.left)
				m.size++
				return 0, false
			}
			node = node.left
		} else if cmpFloat64(key, node.key) > 0 {
			if node.right == nil {
				node.right = &float64Int8TreeNode{key: key, value: value, parent: node, color: float64Int8TreeNodeRed}
				m.fixAfterInsert(node.right)
				m.size++
				return 0, false
			}
			node = node.right
		} else {
			old := node.value
			node.value = value
			return old, true
		}
	}
}

// Get returns the value for the key, or the zero value and false if not found.
func (m *Float64Int8TreeMap) Get(key float64) (int8, bool) {
	node := m.findNode(key)
	if node == nil {
		return 0, false
	}
	return node.value, true
}

// GetOrDefault returns the value for the key if present, or the default value otherwise.
func (m *Float64Int8TreeMap) GetOrDefault(key float64, defaultValue int8) int8 {
	if v, ok := m.Get(key); ok {
		return v
	}
	return defaultValue
}

// ContainsKey returns true if the map contains the given key.
func (m *Float64Int8TreeMap) ContainsKey(key float64) bool {
	return m.findNode(key) != nil
}

// Remove removes the entry for the given key. Returns the previous value and true if found.
func (m *Float64Int8TreeMap) Remove(key float64) (int8, bool) {
	node := m.findNode(key)
	if node == nil {
		return 0, false
	}
	old := node.value
	m.deleteNode(node)
	m.size--
	return old, true
}

// Size returns the number of entries.
func (m *Float64Int8TreeMap) Size() int {
	return m.size
}

// IsEmpty returns true if the map is empty.
func (m *Float64Int8TreeMap) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries.
func (m *Float64Int8TreeMap) Clear() {
	m.root = nil
	m.size = 0
}

// Min returns the smallest key and its value, or zero values and false if empty.
func (m *Float64Int8TreeMap) Min() (float64, int8, bool) {
	if m.root == nil {
		return 0, 0, false
	}
	node := m.minNode(m.root)
	return node.key, node.value, true
}

// Max returns the largest key and its value, or zero values and false if empty.
func (m *Float64Int8TreeMap) Max() (float64, int8, bool) {
	if m.root == nil {
		return 0, 0, false
	}
	node := m.maxNode(m.root)
	return node.key, node.value, true
}

// Floor returns the largest key <= the given key, or zero values and false.
func (m *Float64Int8TreeMap) Floor(key float64) (float64, int8, bool) {
	var result *float64Int8TreeNode
	node := m.root
	for node != nil {
		if key == node.key {
			return node.key, node.value, true
		}
		if cmpFloat64(key, node.key) > 0 {
			result = node
			node = node.right
		} else {
			node = node.left
		}
	}
	if result == nil {
		return 0, 0, false
	}
	return result.key, result.value, true
}

// Ceiling returns the smallest key >= the given key, or zero values and false.
func (m *Float64Int8TreeMap) Ceiling(key float64) (float64, int8, bool) {
	var result *float64Int8TreeNode
	node := m.root
	for node != nil {
		if key == node.key {
			return node.key, node.value, true
		}
		if cmpFloat64(key, node.key) < 0 {
			result = node
			node = node.left
		} else {
			node = node.right
		}
	}
	if result == nil {
		return 0, 0, false
	}
	return result.key, result.value, true
}

// All returns an iter.Seq2 that yields all key-value pairs in ascending key order.
func (m *Float64Int8TreeMap) All() iter.Seq2[float64, int8] {
	return func(yield func(float64, int8) bool) {
		var inorder func(node *float64Int8TreeNode) bool
		inorder = func(node *float64Int8TreeNode) bool {
			if node == nil {
				return true
			}
			if !inorder(node.left) {
				return false
			}
			if !yield(node.key, node.value) {
				return false
			}
			return inorder(node.right)
		}
		inorder(m.root)
	}
}

// Keys returns an iter.Seq that yields all keys in ascending order.
func (m *Float64Int8TreeMap) Keys() iter.Seq[float64] {
	return func(yield func(float64) bool) {
		for k, _ := range m.All() {
			if !yield(k) {
				return
			}
		}
	}
}

// Values returns an iter.Seq that yields all values in key order.
func (m *Float64Int8TreeMap) Values() iter.Seq[int8] {
	return func(yield func(int8) bool) {
		for _, v := range m.All() {
			if !yield(v) {
				return
			}
		}
	}
}

// RangeKeys returns an iter.Seq2 that yields entries with keys in [fromKey, toKey).
func (m *Float64Int8TreeMap) RangeKeys(fromKey, toKey float64) iter.Seq2[float64, int8] {
	return func(yield func(float64, int8) bool) {
		for k, v := range m.All() {
			if k < fromKey {
				continue
			}
			if k >= toKey {
				return
			}
			if !yield(k, v) {
				return
			}
		}
	}
}

// Higher returns the smallest key strictly greater than `key` (and its value),
// or zero values and false. Unlike Ceiling, never returns `key` itself.
func (m *Float64Int8TreeMap) Higher(key float64) (float64, int8, bool) {
	var result *float64Int8TreeNode
	node := m.root
	for node != nil {
		if cmpFloat64(key, node.key) < 0 {
			result = node
			node = node.left
		} else {
			node = node.right
		}
	}
	if result == nil {
		return 0, 0, false
	}
	return result.key, result.value, true
}

// Lower returns the largest key strictly less than `key` (and its value),
// or zero values and false. Unlike Floor, never returns `key` itself.
func (m *Float64Int8TreeMap) Lower(key float64) (float64, int8, bool) {
	var result *float64Int8TreeNode
	node := m.root
	for node != nil {
		if cmpFloat64(key, node.key) > 0 {
			result = node
			node = node.right
		} else {
			node = node.left
		}
	}
	if result == nil {
		return 0, 0, false
	}
	return result.key, result.value, true
}

// HeadMap returns an iter.Seq2 over entries with keys strictly less than toKey.
// Matches Java NavigableMap.headMap(toKey) (exclusive by default).
func (m *Float64Int8TreeMap) HeadMap(toKey float64) iter.Seq2[float64, int8] {
	return func(yield func(float64, int8) bool) {
		for k, v := range m.All() {
			if k >= toKey {
				return
			}
			if !yield(k, v) {
				return
			}
		}
	}
}

// TailMap returns an iter.Seq2 over entries with keys >= fromKey.
// Matches Java NavigableMap.tailMap(fromKey) (inclusive by default).
func (m *Float64Int8TreeMap) TailMap(fromKey float64) iter.Seq2[float64, int8] {
	return func(yield func(float64, int8) bool) {
		for k, v := range m.All() {
			if k < fromKey {
				continue
			}
			if !yield(k, v) {
				return
			}
		}
	}
}

// SubMap returns an iter.Seq2 over entries with keys in [fromKey, toKey).
// Alias for RangeKeys; exists for Java-NavigableMap API parity.
func (m *Float64Int8TreeMap) SubMap(fromKey, toKey float64) iter.Seq2[float64, int8] {
	return m.RangeKeys(fromKey, toKey)
}

// FirstEntry is an alias of Min — the smallest key and its value, or zero/false.
func (m *Float64Int8TreeMap) FirstEntry() (float64, int8, bool) { return m.Min() }

// LastEntry is an alias of Max — the largest key and its value, or zero/false.
func (m *Float64Int8TreeMap) LastEntry() (float64, int8, bool) { return m.Max() }

// PollFirstEntry removes and returns the smallest entry, or zero/false if empty.
func (m *Float64Int8TreeMap) PollFirstEntry() (float64, int8, bool) {
	k, v, ok := m.Min()
	if !ok {
		return 0, 0, false
	}
	m.Remove(k)
	return k, v, true
}

// PollLastEntry removes and returns the largest entry, or zero/false if empty.
func (m *Float64Int8TreeMap) PollLastEntry() (float64, int8, bool) {
	k, v, ok := m.Max()
	if !ok {
		return 0, 0, false
	}
	m.Remove(k)
	return k, v, true
}

// DescendingMap returns an iter.Seq2 over entries in descending key order.
func (m *Float64Int8TreeMap) DescendingMap() iter.Seq2[float64, int8] {
	return func(yield func(float64, int8) bool) {
		var reverse func(node *float64Int8TreeNode) bool
		reverse = func(node *float64Int8TreeNode) bool {
			if node == nil {
				return true
			}
			if !reverse(node.right) {
				return false
			}
			if !yield(node.key, node.value) {
				return false
			}
			return reverse(node.left)
		}
		reverse(m.root)
	}
}

// DescendingKeys returns an iter.Seq over keys in descending order.
func (m *Float64Int8TreeMap) DescendingKeys() iter.Seq[float64] {
	return func(yield func(float64) bool) {
		for k := range m.DescendingMap() {
			if !yield(k) {
				return
			}
		}
	}
}

// ForEach calls the function for each key-value pair in ascending order.
func (m *Float64Int8TreeMap) ForEach(f func(float64, int8)) {
	for k, v := range m.All() {
		f(k, v)
	}
}

// Select returns a new TreeMap with entries satisfying the predicate.
func (m *Float64Int8TreeMap) Select(predicate func(float64, int8) bool) *Float64Int8TreeMap {
	result := NewFloat64Int8TreeMap()
	for k, v := range m.All() {
		if predicate(k, v) {
			result.Put(k, v)
		}
	}
	return result
}

// Reject returns a new TreeMap with entries NOT satisfying the predicate.
func (m *Float64Int8TreeMap) Reject(predicate func(float64, int8) bool) *Float64Int8TreeMap {
	result := NewFloat64Int8TreeMap()
	for k, v := range m.All() {
		if !predicate(k, v) {
			result.Put(k, v)
		}
	}
	return result
}

// Detect returns the first entry satisfying the predicate (in key order), or (zero, zero, false).
func (m *Float64Int8TreeMap) Detect(predicate func(float64, int8) bool) (float64, int8, bool) {
	for k, v := range m.All() {
		if predicate(k, v) {
			return k, v, true
		}
	}
	var zk float64
	var zv int8
	return zk, zv, false
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *Float64Int8TreeMap) AnySatisfy(predicate func(float64, int8) bool) bool {
	for k, v := range m.All() {
		if predicate(k, v) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all entries satisfy the predicate.
func (m *Float64Int8TreeMap) AllSatisfy(predicate func(float64, int8) bool) bool {
	for k, v := range m.All() {
		if !predicate(k, v) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *Float64Int8TreeMap) NoneSatisfy(predicate func(float64, int8) bool) bool {
	for k, v := range m.All() {
		if predicate(k, v) {
			return false
		}
	}
	return true
}

// Count returns the number of entries satisfying the predicate.
func (m *Float64Int8TreeMap) Count(predicate func(float64, int8) bool) int {
	c := 0
	for k, v := range m.All() {
		if predicate(k, v) {
			c++
		}
	}
	return c
}

// String returns a string representation with entries in sorted key order.
func (m *Float64Int8TreeMap) String() string {
	if m.size == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for k, v := range m.All() {
		if !first {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v: %v", k, v)
		first = false
	}
	sb.WriteString("}")
	return sb.String()
}

// --- Red-black tree internals ---

func (m *Float64Int8TreeMap) findNode(key float64) *float64Int8TreeNode {
	node := m.root
	for node != nil {
		if cmpFloat64(key, node.key) < 0 {
			node = node.left
		} else if cmpFloat64(key, node.key) > 0 {
			node = node.right
		} else {
			return node
		}
	}
	return nil
}

func (m *Float64Int8TreeMap) minNode(node *float64Int8TreeNode) *float64Int8TreeNode {
	for node.left != nil {
		node = node.left
	}
	return node
}

func (m *Float64Int8TreeMap) maxNode(node *float64Int8TreeNode) *float64Int8TreeNode {
	for node.right != nil {
		node = node.right
	}
	return node
}

func (m *Float64Int8TreeMap) rotateLeft(x *float64Int8TreeNode) {
	y := x.right
	x.right = y.left
	if y.left != nil {
		y.left.parent = x
	}
	y.parent = x.parent
	if x.parent == nil {
		m.root = y
	} else if x == x.parent.left {
		x.parent.left = y
	} else {
		x.parent.right = y
	}
	y.left = x
	x.parent = y
}

func (m *Float64Int8TreeMap) rotateRight(x *float64Int8TreeNode) {
	y := x.left
	x.left = y.right
	if y.right != nil {
		y.right.parent = x
	}
	y.parent = x.parent
	if x.parent == nil {
		m.root = y
	} else if x == x.parent.right {
		x.parent.right = y
	} else {
		x.parent.left = y
	}
	y.right = x
	x.parent = y
}

func (m *Float64Int8TreeMap) fixAfterInsert(z *float64Int8TreeNode) {
	for z.parent != nil && z.parent.color == float64Int8TreeNodeRed {
		if z.parent == z.parent.parent.left {
			y := z.parent.parent.right
			if y != nil && y.color == float64Int8TreeNodeRed {
				z.parent.color = float64Int8TreeNodeBlack
				y.color = float64Int8TreeNodeBlack
				z.parent.parent.color = float64Int8TreeNodeRed
				z = z.parent.parent
			} else {
				if z == z.parent.right {
					z = z.parent
					m.rotateLeft(z)
				}
				z.parent.color = float64Int8TreeNodeBlack
				z.parent.parent.color = float64Int8TreeNodeRed
				m.rotateRight(z.parent.parent)
			}
		} else {
			y := z.parent.parent.left
			if y != nil && y.color == float64Int8TreeNodeRed {
				z.parent.color = float64Int8TreeNodeBlack
				y.color = float64Int8TreeNodeBlack
				z.parent.parent.color = float64Int8TreeNodeRed
				z = z.parent.parent
			} else {
				if z == z.parent.left {
					z = z.parent
					m.rotateRight(z)
				}
				z.parent.color = float64Int8TreeNodeBlack
				z.parent.parent.color = float64Int8TreeNodeRed
				m.rotateLeft(z.parent.parent)
			}
		}
	}
	m.root.color = float64Int8TreeNodeBlack
}

func (m *Float64Int8TreeMap) deleteNode(z *float64Int8TreeNode) {
	if z.left != nil && z.right != nil {
		succ := m.minNode(z.right)
		z.key = succ.key
		z.value = succ.value
		z = succ
	}
	var child *float64Int8TreeNode
	if z.left != nil {
		child = z.left
	} else {
		child = z.right
	}
	if child != nil {
		child.parent = z.parent
		if z.parent == nil {
			m.root = child
		} else if z == z.parent.left {
			z.parent.left = child
		} else {
			z.parent.right = child
		}
		if z.color == float64Int8TreeNodeBlack {
			m.fixAfterDelete(child)
		}
	} else if z.parent == nil {
		m.root = nil
	} else {
		if z.color == float64Int8TreeNodeBlack {
			m.fixAfterDelete(z)
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

func (m *Float64Int8TreeMap) fixAfterDelete(x *float64Int8TreeNode) {
	for x != m.root && x.color == float64Int8TreeNodeBlack {
		if x == x.parent.left {
			w := x.parent.right
			if w == nil {
				x = x.parent
				continue
			}
			if w.color == float64Int8TreeNodeRed {
				w.color = float64Int8TreeNodeBlack
				x.parent.color = float64Int8TreeNodeRed
				m.rotateLeft(x.parent)
				w = x.parent.right
			}
			if w == nil {
				x = x.parent
				continue
			}
			leftBlack := w.left == nil || w.left.color == float64Int8TreeNodeBlack
			rightBlack := w.right == nil || w.right.color == float64Int8TreeNodeBlack
			if leftBlack && rightBlack {
				w.color = float64Int8TreeNodeRed
				x = x.parent
			} else {
				if rightBlack {
					if w.left != nil {
						w.left.color = float64Int8TreeNodeBlack
					}
					w.color = float64Int8TreeNodeRed
					m.rotateRight(w)
					w = x.parent.right
				}
				w.color = x.parent.color
				x.parent.color = float64Int8TreeNodeBlack
				if w.right != nil {
					w.right.color = float64Int8TreeNodeBlack
				}
				m.rotateLeft(x.parent)
				x = m.root
			}
		} else {
			w := x.parent.left
			if w == nil {
				x = x.parent
				continue
			}
			if w.color == float64Int8TreeNodeRed {
				w.color = float64Int8TreeNodeBlack
				x.parent.color = float64Int8TreeNodeRed
				m.rotateRight(x.parent)
				w = x.parent.left
			}
			if w == nil {
				x = x.parent
				continue
			}
			leftBlack := w.left == nil || w.left.color == float64Int8TreeNodeBlack
			rightBlack := w.right == nil || w.right.color == float64Int8TreeNodeBlack
			if leftBlack && rightBlack {
				w.color = float64Int8TreeNodeRed
				x = x.parent
			} else {
				if leftBlack {
					if w.right != nil {
						w.right.color = float64Int8TreeNodeBlack
					}
					w.color = float64Int8TreeNodeRed
					m.rotateLeft(w)
					w = x.parent.left
				}
				w.color = x.parent.color
				x.parent.color = float64Int8TreeNodeBlack
				if w.left != nil {
					w.left.color = float64Int8TreeNodeBlack
				}
				m.rotateRight(x.parent)
				x = m.root
			}
		}
	}
	x.color = float64Int8TreeNodeBlack
}
