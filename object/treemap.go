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
	// size is the number of nodes in the subtree rooted here (this node plus
	// both children's subtrees). Maintained in O(1) on every structural change
	// — insert, remove, and all rotations — so order-statistic Rank/Select run
	// in O(log n). Invariant after any operation: size == 1 + size(left) + size(right).
	size int
}

// tmSize returns the subtree size of a node link (0 if nil).
func tmSize[K any, V any](n *tmNode[K, V]) int {
	if n == nil {
		return 0
	}
	return n.size
}

// tmFixSize recomputes a node's cached subtree size from its children.
func tmFixSize[K any, V any](n *tmNode[K, V]) {
	n.size = 1 + tmSize(n.left) + tmSize(n.right)
}

// NewTreeMap creates an empty TreeMap using the given comparator.
func NewTreeMap[K any, V any](cmp Comparator[K]) *TreeMap[K, V] {
	return &TreeMap[K, V]{cmp: cmp}
}

func (m *TreeMap[K, V]) Put(key K, value V) (V, bool) {
	if m.root == nil {
		m.root = &tmNode[K, V]{key: key, value: value, size: 1}
		m.size++
		var zero V
		return zero, false
	}
	n := m.root
	for {
		c := m.cmp(key, n.key)
		if c < 0 {
			if n.left == nil {
				n.left = &tmNode[K, V]{key: key, value: value, parent: n, red: true, size: 1}
				m.incSizeToRoot(n)
				m.fixAfterInsert(n.left)
				m.size++
				var zero V
				return zero, false
			}
			n = n.left
		} else if c > 0 {
			if n.right == nil {
				n.right = &tmNode[K, V]{key: key, value: value, parent: n, red: true, size: 1}
				m.incSizeToRoot(n)
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

// Rank returns the number of keys strictly less than key under the map's
// comparator — the 0-based lower-bound index the key occupies (if present) or
// would occupy (if absent). Defined for present and absent keys alike; the
// result is in 0..=Len() (Len() for any key greater than the maximum). Pure
// query; never mutates.
func (m *TreeMap[K, V]) Rank(key K) int {
	rank := 0
	n := m.root
	for n != nil {
		c := m.cmp(key, n.key)
		if c < 0 {
			n = n.left
		} else if c > 0 {
			rank += 1 + tmSize(n.left)
			n = n.right
		} else {
			return rank + tmSize(n.left)
		}
	}
	return rank
}

// SelectKey returns the i-th smallest key (0-based) and true, or the zero value
// and false if i >= Len() or i < 0. Out-of-range indices (including on an empty
// map and negative i) return absence and do not trap. Round-trips with Rank.
func (m *TreeMap[K, V]) SelectKey(i int) (K, bool) {
	n := m.selectNode(i)
	if n == nil {
		var zero K
		return zero, false
	}
	return n.key, true
}

// SelectEntry returns the i-th smallest entry (0-based) as (key, value, true),
// or zero values and false if i >= Len() or i < 0. Same index domain as
// SelectKey; no trap on out-of-range or negative i.
func (m *TreeMap[K, V]) SelectEntry(i int) (K, V, bool) {
	n := m.selectNode(i)
	if n == nil {
		var zk K
		var zv V
		return zk, zv, false
	}
	return n.key, n.value, true
}

// selectNode walks to the node at 0-based sorted index i, or nil if out of
// range (i < 0 or i >= Len()). O(log n) via the subtree-size augmentation.
func (m *TreeMap[K, V]) selectNode(i int) *tmNode[K, V] {
	if i < 0 {
		return nil
	}
	n := m.root
	for n != nil {
		left := tmSize(n.left)
		if i < left {
			n = n.left
		} else if i == left {
			return n
		} else {
			i -= left + 1
			n = n.right
		}
	}
	return nil
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

// SubMapCopy returns a new independent TreeMap of the entries whose key is in
// [fromKey, toKey) under this map's comparator. It is a materialized snapshot,
// not a live view: mutating it never affects the original and vice versa. The
// snapshot PRESERVES the source map's comparator, so a reverse/custom/float
// total-order map keeps its ordering semantics in the slice (it never resets to
// natural key order).
func (m *TreeMap[K, V]) SubMapCopy(fromKey, toKey K) *TreeMap[K, V] {
	out := NewTreeMap[K, V](m.cmp)
	for k, v := range m.All() {
		if m.cmp(k, fromKey) < 0 {
			continue
		}
		if m.cmp(k, toKey) >= 0 {
			break
		}
		out.Put(k, v)
	}
	return out
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
	// r took n's former position so it inherits n's old subtree size; recompute
	// bottom-up: the demoted n first (now r's left child), then the promoted r.
	tmFixSize(n)
	tmFixSize(r)
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
	// Symmetric to rotateLeft: recompute the demoted n, then the promoted l.
	tmFixSize(n)
	tmFixSize(l)
}

// incSizeToRoot walks from n up to the root, bumping each ancestor's cached
// subtree size by one after a new leaf was linked below n.
func (m *TreeMap[K, V]) incSizeToRoot(n *tmNode[K, V]) {
	for ; n != nil; n = n.parent {
		n.size++
	}
}

// fixSizeToRoot walks from n up to the root recomputing each node's cached
// subtree size from its children. Used after a delete splice (the rotations
// inside fixAfterDelete already maintain their own sizes).
func (m *TreeMap[K, V]) fixSizeToRoot(n *tmNode[K, V]) {
	for ; n != nil; n = n.parent {
		tmFixSize(n)
	}
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
	// n is now the node physically spliced out. fixSizeFrom is the lowest node
	// whose cached subtree size must be refreshed (the removed node's surviving
	// parent at unlink time); recomputing that path to the root once the structure
	// is final restores the invariant. Rotations inside fixAfterDelete maintain
	// their own sizes and everything below fixSizeFrom is left consistent.
	var fixSizeFrom *tmNode[K, V]
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
		fixSizeFrom = child
		if !n.red {
			m.fixAfterDelete(child)
		}
	} else if n.parent == nil {
		m.root = nil
	} else {
		if !n.red {
			m.fixAfterDelete(n)
		}
		// fixAfterDelete may have rotated n to a new parent; read it now.
		fixSizeFrom = n.parent
		if n.parent != nil {
			if n == n.parent.left {
				n.parent.left = nil
			} else {
				n.parent.right = nil
			}
		}
	}
	m.fixSizeToRoot(fixSizeFrom)
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
