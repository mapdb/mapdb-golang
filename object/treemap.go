// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package object

import (
	"fmt"
	"iter"
	"strings"
)

// TreeMap is a sorted map backed by a red-black tree with a pluggable Comparator.
// Keys are maintained in the order defined by the comparator.
type TreeMap[K any, V any] struct {
	root *tmNode[K, V]
	size int
	cmp  Comparator[K]
}

type tmNode[K any, V any] struct {
	key    K
	value  V
	left   *tmNode[K, V]
	right  *tmNode[K, V]
	parent *tmNode[K, V]
	red    bool
}

// NewTreeMap creates an empty TreeMap using the given comparator.
func NewTreeMap[K any, V any](cmp Comparator[K]) *TreeMap[K, V] {
	return &TreeMap[K, V]{cmp: cmp}
}

func (m *TreeMap[K, V]) Put(key K, value V) (V, bool) {
	if m.root == nil {
		m.root = &tmNode[K, V]{key: key, value: value}
		m.size++
		var zero V
		return zero, false
	}
	n := m.root
	for {
		c := m.cmp(key, n.key)
		if c < 0 {
			if n.left == nil {
				n.left = &tmNode[K, V]{key: key, value: value, parent: n, red: true}
				m.fixAfterInsert(n.left)
				m.size++
				var zero V
				return zero, false
			}
			n = n.left
		} else if c > 0 {
			if n.right == nil {
				n.right = &tmNode[K, V]{key: key, value: value, parent: n, red: true}
				m.fixAfterInsert(n.right)
				m.size++
				var zero V
				return zero, false
			}
			n = n.right
		} else {
			old := n.value
			n.value = value
			return old, true
		}
	}
}

func (m *TreeMap[K, V]) Get(key K) (V, bool) {
	n := m.findNode(key)
	if n == nil {
		var zero V
		return zero, false
	}
	return n.value, true
}

func (m *TreeMap[K, V]) ContainsKey(key K) bool {
	return m.findNode(key) != nil
}

func (m *TreeMap[K, V]) Remove(key K) (V, bool) {
	n := m.findNode(key)
	if n == nil {
		var zero V
		return zero, false
	}
	old := n.value
	m.deleteNode(n)
	m.size--
	return old, true
}

// Len returns the number of entries. Use m.Len() == 0 to test for emptiness.
func (m *TreeMap[K, V]) Len() int { return m.size }

func (m *TreeMap[K, V]) Clear() {
	m.root = nil
	m.size = 0
}

// Min returns the smallest key and its value.
func (m *TreeMap[K, V]) Min() (K, V, bool) {
	if m.root == nil {
		var zk K
		var zv V
		return zk, zv, false
	}
	n := m.minNode(m.root)
	return n.key, n.value, true
}

// Max returns the largest key and its value.
func (m *TreeMap[K, V]) Max() (K, V, bool) {
	if m.root == nil {
		var zk K
		var zv V
		return zk, zv, false
	}
	n := m.maxNode(m.root)
	return n.key, n.value, true
}

// All returns an iterator over key-value pairs in sorted order.
func (m *TreeMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		m.inOrder(m.root, yield)
	}
}

func (m *TreeMap[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		m.inOrder(m.root, func(k K, v V) bool { return yield(k) })
	}
}

func (m *TreeMap[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		m.inOrder(m.root, func(k K, v V) bool { return yield(v) })
	}
}

func (m *TreeMap[K, V]) ForEach(f func(K, V)) {
	m.inOrder(m.root, func(k K, v V) bool { f(k, v); return true })
}

func (m *TreeMap[K, V]) Select(predicate func(K, V) bool) *TreeMap[K, V] {
	result := NewTreeMap[K, V](m.cmp)
	m.ForEach(func(k K, v V) {
		if predicate(k, v) {
			result.Put(k, v)
		}
	})
	return result
}

func (m *TreeMap[K, V]) Reject(predicate func(K, V) bool) *TreeMap[K, V] {
	result := NewTreeMap[K, V](m.cmp)
	m.ForEach(func(k K, v V) {
		if !predicate(k, v) {
			result.Put(k, v)
		}
	})
	return result
}

func (m *TreeMap[K, V]) String() string {
	var b strings.Builder
	b.WriteByte('{')
	first := true
	m.ForEach(func(k K, v V) {
		if !first {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%v: %v", k, v)
		first = false
	})
	b.WriteByte('}')
	return b.String()
}

// ── navigable surface ────────────────────────────────────────────────
//
// Mirrors Java's NavigableMap: Floor/Ceiling (inclusive bounds),
// Higher/Lower (strict bounds), Head/Tail/SubMap as lazy range
// iterators, First/LastEntry and PollFirst/PollLast, Descending views.
//
// The Head/Tail/SubMap "views" are iter.Seq2 rather than true live
// views. This matches Java's read-only use case (iteration, counting,
// search) without paying for the view-bookkeeping Java's TreeSubMap
// wrapper requires. For mutable range operations, snapshot via
// SelectInRange below.

// Floor returns the largest key <= given, with its value, or zero/false.
func (m *TreeMap[K, V]) Floor(key K) (K, V, bool) {
	var result *tmNode[K, V]
	for n := m.root; n != nil; {
		c := m.cmp(key, n.key)
		switch {
		case c == 0:
			return n.key, n.value, true
		case c > 0:
			result = n
			n = n.right
		default:
			n = n.left
		}
	}
	if result == nil {
		var zk K
		var zv V
		return zk, zv, false
	}
	return result.key, result.value, true
}

// Ceiling returns the smallest key >= given, with its value, or zero/false.
func (m *TreeMap[K, V]) Ceiling(key K) (K, V, bool) {
	var result *tmNode[K, V]
	for n := m.root; n != nil; {
		c := m.cmp(key, n.key)
		switch {
		case c == 0:
			return n.key, n.value, true
		case c < 0:
			result = n
			n = n.left
		default:
			n = n.right
		}
	}
	if result == nil {
		var zk K
		var zv V
		return zk, zv, false
	}
	return result.key, result.value, true
}

// Higher returns the smallest key strictly > given, with its value, or zero/false.
func (m *TreeMap[K, V]) Higher(key K) (K, V, bool) {
	var result *tmNode[K, V]
	for n := m.root; n != nil; {
		if m.cmp(key, n.key) < 0 {
			result = n
			n = n.left
		} else {
			n = n.right
		}
	}
	if result == nil {
		var zk K
		var zv V
		return zk, zv, false
	}
	return result.key, result.value, true
}

// Lower returns the largest key strictly < given, with its value, or zero/false.
func (m *TreeMap[K, V]) Lower(key K) (K, V, bool) {
	var result *tmNode[K, V]
	for n := m.root; n != nil; {
		if m.cmp(key, n.key) > 0 {
			result = n
			n = n.right
		} else {
			n = n.left
		}
	}
	if result == nil {
		var zk K
		var zv V
		return zk, zv, false
	}
	return result.key, result.value, true
}

// HeadMap yields entries with keys strictly less than toKey.
func (m *TreeMap[K, V]) HeadMap(toKey K) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range m.All() {
			if m.cmp(k, toKey) >= 0 {
				return
			}
			if !yield(k, v) {
				return
			}
		}
	}
}

// TailMap yields entries with keys >= fromKey.
func (m *TreeMap[K, V]) TailMap(fromKey K) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range m.All() {
			if m.cmp(k, fromKey) < 0 {
				continue
			}
			if !yield(k, v) {
				return
			}
		}
	}
}

// SubMap yields entries with keys in [fromKey, toKey).
func (m *TreeMap[K, V]) SubMap(fromKey, toKey K) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range m.All() {
			if m.cmp(k, fromKey) < 0 {
				continue
			}
			if m.cmp(k, toKey) >= 0 {
				return
			}
			if !yield(k, v) {
				return
			}
		}
	}
}

// FirstEntry is an alias of Min.
func (m *TreeMap[K, V]) FirstEntry() (K, V, bool) { return m.Min() }

// LastEntry is an alias of Max.
func (m *TreeMap[K, V]) LastEntry() (K, V, bool) { return m.Max() }

// PollFirstEntry removes and returns the smallest entry.
func (m *TreeMap[K, V]) PollFirstEntry() (K, V, bool) {
	k, v, ok := m.Min()
	if !ok {
		return k, v, false
	}
	m.Remove(k)
	return k, v, true
}

// PollLastEntry removes and returns the largest entry.
func (m *TreeMap[K, V]) PollLastEntry() (K, V, bool) {
	k, v, ok := m.Max()
	if !ok {
		return k, v, false
	}
	m.Remove(k)
	return k, v, true
}

// DescendingMap yields entries in descending key order.
func (m *TreeMap[K, V]) DescendingMap() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		m.reverseInOrder(m.root, yield)
	}
}

// DescendingKeys yields keys in descending order.
func (m *TreeMap[K, V]) DescendingKeys() iter.Seq[K] {
	return func(yield func(K) bool) {
		m.reverseInOrder(m.root, func(k K, _ V) bool { return yield(k) })
	}
}

// reverseInOrder is the mirror of inOrder — right subtree first.
func (m *TreeMap[K, V]) reverseInOrder(n *tmNode[K, V], yield func(K, V) bool) bool {
	if n == nil {
		return true
	}
	if !m.reverseInOrder(n.right, yield) {
		return false
	}
	if !yield(n.key, n.value) {
		return false
	}
	return m.reverseInOrder(n.left, yield)
}

// ── internal: lookup ──────────────────────────────────────────────────

func (m *TreeMap[K, V]) findNode(key K) *tmNode[K, V] {
	n := m.root
	for n != nil {
		c := m.cmp(key, n.key)
		if c < 0 {
			n = n.left
		} else if c > 0 {
			n = n.right
		} else {
			return n
		}
	}
	return nil
}

func (m *TreeMap[K, V]) minNode(n *tmNode[K, V]) *tmNode[K, V] {
	for n.left != nil {
		n = n.left
	}
	return n
}

func (m *TreeMap[K, V]) maxNode(n *tmNode[K, V]) *tmNode[K, V] {
	for n.right != nil {
		n = n.right
	}
	return n
}

func (m *TreeMap[K, V]) inOrder(n *tmNode[K, V], yield func(K, V) bool) bool {
	if n == nil {
		return true
	}
	if !m.inOrder(n.left, yield) {
		return false
	}
	if !yield(n.key, n.value) {
		return false
	}
	return m.inOrder(n.right, yield)
}

// ── internal: red-black tree operations ───────────────────────────────

func (m *TreeMap[K, V]) isRed(n *tmNode[K, V]) bool { return n != nil && n.red }

func (m *TreeMap[K, V]) rotateLeft(n *tmNode[K, V]) {
	r := n.right
	n.right = r.left
	if r.left != nil {
		r.left.parent = n
	}
	r.parent = n.parent
	if n.parent == nil {
		m.root = r
	} else if n == n.parent.left {
		n.parent.left = r
	} else {
		n.parent.right = r
	}
	r.left = n
	n.parent = r
}

func (m *TreeMap[K, V]) rotateRight(n *tmNode[K, V]) {
	l := n.left
	n.left = l.right
	if l.right != nil {
		l.right.parent = n
	}
	l.parent = n.parent
	if n.parent == nil {
		m.root = l
	} else if n == n.parent.right {
		n.parent.right = l
	} else {
		n.parent.left = l
	}
	l.right = n
	n.parent = l
}

func (m *TreeMap[K, V]) fixAfterInsert(n *tmNode[K, V]) {
	n.red = true
	for n != nil && n != m.root && n.parent.red {
		if n.parent == n.parent.parent.left {
			uncle := n.parent.parent.right
			if m.isRed(uncle) {
				n.parent.red = false
				uncle.red = false
				n.parent.parent.red = true
				n = n.parent.parent
			} else {
				if n == n.parent.right {
					n = n.parent
					m.rotateLeft(n)
				}
				n.parent.red = false
				n.parent.parent.red = true
				m.rotateRight(n.parent.parent)
			}
		} else {
			uncle := n.parent.parent.left
			if m.isRed(uncle) {
				n.parent.red = false
				uncle.red = false
				n.parent.parent.red = true
				n = n.parent.parent
			} else {
				if n == n.parent.left {
					n = n.parent
					m.rotateRight(n)
				}
				n.parent.red = false
				n.parent.parent.red = true
				m.rotateLeft(n.parent.parent)
			}
		}
	}
	m.root.red = false
}

func (m *TreeMap[K, V]) deleteNode(n *tmNode[K, V]) {
	if n.left != nil && n.right != nil {
		succ := m.minNode(n.right)
		n.key = succ.key
		n.value = succ.value
		n = succ
	}
	var child *tmNode[K, V]
	if n.left != nil {
		child = n.left
	} else {
		child = n.right
	}
	if child != nil {
		child.parent = n.parent
		if n.parent == nil {
			m.root = child
		} else if n == n.parent.left {
			n.parent.left = child
		} else {
			n.parent.right = child
		}
		if !n.red {
			m.fixAfterDelete(child)
		}
	} else if n.parent == nil {
		m.root = nil
	} else {
		if !n.red {
			m.fixAfterDelete(n)
		}
		if n.parent != nil {
			if n == n.parent.left {
				n.parent.left = nil
			} else {
				n.parent.right = nil
			}
		}
	}
}

func (m *TreeMap[K, V]) fixAfterDelete(n *tmNode[K, V]) {
	for n != m.root && !m.isRed(n) {
		if n == n.parent.left {
			sib := n.parent.right
			if m.isRed(sib) {
				sib.red = false
				n.parent.red = true
				m.rotateLeft(n.parent)
				sib = n.parent.right
			}
			if sib == nil {
				n = n.parent
				continue
			}
			if !m.isRed(sib.left) && !m.isRed(sib.right) {
				sib.red = true
				n = n.parent
			} else {
				if !m.isRed(sib.right) {
					if sib.left != nil {
						sib.left.red = false
					}
					sib.red = true
					m.rotateRight(sib)
					sib = n.parent.right
				}
				sib.red = n.parent.red
				n.parent.red = false
				if sib.right != nil {
					sib.right.red = false
				}
				m.rotateLeft(n.parent)
				n = m.root
			}
		} else {
			sib := n.parent.left
			if m.isRed(sib) {
				sib.red = false
				n.parent.red = true
				m.rotateRight(n.parent)
				sib = n.parent.left
			}
			if sib == nil {
				n = n.parent
				continue
			}
			if !m.isRed(sib.right) && !m.isRed(sib.left) {
				sib.red = true
				n = n.parent
			} else {
				if !m.isRed(sib.left) {
					if sib.right != nil {
						sib.right.red = false
					}
					sib.red = true
					m.rotateLeft(sib)
					sib = n.parent.left
				}
				sib.red = n.parent.red
				n.parent.red = false
				if sib.left != nil {
					sib.left.red = false
				}
				m.rotateRight(n.parent)
				n = m.root
			}
		}
	}
	if n != nil {
		n.red = false
	}
}
