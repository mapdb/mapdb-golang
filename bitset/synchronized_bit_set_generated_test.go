
package bitset

import (
	"sync"
	"testing"
)

func TestSynchronizedBitSet_Generated_SetGetClear(t *testing.T) {
	b := NewSynchronizedBitSet()
	b.Set(3)
	b.Set(65)
	if !b.Get(3) {
		t.Error("Get(3) should be true")
	}
	if !b.Get(65) {
		t.Error("Get(65) should be true")
	}
	if b.Get(4) {
		t.Error("Get(4) should be false")
	}
	b.Clear(3)
	if b.Get(3) {
		t.Error("Get(3) should be false after Clear")
	}
}

func TestSynchronizedBitSet_Generated_FlipCardinality(t *testing.T) {
	b := NewSynchronizedBitSet()
	b.Flip(0)
	b.Flip(1)
	b.Flip(2)
	if b.Cardinality() != 3 {
		t.Errorf("Cardinality = %d", b.Cardinality())
	}
	b.Flip(1)
	if b.Cardinality() != 2 {
		t.Errorf("Cardinality after flip = %d", b.Cardinality())
	}
}

func TestSynchronizedBitSet_Generated_IsEmptyClearAll(t *testing.T) {
	b := NewSynchronizedBitSet()
	if !b.IsEmpty() {
		t.Error("Should be empty")
	}
	b.Set(5)
	if b.IsEmpty() {
		t.Error("Should not be empty")
	}
	b.ClearAll()
	if !b.IsEmpty() {
		t.Error("Should be empty after ClearAll")
	}
}

func TestSynchronizedBitSet_Generated_Bitwise(t *testing.T) {
	a := NewSynchronizedBitSet()
	a.Set(1)
	a.Set(2)
	c := NewSynchronizedBitSet()
	c.Set(2)
	c.Set(3)
	if !a.Intersects(c) {
		t.Error("a should intersect c at bit 2")
	}
	d := NewSynchronizedBitSet()
	d.Set(100)
	if a.Intersects(d) {
		t.Error("a should not intersect d")
	}
}

func TestSynchronizedBitSet_Generated_AndOrXor(t *testing.T) {
	a := NewSynchronizedBitSet()
	a.Set(1)
	a.Set(2)
	a.Set(3)
	c := NewSynchronizedBitSet()
	c.Set(2)
	c.Set(3)
	c.Set(4)
	a.AndInPlace(c)
	if a.Cardinality() != 2 {
		t.Errorf("AndInPlace cardinality = %d", a.Cardinality())
	}

	e := NewSynchronizedBitSet()
	e.Set(1)
	f := NewSynchronizedBitSet()
	f.Set(2)
	e.OrInPlace(f)
	if e.Cardinality() != 2 {
		t.Errorf("OrInPlace cardinality = %d", e.Cardinality())
	}

	g := NewSynchronizedBitSet()
	g.Set(1)
	g.Set(2)
	h := NewSynchronizedBitSet()
	h.Set(2)
	h.Set(3)
	g.XorInPlace(h)
	if g.Cardinality() != 2 {
		t.Errorf("XorInPlace cardinality = %d", g.Cardinality())
	}
}

func TestSynchronizedBitSet_Generated_ToSliceEquals(t *testing.T) {
	a := NewSynchronizedBitSet()
	a.Set(0)
	a.Set(5)
	a.Set(10)
	if len(a.ToSlice()) != 3 {
		t.Errorf("ToSlice len = %d", len(a.ToSlice()))
	}
	c := NewSynchronizedBitSet()
	c.Set(0)
	c.Set(5)
	c.Set(10)
	if !a.Equals(c) {
		t.Error("a should equal c")
	}
}

func TestSynchronizedBitSet_Generated_NextSetBit(t *testing.T) {
	b := NewSynchronizedBitSet()
	b.Set(10)
	b.Set(100)
	if b.NextSetBit(0) != 10 {
		t.Errorf("NextSetBit(0) = %d", b.NextSetBit(0))
	}
	if b.NextSetBit(11) != 100 {
		t.Errorf("NextSetBit(11) = %d", b.NextSetBit(11))
	}
}

func TestSynchronizedBitSet_Generated_ConcurrentAccess(t *testing.T) {
	b := NewSynchronizedBitSet()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(base int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				b.Set(base*100 + j)
				_ = b.Get(base*100 + j)
				_ = b.Cardinality()
			}
		}(i)
	}
	wg.Wait()
	if b.Cardinality() != 800 {
		t.Errorf("Cardinality = %d, want 800", b.Cardinality())
	}
}

func TestSynchronizedBitSet_Generated_String(t *testing.T) {
	b := NewSynchronizedBitSet()
	b.Set(1)
	if b.String() == "" {
		t.Error("empty string")
	}
}
