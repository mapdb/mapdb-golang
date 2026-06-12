package stack

import "testing"

func TestImmutableInt64_Generated_PeekAndSize(t *testing.T) {
	m := NewInt64()
	m.Push(1)
	m.Push(2)
	m.Push(3)
	im := m.ToImmutable()
	if im.Len() != 3 {
		t.Errorf("Size = %d, want 3", im.Len())
	}
	if p, ok := im.Peek(); !ok || p != 3 {
		t.Errorf("Peek = (%v, %v)", p, ok)
	}
}
func TestImmutableInt64_Generated_IsEmpty(t *testing.T) {
	im := NewInt64().ToImmutable()
	if im.Len() != 0 {
		t.Error("Should be empty")
	}
}
func TestImmutableInt64_Generated_Contains(t *testing.T) {
	m := NewInt64()
	m.Push(1)
	m.Push(2)
	im := m.ToImmutable()
	if !im.Contains(1) {
		t.Error("Contains should be true")
	}
	if im.Contains(3) {
		t.Error("Contains should be false")
	}
}
func TestImmutableInt64_Generated_Push(t *testing.T) {
	m := NewInt64()
	m.Push(1)
	im := m.ToImmutable()
	im2 := im.Push(2)
	if im.Len() != 1 {
		t.Error("Original should not change")
	}
	if im2.Len() != 2 {
		t.Errorf("New size = %d, want 2", im2.Len())
	}
	if p, ok := im2.Peek(); !ok || p != 2 {
		t.Errorf("New peek = (%v, %v)", p, ok)
	}
}
func TestImmutableInt64_Generated_Pop(t *testing.T) {
	m := NewInt64()
	m.Push(1)
	m.Push(2)
	im := m.ToImmutable()
	im2, val, ok := im.Pop()
	if !ok || val != 2 {
		t.Errorf("Pop = (%v, %v)", val, ok)
	}
	if im2.Len() != 1 {
		t.Errorf("After pop size = %d, want 1", im2.Len())
	}
	if im.Len() != 2 {
		t.Error("Original should not change")
	}
}
func TestImmutableInt64_Generated_PopEmpty(t *testing.T) {
	im := NewInt64().ToImmutable()
	if _, _, ok := im.Pop(); ok {
		t.Error("Pop on empty immutable stack should report not-ok")
	}
}
func TestImmutableInt64_Generated_All(t *testing.T) {
	m := NewInt64()
	m.Push(1)
	m.Push(2)
	im := m.ToImmutable()
	count := 0
	for range im.All() {
		count++
	}
	if count != 2 {
		t.Errorf("All count = %d, want 2", count)
	}
}
func TestImmutableInt64_Generated_Select(t *testing.T) {
	m := NewInt64()
	m.Push(1)
	m.Push(2)
	m.Push(3)
	im := m.ToImmutable()
	sel := im.Select(func(v int64) bool { return v > 1 })
	if sel.Len() != 2 {
		t.Errorf("Select size = %d, want 2", sel.Len())
	}
}
func TestImmutableInt64_Generated_ToSlice(t *testing.T) {
	m := NewInt64()
	m.Push(1)
	m.Push(2)
	im := m.ToImmutable()
	if len(im.ToSlice()) != 2 {
		t.Error("ToSlice wrong len")
	}
}
func TestImmutableInt64_Generated_Equals(t *testing.T) {
	m1 := NewInt64()
	m1.Push(1)
	m1.Push(2)
	m2 := NewInt64()
	m2.Push(1)
	m2.Push(2)
	if !m1.ToImmutable().Equals(m2.ToImmutable()) {
		t.Error("Should be equal")
	}
}
func TestImmutableInt64_Generated_ToMutable(t *testing.T) {
	m := NewInt64()
	m.Push(1)
	im := m.ToImmutable()
	m2 := im.ToMutable()
	m2.Push(2)
	if im.Len() != 1 {
		t.Error("Immutable should not change")
	}
}
func TestImmutableInt64_Generated_String(t *testing.T) {
	m := NewInt64()
	m.Push(1)
	if m.ToImmutable().String() == "" {
		t.Error("String empty")
	}
}

func TestImmutableInt64_Generated_PeekAtPanics(t *testing.T) {
	s := NewImmutableInt64(1, 2, 3)
	assertPanics(t, func() { _ = s.PeekAt(99) })
}
