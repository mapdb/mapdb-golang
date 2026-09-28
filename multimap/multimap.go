package multimap

import (
	"fmt"
	"iter"
	"reflect"
	"strings"

	"github.com/mapdb/mapdb-golang/internal/floatid"
)

// Multimap is a generic one-to-many map: each key maps to a slice of values.
// Uses Go generics — works with any comparable key and any value type.
//
// This is the Go equivalent of Eclipse Collections' Multimap / ArrayListMultimap.
//
// Float keys: when K's underlying type is float32 or float64 (named types
// included) keys are identified by their IEEE 754 bit pattern, as the
// collection spec requires (algorithms.md "NaN must hash and compare by bit
// pattern"): a NaN key finds itself, distinct NaN payloads are distinct keys
// and -0.0 and +0.0 are distinct keys. Every other K uses Go ==. Composite
// keys that merely contain floats keep == semantics.
type Multimap[K comparable, V any] struct {
	data  map[K][]V                  // non-float K
	fdata map[uint64]floatSlot[K, V] // float K, keyed by bit pattern
	size  int                        // total values across all keys
}

// floatSlot keeps the original float key next to its values so iteration can
// yield it.
type floatSlot[K comparable, V any] struct {
	key  K
	vals []V
}

// NewMultimap creates a new empty Multimap.
func NewMultimap[K comparable, V any]() *Multimap[K, V] {
	m := &Multimap[K, V]{}
	m.init()
	return m
}

func (m *Multimap[K, V]) init() {
	if floatid.IsFloat[K]() {
		m.fdata = make(map[uint64]floatSlot[K, V])
	} else {
		m.data = make(map[K][]V)
	}
}

func (m *Multimap[K, V]) lookup(key K) ([]V, bool) {
	if k := floatid.Kind[K](); k != reflect.Invalid {
		s, ok := m.fdata[floatid.Bits(k, key)]
		return s.vals, ok
	}
	vals, ok := m.data[key]
	return vals, ok
}

func (m *Multimap[K, V]) store(key K, vals []V) {
	if m.data == nil && m.fdata == nil {
		m.init()
	}
	if k := floatid.Kind[K](); k != reflect.Invalid {
		m.fdata[floatid.Bits(k, key)] = floatSlot[K, V]{key: key, vals: vals}
		return
	}
	m.data[key] = vals
}

func (m *Multimap[K, V]) delete(key K) {
	if k := floatid.Kind[K](); k != reflect.Invalid {
		delete(m.fdata, floatid.Bits(k, key))
		return
	}
	delete(m.data, key)
}

// entries iterates (key, stored values) in unspecified order. The yielded
// slice is the internal one; callers copy before exposing it.
func (m *Multimap[K, V]) entries() iter.Seq2[K, []V] {
	return func(yield func(K, []V) bool) {
		if m.fdata != nil {
			for _, s := range m.fdata {
				if !yield(s.key, s.vals) {
					return
				}
			}
			return
		}
		for k, vals := range m.data {
			if !yield(k, vals) {
				return
			}
		}
	}
}

// Put adds a value for the key. Always appends (does not check for duplicates).
func (m *Multimap[K, V]) Put(key K, value V) {
	vals, _ := m.lookup(key)
	m.store(key, append(vals, value))
	m.size++
}

// PutAll adds multiple values for the key.
func (m *Multimap[K, V]) PutAll(key K, values ...V) {
	vals, _ := m.lookup(key)
	m.store(key, append(vals, values...))
	m.size += len(values)
}

// Get returns a copy of all values for the key. Returns nil if key not present.
func (m *Multimap[K, V]) Get(key K) []V {
	vals, _ := m.lookup(key)
	if vals == nil {
		return nil
	}
	result := make([]V, len(vals))
	copy(result, vals)
	return result
}

// ContainsKey returns true if the key has at least one value.
func (m *Multimap[K, V]) ContainsKey(key K) bool {
	vals, ok := m.lookup(key)
	return ok && len(vals) > 0
}

// RemoveAll removes all values for the key. Returns the removed values.
func (m *Multimap[K, V]) RemoveAll(key K) []V {
	vals, _ := m.lookup(key)
	if vals != nil {
		m.size -= len(vals)
		m.delete(key)
	}
	return vals
}

// Len returns the total number of values across all keys. Use m.Len() == 0 to
// test for emptiness.
func (m *Multimap[K, V]) Len() int { return m.size }

// SizeDistinct returns the number of distinct keys.
func (m *Multimap[K, V]) SizeDistinct() int { return len(m.data) + len(m.fdata) }

// Clear removes all entries.
func (m *Multimap[K, V]) Clear() { m.data, m.fdata = nil, nil; m.init(); m.size = 0 }

// Keys returns an iter.Seq of distinct keys.
func (m *Multimap[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range m.entries() {
			if !yield(k) {
				return
			}
		}
	}
}

// All returns an iter.Seq2 that yields each (key, value) pair.
// Keys with multiple values appear multiple times.
func (m *Multimap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, vals := range m.entries() {
			for _, v := range vals {
				if !yield(k, v) {
					return
				}
			}
		}
	}
}

// ForEachKey calls f for each distinct key with a copy of its values.
func (m *Multimap[K, V]) ForEachKey(f func(K, []V)) {
	for k, vals := range m.entries() {
		copied := make([]V, len(vals))
		copy(copied, vals)
		f(k, copied)
	}
}

// ForEach calls f for each (key, value) pair.
func (m *Multimap[K, V]) ForEach(f func(K, V)) {
	for k, vals := range m.entries() {
		for _, v := range vals {
			f(k, v)
		}
	}
}

// String returns a string representation.
func (m *Multimap[K, V]) String() string {
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for k, vals := range m.entries() {
		if !first {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v: %v", k, vals)
		first = false
	}
	sb.WriteString("}")
	return sb.String()
}
