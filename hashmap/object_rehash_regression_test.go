package hashmap

import (
	"math/rand"
	"testing"
)

// Regression for HIGH-2 (todo/fable-golang/01-critical-bugs.md): the object-keyed
// and object-valued hashmap templates had an inverted backward-shift condition in
// rehashFrom (and a missing wraparound guard), so Remove silently lost keys and
// created ghost duplicates. A random Put/Remove differential test vs a builtin map
// exposed it (Len() 328 vs true 211 in one reported run). The prim×prim maps, which
// carried the correct logic, passed the same test.
//
// These tests drive far more churn than the default table capacity so the delete
// path and its rehash run repeatedly.

func TestObjectInt32_RemoveDifferential(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	m := NewObjectInt32[int32]()
	ref := map[int32]int32{}

	// Small key domain forces frequent collisions, tombstones, and rehashes.
	const domain = 64
	for op := 0; op < 200_000; op++ {
		k := int32(rng.Intn(domain))
		if rng.Intn(2) == 0 {
			v := rng.Int31()
			m.Put(k, v)
			ref[k] = v
		} else {
			mo, mok := m.Remove(k)
			ro, rok := ref[k]
			delete(ref, k)
			if mok != rok || (mok && mo != ro) {
				t.Fatalf("op %d Remove(%d) = (%d,%v), ref (%d,%v)", op, k, mo, mok, ro, rok)
			}
		}
		if m.Len() != len(ref) {
			t.Fatalf("op %d Len() = %d, ref %d", op, m.Len(), len(ref))
		}
	}
	// Every reference key must be present with the right value, and no ghosts.
	for k, v := range ref {
		if got, ok := m.Get(k); !ok || got != v {
			t.Fatalf("Get(%d) = (%d,%v), want (%d,true)", k, got, ok, v)
		}
	}
	seen := 0
	for k, v := range m.All() {
		seen++
		if rv, ok := ref[k]; !ok || rv != v {
			t.Fatalf("All yielded ghost/mismatched (%d,%d); ref has (%d,%v)", k, v, rv, ok)
		}
	}
	if seen != len(ref) {
		t.Fatalf("All yielded %d entries, want %d", seen, len(ref))
	}
}

func TestInt32Object_RemoveDifferential(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	m := NewInt32Object[int32]()
	ref := map[int32]int32{}

	const domain = 64
	for op := 0; op < 200_000; op++ {
		k := int32(rng.Intn(domain))
		if rng.Intn(2) == 0 {
			v := rng.Int31()
			m.Put(k, v)
			ref[k] = v
		} else {
			mo, mok := m.Remove(k)
			ro, rok := ref[k]
			delete(ref, k)
			if mok != rok || (mok && mo != ro) {
				t.Fatalf("op %d Remove(%d) = (%d,%v), ref (%d,%v)", op, k, mo, mok, ro, rok)
			}
		}
		if m.Len() != len(ref) {
			t.Fatalf("op %d Len() = %d, ref %d", op, m.Len(), len(ref))
		}
	}
	for k, v := range ref {
		if got, ok := m.Get(k); !ok || got != v {
			t.Fatalf("Get(%d) = (%d,%v), want (%d,true)", k, got, ok, v)
		}
	}
	seen := 0
	for k, v := range m.All() {
		seen++
		if rv, ok := ref[k]; !ok || rv != v {
			t.Fatalf("All yielded ghost/mismatched (%d,%d); ref has (%d,%v)", k, v, rv, ok)
		}
	}
	if seen != len(ref) {
		t.Fatalf("All yielded %d entries, want %d", seen, len(ref))
	}
}
