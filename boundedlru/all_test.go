// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package boundedlru

import (
	"testing"
)

// TestAllYieldsLruOrder pins All(): it yields every pair in LRU order (least-
// recently-used first), parallel to Keys()/Values(), and the count matches Len.
func TestAllYieldsLruOrder(t *testing.T) {
	m := NewBoundedLruInt32Int32Map(10)
	m.Put(1, 10)
	m.Put(2, 20)
	m.Put(3, 30)
	m.Get(1) // touch 1 → moves it to most-recently-used (tail); LRU now 2,3,1

	var gotK, gotV []int32
	for k, v := range m.All() {
		gotK = append(gotK, k)
		gotV = append(gotV, v)
	}

	wantK := m.Keys()
	wantV := m.Values()
	if len(gotK) != len(wantK) || len(gotV) != len(wantV) {
		t.Fatalf("All() len (%d) != Keys/Values (%d/%d)", len(gotK), len(wantK), len(wantV))
	}
	for i := range wantK {
		if gotK[i] != wantK[i] || gotV[i] != wantV[i] {
			t.Fatalf("All()[%d]=(%d,%d), want (%d,%d)", i, gotK[i], gotV[i], wantK[i], wantV[i])
		}
	}
}

// TestAllSnapshotSafeUnderTouch confirms All() is a snapshot: touching the map
// (Get, which reorders recency) during the range does not corrupt or truncate
// the iteration.
func TestAllSnapshotSafeUnderTouch(t *testing.T) {
	m := NewBoundedLruInt32Int32Map(10)
	for i := int32(1); i <= 5; i++ {
		m.Put(i, i*10)
	}
	seen := 0
	for k := range m.All() {
		m.Get(k) // reorder the live recency list mid-range
		seen++
	}
	if seen != 5 {
		t.Fatalf("All() under concurrent touch saw %d pairs, want 5", seen)
	}
}

// TestAllEarlyBreak confirms the iterator honors an early break.
func TestAllEarlyBreak(t *testing.T) {
	m := NewBoundedLruInt32Int32Map(10)
	m.Put(1, 10)
	m.Put(2, 20)
	m.Put(3, 30)
	seen := 0
	for range m.All() {
		seen++
		break
	}
	if seen != 1 {
		t.Fatalf("early-break All() saw %d, want 1", seen)
	}
}

// TestAllEmpty confirms All() over an empty map yields nothing.
func TestAllEmpty(t *testing.T) {
	seen := 0
	for range NewBoundedLruInt32Int32Map(10).All() {
		seen++
	}
	if seen != 0 {
		t.Fatalf("All() over empty map saw %d, want 0", seen)
	}
}
