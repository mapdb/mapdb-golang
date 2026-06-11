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

// LinkedHashMap is a generic insertion-ordered map backed by a Go builtin map
// and a doubly-linked list of entries. Iteration follows insertion order.
// It implements MutableMap[K, V].
type LinkedHashMap[K comparable, V any] struct {
	m    map[K]*lhmEntry[K, V]
	head *lhmEntry[K, V]
	tail *lhmEntry[K, V]
}

type lhmEntry[K comparable, V any] struct {
	key        K
	value      V
	prev, next *lhmEntry[K, V]
}

// NewLinkedHashMap creates an empty LinkedHashMap.
func NewLinkedHashMap[K comparable, V any]() *LinkedHashMap[K, V] {
	return &LinkedHashMap[K, V]{m: make(map[K]*lhmEntry[K, V])}
}

// ── MapIterable ───────────────────────────────────────────────────────

func (h *LinkedHashMap[K, V]) Get(key K) (V, bool) {
	if e, ok := h.m[key]; ok {
		return e.value, true
	}
	var zero V
	return zero, false
}

func (h *LinkedHashMap[K, V]) GetOrDefault(key K, defaultValue V) V {
	if e, ok := h.m[key]; ok {
		return e.value
	}
	return defaultValue
}

func (h *LinkedHashMap[K, V]) ContainsKey(key K) bool {
	_, ok := h.m[key]
	return ok
}

func (h *LinkedHashMap[K, V]) Size() int { return len(h.m) }

// Len returns the number of elements. It is an alias for Size, matching
// Go convention (sort.Interface, container/list, bytes.Buffer).
func (h *LinkedHashMap[K, V]) Len() int      { return h.Size() }
func (h *LinkedHashMap[K, V]) IsEmpty() bool { return len(h.m) == 0 }

func (h *LinkedHashMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for e := h.head; e != nil; e = e.next {
			if !yield(e.key, e.value) {
				return
			}
		}
	}
}

func (h *LinkedHashMap[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for e := h.head; e != nil; e = e.next {
			if !yield(e.key) {
				return
			}
		}
	}
}

func (h *LinkedHashMap[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for e := h.head; e != nil; e = e.next {
			if !yield(e.value) {
				return
			}
		}
	}
}

func (h *LinkedHashMap[K, V]) ForEach(f func(K, V)) {
	for e := h.head; e != nil; e = e.next {
		f(e.key, e.value)
	}
}

func (h *LinkedHashMap[K, V]) AnySatisfy(predicate func(K, V) bool) bool {
	for e := h.head; e != nil; e = e.next {
		if predicate(e.key, e.value) {
			return true
		}
	}
	return false
}

func (h *LinkedHashMap[K, V]) AllSatisfy(predicate func(K, V) bool) bool {
	for e := h.head; e != nil; e = e.next {
		if !predicate(e.key, e.value) {
			return false
		}
	}
	return true
}

func (h *LinkedHashMap[K, V]) NoneSatisfy(predicate func(K, V) bool) bool {
	for e := h.head; e != nil; e = e.next {
		if predicate(e.key, e.value) {
			return false
		}
	}
	return true
}

// ── MutableMap ────────────────────────────────────────────────────────

func (h *LinkedHashMap[K, V]) Put(key K, value V) (V, bool) {
	if h.m == nil {
		h.m = make(map[K]*lhmEntry[K, V])
	}
	if e, ok := h.m[key]; ok {
		old := e.value
		e.value = value
		return old, true
	}
	e := &lhmEntry[K, V]{key: key, value: value, prev: h.tail}
	if h.tail != nil {
		h.tail.next = e
	} else {
		h.head = e
	}
	h.tail = e
	h.m[key] = e
	var zero V
	return zero, false
}

func (h *LinkedHashMap[K, V]) Remove(key K) (V, bool) {
	e, ok := h.m[key]
	if !ok {
		var zero V
		return zero, false
	}
	h.unlink(e)
	delete(h.m, key)
	return e.value, true
}

func (h *LinkedHashMap[K, V]) Clear() {
	clear(h.m)
	h.head = nil
	h.tail = nil
}

func (h *LinkedHashMap[K, V]) unlink(e *lhmEntry[K, V]) {
	if e.prev != nil {
		e.prev.next = e.next
	} else {
		h.head = e.next
	}
	if e.next != nil {
		e.next.prev = e.prev
	} else {
		h.tail = e.prev
	}
}

// ── Functional operations ─────────────────────────────────────────────

func (h *LinkedHashMap[K, V]) Select(predicate func(K, V) bool) *LinkedHashMap[K, V] {
	result := NewLinkedHashMap[K, V]()
	for e := h.head; e != nil; e = e.next {
		if predicate(e.key, e.value) {
			result.Put(e.key, e.value)
		}
	}
	return result
}

func (h *LinkedHashMap[K, V]) Reject(predicate func(K, V) bool) *LinkedHashMap[K, V] {
	result := NewLinkedHashMap[K, V]()
	for e := h.head; e != nil; e = e.next {
		if !predicate(e.key, e.value) {
			result.Put(e.key, e.value)
		}
	}
	return result
}

func (h *LinkedHashMap[K, V]) Detect(predicate func(K, V) bool) (K, V, bool) {
	for e := h.head; e != nil; e = e.next {
		if predicate(e.key, e.value) {
			return e.key, e.value, true
		}
	}
	var zk K
	var zv V
	return zk, zv, false
}

func (h *LinkedHashMap[K, V]) Count(predicate func(K, V) bool) int {
	n := 0
	for e := h.head; e != nil; e = e.next {
		if predicate(e.key, e.value) {
			n++
		}
	}
	return n
}

func (h *LinkedHashMap[K, V]) KeysToSlice() []K {
	result := make([]K, 0, len(h.m))
	for e := h.head; e != nil; e = e.next {
		result = append(result, e.key)
	}
	return result
}

func (h *LinkedHashMap[K, V]) ValuesToSlice() []V {
	result := make([]V, 0, len(h.m))
	for e := h.head; e != nil; e = e.next {
		result = append(result, e.value)
	}
	return result
}

// ── Stringer ──────────────────────────────────────────────────────────

func (h *LinkedHashMap[K, V]) String() string {
	var b strings.Builder
	b.WriteByte('{')
	first := true
	for e := h.head; e != nil; e = e.next {
		if !first {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%v: %v", e.key, e.value)
		first = false
	}
	b.WriteByte('}')
	return b.String()
}

// ── Interface compliance ──────────────────────────────────────────────

var _ MutableMap[string, int] = (*LinkedHashMap[string, int])(nil)
var _ MutableMap[int, string] = (*LinkedHashMap[int, string])(nil)
