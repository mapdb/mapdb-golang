package stack

import "testing"

func TestInt32WithCapacity_BulkPush(t *testing.T) {
	s := NewInt32WithCapacity(100)
	if s.Len() != 0 {
		t.Fatalf("new with-capacity stack not empty")
	}
	for i := int32(0); i < 100; i++ {
		s.Push(i)
	}
	if s.Len() != 100 {
		t.Fatalf("len %d want 100", s.Len())
	}
	if v, ok := s.Pop(); !ok || v != 99 {
		t.Fatalf("top %v %v want 99", v, ok)
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
