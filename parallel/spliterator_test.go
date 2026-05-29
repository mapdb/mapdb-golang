// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package parallel

import (
	"slices"
	"testing"
)

func TestSpliterator_TryAdvanceWalksEveryElement(t *testing.T) {
	sp := NewSliceSpliterator([]int{1, 2, 3})
	var seen []int
	for sp.TryAdvance(func(v int) { seen = append(seen, v) }) {
	}
	if !slices.Equal(seen, []int{1, 2, 3}) {
		t.Fatalf("expected [1 2 3], got %v", seen)
	}
	// Exhausted: a further advance is a no-op returning false.
	if sp.TryAdvance(func(int) { t.Fatal("should not be called") }) {
		t.Fatal("expected false from exhausted spliterator")
	}
}

func TestSpliterator_ForEachRemainingVisitsAll(t *testing.T) {
	sp := NewSliceSpliterator([]int{10, 20, 30, 40})
	sum := 0
	sp.ForEachRemaining(func(v int) { sum += v })
	if sum != 100 {
		t.Fatalf("expected 100, got %d", sum)
	}
	if sp.EstimateSize() != 0 {
		t.Fatalf("expected exhausted, got size %d", sp.EstimateSize())
	}
}

func TestSpliterator_TrySplitReturnsPrefixKeepsSuffix(t *testing.T) {
	sp := NewSliceSpliterator([]int{1, 2, 3, 4, 5})
	prefix, ok := sp.TrySplit()
	if !ok {
		t.Fatal("expected splittable")
	}
	// Java convention: prefix covers the front, receiver keeps the back.
	if got := prefix.(*SliceSpliterator[int]).Remainder(); !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("prefix: expected [1 2], got %v", got)
	}
	if got := sp.Remainder(); !slices.Equal(got, []int{3, 4, 5}) {
		t.Fatalf("suffix: expected [3 4 5], got %v", got)
	}
}

func TestSpliterator_SplitRecursivelyCoversEveryElementOnce(t *testing.T) {
	data := makeRange(1000)
	var collected []int
	work := []Spliterator[int]{NewSliceSpliterator(data)}
	for len(work) > 0 {
		sp := work[len(work)-1]
		work = work[:len(work)-1]
		if prefix, ok := sp.TrySplit(); ok {
			work = append(work, prefix, sp)
		} else {
			sp.ForEachRemaining(func(v int) { collected = append(collected, v) })
		}
	}
	slices.Sort(collected)
	if !slices.Equal(collected, data) {
		t.Fatal("recursive split did not cover every element exactly once")
	}
}

func TestSpliterator_SingletonsAndEmptiesDoNotSplit(t *testing.T) {
	if _, ok := NewSliceSpliterator([]int{42}).TrySplit(); ok {
		t.Fatal("singleton should not split")
	}
	if _, ok := NewSliceSpliterator([]int{}).TrySplit(); ok {
		t.Fatal("empty should not split")
	}
}

func TestSpliterator_CharacteristicsAndExactSize(t *testing.T) {
	sp := NewSliceSpliterator([]int{1, 2, 3})
	if !HasCharacteristics[int](sp, Sized) {
		t.Fatal("expected Sized")
	}
	if !HasCharacteristics[int](sp, Ordered|Subsized) {
		t.Fatal("expected Ordered|Subsized")
	}
	if HasCharacteristics[int](sp, Sorted) {
		t.Fatal("did not expect Sorted")
	}
	if n, ok := ExactSize[int](sp); !ok || n != 3 {
		t.Fatalf("expected exact size 3, got %d ok=%v", n, ok)
	}
	if sp.EstimateSize() != 3 {
		t.Fatalf("expected estimate 3, got %d", sp.EstimateSize())
	}
}

func TestSpliterator_CharacteristicBitValuesMatchJava(t *testing.T) {
	// Bit values lifted directly from java.util.Spliterator.
	cases := []struct {
		got  uint
		want uint
	}{
		{Distinct, 0x01},
		{Sorted, 0x04},
		{Ordered, 0x10},
		{Sized, 0x40},
		{NonNull, 0x100},
		{Immutable, 0x400},
		{Concurrent, 0x1000},
		{Subsized, 0x4000},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Fatalf("characteristic mismatch: got 0x%x want 0x%x", c.got, c.want)
		}
	}
}
