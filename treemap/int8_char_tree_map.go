
package treemap

import (
	"fmt"
	"iter"
	"strings"
)

const (
	int8CharTreeNodeRed   = false
	int8CharTreeNodeBlack = true
)

type int8CharTreeNode struct {
	key    int8
	value  uint16
	left   *int8CharTreeNode
	right  *int8CharTreeNode
	parent *int8CharTreeNode
	color  bool
}

// Int8CharTreeMap is a sorted map with int8 keys and uint16 values, backed by a red-black tree.
// Keys are maintained in ascending order.
type Int8CharTreeMap struct {
	root *int8CharTreeNode
	size int
}

// NewInt8CharTreeMap creates a new empty sorted map.
func NewInt8CharTreeMap() *Int8CharTreeMap {
	return &Int8CharTreeMap{}
}

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *Int8CharTreeMap) Put(key int8, value uint16) (uint16, bool) {
	if m.root == nil {
		m.root = &int8CharTreeNode{key: key, value: value, color: int8CharTreeNodeBlack}
		m.size++
		return 0, false
	}
	node := m.root
	for {
		if key < node.key {
			if node.left == nil {
				node.left = &int8CharTreeNode{key: key, value: value, parent: node, color: int8CharTreeNodeRed}
				m.fixAfterInsert(node.left)
				m.size++
				return 0, false
			}
			node = node.left
		} else if key > node.key {
			if node.right == nil {
				node.right = &int8CharTreeNode{key: key, value: value, parent: node, color: int8CharTreeNodeRed}
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
func (m *Int8CharTreeMap) Get(key int8) (uint16, bool) {
	node := m.findNode(key)
	if node == nil {
		return 0, false
	}
	return node.value, true
}

// GetOrDefault returns the value for the key if present, or the default value otherwise.
func (m *Int8CharTreeMap) GetOrDefault(key int8, defaultValue uint16) uint16 {
	if v, ok := m.Get(key); ok {
		return v
	}
	return defaultValue
}

// ContainsKey returns true if the map contains the given key.
func (m *Int8CharTreeMap) ContainsKey(key int8) bool {
	return m.findNode(key) != nil
}

// Remove removes the entry for the given key. Returns the previous value and true if found.
func (m *Int8CharTreeMap) Remove(key int8) (uint16, bool) {
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
func (m *Int8CharTreeMap) Size() int {
	return m.size
}

// IsEmpty returns true if the map is empty.
func (m *Int8CharTreeMap) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries.
func (m *Int8CharTreeMap) Clear() {
	m.root = nil
	m.size = 0
}

// Min returns the smallest key and its value, or zero values and false if empty.
func (m *Int8CharTreeMap) Min() (int8, uint16, bool) {
	if m.root == nil {
		return 0, 0, false
	}
	node := m.minNode(m.root)
	return node.key, node.value, true
}

// Max returns the largest key and its value, or zero values and false if empty.
func (m *Int8CharTreeMap) Max() (int8, uint16, bool) {
	if m.root == nil {
		return 0, 0, false
	}
	node := m.maxNode(m.root)
	return node.key, node.value, true
}

// Floor returns the largest key <= the given key, or zero values and false.
func (m *Int8CharTreeMap) Floor(key int8) (int8, uint16, bool) {
	var result *int8CharTreeNode
	node := m.root
	for node != nil {
		if key == node.key {
			return node.key, node.value, true
		}
		if key > node.key {
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
func (m *Int8CharTreeMap) Ceiling(key int8) (int8, uint16, bool) {
	var result *int8CharTreeNode
	node := m.root
	for node != nil {
		if key == node.key {
			return node.key, node.value, true
		}
		if key < node.key {
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
func (m *Int8CharTreeMap) All() iter.Seq2[int8, uint16] {
	return func(yield func(int8, uint16) bool) {
		var inorder func(node *int8CharTreeNode) bool
		inorder = func(node *int8CharTreeNode) bool {
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
func (m *Int8CharTreeMap) Keys() iter.Seq[int8] {
	return func(yield func(int8) bool) {
		for k, _ := range m.All() {
			if !yield(k) {
				return
			}
		}
	}
}

// Values returns an iter.Seq that yields all values in key order.
func (m *Int8CharTreeMap) Values() iter.Seq[uint16] {
	return func(yield func(uint16) bool) {
		for _, v := range m.All() {
			if !yield(v) {
				return
			}
		}
	}
}

// RangeKeys returns an iter.Seq2 that yields entries with keys in [fromKey, toKey).
func (m *Int8CharTreeMap) RangeKeys(fromKey, toKey int8) iter.Seq2[int8, uint16] {
	return func(yield func(int8, uint16) bool) {
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
func (m *Int8CharTreeMap) Higher(key int8) (int8, uint16, bool) {
	var result *int8CharTreeNode
	node := m.root
	for node != nil {
		if key < node.key {
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
func (m *Int8CharTreeMap) Lower(key int8) (int8, uint16, bool) {
	var result *int8CharTreeNode
	node := m.root
	for node != nil {
		if key > node.key {
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
func (m *Int8CharTreeMap) HeadMap(toKey int8) iter.Seq2[int8, uint16] {
	return func(yield func(int8, uint16) bool) {
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
func (m *Int8CharTreeMap) TailMap(fromKey int8) iter.Seq2[int8, uint16] {
	return func(yield func(int8, uint16) bool) {
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
func (m *Int8CharTreeMap) SubMap(fromKey, toKey int8) iter.Seq2[int8, uint16] {
	return m.RangeKeys(fromKey, toKey)
}

// FirstEntry is an alias of Min — the smallest key and its value, or zero/false.
func (m *Int8CharTreeMap) FirstEntry() (int8, uint16, bool) { return m.Min() }

// LastEntry is an alias of Max — the largest key and its value, or zero/false.
func (m *Int8CharTreeMap) LastEntry() (int8, uint16, bool) { return m.Max() }

// PollFirstEntry removes and returns the smallest entry, or zero/false if empty.
func (m *Int8CharTreeMap) PollFirstEntry() (int8, uint16, bool) {
	k, v, ok := m.Min()
	if !ok {
		return 0, 0, false
	}
	m.Remove(k)
	return k, v, true
}

// PollLastEntry removes and returns the largest entry, or zero/false if empty.
func (m *Int8CharTreeMap) PollLastEntry() (int8, uint16, bool) {
	k, v, ok := m.Max()
	if !ok {
		return 0, 0, false
	}
	m.Remove(k)
	return k, v, true
}

// DescendingMap returns an iter.Seq2 over entries in descending key order.
func (m *Int8CharTreeMap) DescendingMap() iter.Seq2[int8, uint16] {
	return func(yield func(int8, uint16) bool) {
		var reverse func(node *int8CharTreeNode) bool
		reverse = func(node *int8CharTreeNode) bool {
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
func (m *Int8CharTreeMap) DescendingKeys() iter.Seq[int8] {
	return func(yield func(int8) bool) {
		for k := range m.DescendingMap() {
			if !yield(k) {
				return
			}
		}
	}
}

// ForEach calls the function for each key-value pair in ascending order.
func (m *Int8CharTreeMap) ForEach(f func(int8, uint16)) {
	for k, v := range m.All() {
		f(k, v)
	}
}

// Select returns a new TreeMap with entries satisfying the predicate.
func (m *Int8CharTreeMap) Select(predicate func(int8, uint16) bool) *Int8CharTreeMap {
	result := NewInt8CharTreeMap()
	for k, v := range m.All() {
		if predicate(k, v) {
			result.Put(k, v)
		}
	}
	return result
}

// Reject returns a new TreeMap with entries NOT satisfying the predicate.
func (m *Int8CharTreeMap) Reject(predicate func(int8, uint16) bool) *Int8CharTreeMap {
	result := NewInt8CharTreeMap()
	for k, v := range m.All() {
		if !predicate(k, v) {
			result.Put(k, v)
		}
	}
	return result
}

// Detect returns the first entry satisfying the predicate (in key order), or (zero, zero, false).
func (m *Int8CharTreeMap) Detect(predicate func(int8, uint16) bool) (int8, uint16, bool) {
	for k, v := range m.All() {
		if predicate(k, v) {
			return k, v, true
		}
	}
	var zk int8
	var zv uint16
	return zk, zv, false
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *Int8CharTreeMap) AnySatisfy(predicate func(int8, uint16) bool) bool {
	for k, v := range m.All() {
		if predicate(k, v) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all entries satisfy the predicate.
func (m *Int8CharTreeMap) AllSatisfy(predicate func(int8, uint16) bool) bool {
	for k, v := range m.All() {
		if !predicate(k, v) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *Int8CharTreeMap) NoneSatisfy(predicate func(int8, uint16) bool) bool {
	for k, v := range m.All() {
		if predicate(k, v) {
			return false
		}
	}
	return true
}

// Count returns the number of entries satisfying the predicate.
func (m *Int8CharTreeMap) Count(predicate func(int8, uint16) bool) int {
	c := 0
	for k, v := range m.All() {
		if predicate(k, v) {
			c++
		}
	}
	return c
}

// String returns a string representation with entries in sorted key order.
func (m *Int8CharTreeMap) String() string {
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

func (m *Int8CharTreeMap) findNode(key int8) *int8CharTreeNode {
	node := m.root
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

func (m *Int8CharTreeMap) minNode(node *int8CharTreeNode) *int8CharTreeNode {
	for node.left != nil {
		node = node.left
	}
	return node
}

func (m *Int8CharTreeMap) maxNode(node *int8CharTreeNode) *int8CharTreeNode {
	for node.right != nil {
		node = node.right
	}
	return node
}

func (m *Int8CharTreeMap) rotateLeft(x *int8CharTreeNode) {
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

func (m *Int8CharTreeMap) rotateRight(x *int8CharTreeNode) {
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

func (m *Int8CharTreeMap) fixAfterInsert(z *int8CharTreeNode) {
	for z.parent != nil && z.parent.color == int8CharTreeNodeRed {
		if z.parent == z.parent.parent.left {
			y := z.parent.parent.right
			if y != nil && y.color == int8CharTreeNodeRed {
				z.parent.color = int8CharTreeNodeBlack
				y.color = int8CharTreeNodeBlack
				z.parent.parent.color = int8CharTreeNodeRed
				z = z.parent.parent
			} else {
				if z == z.parent.right {
					z = z.parent
					m.rotateLeft(z)
				}
				z.parent.color = int8CharTreeNodeBlack
				z.parent.parent.color = int8CharTreeNodeRed
				m.rotateRight(z.parent.parent)
			}
		} else {
			y := z.parent.parent.left
			if y != nil && y.color == int8CharTreeNodeRed {
				z.parent.color = int8CharTreeNodeBlack
				y.color = int8CharTreeNodeBlack
				z.parent.parent.color = int8CharTreeNodeRed
				z = z.parent.parent
			} else {
				if z == z.parent.left {
					z = z.parent
					m.rotateRight(z)
				}
				z.parent.color = int8CharTreeNodeBlack
				z.parent.parent.color = int8CharTreeNodeRed
				m.rotateLeft(z.parent.parent)
			}
		}
	}
	m.root.color = int8CharTreeNodeBlack
}

func (m *Int8CharTreeMap) deleteNode(z *int8CharTreeNode) {
	if z.left != nil && z.right != nil {
		succ := m.minNode(z.right)
		z.key = succ.key
		z.value = succ.value
		z = succ
	}
	var child *int8CharTreeNode
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
		if z.color == int8CharTreeNodeBlack {
			m.fixAfterDelete(child)
		}
	} else if z.parent == nil {
		m.root = nil
	} else {
		if z.color == int8CharTreeNodeBlack {
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

func (m *Int8CharTreeMap) fixAfterDelete(x *int8CharTreeNode) {
	for x != m.root && x.color == int8CharTreeNodeBlack {
		if x == x.parent.left {
			w := x.parent.right
			if w == nil {
				x = x.parent
				continue
			}
			if w.color == int8CharTreeNodeRed {
				w.color = int8CharTreeNodeBlack
				x.parent.color = int8CharTreeNodeRed
				m.rotateLeft(x.parent)
				w = x.parent.right
			}
			if w == nil {
				x = x.parent
				continue
			}
			leftBlack := w.left == nil || w.left.color == int8CharTreeNodeBlack
			rightBlack := w.right == nil || w.right.color == int8CharTreeNodeBlack
			if leftBlack && rightBlack {
				w.color = int8CharTreeNodeRed
				x = x.parent
			} else {
				if rightBlack {
					if w.left != nil {
						w.left.color = int8CharTreeNodeBlack
					}
					w.color = int8CharTreeNodeRed
					m.rotateRight(w)
					w = x.parent.right
				}
				w.color = x.parent.color
				x.parent.color = int8CharTreeNodeBlack
				if w.right != nil {
					w.right.color = int8CharTreeNodeBlack
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
			if w.color == int8CharTreeNodeRed {
				w.color = int8CharTreeNodeBlack
				x.parent.color = int8CharTreeNodeRed
				m.rotateRight(x.parent)
				w = x.parent.left
			}
			if w == nil {
				x = x.parent
				continue
			}
			leftBlack := w.left == nil || w.left.color == int8CharTreeNodeBlack
			rightBlack := w.right == nil || w.right.color == int8CharTreeNodeBlack
			if leftBlack && rightBlack {
				w.color = int8CharTreeNodeRed
				x = x.parent
			} else {
				if leftBlack {
					if w.right != nil {
						w.right.color = int8CharTreeNodeBlack
					}
					w.color = int8CharTreeNodeRed
					m.rotateLeft(w)
					w = x.parent.left
				}
				w.color = x.parent.color
				x.parent.color = int8CharTreeNodeBlack
				if w.left != nil {
					w.left.color = int8CharTreeNodeBlack
				}
				m.rotateRight(x.parent)
				x = m.root
			}
		}
	}
	x.color = int8CharTreeNodeBlack
}
