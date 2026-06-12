package deque

import (
	"testing"
)

func TestInt8_Generated_AddLastRemoveFirst(t *testing.T) {
	d := NewInt8()
	d.AddLast(1)
	d.AddLast(2)
	d.AddLast(3)
	if d.Len() != 3 {
		t.Errorf("Size = %d, want 3", d.Len())
	}
	v0, ok := d.RemoveFirst()
	if !ok || !(v0 == 1) {
		t.Errorf("RemoveFirst = %v, ok=%v, want 1", v0, ok)
	}
	v1, ok := d.RemoveFirst()
	if !ok || !(v1 == 2) {
		t.Errorf("RemoveFirst = %v, ok=%v, want 2", v1, ok)
	}
	v2, ok := d.RemoveFirst()
	if !ok || !(v2 == 3) {
		t.Errorf("RemoveFirst = %v, ok=%v, want 3", v2, ok)
	}
	if d.Len() != 0 {
		t.Errorf("IsEmpty = false, want true")
	}
}

func TestInt8_Generated_AddFirstRemoveLast(t *testing.T) {
	d := NewInt8()
	d.AddFirst(1)
	d.AddFirst(2)
	d.AddFirst(3)
	got, ok := d.PeekFirst()
	if !ok || !(got == 3) {
		t.Errorf("PeekFirst = %v, ok=%v, want 3", got, ok)
	}
	got, ok = d.PeekLast()
	if !ok || !(got == 1) {
		t.Errorf("PeekLast = %v, ok=%v, want 1", got, ok)
	}
	got, _ = d.RemoveLast()
	if !(got == 1) {
		t.Errorf("RemoveLast = %v, want 1", got)
	}
	got, _ = d.RemoveLast()
	if !(got == 2) {
		t.Errorf("RemoveLast = %v, want 2", got)
	}
	got, _ = d.RemoveLast()
	if !(got == 3) {
		t.Errorf("RemoveLast = %v, want 3", got)
	}
}

func TestInt8_Generated_RemoveEmpty(t *testing.T) {
	d := NewInt8()
	if _, ok := d.RemoveFirst(); ok {
		t.Errorf("RemoveFirst on empty: want not-ok")
	}
	if _, ok := d.RemoveLast(); ok {
		t.Errorf("RemoveLast on empty: want not-ok")
	}
	if _, ok := d.PeekFirst(); ok {
		t.Errorf("PeekFirst on empty: want not-ok")
	}
	if _, ok := d.PeekLast(); ok {
		t.Errorf("PeekLast on empty: want not-ok")
	}
}

func TestInt8_Generated_MixedOps(t *testing.T) {
	d := NewInt8()
	d.AddLast(2)
	d.AddFirst(1)
	d.AddLast(3)
	got, _ := d.RemoveFirst()
	if !(got == 1) {
		t.Errorf("RemoveFirst after mixed ops = %v, want 1", got)
	}
	got, _ = d.RemoveLast()
	if !(got == 3) {
		t.Errorf("RemoveLast after mixed ops = %v, want 3", got)
	}
	got, _ = d.RemoveFirst()
	if !(got == 2) {
		t.Errorf("RemoveFirst last = %v, want 2", got)
	}
}

func TestInt8_Generated_ContainsAndClear(t *testing.T) {
	d := NewInt8()
	d.AddLast(1)
	d.AddLast(2)
	if !d.Contains(1) {
		t.Errorf("Contains(1) = false, want true")
	}
	d.Clear()
	if d.Len() != 0 {
		t.Errorf("IsEmpty after Clear = false, want true")
	}
}

func TestInt8_Generated_ToSlice(t *testing.T) {
	d := Int8Of(1, 2, 3)
	out := d.ToSlice()
	if len(out) != 3 {
		t.Errorf("ToSlice len = %d, want 3", len(out))
	}
}

func TestInt8_Generated_Equals(t *testing.T) {
	a := Int8Of(1, 2)
	b := Int8Of(1, 2)
	if !a.Equals(b) {
		t.Errorf("Equals = false, want true")
	}
}

func TestInt8_Generated_String(t *testing.T) {
	d := NewInt8()
	d.AddLast(1)
	if d.String() == "" {
		t.Errorf("String is empty")
	}
}

// TestInt8_Generated_RingBufferWrapAround exercises the ring buffer's
// wrap-around logic: insert enough at the back, drain from the front,
// then insert again so head+size straddles the end of the underlying
// slice. ToSlice and iteration must still yield logical order.
func TestInt8_Generated_RingBufferWrapAround(t *testing.T) {
	d := NewInt8()
	// Initial capacity is 16. Fill past it to force a grow, then drain
	// and refill in a pattern that leaves head deep inside the buffer.
	for i := 0; i < 10; i++ {
		d.AddLast(1)
	}
	for i := 0; i < 8; i++ {
		if _, ok := d.RemoveFirst(); !ok {
			t.Fatalf("RemoveFirst during setup: %v", ok)
		}
	}
	// Now head ~= 8, size == 2. Add 10 more via AddLast — this must
	// wrap through index 0 without corrupting logical order.
	for i := 0; i < 10; i++ {
		d.AddLast(2)
	}
	if d.Len() != 12 {
		t.Fatalf("Size after wrap = %d, want 12", d.Len())
	}
	// ToSlice must yield the 2 original tail elements followed by 10 new ones.
	out := d.ToSlice()
	if len(out) != 12 {
		t.Fatalf("ToSlice len after wrap = %d, want 12", len(out))
	}
	for i := 0; i < 2; i++ {
		if !(out[i] == 1) {
			t.Errorf("ToSlice()[%d] = %v, want 1", i, out[i])
		}
	}
	for i := 2; i < 12; i++ {
		if !(out[i] == 2) {
			t.Errorf("ToSlice()[%d] = %v, want 2", i, out[i])
		}
	}
}

// TestInt8_Generated_AddFirstIsO1 is a complexity sanity check.
// Under the old shift-based implementation, N AddFirst calls cost
// O(N^2) element moves; under the ring buffer it's O(N) amortised.
// We don't measure time (flaky in CI) — we just assert the deque
// stays functional at a size that would have been quadratic before.
func TestInt8_Generated_AddFirstIsO1(t *testing.T) {
	const N = 10000
	d := NewInt8()
	for i := 0; i < N; i++ {
		d.AddFirst(1)
	}
	if d.Len() != N {
		t.Fatalf("Size after %d AddFirst = %d", N, d.Len())
	}
	for i := 0; i < N; i++ {
		if _, ok := d.RemoveLast(); !ok {
			t.Fatalf("RemoveLast at %d: %v", i, ok)
		}
	}
	if d.Len() != 0 {
		t.Fatalf("not empty after full drain, size=%d", d.Len())
	}
}

// TestInt8_Generated_AlternatingEndsGrow exercises the grow path while
// head is non-zero: grow() must unwrap the logical order into a
// contiguous prefix so that subsequent mask arithmetic stays correct.
func TestInt8_Generated_AlternatingEndsGrow(t *testing.T) {
	d := NewInt8()
	// Drive head to a non-zero value via AddFirst.
	for i := 0; i < 5; i++ {
		d.AddFirst(3)
	}
	// Then pile enough in to force multiple grows.
	for i := 0; i < 100; i++ {
		d.AddLast(2)
	}
	if d.Len() != 105 {
		t.Fatalf("Size = %d, want 105", d.Len())
	}
	out := d.ToSlice()
	for i := 0; i < 5; i++ {
		if !(out[i] == 3) {
			t.Errorf("prefix[%d] = %v, want 3", i, out[i])
		}
	}
	for i := 5; i < 105; i++ {
		if !(out[i] == 2) {
			t.Errorf("suffix[%d] = %v, want 2", i, out[i])
		}
	}
}
