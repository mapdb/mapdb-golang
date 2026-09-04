// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package boundedlru

import (
	"reflect"
	"testing"
)

// evRec is one recorded eviction-callback invocation.
type evRec struct {
	key   int32
	value int32
	cause EvictionCause
}

// mapWithLog builds a pure max-size map of capacity n plus a recording
// eviction callback.
func mapWithLog(n int) (*BoundedLruInt32Int32Map, *[]evRec) {
	var log []evRec
	m := NewBuilderBoundedLruInt32Int32Map().
		MaxSize(n).
		OnEvict(func(k, v int32, c EvictionCause) { log = append(log, evRec{k, v, c}) }).
		Build()
	return m, &log
}

// mapWithLogTTL builds a max-size+TTL map plus a recording callback.
func mapWithLogTTL(n int, ttl uint64) (*BoundedLruInt32Int32Map, *[]evRec) {
	var log []evRec
	m := NewBuilderBoundedLruInt32Int32Map().
		MaxSize(n).
		TTL(ttl).
		OnEvict(func(k, v int32, c EvictionCause) { log = append(log, evRec{k, v, c}) }).
		Build()
	return m, &log
}

func wantKeys(t *testing.T, m *BoundedLruInt32Int32Map, want []int32) {
	t.Helper()
	got := m.Keys()
	if len(want) == 0 && len(got) == 0 {
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("keys = %v, want %v", got, want)
	}
}

func wantValues(t *testing.T, m *BoundedLruInt32Int32Map, want []int32) {
	t.Helper()
	got := m.Values()
	if len(want) == 0 && len(got) == 0 {
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("values = %v, want %v", got, want)
	}
}

func wantLog(t *testing.T, log *[]evRec, want []evRec) {
	t.Helper()
	if len(*log) == 0 && len(want) == 0 {
		return
	}
	if !reflect.DeepEqual(*log, want) {
		t.Fatalf("eviction log = %v, want %v", *log, want)
	}
}

func TestEvictBasicVictimIsLRU(t *testing.T) {
	m, log := mapWithLog(2)
	m.Put(1, 10)
	m.Put(2, 20)
	m.Put(3, 30) // evicts 1 (LRU)
	wantKeys(t, m, []int32{2, 3})
	wantValues(t, m, []int32{20, 30})
	wantLog(t, log, []evRec{{1, 10, CauseSize}})
}

func TestGetRefreshesRecency(t *testing.T) {
	m, log := mapWithLog(2)
	m.Put(1, 10)
	m.Put(2, 20)
	if v, ok := m.Get(1); !ok || v != 10 { // 1 now MRU, 2 is LRU
		t.Fatalf("Get(1) = %v,%v", v, ok)
	}
	m.Put(3, 30) // evicts 2
	wantKeys(t, m, []int32{1, 3})
	wantLog(t, log, []evRec{{2, 20, CauseSize}})
}

func TestGetOrDefaultHitRefreshesMissDoesNot(t *testing.T) {
	m, _ := mapWithLog(2)
	m.Put(1, 10)
	m.Put(2, 20)
	if got := m.GetOrDefault(1, -1); got != 10 { // hit: 1 MRU
		t.Fatalf("GetOrDefault(1) = %d", got)
	}
	if got := m.GetOrDefault(99, -1); got != -1 { // miss: no insert, no refresh
		t.Fatalf("GetOrDefault(99) = %d", got)
	}
	if m.Size() != 2 {
		t.Fatalf("size = %d", m.Size())
	}
	if m.ContainsKey(99) {
		t.Fatal("99 must not be present")
	}
	m.Put(3, 30) // evicts 2
	wantKeys(t, m, []int32{1, 3})
}

func TestContainsKeyDoesNotRefresh(t *testing.T) {
	m, log := mapWithLog(2)
	m.Put(1, 10)
	m.Put(2, 20)
	if !m.ContainsKey(1) { // must NOT refresh 1
		t.Fatal("ContainsKey(1) false")
	}
	m.Put(3, 30) // evicts 1 (still LRU)
	wantKeys(t, m, []int32{2, 3})
	wantLog(t, log, []evRec{{1, 10, CauseSize}})
}

func TestUpdateAtCapacityDoesNotEvict(t *testing.T) {
	m, log := mapWithLog(2)
	m.Put(1, 10)
	m.Put(2, 20)
	if old, ok := m.Put(1, 11); !ok || old != 10 { // update: no evict, 1 becomes MRU
		t.Fatalf("Put(1,11) = %v,%v", old, ok)
	}
	wantLog(t, log, nil)
	wantKeys(t, m, []int32{2, 1})
	m.Put(3, 30) // now evicts 2 (LRU)
	wantKeys(t, m, []int32{1, 3})
	wantLog(t, log, []evRec{{2, 20, CauseSize}})
}

func TestIterationDoesNotRefreshOrEvict(t *testing.T) {
	m, log := mapWithLog(2)
	m.Put(1, 10)
	m.Put(2, 20)
	if snap := m.Keys(); !reflect.DeepEqual(snap, []int32{1, 2}) {
		t.Fatalf("snapshot = %v", snap)
	}
	wantLog(t, log, nil)
	m.Put(3, 30) // 1 still LRU -> evicted
	wantKeys(t, m, []int32{2, 3})
	wantLog(t, log, []evRec{{1, 10, CauseSize}})
}

func TestRemoveNoCallbackNoOtherRecencyChange(t *testing.T) {
	m, log := mapWithLog(3)
	m.Put(1, 10)
	m.Put(2, 20)
	m.Put(3, 30)
	if v, ok := m.Remove(2); !ok || v != 20 {
		t.Fatalf("Remove(2) = %v,%v", v, ok)
	}
	wantLog(t, log, nil)
	wantKeys(t, m, []int32{1, 3})
	m.Put(4, 40)
	m.Put(5, 50) // capacity 3: now full {1,3,4} -> 5 evicts 1
	wantKeys(t, m, []int32{3, 4, 5})
	wantLog(t, log, []evRec{{1, 10, CauseSize}})
}

func TestClearFiresNoCallback(t *testing.T) {
	m, log := mapWithLog(3)
	m.Put(1, 10)
	m.Put(2, 20)
	m.Clear()
	if !m.IsEmpty() || m.Size() != 0 {
		t.Fatal("not empty after clear")
	}
	wantLog(t, log, nil)
	// After clear the arena/free-list is sane: reuse works.
	m.Put(7, 70)
	wantKeys(t, m, []int32{7})
}

func TestCapacityZeroDropsEverything(t *testing.T) {
	m, log := mapWithLog(0)
	if _, ok := m.Put(1, 10); ok {
		t.Fatal("put into cap-0 returned present")
	}
	m.Put(2, 20)
	m.Put(3, 30)
	if m.Size() != 0 || !m.IsEmpty() {
		t.Fatal("cap-0 not empty")
	}
	if _, ok := m.Get(1); ok {
		t.Fatal("get hit in cap-0")
	}
	wantLog(t, log, nil) // nothing was ever resident
}

func TestCapacityOneEvictsThenInserts(t *testing.T) {
	m, log := mapWithLog(1)
	m.Put(1, 10)
	m.Put(2, 20) // evicts 1
	wantKeys(t, m, []int32{2})
	wantLog(t, log, []evRec{{1, 10, CauseSize}})
	if old, ok := m.Put(2, 22); !ok || old != 20 { // update: no new log entry
		t.Fatalf("Put(2,22) = %v,%v", old, ok)
	}
	if len(*log) != 1 {
		t.Fatalf("log len = %d", len(*log))
	}
	wantKeys(t, m, []int32{2})
	wantValues(t, m, []int32{22})
}

func TestEvictBeforeInsertNewKeyNeverSelfVictim(t *testing.T) {
	m, log := mapWithLog(1)
	m.Put(1, 10)
	m.Put(2, 20) // 2 inserted, 1 evicted -- 2 is never its own victim
	if !m.ContainsKey(2) || m.ContainsKey(1) {
		t.Fatal("evict-before-insert wrong")
	}
	wantLog(t, log, []evRec{{1, 10, CauseSize}})
}

func TestSameNowRecencyUsesPositionNotNow(t *testing.T) {
	m, log := mapWithLogTTL(2, 100)
	m.PutAt(1, 10, 5)
	m.PutAt(2, 20, 5) // both written at now=5
	if v, ok := m.Get(1); !ok || v != 10 {
		t.Fatalf("Get(1) = %v,%v", v, ok)
	}
	m.PutAt(3, 30, 5) // evicts 2, NOT 1 (recency != now)
	wantKeys(t, m, []int32{1, 3})
	wantLog(t, log, []evRec{{2, 20, CauseSize}})
}

func TestExpireBasicInclusive(t *testing.T) {
	m, log := mapWithLogTTL(10, 10)
	m.PutAt(1, 10, 0) // expire_at 10
	m.PutAt(2, 20, 0) // expire_at 10
	m.PutAt(3, 30, 5) // expire_at 15
	if n := m.ExpireEntries(10); n != 2 {
		t.Fatalf("expired = %d", n)
	}
	wantKeys(t, m, []int32{3})
	wantLog(t, log, []evRec{{1, 10, CauseExpired}, {2, 20, CauseExpired}})
}

func TestExpireTiebreakByLastUse(t *testing.T) {
	m, log := mapWithLogTTL(10, 10)
	m.PutAt(1, 10, 0)
	m.PutAt(2, 20, 0)
	m.PutAt(3, 30, 0) // all expire_at 10
	m.Get(2)
	m.Get(3)
	m.Get(1) // last_use order (asc): 2 < 3 < 1
	if n := m.ExpireEntries(10); n != 3 {
		t.Fatalf("expired = %d", n)
	}
	wantLog(t, log, []evRec{
		{2, 20, CauseExpired},
		{3, 30, CauseExpired},
		{1, 10, CauseExpired},
	})
}

func TestExpireOrdersByExpireAtThenLastUse(t *testing.T) {
	m, log := mapWithLogTTL(10, 0)
	m.PutAt(1, 10, 5) // expire_at 5
	m.PutAt(2, 20, 3) // expire_at 3
	m.PutAt(3, 30, 5) // expire_at 5
	m.PutAt(4, 40, 3) // expire_at 3
	if n := m.ExpireEntries(5); n != 4 {
		t.Fatalf("expired = %d", n)
	}
	wantLog(t, log, []evRec{
		{2, 20, CauseExpired},
		{4, 40, CauseExpired},
		{1, 10, CauseExpired},
		{3, 30, CauseExpired},
	})
}

func TestExpireInclusiveAndSaturation(t *testing.T) {
	m, log := mapWithLogTTL(10, neverExpire) // ttl huge: now+ttl saturates -> never
	m.PutAt(1, 10, 5)
	if n := m.ExpireEntries(neverExpire - 1); n != 0 {
		t.Fatalf("expired = %d", n)
	}
	if !m.ContainsKey(1) {
		t.Fatal("1 should survive")
	}
	wantLog(t, log, nil)
}

func TestTTLZeroBoundary(t *testing.T) {
	m, _ := mapWithLogTTL(10, 0)
	m.PutAt(1, 10, 5) // expire_at = 5
	if n := m.ExpireEntries(4); n != 0 {
		t.Fatalf("4: expired = %d", n)
	}
	if !m.ContainsKey(1) {
		t.Fatal("1 should survive at now=4")
	}
	if n := m.ExpireEntries(5); n != 1 {
		t.Fatalf("5: expired = %d", n)
	}
	if m.ContainsKey(1) {
		t.Fatal("1 should be gone at now=5")
	}
}

func TestNoTTLPureLRU(t *testing.T) {
	m, _ := mapWithLog(2) // no ttl
	m.Put(1, 10)
	m.Put(2, 20)
	if n := m.ExpireEntries(neverExpire); n != 0 {
		t.Fatalf("expired = %d", n)
	}
	wantKeys(t, m, []int32{1, 2})
}

func TestU64MaxExpireAtIsNeverSentinel(t *testing.T) {
	m, log := mapWithLogTTL(10, 1)
	m.PutAt(1, 10, neverExpire-1) // now+ttl = MaxUint64 exactly -> sentinel
	if n := m.ExpireEntries(neverExpire); n != 0 {
		t.Fatalf("expired = %d", n)
	}
	if !m.ContainsKey(1) {
		t.Fatal("sentinel entry must never expire")
	}
	wantLog(t, log, nil)
}

func TestExpireThenSizeInteraction(t *testing.T) {
	m, log := mapWithLogTTL(2, 10)
	m.PutAt(1, 10, 0)
	m.PutAt(2, 20, 0)
	if n := m.ExpireEntries(10); n != 2 {
		t.Fatalf("expired = %d", n)
	}
	if !m.IsEmpty() {
		t.Fatal("not empty after expire")
	}
	m.PutAt(3, 30, 20) // below capacity now: no SIZE eviction
	wantKeys(t, m, []int32{3})
	wantLog(t, log, []evRec{{1, 10, CauseExpired}, {2, 20, CauseExpired}})
}

func TestUpdateBeforeExpireResetsExpiryAndValue(t *testing.T) {
	m, log := mapWithLogTTL(10, 10)
	m.PutAt(1, 10, 0) // expire_at 10
	if old, ok := m.PutAt(1, 11, 5); !ok || old != 10 {
		t.Fatalf("PutAt update = %v,%v", old, ok)
	}
	if n := m.ExpireEntries(10); n != 0 { // survives (15 > 10)
		t.Fatalf("expired = %d", n)
	}
	if !m.ContainsKey(1) {
		t.Fatal("1 should survive")
	}
	if n := m.ExpireEntries(15); n != 1 { // now expires with UPDATED value 11
		t.Fatalf("expired = %d", n)
	}
	wantLog(t, log, []evRec{{1, 11, CauseExpired}})
}

func TestMissDoesNotRefreshOrInsert(t *testing.T) {
	m, _ := mapWithLog(2)
	m.Put(1, 10)
	m.Put(2, 20) // {1(LRU), 2}
	if _, ok := m.Get(99); ok {
		t.Fatal("miss returned hit")
	}
	if got := m.GetOrDefault(99, -1); got != -1 {
		t.Fatalf("GetOrDefault(99) = %d", got)
	}
	if m.Size() != 2 {
		t.Fatalf("size = %d", m.Size())
	}
	m.Put(4, 40) // 1 still LRU -> evicted
	wantKeys(t, m, []int32{2, 4})
}

func TestRemoveReinsertGetsFreshRecency(t *testing.T) {
	m, log := mapWithLog(3)
	m.Put(1, 10)
	m.Put(2, 20)
	m.Put(3, 30) // {1,2,3}
	m.Remove(1)
	m.Put(1, 11) // fresh insert: 1 is MRU now, order {2,3,1}
	m.Put(4, 40) // evicts 2 (LRU)
	wantKeys(t, m, []int32{3, 1, 4})
	wantLog(t, log, []evRec{{2, 20, CauseSize}})
}

func TestSlotReuseAfterEvictionNoDangling(t *testing.T) {
	m, _ := mapWithLog(3)
	for k := int32(0); k < 1000; k++ {
		m.Put(k, k*10)
		if m.Size() > 3 {
			t.Fatalf("size %d exceeds cap at k=%d", m.Size(), k)
		}
	}
	wantKeys(t, m, []int32{997, 998, 999})
	wantValues(t, m, []int32{9970, 9980, 9990})
	if len(m.arena) > 4 {
		t.Fatalf("arena grew to %d slots (free-list not reused)", len(m.arena))
	}
}

func TestEntriesLRUOrderPairing(t *testing.T) {
	m, _ := mapWithLog(3)
	m.Put(1, 10)
	m.Put(2, 20)
	m.Put(3, 30)
	m.Get(1) // order becomes 2,3,1
	got := m.Entries()
	want := []Entry{{2, 20}, {3, 30}, {1, 10}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("entries = %v, want %v", got, want)
	}
}

func TestSnapshotIndependence(t *testing.T) {
	m, _ := mapWithLog(3)
	m.Put(1, 10)
	m.Put(2, 20)
	snap := m.Keys()
	m.Put(3, 30) // mutate after snapshot
	m.Get(1)
	if !reflect.DeepEqual(snap, []int32{1, 2}) {
		t.Fatalf("snapshot mutated to %v", snap)
	}
}

func TestNegativeCapacityClampsToZeroDropEverything(t *testing.T) {
	// A negative capacity is meaningless; it must clamp to 0 (drop-everything)
	// rather than panic on first insert (Rust's usize cannot go negative).
	m, log := mapWithLog(-1)
	if m.Capacity() != 0 {
		t.Fatalf("capacity = %d, want 0", m.Capacity())
	}
	if _, ok := m.Put(1, 10); ok { // must not panic; entry is dropped
		t.Fatal("put into negative-cap returned present")
	}
	if m.Size() != 0 || !m.IsEmpty() {
		t.Fatal("negative-cap map not empty")
	}
	wantLog(t, log, nil) // nothing was ever resident
	// The builder constructor path clamps too.
	if NewBoundedLruInt32Int32Map(-5).Capacity() != 0 {
		t.Fatal("constructor did not clamp negative capacity")
	}
}

func TestCapacityAccessor(t *testing.T) {
	m := NewBoundedLruInt32Int32Map(7)
	if m.Capacity() != 7 {
		t.Fatalf("capacity = %d", m.Capacity())
	}
	if m.Len() != 0 {
		t.Fatalf("len = %d", m.Len())
	}
}

// TestTieFreeDeterminismOverRandomSequence replays a deterministic
// pseudo-random op sequence twice and requires identical contents (the model
// is tie-free by construction).
func TestTieFreeDeterminismOverRandomSequence(t *testing.T) {
	replay := func() ([]int32, []int32, []evRec) {
		m, log := mapWithLog(5)
		state := uint64(0x12345678)
		for i := 0; i < 2000; i++ {
			state = state*6364136223846793005 + 1
			k := int32((state >> 33) % 20)
			switch (state >> 30) & 3 {
			case 0:
				m.Put(k, k*100)
			case 1:
				m.Get(k)
			case 2:
				m.ContainsKey(k)
			default:
				m.Remove(k)
			}
		}
		return m.Keys(), m.Values(), *log
	}
	k1, v1, l1 := replay()
	k2, v2, l2 := replay()
	if !reflect.DeepEqual(k1, k2) || !reflect.DeepEqual(v1, v2) || !reflect.DeepEqual(l1, l2) {
		t.Fatal("replay not deterministic")
	}
}

// TestPeekDoesNotRefreshRecency pins the contract that distinguishes Peek from
// Get: a Peek hit must leave the LRU order exactly as it was, so the next
// size-eviction still evicts the entry Get would have rescued.
func TestPeekDoesNotRefreshRecency(t *testing.T) {
	m := NewBoundedLruInt32Int32Map(2)
	m.Put(1, 10)
	m.Put(2, 20)

	if v, ok := m.Peek(1); !ok || v != 10 {
		t.Fatalf("Peek(1) = (%d, %v), want (10, true)", v, ok)
	}
	if _, ok := m.Peek(99); ok {
		t.Errorf("Peek(99) reported a hit on an absent key")
	}
	// Peek must NOT have rescued key 1: it is still the least-recently-used, so
	// inserting a third key evicts it. (With Get in place of Peek, key 2 would
	// be evicted instead -- that is exactly the perturbation Peek avoids.)
	m.Put(3, 30)
	if m.ContainsKey(1) {
		t.Errorf("Peek refreshed recency: key 1 survived the size eviction")
	}
	if !m.ContainsKey(2) || !m.ContainsKey(3) {
		t.Errorf("unexpected eviction: keys = %v", m.Keys())
	}

	// Contrast: Get DOES refresh, so the same sequence rescues the peeked key.
	n := NewBoundedLruInt32Int32Map(2)
	n.Put(1, 10)
	n.Put(2, 20)
	if v, ok := n.Get(1); !ok || v != 10 {
		t.Fatalf("Get(1) = (%d, %v), want (10, true)", v, ok)
	}
	n.Put(3, 30)
	if !n.ContainsKey(1) {
		t.Errorf("Get did not refresh recency: key 1 was evicted")
	}
	if n.ContainsKey(2) {
		t.Errorf("expected key 2 to be evicted after Get(1) refreshed key 1")
	}
}
