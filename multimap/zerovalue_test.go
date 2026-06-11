package multimap

import "testing"

// Zero-value (nil-map) usability: a freshly declared multimap must accept Put
// without panicking ("assignment to entry in nil map"). Phase 7a.

func TestZeroValueInt32Int32ListMultimap(t *testing.T) {
	var m Int32Int32ListMultimap
	m.Put(1, 10)
	m.Put(1, 20)
	if got := m.Size(); got != 2 {
		t.Fatalf("Size() = %d, want 2", got)
	}
	if got := m.GetAll(1); len(got) != 2 {
		t.Fatalf("GetAll(1) len = %d, want 2", len(got))
	}
}

func TestZeroValueInt32Int32SetMultimap(t *testing.T) {
	var m Int32Int32SetMultimap
	m.Put(1, 10)
	m.Put(1, 10) // dedup
	if got := m.Size(); got != 1 {
		t.Fatalf("Size() = %d, want 1", got)
	}
}

func TestZeroValueGenericMultimap(t *testing.T) {
	var m Multimap[int, int]
	m.Put(1, 10)
	m.PutAll(1, 20, 30)
	if got := m.Size(); got != 3 {
		t.Fatalf("Size() = %d, want 3", got)
	}
}

func TestZeroValueFloat64Int32Multimaps(t *testing.T) {
	var l Float64Int32ListMultimap
	l.Put(1.5, 10)
	if got := l.Size(); got != 1 {
		t.Fatalf("list Size() = %d, want 1", got)
	}
	var s Float64Int32SetMultimap
	s.Put(1.5, 10)
	if got := s.Size(); got != 1 {
		t.Fatalf("set Size() = %d, want 1", got)
	}
}
