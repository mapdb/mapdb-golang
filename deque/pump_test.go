package deque

import "testing"

func TestInt32WithCapacity_BulkAdd(t *testing.T) {
	d := NewInt32WithCapacity(100)
	if d.Len() != 0 {
		t.Fatalf("new with-capacity deque not empty")
	}
	for i := int32(0); i < 100; i++ {
		d.AddLast(i)
	}
	if d.Len() != 100 {
		t.Fatalf("len %d want 100", d.Len())
	}
	// backing buffer must already hold 100 (power-of-two >= 100 => 128); no resize
	if cap := len(d.items); cap < 100 {
		t.Fatalf("backing buffer %d too small", cap)
	}
	if v, ok := d.PeekFirst(); !ok || v != 0 {
		t.Fatalf("front %v %v", v, ok)
	}
}

func TestInt32WithCapacity_NegativePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	NewInt32WithCapacity(-1)
}
