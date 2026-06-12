package multimap

import "testing"

// These are scaled KEY-IDENTITY stress tests, not hash-distribution tests. The
// i64 multimaps back onto the Go builtin map[int64][]int32, whose bucket
// distribution is not observable and whose collisions are functionally
// invisible (the runtime resolves them), so distribution quality cannot be
// asserted from behaviour. What IS observable — and what the cross-language
// validator only checks at 8 keys — is that a large family of keys sharing
// identical low 32 bits and differing only in the high 32 bits (the family that
// would collapse under a low-bits-only hash) each survives as a distinct,
// correctly-valued entry through the map's resize path. A key-equivalence or
// resize bug would lose or conflate keys here.

const i64StressN = 5000

// highBitKey returns i*2^32 + 1 — the i-th member of the {1, 2^32+1, 2*2^32+1,
// ...} family. Every member shares the low 32 bits (==1).
func highBitKey(i int) int64 { return (int64(i) << 32) | 1 }

func TestInt64ListHighBitKeyIdentityAtScale(t *testing.T) {
	m := NewInt64Int32List()
	for i := 0; i < i64StressN; i++ {
		m.Put(highBitKey(i), int32(i))
	}
	if got := m.KeysCount(); got != i64StressN {
		t.Fatalf("KeysCount() = %d, want %d (high-bit family collapsed)", got, i64StressN)
	}
	for i := 0; i < i64StressN; i++ {
		key := highBitKey(i)
		if !m.ContainsKey(key) {
			t.Fatalf("ContainsKey(%d) = false, want true", key)
		}
		vals := m.Get(key)
		if len(vals) != 1 || vals[0] != int32(i) {
			t.Fatalf("Get(%d) = %v, want [%d]", key, vals, i)
		}
	}
	if m.ContainsKey(highBitKey(i64StressN)) {
		t.Errorf("ContainsKey of an un-inserted same-family key returned true")
	}
}

func TestInt64SetHighBitKeyIdentityAtScale(t *testing.T) {
	m := NewInt64Int32Set()
	for i := 0; i < i64StressN; i++ {
		m.Put(highBitKey(i), int32(i))
	}
	if got := m.KeysCount(); got != i64StressN {
		t.Fatalf("KeysCount() = %d, want %d (high-bit family collapsed)", got, i64StressN)
	}
	for i := 0; i < i64StressN; i++ {
		key := highBitKey(i)
		if !m.ContainsKey(key) {
			t.Fatalf("ContainsKey(%d) = false, want true", key)
		}
		vals := m.Get(key)
		if len(vals) != 1 || vals[0] != int32(i) {
			t.Fatalf("Get(%d) = %v, want [%d]", key, vals, i)
		}
	}
	if m.ContainsKey(highBitKey(i64StressN)) {
		t.Errorf("ContainsKey of an un-inserted same-family key returned true")
	}
}
