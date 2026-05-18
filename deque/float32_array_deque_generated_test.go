
package deque

import (
	"math"
	"testing"
)

func TestFloat32ArrayDeque_Generated_AddLastRemoveFirst(t *testing.T) {
	d := NewFloat32ArrayDeque()
	d.AddLast(1.0)
	d.AddLast(2.0)
	d.AddLast(3.0)
	if d.Size() != 3 {
		t.Errorf("Size = %d, want 3", d.Size())
	}
	v0, err := d.RemoveFirst()
	if err != nil || !(math.Float32bits(v0) == math.Float32bits(1.0)) {
		t.Errorf("RemoveFirst = %v, err=%v, want 1.0", v0, err)
	}
	v1, err := d.RemoveFirst()
	if err != nil || !(math.Float32bits(v1) == math.Float32bits(2.0)) {
		t.Errorf("RemoveFirst = %v, err=%v, want 2.0", v1, err)
	}
	v2, err := d.RemoveFirst()
	if err != nil || !(math.Float32bits(v2) == math.Float32bits(3.0)) {
		t.Errorf("RemoveFirst = %v, err=%v, want 3.0", v2, err)
	}
	if !d.IsEmpty() {
		t.Errorf("IsEmpty = false, want true")
	}
}

func TestFloat32ArrayDeque_Generated_AddFirstRemoveLast(t *testing.T) {
	d := NewFloat32ArrayDeque()
	d.AddFirst(1.0)
	d.AddFirst(2.0)
	d.AddFirst(3.0)
	got, err := d.PeekFirst()
	if err != nil || !(math.Float32bits(got) == math.Float32bits(3.0)) {
		t.Errorf("PeekFirst = %v, err=%v, want 3.0", got, err)
	}
	got, err = d.PeekLast()
	if err != nil || !(math.Float32bits(got) == math.Float32bits(1.0)) {
		t.Errorf("PeekLast = %v, err=%v, want 1.0", got, err)
	}
	got, _ = d.RemoveLast()
	if !(math.Float32bits(got) == math.Float32bits(1.0)) {
		t.Errorf("RemoveLast = %v, want 1.0", got)
	}
	got, _ = d.RemoveLast()
	if !(math.Float32bits(got) == math.Float32bits(2.0)) {
		t.Errorf("RemoveLast = %v, want 2.0", got)
	}
	got, _ = d.RemoveLast()
	if !(math.Float32bits(got) == math.Float32bits(3.0)) {
		t.Errorf("RemoveLast = %v, want 3.0", got)
	}
}

func TestFloat32ArrayDeque_Generated_RemoveEmpty(t *testing.T) {
	d := NewFloat32ArrayDeque()
	if _, err := d.RemoveFirst(); err == nil {
		t.Errorf("RemoveFirst on empty: want error")
	}
	if _, err := d.RemoveLast(); err == nil {
		t.Errorf("RemoveLast on empty: want error")
	}
	if _, err := d.PeekFirst(); err == nil {
		t.Errorf("PeekFirst on empty: want error")
	}
	if _, err := d.PeekLast(); err == nil {
		t.Errorf("PeekLast on empty: want error")
	}
}

func TestFloat32ArrayDeque_Generated_MixedOps(t *testing.T) {
	d := NewFloat32ArrayDeque()
	d.AddLast(2.0)
	d.AddFirst(1.0)
	d.AddLast(3.0)
	got, _ := d.RemoveFirst()
	if !(math.Float32bits(got) == math.Float32bits(1.0)) {
		t.Errorf("RemoveFirst after mixed ops = %v, want 1.0", got)
	}
	got, _ = d.RemoveLast()
	if !(math.Float32bits(got) == math.Float32bits(3.0)) {
		t.Errorf("RemoveLast after mixed ops = %v, want 3.0", got)
	}
	got, _ = d.RemoveFirst()
	if !(math.Float32bits(got) == math.Float32bits(2.0)) {
		t.Errorf("RemoveFirst last = %v, want 2.0", got)
	}
}

func TestFloat32ArrayDeque_Generated_ContainsAndClear(t *testing.T) {
	d := NewFloat32ArrayDeque()
	d.AddLast(1.0)
	d.AddLast(2.0)
	if !d.Contains(1.0) {
		t.Errorf("Contains(1.0) = false, want true")
	}
	d.Clear()
	if !d.IsEmpty() {
		t.Errorf("IsEmpty after Clear = false, want true")
	}
}

func TestFloat32ArrayDeque_Generated_ToSlice(t *testing.T) {
	d := Float32ArrayDequeOf(1.0, 2.0, 3.0)
	out := d.ToSlice()
	if len(out) != 3 {
		t.Errorf("ToSlice len = %d, want 3", len(out))
	}
}

func TestFloat32ArrayDeque_Generated_Equals(t *testing.T) {
	a := Float32ArrayDequeOf(1.0, 2.0)
	b := Float32ArrayDequeOf(1.0, 2.0)
	if !a.Equals(b) {
		t.Errorf("Equals = false, want true")
	}
}

func TestFloat32ArrayDeque_Generated_String(t *testing.T) {
	d := NewFloat32ArrayDeque()
	d.AddLast(1.0)
	if d.String() == "" {
		t.Errorf("String is empty")
	}
}

// TestFloat32ArrayDeque_Generated_RingBufferWrapAround exercises the ring buffer's
// wrap-around logic: insert enough at the back, drain from the front,
// then insert again so head+size straddles the end of the underlying
// slice. ToSlice and iteration must still yield logical order.
func TestFloat32ArrayDeque_Generated_RingBufferWrapAround(t *testing.T) {
	d := NewFloat32ArrayDeque()
	// Initial capacity is 16. Fill past it to force a grow, then drain
	// and refill in a pattern that leaves head deep inside the buffer.
	for i := 0; i < 10; i++ {
		d.AddLast(1.0)
	}
	for i := 0; i < 8; i++ {
		if _, err := d.RemoveFirst(); err != nil {
			t.Fatalf("RemoveFirst during setup: %v", err)
		}
	}
	// Now head ~= 8, size == 2. Add 10 more via AddLast — this must
	// wrap through index 0 without corrupting logical order.
	for i := 0; i < 10; i++ {
		d.AddLast(2.0)
	}
	if d.Size() != 12 {
		t.Fatalf("Size after wrap = %d, want 12", d.Size())
	}
	// ToSlice must yield the 2 original tail elements followed by 10 new ones.
	out := d.ToSlice()
	if len(out) != 12 {
		t.Fatalf("ToSlice len after wrap = %d, want 12", len(out))
	}
	for i := 0; i < 2; i++ {
		if !(math.Float32bits(out[i]) == math.Float32bits(1.0)) {
			t.Errorf("ToSlice()[%d] = %v, want 1.0", i, out[i])
		}
	}
	for i := 2; i < 12; i++ {
		if !(math.Float32bits(out[i]) == math.Float32bits(2.0)) {
			t.Errorf("ToSlice()[%d] = %v, want 2.0", i, out[i])
		}
	}
}

// TestFloat32ArrayDeque_Generated_AddFirstIsO1 is a complexity sanity check.
// Under the old shift-based implementation, N AddFirst calls cost
// O(N^2) element moves; under the ring buffer it's O(N) amortised.
// We don't measure time (flaky in CI) — we just assert the deque
// stays functional at a size that would have been quadratic before.
func TestFloat32ArrayDeque_Generated_AddFirstIsO1(t *testing.T) {
	const N = 10000
	d := NewFloat32ArrayDeque()
	for i := 0; i < N; i++ {
		d.AddFirst(1.0)
	}
	if d.Size() != N {
		t.Fatalf("Size after %d AddFirst = %d", N, d.Size())
	}
	for i := 0; i < N; i++ {
		if _, err := d.RemoveLast(); err != nil {
			t.Fatalf("RemoveLast at %d: %v", i, err)
		}
	}
	if !d.IsEmpty() {
		t.Fatalf("not empty after full drain, size=%d", d.Size())
	}
}

// TestFloat32ArrayDeque_Generated_AlternatingEndsGrow exercises the grow path while
// head is non-zero: grow() must unwrap the logical order into a
// contiguous prefix so that subsequent mask arithmetic stays correct.
func TestFloat32ArrayDeque_Generated_AlternatingEndsGrow(t *testing.T) {
	d := NewFloat32ArrayDeque()
	// Drive head to a non-zero value via AddFirst.
	for i := 0; i < 5; i++ {
		d.AddFirst(3.0)
	}
	// Then pile enough in to force multiple grows.
	for i := 0; i < 100; i++ {
		d.AddLast(2.0)
	}
	if d.Size() != 105 {
		t.Fatalf("Size = %d, want 105", d.Size())
	}
	out := d.ToSlice()
	for i := 0; i < 5; i++ {
		if !(math.Float32bits(out[i]) == math.Float32bits(3.0)) {
			t.Errorf("prefix[%d] = %v, want 3.0", i, out[i])
		}
	}
	for i := 5; i < 105; i++ {
		if !(math.Float32bits(out[i]) == math.Float32bits(2.0)) {
			t.Errorf("suffix[%d] = %v, want 2.0", i, out[i])
		}
	}
}
