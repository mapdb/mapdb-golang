package sentinelhashmap

import "testing"

// Regression for HIGH-1 (todo/fable-golang/01-critical-bugs.md): tombstones left
// by Remove were never counted toward the load factor, so a Put/Remove churn
// workload filled the table with tombstones until no empty slot remained — then
// any probe (Put, or Get of an absent key) spun forever. A default 16-slot table
// hung at ~22 churn cycles. With the fix, tombstones count toward occupancy and
// trigger a rehash that reclaims them.
//
// The test is bounded work; before the fix it does not fail, it hangs — so the
// package test timeout is the backstop.
func TestInt32Int32_ChurnDoesNotHang(t *testing.T) {
	m := NewInt32Int32()
	// Keep a small live set while churning distinct keys through the same few
	// slots. Far more cycles than the ~22 that used to hang.
	const cycles = 100_000
	for i := int32(2); i < cycles; i++ {
		m.Put(i, i*2)
		// Probe an absent key — this is the operation that spins forever once the
		// table is saturated with tombstones.
		if _, ok := m.Get(-1); ok {
			t.Fatal("Get(-1) should be absent")
		}
		m.Remove(i)
	}
	if m.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", m.Len())
	}
	// The table must NOT grow without bound under pure churn: tombstone-pressure
	// resizes rehash at the same capacity instead of doubling. With a live set of
	// at most one entry the table stays tiny; before the resize fix it grew to
	// O(total removes) (~131072 slots for this loop).
	if got := len(m.keys); got > 64 {
		t.Fatalf("table grew to %d slots under pure churn (live set ~1); resize should reclaim tombstones in place", got)
	}

	// Interleaved churn that keeps a rolling window of live keys, exercising both
	// tombstone reuse and rehash-driven reclamation.
	const window = 8
	for i := int32(2); i < cycles; i++ {
		m.Put(i, i)
		if i >= window {
			m.Remove(i - window)
		}
	}
	if got := m.Len(); got != window {
		t.Fatalf("Len() = %d, want %d", got, window)
	}
	for i := int32(cycles - window); i < cycles; i++ {
		if v, ok := m.Get(i); !ok || v != i {
			t.Errorf("Get(%d) = (%d, %v), want (%d, true)", i, v, ok, i)
		}
	}
}
