// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package object

import (
	"math/rand"
	"testing"
)

const propTrials = 1000

// Property: After adding N elements, Len() == N
func TestProperty_ArrayList_LenAfterAdd(t *testing.T) {
	for trial := 0; trial < propTrials; trial++ {
		n := rand.Intn(200)
		l := NewArrayList[int]()
		for i := 0; i < n; i++ {
			l.Add(i)
		}
		if l.Size() != n {
			t.Fatalf("trial %d: expected len %d, got %d", trial, n, l.Size())
		}
	}
}

// Property: Contains(v) == true iff v was added
func TestProperty_HashSet_ContainsAfterAdd(t *testing.T) {
	for trial := 0; trial < propTrials; trial++ {
		n := rand.Intn(200)
		s := NewHashSet[int]()
		added := make(map[int]bool)
		for i := 0; i < n; i++ {
			v := rand.Intn(500)
			s.Add(v)
			added[v] = true
		}
		for v := range added {
			if !s.Contains(v) {
				t.Fatalf("trial %d: set should contain %d", trial, v)
			}
		}
		// check a value not added
		for probe := 500; probe < 600; probe++ {
			if !added[probe] && s.Contains(probe) {
				t.Fatalf("trial %d: set should not contain %d", trial, probe)
			}
		}
	}
}

// Property: Bag occurrences match count of adds
func TestProperty_HashBag_OccurrencesMatchAdds(t *testing.T) {
	for trial := 0; trial < propTrials; trial++ {
		n := rand.Intn(200)
		b := NewHashBag[int]()
		counts := make(map[int]int)
		for i := 0; i < n; i++ {
			v := rand.Intn(50)
			b.Add(v)
			counts[v]++
		}
		for v, expected := range counts {
			got := b.OccurrencesOf(v)
			if got != expected {
				t.Fatalf("trial %d: expected %d occurrences of %d, got %d", trial, expected, v, got)
			}
		}
		if b.Size() != n {
			t.Fatalf("trial %d: expected total size %d, got %d", trial, n, b.Size())
		}
	}
}

// Property: HashMap get returns what was put
func TestProperty_HashMap_GetAfterPut(t *testing.T) {
	for trial := 0; trial < propTrials; trial++ {
		n := rand.Intn(200)
		m := NewHashMap[int, int]()
		expected := make(map[int]int)
		for i := 0; i < n; i++ {
			k := rand.Intn(100)
			v := rand.Intn(10000)
			m.Put(k, v)
			expected[k] = v
		}
		for k, v := range expected {
			got, ok := m.Get(k)
			if !ok || got != v {
				t.Fatalf("trial %d: expected m[%d]=%d, got %d (ok=%v)", trial, k, v, got, ok)
			}
		}
		if m.Size() != len(expected) {
			t.Fatalf("trial %d: expected size %d, got %d", trial, len(expected), m.Size())
		}
	}
}

// Property: Stack LIFO order — pop returns elements in reverse push order
func TestProperty_ArrayStack_LIFO(t *testing.T) {
	for trial := 0; trial < propTrials; trial++ {
		n := rand.Intn(100)
		s := NewArrayStack[int]()
		values := make([]int, n)
		for i := 0; i < n; i++ {
			v := rand.Intn(10000)
			values[i] = v
			s.Push(v)
		}
		for i := n - 1; i >= 0; i-- {
			v, err := s.Pop()
			if err != nil || v != values[i] {
				t.Fatalf("trial %d: expected pop %d, got %d (err=%v)", trial, values[i], v, err)
			}
		}
		if s.Size() != 0 {
			t.Fatalf("trial %d: stack should be empty after popping all", trial)
		}
	}
}

// Property: ForEach visits exactly Len() elements
func TestProperty_ArrayList_ForEachCount(t *testing.T) {
	for trial := 0; trial < propTrials; trial++ {
		n := rand.Intn(200)
		l := NewArrayList[int]()
		for i := 0; i < n; i++ {
			l.Add(rand.Intn(10000))
		}
		count := 0
		l.ForEach(func(v int) { count++ })
		if count != n {
			t.Fatalf("trial %d: forEach visited %d elements, expected %d", trial, count, n)
		}
	}
}

// Property: BiMap bijection — keys and values are both unique
func TestProperty_HashBiMap_Bijection(t *testing.T) {
	for trial := 0; trial < propTrials; trial++ {
		n := rand.Intn(100)
		bm := NewHashBiMap[int, int]()
		for i := 0; i < n; i++ {
			k := rand.Intn(200)
			v := rand.Intn(200)
			bm.Put(k, v)
		}
		// All keys map to unique values
		seenValues := make(map[int]bool)
		bm.ForEach(func(k, v int) {
			if seenValues[v] {
				t.Fatalf("trial %d: duplicate value %d in bimap", trial, v)
			}
			seenValues[v] = true
		})
	}
}
