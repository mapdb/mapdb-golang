
package treemap

import (
	"fmt"
	"iter"
	"strings"
)

const (
	int32Int16TreeNodeRed   = false
	int32Int16TreeNodeBlack = true
)

type int32Int16TreeNode struct {
	key    int32
	value  int16
	left   *int32Int16TreeNode
	right  *int32Int16TreeNode
	parent *int32Int16TreeNode
	color  bool
}

// Int32Int16TreeMap is a sorted map with int32 keys and int16 values, backed by a red-black tree.
// Keys are maintained in ascending order.
type Int32Int16TreeMap struct {
	root *int32Int16TreeNode
	size int
}

// NewInt32Int16TreeMap creates a new empty sorted map.
func NewInt32Int16TreeMap() *Int32Int16TreeMap {
	return &Int32Int16TreeMap{}
}

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *Int32Int16TreeMap) Put(key int32, value int16) (int16, bool) {
	if m.root == nil {
		m.root = &int32Int16TreeNode{key: key, value: value, color: int32Int16TreeNodeBlack}
		m.size++
		return 0, false
	}
	node := m.root
	for {
		if key < node.key {
			if node.left == nil {
				node.left = &int32Int16TreeNode{key: key, value: value, parent: node, color: int32Int16TreeNodeRed}
				m.fixAfterInsert(node.left)
				m.size++
				return 0, false
			}
			node = node.left
		} else if key > node.key {
			if node.right == nil {
				node.right = &int32Int16TreeNode{key: key, value: value, parent: node, color: int32Int16TreeNodeRed}
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
func (m *Int32Int16TreeMap) Get(key int32) (int16, bool) {
	node := m.findNode(key)
	if node == nil {
		return 0, false
	}
	return node.value, true
}

// GetOrDefault returns the value for the key if present, or the default value otherwise.
func (m *Int32Int16TreeMap) GetOrDefault(key int32, defaultValue int16) int16 {
	if v, ok := m.Get(key); ok {
		return v
	}
	return defaultValue
}

// ContainsKey returns true if the map contains the given key.
func (m *Int32Int16TreeMap) ContainsKey(key int32) bool {
	return m.findNode(key) != nil
}

// Remove removes the entry for the given key. Returns the previous value and true if found.
func (m *Int32Int16TreeMap) Remove(key int32) (int16, bool) {
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
func (m *Int32Int16TreeMap) Size() int {
	return m.size
}

// IsEmpty returns true if the map is empty.
func (m *Int32Int16TreeMap) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries.
func (m *Int32Int16TreeMap) Clear() {
	m.root = nil
	m.size = 0
}

// Min returns the smallest key and its value, or zero values and false if empty.
func (m *Int32Int16TreeMap) Min() (int32, int16, bool) {
	if m.root == nil {
		return 0, 0, false
	}
	node := m.minNode(m.root)
	return node.key, node.value, true
}

// Max returns the largest key and its value, or zero values and false if empty.
func (m *Int32Int16TreeMap) Max() (int32, int16, bool) {
	if m.root == nil {
		return 0, 0, false
	}
	node := m.maxNode(m.root)
	return node.key, node.value, true
}

// Floor returns the largest key <= the given key, or zero values and false.
func (m *Int32Int16TreeMap) Floor(key int32) (int32, int16, bool) {
	var result *int32Int16TreeNode
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
func (m *Int32Int16TreeMap) Ceiling(key int32) (int32, int16, bool) {
	var result *int32Int16TreeNode
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
func (m *Int32Int16TreeMap) All() iter.Seq2[int32, int16] {
	return func(yield func(int32, int16) bool) {
		var inorder func(node *int32Int16TreeNode) bool
		inorder = func(node *int32Int16TreeNode) bool {
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
func (m *Int32Int16TreeMap) Keys() iter.Seq[int32] {
	return func(yield func(int32) bool) {
		for k, _ := range m.All() {
			if !yield(k) {
				return
			}
		}
	}
}

// Values returns an iter.Seq that yields all values in key order.
func (m *Int32Int16TreeMap) Values() iter.Seq[int16] {
	return func(yield func(int16) bool) {
		for _, v := range m.All() {
			if !yield(v) {
				return
			}
		}
	}
}

// RangeKeys returns an iter.Seq2 that yields entries with keys in [fromKey, toKey).
func (m *Int32Int16TreeMap) RangeKeys(fromKey, toKey int32) iter.Seq2[int32, int16] {
	return func(yield func(int32, int16) bool) {
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
func (m *Int32Int16TreeMap) Higher(key int32) (int32, int16, bool) {
	var result *int32Int16TreeNode
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
func (m *Int32Int16TreeMap) Lower(key int32) (int32, int16, bool) {
	var result *int32Int16TreeNode
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
func (m *Int32Int16TreeMap) HeadMap(toKey int32) iter.Seq2[int32, int16] {
	return func(yield func(int32, int16) bool) {
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
func (m *Int32Int16TreeMap) TailMap(fromKey int32) iter.Seq2[int32, int16] {
	return func(yield func(int32, int16) bool) {
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
func (m *Int32Int16TreeMap) SubMap(fromKey, toKey int32) iter.Seq2[int32, int16] {
	return m.RangeKeys(fromKey, toKey)
}

// FirstEntry is an alias of Min — the smallest key and its value, or zero/false.
func (m *Int32Int16TreeMap) FirstEntry() (int32, int16, bool) { return m.Min() }

// LastEntry is an alias of Max — the largest key and its value, or zero/false.
func (m *Int32Int16TreeMap) LastEntry() (int32, int16, bool) { return m.Max() }

// PollFirstEntry removes and returns the smallest entry, or zero/false if empty.
func (m *Int32Int16TreeMap) PollFirstEntry() (int32, int16, bool) {
	k, v, ok := m.Min()
	if !ok {
		return 0, 0, false
	}
	m.Remove(k)
	return k, v, true
}

// PollLastEntry removes and returns the largest entry, or zero/false if empty.
func (m *Int32Int16TreeMap) PollLastEntry() (int32, int16, bool) {
	k, v, ok := m.Max()
	if !ok {
		return 0, 0, false
	}
	m.Remove(k)
	return k, v, true
}

// DescendingMap returns an iter.Seq2 over entries in descending key order.
func (m *Int32Int16TreeMap) DescendingMap() iter.Seq2[int32, int16] {
	return func(yield func(int32, int16) bool) {
		var reverse func(node *int32Int16TreeNode) bool
		reverse = func(node *int32Int16TreeNode) bool {
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
func (m *Int32Int16TreeMap) DescendingKeys() iter.Seq[int32] {
	return func(yield func(int32) bool) {
		for k := range m.DescendingMap() {
			if !yield(k) {
				return
			}
		}
	}
}

// ForEach calls the function for each key-value pair in ascending order.
func (m *Int32Int16TreeMap) ForEach(f func(int32, int16)) {
	for k, v := range m.All() {
		f(k, v)
	}
}

// Select returns a new TreeMap with entries satisfying the predicate.
func (m *Int32Int16TreeMap) Select(predicate func(int32, int16) bool) *Int32Int16TreeMap {
	result := NewInt32Int16TreeMap()
	for k, v := range m.All() {
		if predicate(k, v) {
			result.Put(k, v)
		}
	}
	return result
}

// Reject returns a new TreeMap with entries NOT satisfying the predicate.
func (m *Int32Int16TreeMap) Reject(predicate func(int32, int16) bool) *Int32Int16TreeMap {
	result := NewInt32Int16TreeMap()
	for k, v := range m.All() {
		if !predicate(k, v) {
			result.Put(k, v)
		}
	}
	return result
}

// Detect returns the first entry satisfying the predicate (in key order), or (zero, zero, false).
func (m *Int32Int16TreeMap) Detect(predicate func(int32, int16) bool) (int32, int16, bool) {
	for k, v := range m.All() {
		if predicate(k, v) {
			return k, v, true
		}
	}
	var zk int32
	var zv int16
	return zk, zv, false
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *Int32Int16TreeMap) AnySatisfy(predicate func(int32, int16) bool) bool {
	for k, v := range m.All() {
		if predicate(k, v) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all entries satisfy the predicate.
func (m *Int32Int16TreeMap) AllSatisfy(predicate func(int32, int16) bool) bool {
	for k, v := range m.All() {
		if !predicate(k, v) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *Int32Int16TreeMap) NoneSatisfy(predicate func(int32, int16) bool) bool {
	for k, v := range m.All() {
		if predicate(k, v) {
			return false
		}
	}
	return true
}

// Count returns the number of entries satisfying the predicate.
func (m *Int32Int16TreeMap) Count(predicate func(int32, int16) bool) int {
	c := 0
	for k, v := range m.All() {
		if predicate(k, v) {
			c++
		}
	}
	return c
}

// String returns a string representation with entries in sorted key order.
func (m *Int32Int16TreeMap) String() string {
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

func (m *Int32Int16TreeMap) findNode(key int32) *int32Int16TreeNode {
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

func (m *Int32Int16TreeMap) minNode(node *int32Int16TreeNode) *int32Int16TreeNode {
	for node.left != nil {
		node = node.left
	}
	return node
}

func (m *Int32Int16TreeMap) maxNode(node *int32Int16TreeNode) *int32Int16TreeNode {
	for node.right != nil {
		node = node.right
	}
	return node
}

func (m *Int32Int16TreeMap) rotateLeft(x *int32Int16TreeNode) {
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

func (m *Int32Int16TreeMap) rotateRight(x *int32Int16TreeNode) {
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

func (m *Int32Int16TreeMap) fixAfterInsert(z *int32Int16TreeNode) {
	for z.parent != nil && z.parent.color == int32Int16TreeNodeRed {
		if z.parent == z.parent.parent.left {
			y := z.parent.parent.right
			if y != nil && y.color == int32Int16TreeNodeRed {
				z.parent.color = int32Int16TreeNodeBlack
				y.color = int32Int16TreeNodeBlack
				z.parent.parent.color = int32Int16TreeNodeRed
				z = z.parent.parent
			} else {
				if z == z.parent.right {
					z = z.parent
					m.rotateLeft(z)
				}
				z.parent.color = int32Int16TreeNodeBlack
				z.parent.parent.color = int32Int16TreeNodeRed
				m.rotateRight(z.parent.parent)
			}
		} else {
			y := z.parent.parent.left
			if y != nil && y.color == int32Int16TreeNodeRed {
				z.parent.color = int32Int16TreeNodeBlack
				y.color = int32Int16TreeNodeBlack
				z.parent.parent.color = int32Int16TreeNodeRed
				z = z.parent.parent
			} else {
				if z == z.parent.left {
					z = z.parent
					m.rotateRight(z)
				}
				z.parent.color = int32Int16TreeNodeBlack
				z.parent.parent.color = int32Int16TreeNodeRed
				m.rotateLeft(z.parent.parent)
			}
		}
	}
	m.root.color = int32Int16TreeNodeBlack
}

func (m *Int32Int16TreeMap) deleteNode(z *int32Int16TreeNode) {
	if z.left != nil && z.right != nil {
		succ := m.minNode(z.right)
		z.key = succ.key
		z.value = succ.value
		z = succ
	}
	var child *int32Int16TreeNode
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
		if z.color == int32Int16TreeNodeBlack {
			m.fixAfterDelete(child)
		}
	} else if z.parent == nil {
		m.root = nil
	} else {
		if z.color == int32Int16TreeNodeBlack {
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

func (m *Int32Int16TreeMap) fixAfterDelete(x *int32Int16TreeNode) {
	for x != m.root && x.color == int32Int16TreeNodeBlack {
		if x == x.parent.left {
			w := x.parent.right
			if w == nil {
				x = x.parent
				continue
			}
			if w.color == int32Int16TreeNodeRed {
				w.color = int32Int16TreeNodeBlack
				x.parent.color = int32Int16TreeNodeRed
				m.rotateLeft(x.parent)
				w = x.parent.right
			}
			if w == nil {
				x = x.parent
				continue
			}
			leftBlack := w.left == nil || w.left.color == int32Int16TreeNodeBlack
			rightBlack := w.right == nil || w.right.color == int32Int16TreeNodeBlack
			if leftBlack && rightBlack {
				w.color = int32Int16TreeNodeRed
				x = x.parent
			} else {
				if rightBlack {
					if w.left != nil {
						w.left.color = int32Int16TreeNodeBlack
					}
					w.color = int32Int16TreeNodeRed
					m.rotateRight(w)
					w = x.parent.right
				}
				w.color = x.parent.color
				x.parent.color = int32Int16TreeNodeBlack
				if w.right != nil {
					w.right.color = int32Int16TreeNodeBlack
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
			if w.color == int32Int16TreeNodeRed {
				w.color = int32Int16TreeNodeBlack
				x.parent.color = int32Int16TreeNodeRed
				m.rotateRight(x.parent)
				w = x.parent.left
			}
			if w == nil {
				x = x.parent
				continue
			}
			leftBlack := w.left == nil || w.left.color == int32Int16TreeNodeBlack
			rightBlack := w.right == nil || w.right.color == int32Int16TreeNodeBlack
			if leftBlack && rightBlack {
				w.color = int32Int16TreeNodeRed
				x = x.parent
			} else {
				if leftBlack {
					if w.right != nil {
						w.right.color = int32Int16TreeNodeBlack
					}
					w.color = int32Int16TreeNodeRed
					m.rotateLeft(w)
					w = x.parent.left
				}
				w.color = x.parent.color
				x.parent.color = int32Int16TreeNodeBlack
				if w.left != nil {
					w.left.color = int32Int16TreeNodeBlack
				}
				m.rotateRight(x.parent)
				x = m.root
			}
		}
	}
	x.color = int32Int16TreeNodeBlack
}
