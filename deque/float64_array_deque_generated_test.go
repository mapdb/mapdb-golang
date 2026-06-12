package deque

import (
	"math"
	"testing"
)

func TestFloat64_Generated_AddLastRemoveFirst(t *testing.T) {
	d := NewFloat64()
	d.AddLast(1.0)
	d.AddLast(2.0)
	d.AddLast(3.0)
	if d.Len() != 3 {
		t.Errorf("Size = %d, want 3", d.Len())
	}
	v0, ok := d.RemoveFirst()
	if !ok || !(math.Float64bits(v0) == math.Float64bits(1.0)) {
		t.Errorf("RemoveFirst = %v, ok=%v, want 1.0", v0, ok)
	}
	v1, ok := d.RemoveFirst()
	if !ok || !(math.Float64bits(v1) == math.Float64bits(2.0)) {
		t.Errorf("RemoveFirst = %v, ok=%v, want 2.0", v1, ok)
	}
	v2, ok := d.RemoveFirst()
	if !ok || !(math.Float64bits(v2) == math.Float64bits(3.0)) {
		t.Errorf("RemoveFirst = %v, ok=%v, want 3.0", v2, ok)
	}
	if d.Len() != 0 {
		t.Errorf("IsEmpty = false, want true")
	}
}

func TestFloat64_Generated_AddFirstRemoveLast(t *testing.T) {
	d := NewFloat64()
	d.AddFirst(1.0)
	d.AddFirst(2.0)
	d.AddFirst(3.0)
	got, ok := d.PeekFirst()
	if !ok || !(math.Float64bits(got) == math.Float64bits(3.0)) {
		t.Errorf("PeekFirst = %v, ok=%v, want 3.0", got, ok)
	}
	got, ok = d.PeekLast()
	if !ok || !(math.Float64bits(got) == math.Float64bits(1.0)) {
		t.Errorf("PeekLast = %v, ok=%v, want 1.0", got, ok)
	}
	got, _ = d.RemoveLast()
	if !(math.Float64bits(got) == math.Float64bits(1.0)) {
		t.Errorf("RemoveLast = %v, want 1.0", got)
	}
	got, _ = d.RemoveLast()
	if !(math.Float64bits(got) == math.Float64bits(2.0)) {
		t.Errorf("RemoveLast = %v, want 2.0", got)
	}
	got, _ = d.RemoveLast()
	if !(math.Float64bits(got) == math.Float64bits(3.0)) {
		t.Errorf("RemoveLast = %v, want 3.0", got)
	}
}

func TestFloat64_Generated_RemoveEmpty(t *testing.T) {
	d := NewFloat64()
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

func TestFloat64_Generated_MixedOps(t *testing.T) {
	d := NewFloat64()
	d.AddLast(2.0)
	d.AddFirst(1.0)
	d.AddLast(3.0)
	got, _ := d.RemoveFirst()
	if !(math.Float64bits(got) == math.Float64bits(1.0)) {
		t.Errorf("RemoveFirst after mixed ops = %v, want 1.0", got)
	}
	got, _ = d.RemoveLast()
	if !(math.Float64bits(got) == math.Float64bits(3.0)) {
		t.Errorf("RemoveLast after mixed ops = %v, want 3.0", got)
	}
	got, _ = d.RemoveFirst()
	if !(math.Float64bits(got) == math.Float64bits(2.0)) {
		t.Errorf("RemoveFirst last = %v, want 2.0", got)
	}
}

func TestFloat64_Generated_ContainsAndClear(t *testing.T) {
	d := NewFloat64()
	d.AddLast(1.0)
	d.AddLast(2.0)
	if !d.Contains(1.0) {
		t.Errorf("Contains(1.0) = false, want true")
	}
	d.Clear()
	if d.Len() != 0 {
		t.Errorf("IsEmpty after Clear = false, want true")
	}
}

func TestFloat64_Generated_ToSlice(t *testing.T) {
	d := Float64Of(1.0, 2.0, 3.0)
	out := d.ToSlice()
	if len(out) != 3 {
		t.Errorf("ToSlice len = %d, want 3", len(out))
	}
}

func TestFloat64_Generated_Equals(t *testing.T) {
	a := Float64Of(1.0, 2.0)
	b := Float64Of(1.0, 2.0)
	if !a.Equals(b) {
		t.Errorf("Equals = false, want true")
	}
}

func TestFloat64_Generated_String(t *testing.T) {
	d := NewFloat64()
	d.AddLast(1.0)
	if d.String() == "" {
		t.Errorf("String is empty")
	}
}

// TestFloat64_Generated_RingBufferWrapAround exercises the ring buffer's
// wrap-around logic: insert enough at the back, drain from the front,
// then insert again so head+size straddles the end of the underlying
// slice. ToSlice and iteration must still yield logical order.
func TestFloat64_Generated_RingBufferWrapAround(t *testing.T) {
	d := NewFloat64()
	// Initial capacity is 16. Fill past it to force a grow, then drain
	// and refill in a pattern that leaves head deep inside the buffer.
	for i := 0; i < 10; i++ {
		d.AddLast(1.0)
	}
	for i := 0; i < 8; i++ {
		if _, ok := d.RemoveFirst(); !ok {
			t.Fatalf("RemoveFirst during setup: %v", ok)
		}
	}
	// Now head ~= 8, size == 2. Add 10 more via AddLast — this must
	// wrap through index 0 without corrupting logical order.
	for i := 0; i < 10; i++ {
		d.AddLast(2.0)
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
		if !(math.Float64bits(out[i]) == math.Float64bits(1.0)) {
			t.Errorf("ToSlice()[%d] = %v, want 1.0", i, out[i])
		}
	}
	for i := 2; i < 12; i++ {
		if !(math.Float64bits(out[i]) == math.Float64bits(2.0)) {
			t.Errorf("ToSlice()[%d] = %v, want 2.0", i, out[i])
		}
	}
}

// TestFloat64_Generated_AddFirstIsO1 is a complexity sanity check.
// Under the old shift-based implementation, N AddFirst calls cost
// O(N^2) element moves; under the ring buffer it's O(N) amortised.
// We don't measure time (flaky in CI) — we just assert the deque
// stays functional at a size that would have been quadratic before.
func TestFloat64_Generated_AddFirstIsO1(t *testing.T) {
	const N = 10000
	d := NewFloat64()
	for i := 0; i < N; i++ {
		d.AddFirst(1.0)
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

// TestFloat64_Generated_AlternatingEndsGrow exercises the grow path while
// head is non-zero: grow() must unwrap the logical order into a
// contiguous prefix so that subsequent mask arithmetic stays correct.
func TestFloat64_Generated_AlternatingEndsGrow(t *testing.T) {
	d := NewFloat64()
	// Drive head to a non-zero value via AddFirst.
	for i := 0; i < 5; i++ {
		d.AddFirst(3.0)
	}
	// Then pile enough in to force multiple grows.
	for i := 0; i < 100; i++ {
		d.AddLast(2.0)
	}
	if d.Len() != 105 {
		t.Fatalf("Size = %d, want 105", d.Len())
	}
	out := d.ToSlice()
	for i := 0; i < 5; i++ {
		if !(math.Float64bits(out[i]) == math.Float64bits(3.0)) {
			t.Errorf("prefix[%d] = %v, want 3.0", i, out[i])
		}
	}
	for i := 5; i < 105; i++ {
		if !(math.Float64bits(out[i]) == math.Float64bits(2.0)) {
			t.Errorf("suffix[%d] = %v, want 2.0", i, out[i])
		}
	}
}
