// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package object_test

// Hand-written conformance stamp (todo 14 §4) for the generic object maps,
// multimaps, and bimap. They expose All() iter.Seq2[K,V] + Len(), so the
// size-accounting law (Len ≡ number of pairs All() yields) applies to all of
// them, and TreeMap additionally gets the KeysAscending law (its keys are unique
// and sorted). Multimaps deliberately get size-accounting only: their Len() is
// the TOTAL pair count (totalSize, not distinct keys), and their keys repeat, so
// strict KeysAscending does not apply even to the sorted TreeMultimap.

import (
	"testing"

	"github.com/mapdb/mapdb-golang/internal/conformance"
	"github.com/mapdb/mapdb-golang/object"
)

// seven distinct keys → Len is a non-trivial 7 for the unique-key maps.
var (
	objMapKeys = []int{3, 1, 4, 5, 9, 2, 6}
	objMapVals = []int{0, 1, 2, 3, 4, 5, 6}
)

func TestConformanceLen2HashMap(t *testing.T) {
	m := object.NewHashMap[int, int]()
	for i, k := range objMapKeys {
		m.Put(k, objMapVals[i])
	}
	conformance.Len2MatchesAll(t, m.Len(), m.All())
}

func TestConformanceLen2LinkedHashMap(t *testing.T) {
	m := object.NewLinkedHashMap[int, int]()
	for i, k := range objMapKeys {
		m.Put(k, objMapVals[i])
	}
	conformance.Len2MatchesAll(t, m.Len(), m.All())
}

func TestConformanceLen2HashBiMap(t *testing.T) {
	// A bimap needs distinct values as well as distinct keys; objMapVals are
	// distinct, so the fixture is a valid bijection.
	m := object.NewHashBiMap[int, int]()
	for i, k := range objMapKeys {
		m.Put(k, objMapVals[i])
	}
	conformance.Len2MatchesAll(t, m.Len(), m.All())
}

func TestConformanceLen2TreeMap(t *testing.T) {
	m := object.NewTreeMap[int, int](object.NaturalComparator[int]())
	for i, k := range objMapKeys {
		m.Put(k, objMapVals[i])
	}
	conformance.Len2MatchesAll(t, m.Len(), m.All())
}

// TreeMap keys are unique and sorted, so the ordering law applies.
func TestConformanceKeysAscendingTreeMap(t *testing.T) {
	m := object.NewTreeMap[int, int](object.NaturalComparator[int]())
	for i, k := range objMapKeys {
		m.Put(k, objMapVals[i])
	}
	conformance.KeysAscending(t, m.All())
}

// objMultimapPairs has repeated keys, so Len (totalSize) exceeds the distinct-key
// count and the size-accounting law meaningfully counts every value.
var objMultimapPairs = [][2]int{{1, 10}, {1, 20}, {2, 30}, {3, 40}, {3, 50}, {3, 60}}

func TestConformanceLen2HashMultimap(t *testing.T) {
	m := object.NewHashMultimap[int, int]()
	for _, p := range objMultimapPairs {
		m.Put(p[0], p[1])
	}
	conformance.Len2MatchesAll(t, m.Len(), m.All())
}

func TestConformanceLen2TreeMultimap(t *testing.T) {
	m := object.NewTreeMultimap[int, int](object.NaturalComparator[int]())
	for _, p := range objMultimapPairs {
		m.Put(p[0], p[1])
	}
	conformance.Len2MatchesAll(t, m.Len(), m.All())
}
