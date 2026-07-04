// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package boundedlru

import "testing"

// Regression for M-5 (todo/fable-golang/01-critical-bugs.md): reentering the map
// from the eviction callback silently corrupted the arena (in ExpireEntries it
// invalidated collected victim indices). The callback now sets a reentrancy flag
// and reentrant mutators panic instead of corrupting.

func mustPanic(t *testing.T, name string, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("%s: expected panic on reentrant call, got none", name)
		}
	}()
	f()
}

func TestReentrantCallbackPanics_Put(t *testing.T) {
	var m *BoundedLruInt32Int32Map
	m = NewBuilderBoundedLruInt32Int32Map().
		MaxSize(1).
		OnEvict(func(k, v int32, c EvictionCause) {
			mustPanic(t, "reentrant Put", func() { m.Put(99, 99) })
			mustPanic(t, "reentrant Remove", func() { m.Remove(k) })
			mustPanic(t, "reentrant Get", func() { m.Get(k) })
			mustPanic(t, "reentrant Clear", func() { m.Clear() })
		}).
		Build()
	m.Put(1, 10)
	m.Put(2, 20) // evicts key 1 -> callback runs the reentrant probes
}

func TestReentrantCallbackPanics_Expire(t *testing.T) {
	var m *BoundedLruInt32Int32Map
	m = NewBuilderBoundedLruInt32Int32Map().
		MaxSize(4).
		TTL(10).
		OnEvict(func(k, v int32, c EvictionCause) {
			mustPanic(t, "reentrant ExpireEntries", func() { m.ExpireEntries(1000) })
		}).
		Build()
	m.PutAt(1, 10, 0)
	m.PutAt(2, 20, 0)
	m.ExpireEntries(100) // both expired -> callback runs, must forbid reentry
}

// After the callback returns, the flag is cleared and normal ops resume.
func TestReentrancyFlagClearedAfterCallback(t *testing.T) {
	m := NewBoundedLruInt32Int32Map(1)
	m.Put(1, 10)
	m.Put(2, 20) // eviction fired and returned
	if _, ok := m.Get(2); !ok {
		t.Fatal("map unusable after callback returned")
	}
	m.Put(3, 30)
	if _, ok := m.Get(3); !ok {
		t.Fatal("Put after callback failed")
	}
}
