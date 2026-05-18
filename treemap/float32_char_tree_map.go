
package treemap

import (
	"fmt"
	"iter"
	"strings"
)

const (
	float32CharTreeNodeRed   = false
	float32CharTreeNodeBlack = true
)

type float32CharTreeNode struct {
	key    float32
	value  uint16
	left   *float32CharTreeNode
	right  *float32CharTreeNode
	parent *float32CharTreeNode
	color  bool
}

// Float32CharTreeMap is a sorted map with float32 keys and uint16 values, backed by a red-black tree.
// Keys are maintained in ascending order.
type Float32CharTreeMap struct {
	root *float32CharTreeNode
	size int
}

// NewFloat32CharTreeMap creates a new empty sorted map.
func NewFloat32CharTreeMap() *Float32CharTreeMap {
	return &Float32CharTreeMap{}
}

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *Float32CharTreeMap) Put(key float32, value uint16) (uint16, bool) {
	if m.root == nil {
		m.root = &float32CharTreeNode{key: key, value: value, color: float32CharTreeNodeBlack}
		m.size++
		return 0, false
	}
	node := m.root
	for {
		if cmpFloat32(key, node.key) < 0 {
			if node.left == nil {
				node.left = &float32CharTreeNode{key: key, value: value, parent: node, color: float32CharTreeNodeRed}
				m.fixAfterInsert(node.left)
				m.size++
				return 0, false
			}
			node = node.left
		} else if cmpFloat32(key, node.key) > 0 {
			if node.right == nil {
				node.right = &float32CharTreeNode{key: key, value: value, parent: node, color: float32CharTreeNodeRed}
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
func (m *Float32CharTreeMap) Get(key float32) (uint16, bool) {
	node := m.findNode(key)
	if node == nil {
		return 0, false
	}
	return node.value, true
}

// GetOrDefault returns the value for the key if present, or the default value otherwise.
func (m *Float32CharTreeMap) GetOrDefault(key float32, defaultValue uint16) uint16 {
	if v, ok := m.Get(key); ok {
		return v
	}
	return defaultValue
}

// ContainsKey returns true if the map contains the given key.
func (m *Float32CharTreeMap) ContainsKey(key float32) bool {
	return m.findNode(key) != nil
}

// Remove removes the entry for the given key. Returns the previous value and true if found.
func (m *Float32CharTreeMap) Remove(key float32) (uint16, bool) {
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
func (m *Float32CharTreeMap) Size() int {
	return m.size
}

// IsEmpty returns true if the map is empty.
func (m *Float32CharTreeMap) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries.
func (m *Float32CharTreeMap) Clear() {
	m.root = nil
	m.size = 0
}

// Min returns the smallest key and its value, or zero values and false if empty.
func (m *Float32CharTreeMap) Min() (float32, uint16, bool) {
	if m.root == nil {
		return 0, 0, false
	}
	node := m.minNode(m.root)
	return node.key, node.value, true
}

// Max returns the largest key and its value, or zero values and false if empty.
func (m *Float32CharTreeMap) Max() (float32, uint16, bool) {
	if m.root == nil {
		return 0, 0, false
	}
	node := m.maxNode(m.root)
	return node.key, node.value, true
}

// Floor returns the largest key <= the given key, or zero values and false.
func (m *Float32CharTreeMap) Floor(key float32) (float32, uint16, bool) {
	var result *float32CharTreeNode
	node := m.root
	for node != nil {
		if key == node.key {
			return node.key, node.value, true
		}
		if cmpFloat32(key, node.key) > 0 {
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
func (m *Float32CharTreeMap) Ceiling(key float32) (float32, uint16, bool) {
	var result *float32CharTreeNode
	node := m.root
	for node != nil {
		if key == node.key {
			return node.key, node.value, true
		}
		if cmpFloat32(key, node.key) < 0 {
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
func (m *Float32CharTreeMap) All() iter.Seq2[float32, uint16] {
	return func(yield func(float32, uint16) bool) {
		var inorder func(node *float32CharTreeNode) bool
		inorder = func(node *float32CharTreeNode) bool {
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
func (m *Float32CharTreeMap) Keys() iter.Seq[float32] {
	return func(yield func(float32) bool) {
		for k, _ := range m.All() {
			if !yield(k) {
				return
			}
		}
	}
}

// Values returns an iter.Seq that yields all values in key order.
func (m *Float32CharTreeMap) Values() iter.Seq[uint16] {
	return func(yield func(uint16) bool) {
		for _, v := range m.All() {
			if !yield(v) {
				return
			}
		}
	}
}

// RangeKeys returns an iter.Seq2 that yields entries with keys in [fromKey, toKey).
func (m *Float32CharTreeMap) RangeKeys(fromKey, toKey float32) iter.Seq2[float32, uint16] {
	return func(yield func(float32, uint16) bool) {
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
func (m *Float32CharTreeMap) Higher(key float32) (float32, uint16, bool) {
	var result *float32CharTreeNode
	node := m.root
	for node != nil {
		if cmpFloat32(key, node.key) < 0 {
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
func (m *Float32CharTreeMap) Lower(key float32) (float32, uint16, bool) {
	var result *float32CharTreeNode
	node := m.root
	for node != nil {
		if cmpFloat32(key, node.key) > 0 {
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
func (m *Float32CharTreeMap) HeadMap(toKey float32) iter.Seq2[float32, uint16] {
	return func(yield func(float32, uint16) bool) {
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
func (m *Float32CharTreeMap) TailMap(fromKey float32) iter.Seq2[float32, uint16] {
	return func(yield func(float32, uint16) bool) {
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
func (m *Float32CharTreeMap) SubMap(fromKey, toKey float32) iter.Seq2[float32, uint16] {
	return m.RangeKeys(fromKey, toKey)
}

// FirstEntry is an alias of Min — the smallest key and its value, or zero/false.
func (m *Float32CharTreeMap) FirstEntry() (float32, uint16, bool) { return m.Min() }

// LastEntry is an alias of Max — the largest key and its value, or zero/false.
func (m *Float32CharTreeMap) LastEntry() (float32, uint16, bool) { return m.Max() }

// PollFirstEntry removes and returns the smallest entry, or zero/false if empty.
func (m *Float32CharTreeMap) PollFirstEntry() (float32, uint16, bool) {
	k, v, ok := m.Min()
	if !ok {
		return 0, 0, false
	}
	m.Remove(k)
	return k, v, true
}

// PollLastEntry removes and returns the largest entry, or zero/false if empty.
func (m *Float32CharTreeMap) PollLastEntry() (float32, uint16, bool) {
	k, v, ok := m.Max()
	if !ok {
		return 0, 0, false
	}
	m.Remove(k)
	return k, v, true
}

// DescendingMap returns an iter.Seq2 over entries in descending key order.
func (m *Float32CharTreeMap) DescendingMap() iter.Seq2[float32, uint16] {
	return func(yield func(float32, uint16) bool) {
		var reverse func(node *float32CharTreeNode) bool
		reverse = func(node *float32CharTreeNode) bool {
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
func (m *Float32CharTreeMap) DescendingKeys() iter.Seq[float32] {
	return func(yield func(float32) bool) {
		for k := range m.DescendingMap() {
			if !yield(k) {
				return
			}
		}
	}
}

// ForEach calls the function for each key-value pair in ascending order.
func (m *Float32CharTreeMap) ForEach(f func(float32, uint16)) {
	for k, v := range m.All() {
		f(k, v)
	}
}

// Select returns a new TreeMap with entries satisfying the predicate.
func (m *Float32CharTreeMap) Select(predicate func(float32, uint16) bool) *Float32CharTreeMap {
	result := NewFloat32CharTreeMap()
	for k, v := range m.All() {
		if predicate(k, v) {
			result.Put(k, v)
		}
	}
	return result
}

// Reject returns a new TreeMap with entries NOT satisfying the predicate.
func (m *Float32CharTreeMap) Reject(predicate func(float32, uint16) bool) *Float32CharTreeMap {
	result := NewFloat32CharTreeMap()
	for k, v := range m.All() {
		if !predicate(k, v) {
			result.Put(k, v)
		}
	}
	return result
}

// Detect returns the first entry satisfying the predicate (in key order), or (zero, zero, false).
func (m *Float32CharTreeMap) Detect(predicate func(float32, uint16) bool) (float32, uint16, bool) {
	for k, v := range m.All() {
		if predicate(k, v) {
			return k, v, true
		}
	}
	var zk float32
	var zv uint16
	return zk, zv, false
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *Float32CharTreeMap) AnySatisfy(predicate func(float32, uint16) bool) bool {
	for k, v := range m.All() {
		if predicate(k, v) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all entries satisfy the predicate.
func (m *Float32CharTreeMap) AllSatisfy(predicate func(float32, uint16) bool) bool {
	for k, v := range m.All() {
		if !predicate(k, v) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *Float32CharTreeMap) NoneSatisfy(predicate func(float32, uint16) bool) bool {
	for k, v := range m.All() {
		if predicate(k, v) {
			return false
		}
	}
	return true
}

// Count returns the number of entries satisfying the predicate.
func (m *Float32CharTreeMap) Count(predicate func(float32, uint16) bool) int {
	c := 0
	for k, v := range m.All() {
		if predicate(k, v) {
			c++
		}
	}
	return c
}

// String returns a string representation with entries in sorted key order.
func (m *Float32CharTreeMap) String() string {
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

func (m *Float32CharTreeMap) findNode(key float32) *float32CharTreeNode {
	node := m.root
	for node != nil {
		if cmpFloat32(key, node.key) < 0 {
			node = node.left
		} else if cmpFloat32(key, node.key) > 0 {
			node = node.right
		} else {
			return node
		}
	}
	return nil
}

func (m *Float32CharTreeMap) minNode(node *float32CharTreeNode) *float32CharTreeNode {
	for node.left != nil {
		node = node.left
	}
	return node
}

func (m *Float32CharTreeMap) maxNode(node *float32CharTreeNode) *float32CharTreeNode {
	for node.right != nil {
		node = node.right
	}
	return node
}

func (m *Float32CharTreeMap) rotateLeft(x *float32CharTreeNode) {
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

func (m *Float32CharTreeMap) rotateRight(x *float32CharTreeNode) {
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

func (m *Float32CharTreeMap) fixAfterInsert(z *float32CharTreeNode) {
	for z.parent != nil && z.parent.color == float32CharTreeNodeRed {
		if z.parent == z.parent.parent.left {
			y := z.parent.parent.right
			if y != nil && y.color == float32CharTreeNodeRed {
				z.parent.color = float32CharTreeNodeBlack
				y.color = float32CharTreeNodeBlack
				z.parent.parent.color = float32CharTreeNodeRed
				z = z.parent.parent
			} else {
				if z == z.parent.right {
					z = z.parent
					m.rotateLeft(z)
				}
				z.parent.color = float32CharTreeNodeBlack
				z.parent.parent.color = float32CharTreeNodeRed
				m.rotateRight(z.parent.parent)
			}
		} else {
			y := z.parent.parent.left
			if y != nil && y.color == float32CharTreeNodeRed {
				z.parent.color = float32CharTreeNodeBlack
				y.color = float32CharTreeNodeBlack
				z.parent.parent.color = float32CharTreeNodeRed
				z = z.parent.parent
			} else {
				if z == z.parent.left {
					z = z.parent
					m.rotateRight(z)
				}
				z.parent.color = float32CharTreeNodeBlack
				z.parent.parent.color = float32CharTreeNodeRed
				m.rotateLeft(z.parent.parent)
			}
		}
	}
	m.root.color = float32CharTreeNodeBlack
}

func (m *Float32CharTreeMap) deleteNode(z *float32CharTreeNode) {
	if z.left != nil && z.right != nil {
		succ := m.minNode(z.right)
		z.key = succ.key
		z.value = succ.value
		z = succ
	}
	var child *float32CharTreeNode
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
		if z.color == float32CharTreeNodeBlack {
			m.fixAfterDelete(child)
		}
	} else if z.parent == nil {
		m.root = nil
	} else {
		if z.color == float32CharTreeNodeBlack {
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

func (m *Float32CharTreeMap) fixAfterDelete(x *float32CharTreeNode) {
	for x != m.root && x.color == float32CharTreeNodeBlack {
		if x == x.parent.left {
			w := x.parent.right
			if w == nil {
				x = x.parent
				continue
			}
			if w.color == float32CharTreeNodeRed {
				w.color = float32CharTreeNodeBlack
				x.parent.color = float32CharTreeNodeRed
				m.rotateLeft(x.parent)
				w = x.parent.right
			}
			if w == nil {
				x = x.parent
				continue
			}
			leftBlack := w.left == nil || w.left.color == float32CharTreeNodeBlack
			rightBlack := w.right == nil || w.right.color == float32CharTreeNodeBlack
			if leftBlack && rightBlack {
				w.color = float32CharTreeNodeRed
				x = x.parent
			} else {
				if rightBlack {
					if w.left != nil {
						w.left.color = float32CharTreeNodeBlack
					}
					w.color = float32CharTreeNodeRed
					m.rotateRight(w)
					w = x.parent.right
				}
				w.color = x.parent.color
				x.parent.color = float32CharTreeNodeBlack
				if w.right != nil {
					w.right.color = float32CharTreeNodeBlack
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
			if w.color == float32CharTreeNodeRed {
				w.color = float32CharTreeNodeBlack
				x.parent.color = float32CharTreeNodeRed
				m.rotateRight(x.parent)
				w = x.parent.left
			}
			if w == nil {
				x = x.parent
				continue
			}
			leftBlack := w.left == nil || w.left.color == float32CharTreeNodeBlack
			rightBlack := w.right == nil || w.right.color == float32CharTreeNodeBlack
			if leftBlack && rightBlack {
				w.color = float32CharTreeNodeRed
				x = x.parent
			} else {
				if leftBlack {
					if w.right != nil {
						w.right.color = float32CharTreeNodeBlack
					}
					w.color = float32CharTreeNodeRed
					m.rotateLeft(w)
					w = x.parent.left
				}
				w.color = x.parent.color
				x.parent.color = float32CharTreeNodeBlack
				if w.left != nil {
					w.left.color = float32CharTreeNodeBlack
				}
				m.rotateRight(x.parent)
				x = m.root
			}
		}
	}
	x.color = float32CharTreeNodeBlack
}
