package bag

import (
	"math"
	"testing"
)

// These tests pin the float bit-pattern identity contract for hash bags:
// NaN must be findable, +0.0 and -0.0 must stay distinct, and Size() must not
// count unreachable entries. They fail against the previous map[float32]int /
// map[float64]int backing (Go float map equality makes every NaN a fresh
// unreachable key and collapses signed zero).

func TestHashFloat32NaNFindable(t *testing.T) {
	b := NewHashFloat32()
	nan := float32(math.NaN())
	b.Add(nan)
	b.Add(nan)
	if !b.Contains(nan) {
		t.Fatal("Contains(NaN) = false, want true")
	}
	if got := b.OccurrencesOf(nan); got != 2 {
		t.Fatalf("OccurrencesOf(NaN) = %d, want 2", got)
	}
	if got := b.Len(); got != 2 {
		t.Fatalf("Size() = %d, want 2 (no unreachable NaN entries)", got)
	}
	if got := b.SizeDistinct(); got != 1 {
		t.Fatalf("SizeDistinct() = %d, want 1", got)
	}
}

func TestHashFloat32SignedZeroDistinct(t *testing.T) {
	b := NewHashFloat32()
	pz := float32(0)
	nz := float32(math.Copysign(0, -1))
	b.Add(pz)
	b.AddOccurrences(nz, 3)
	if got := b.OccurrencesOf(pz); got != 1 {
		t.Fatalf("OccurrencesOf(+0) = %d, want 1", got)
	}
	if got := b.OccurrencesOf(nz); got != 3 {
		t.Fatalf("OccurrencesOf(-0) = %d, want 3", got)
	}
	if got := b.SizeDistinct(); got != 2 {
		t.Fatalf("SizeDistinct() = %d, want 2 (+0 and -0 distinct)", got)
	}
	if got := b.Len(); got != 4 {
		t.Fatalf("Size() = %d, want 4", got)
	}
}

func TestHashFloat32SizeAfterRemove(t *testing.T) {
	b := NewHashFloat32()
	nan := float32(math.NaN())
	b.Add(nan)
	if !b.Remove(nan) {
		t.Fatal("Remove(NaN) = false, want true")
	}
	if got := b.Len(); got != 0 {
		t.Fatalf("Size() after remove = %d, want 0", got)
	}
	if b.Contains(nan) {
		t.Fatal("Contains(NaN) after remove = true, want false")
	}
}

func TestImmutableHashFloat32DelegatesIdentity(t *testing.T) {
	nan := float32(math.NaN())
	nz := float32(math.Copysign(0, -1))
	b := NewHashFloat32()
	b.Add(nan)
	b.Add(float32(0))
	b.Add(nz)
	im := b.ToImmutable()
	if !im.Contains(nan) {
		t.Fatal("immutable Contains(NaN) = false")
	}
	if im.SizeDistinct() != 3 {
		t.Fatalf("immutable SizeDistinct() = %d, want 3", im.SizeDistinct())
	}
	if im.Len() != 3 {
		t.Fatalf("immutable Size() = %d, want 3", im.Len())
	}
}

func TestHashFloat64NaNFindable(t *testing.T) {
	b := NewHashFloat64()
	nan := math.NaN()
	b.Add(nan)
	b.Add(nan)
	if !b.Contains(nan) {
		t.Fatal("Contains(NaN) = false, want true")
	}
	if got := b.OccurrencesOf(nan); got != 2 {
		t.Fatalf("OccurrencesOf(NaN) = %d, want 2", got)
	}
	if got := b.Len(); got != 2 {
		t.Fatalf("Size() = %d, want 2", got)
	}
}

func TestHashFloat64SignedZeroDistinct(t *testing.T) {
	b := NewHashFloat64()
	pz := 0.0
	nz := math.Copysign(0, -1)
	b.Add(pz)
	b.Add(nz)
	if b.SizeDistinct() != 2 {
		t.Fatalf("SizeDistinct() = %d, want 2", b.SizeDistinct())
	}
	if b.OccurrencesOf(pz) != 1 || b.OccurrencesOf(nz) != 1 {
		t.Fatalf("signed zero collapsed: +0=%d -0=%d", b.OccurrencesOf(pz), b.OccurrencesOf(nz))
	}
}
