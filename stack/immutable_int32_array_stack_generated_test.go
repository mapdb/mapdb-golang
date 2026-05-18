
package stack

import "testing"

func TestImmutableInt32ArrayStack_Generated_PeekAndSize(t *testing.T) {
	m := NewInt32ArrayStack()
	m.Push(1)
	m.Push(2)
	m.Push(3)
	im := m.ToImmutable()
	if im.Size() != 3 {
		t.Errorf("Size = %d, want 3", im.Size())
	}
	if p, err := im.Peek(); err != nil || p != 3 {
		t.Errorf("Peek = (%v, %v)", p, err)
	}
}
func TestImmutableInt32ArrayStack_Generated_IsEmpty(t *testing.T) {
	im := NewInt32ArrayStack().ToImmutable()
	if !im.IsEmpty() {
		t.Error("Should be empty")
	}
}
func TestImmutableInt32ArrayStack_Generated_Contains(t *testing.T) {
	m := NewInt32ArrayStack()
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
func TestImmutableInt32ArrayStack_Generated_Push(t *testing.T) {
	m := NewInt32ArrayStack()
	m.Push(1)
	im := m.ToImmutable()
	im2 := im.Push(2)
	if im.Size() != 1 {
		t.Error("Original should not change")
	}
	if im2.Size() != 2 {
		t.Errorf("New size = %d, want 2", im2.Size())
	}
	if p, err := im2.Peek(); err != nil || p != 2 {
		t.Errorf("New peek = (%v, %v)", p, err)
	}
}
func TestImmutableInt32ArrayStack_Generated_Pop(t *testing.T) {
	m := NewInt32ArrayStack()
	m.Push(1)
	m.Push(2)
	im := m.ToImmutable()
	im2, val, err := im.Pop()
	if err != nil || val != 2 {
		t.Errorf("Pop = (%v, %v)", val, err)
	}
	if im2.Size() != 1 {
		t.Errorf("After pop size = %d, want 1", im2.Size())
	}
	if im.Size() != 2 {
		t.Error("Original should not change")
	}
}
func TestImmutableInt32ArrayStack_Generated_PopEmpty(t *testing.T) {
	im := NewInt32ArrayStack().ToImmutable()
	if _, _, err := im.Pop(); err == nil {
		t.Error("Pop on empty immutable stack should return an error")
	}
}
func TestImmutableInt32ArrayStack_Generated_All(t *testing.T) {
	m := NewInt32ArrayStack()
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
func TestImmutableInt32ArrayStack_Generated_Select(t *testing.T) {
	m := NewInt32ArrayStack()
	m.Push(1)
	m.Push(2)
	m.Push(3)
	im := m.ToImmutable()
	sel := im.Select(func(v int32) bool { return v > 1 })
	if sel.Size() != 2 {
		t.Errorf("Select size = %d, want 2", sel.Size())
	}
}
func TestImmutableInt32ArrayStack_Generated_ToSlice(t *testing.T) {
	m := NewInt32ArrayStack()
	m.Push(1)
	m.Push(2)
	im := m.ToImmutable()
	if len(im.ToSlice()) != 2 {
		t.Error("ToSlice wrong len")
	}
}
func TestImmutableInt32ArrayStack_Generated_Equals(t *testing.T) {
	m1 := NewInt32ArrayStack()
	m1.Push(1)
	m1.Push(2)
	m2 := NewInt32ArrayStack()
	m2.Push(1)
	m2.Push(2)
	if !m1.ToImmutable().Equals(m2.ToImmutable()) {
		t.Error("Should be equal")
	}
}
func TestImmutableInt32ArrayStack_Generated_ToMutable(t *testing.T) {
	m := NewInt32ArrayStack()
	m.Push(1)
	im := m.ToImmutable()
	m2 := im.ToMutable()
	m2.Push(2)
	if im.Size() != 1 {
		t.Error("Immutable should not change")
	}
}
func TestImmutableInt32ArrayStack_Generated_String(t *testing.T) {
	m := NewInt32ArrayStack()
	m.Push(1)
	if m.ToImmutable().String() == "" {
		t.Error("String empty")
	}
}
