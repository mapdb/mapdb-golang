package hashmap

import (
	"strings"
	"testing"
)

// namedStr is a named string type. The old raw-memory hasher's string fast path
// used a type assertion any(key).(string), which fails for a named type, so it
// fell through to hashing the {data ptr, len} header — keying on the backing
// array's address. Two ==-equal named strings with distinct backings hashed
// apart, making the second unfindable. maphash.Comparable fixes this.
type namedStr string

// personKey is a pointer-bearing struct: its Name field is a string header.
// Raw-memory hashing folded in the backing pointer, so two ==-equal structs
// with distinctly-backed Name strings hashed apart.
type personKey struct {
	Name string
	Age  int
}

// freshBacking returns a string with the same content as s but a guaranteed
// distinct backing array.
func freshBacking(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		b.WriteByte(s[i])
	}
	return b.String()
}

func TestHashComparable_NamedStringEqualityConsistent(t *testing.T) {
	m := NewObjectInt32HashMap[namedStr]()

	a := namedStr("hello")
	b := namedStr(freshBacking("hello"))

	if a != b {
		t.Fatalf("precondition: a and b should be == equal")
	}

	// Direct invariant: ==-equal keys MUST hash identically. This is the core
	// guarantee and fails deterministically under the old raw-memory hasher,
	// independent of how the two values happen to fall into buckets.
	if hashComparable(a) != hashComparable(b) {
		t.Errorf("hashComparable(a) != hashComparable(b) for ==-equal named strings")
	}

	m.Put(a, 1)

	if !m.ContainsKey(b) {
		t.Errorf("ContainsKey(b) = false, want true (a == b but not found)")
	}
	if v, ok := m.Get(b); !ok || v != 1 {
		t.Errorf("Get(b) = (%d, %v), want (1, true)", v, ok)
	}

	sizeBefore := m.Size()
	m.Put(b, 2)
	if m.Size() != sizeBefore {
		t.Errorf("Put(b) grew size from %d to %d; a == b so it must be the same logical key", sizeBefore, m.Size())
	}
	if v, ok := m.Get(a); !ok || v != 2 {
		t.Errorf("Get(a) after Put(b,2) = (%d, %v), want (2, true)", v, ok)
	}
}

func TestHashComparable_PointerBearingStructConsistent(t *testing.T) {
	m := NewObjectInt32HashMap[personKey]()

	a := personKey{Name: "Alice", Age: 30}
	b := personKey{Name: freshBacking("Alice"), Age: 30}

	if a != b {
		t.Fatalf("precondition: a and b should be == equal")
	}

	// Direct invariant: ==-equal struct keys MUST hash identically regardless of
	// their Name fields' backing arrays. Deterministic fail under the old hasher.
	if hashComparable(a) != hashComparable(b) {
		t.Errorf("hashComparable(a) != hashComparable(b) for ==-equal structs")
	}

	m.Put(a, 10)

	if !m.ContainsKey(b) {
		t.Errorf("ContainsKey(b) = false, want true (a == b but not found)")
	}
	if v, ok := m.Get(b); !ok || v != 10 {
		t.Errorf("Get(b) = (%d, %v), want (10, true)", v, ok)
	}

	sizeBefore := m.Size()
	m.Put(b, 20)
	if m.Size() != sizeBefore {
		t.Errorf("Put(b) grew size from %d to %d; a == b so it must be the same logical key", sizeBefore, m.Size())
	}
	if v, ok := m.Get(a); !ok || v != 20 {
		t.Errorf("Get(a) after Put(b,20) = (%d, %v), want (20, true)", v, ok)
	}
}

func TestHashComparable_PlainStringConsistent(t *testing.T) {
	m := NewObjectInt32HashMap[string]()

	a := "world"
	b := freshBacking("world")

	m.Put(a, 7)

	if !m.ContainsKey(b) {
		t.Errorf("ContainsKey(b) = false, want true")
	}
	if v, ok := m.Get(b); !ok || v != 7 {
		t.Errorf("Get(b) = (%d, %v), want (7, true)", v, ok)
	}

	sizeBefore := m.Size()
	m.Put(b, 8)
	if m.Size() != sizeBefore {
		t.Errorf("Put(b) grew size from %d to %d; same logical key expected", sizeBefore, m.Size())
	}
}

func TestHashComparable_DeterministicWithinProcess(t *testing.T) {
	k := namedStr("repeatable")
	if hashComparable(k) != hashComparable(k) {
		t.Errorf("hashComparable is not deterministic within a process for the same key")
	}
}
