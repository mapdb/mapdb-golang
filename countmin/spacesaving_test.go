// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package countmin

import (
	"math"
	"reflect"
	"testing"
)

func tri(item int32, count, err uint64) SSEntry {
	return SSEntry{Item: item, Count: count, Error: err}
}

func TestAdmitUnderCapacityNoEviction(t *testing.T) {
	s := NewSpaceSaving(3)
	s.AddOne(7)
	s.AddOne(7)
	s.AddOne(-1)
	if s.Size() != 2 {
		t.Fatalf("size = %d, want 2", s.Size())
	}
	if s.Count(7) != 2 || s.Error(7) != 0 || s.Count(-1) != 1 {
		t.Fatal("unexpected counts")
	}
	// 7 (count 2) before -1 (count 1); both error 0.
	want := []SSEntry{tri(7, 2, 0), tri(-1, 1, 0)}
	if got := s.MonitoredSet(); !reflect.DeepEqual(got, want) {
		t.Fatalf("monitored_set = %v, want %v", got, want)
	}
	if got := s.TopK(1); !reflect.DeepEqual(got, []SSEntry{tri(7, 2, 0)}) {
		t.Fatalf("top_k(1) = %v", got)
	}
}

func TestEvictMinTiebreakSmallerSignedItem(t *testing.T) {
	// capacity 2: add(1), add(2) -> both count 1 (full); add(3) evicts the
	// min-count item; tie -> smallest signed item = 1 evicted. 3 admitted with
	// count = 1 + 1 = 2, error = 1.
	s := NewSpaceSaving(2)
	s.AddOne(1)
	s.AddOne(2)
	s.AddOne(3)
	want := []SSEntry{tri(3, 2, 1), tri(2, 1, 0)}
	if got := s.MonitoredSet(); !reflect.DeepEqual(got, want) {
		t.Fatalf("monitored_set = %v, want %v", got, want)
	}
	if s.Count(1) != 0 || s.IsMonitored(1) {
		t.Fatal("1 must be evicted")
	}
	if s.Count(3) != 2 || s.Error(3) != 1 {
		t.Fatal("3 count/error wrong")
	}
}

func TestEvictTiebreakNegativeBeatsPositive(t *testing.T) {
	// Monitored -5 (count 1) and 2 (count 1); add a new item -> -5 < 2 (SIGNED),
	// so -5 is evicted (an unsigned comparison would evict 2).
	s := NewSpaceSaving(2)
	s.AddOne(-5)
	s.AddOne(2)
	s.AddOne(9)
	if s.IsMonitored(-5) {
		t.Fatal("smaller signed item -5 must be evicted")
	}
	if !s.IsMonitored(2) || !s.IsMonitored(9) {
		t.Fatal("2 and 9 must be monitored")
	}
	if s.Count(9) != 2 || s.Error(9) != 1 {
		t.Fatal("9 count/error wrong")
	}
}

func TestAlreadyMonitoredErrorNeverChanges(t *testing.T) {
	s := NewSpaceSaving(2)
	s.AddOne(1)
	s.AddOne(2)
	s.AddOne(3) // evicts 1; 3 -> count 2, error 1
	if s.Error(3) != 1 {
		t.Fatal("3 error should be 1")
	}
	s.Add(3, 100) // re-add of a monitored item: error unchanged.
	if s.Count(3) != 102 || s.Error(3) != 1 {
		t.Fatalf("count=%d error=%d, want 102/1", s.Count(3), s.Error(3))
	}
}

func TestAdmittedWithRoomHasZeroError(t *testing.T) {
	s := NewSpaceSaving(5)
	s.Add(7, 9)
	if s.Error(7) != 0 || s.Count(7) != 9 {
		t.Fatal("admitted-with-room must have error 0")
	}
}

func TestCountZeroIsNoop(t *testing.T) {
	s := NewSpaceSaving(1)
	s.AddOne(1)
	s.Add(2, 0) // must NOT evict 1.
	if !s.IsMonitored(1) || s.IsMonitored(2) || s.Size() != 1 {
		t.Fatal("count=0 add must be a full no-op")
	}
}

func TestEmptySummary(t *testing.T) {
	s := NewSpaceSaving(3)
	if s.Size() != 0 || s.Capacity() != 3 {
		t.Fatal("empty size/capacity wrong")
	}
	if len(s.MonitoredSet()) != 0 {
		t.Fatal("empty monitored_set must be empty")
	}
	if s.Count(7) != 0 || s.Error(7) != 0 {
		t.Fatal("unmonitored count/error must be 0")
	}
	if len(s.TopK(3)) != 0 {
		t.Fatal("empty top_k must be empty")
	}
}

func TestTopKCanonicalOrderAndBounds(t *testing.T) {
	s := NewSpaceSaving(10)
	s.Add(1, 5)
	s.Add(2, 3)
	s.Add(3, 3) // tie at count 3 -> 2 before 3 (signed asc)
	s.Add(4, 1)
	full := []SSEntry{tri(1, 5, 0), tri(2, 3, 0), tri(3, 3, 0), tri(4, 1, 0)}
	if got := s.MonitoredSet(); !reflect.DeepEqual(got, full) {
		t.Fatalf("monitored_set = %v, want %v", got, full)
	}
	if got := s.TopK(1); !reflect.DeepEqual(got, []SSEntry{tri(1, 5, 0)}) {
		t.Fatalf("top_k(1) = %v", got)
	}
	if got := s.TopK(2); !reflect.DeepEqual(got, []SSEntry{tri(1, 5, 0), tri(2, 3, 0)}) {
		t.Fatalf("top_k(2) = %v", got)
	}
	if got := s.TopK(4); !reflect.DeepEqual(got, full) {
		t.Fatalf("top_k(4) = %v", got)
	}
	if got := s.TopK(99); !reflect.DeepEqual(got, full) { // k > size -> all, no padding
		t.Fatalf("top_k(99) = %v", got)
	}
	if got := s.TopK(0); len(got) != 0 {
		t.Fatalf("top_k(0) = %v, want empty", got)
	}
}

func TestCountUnmonitoredIsZero(t *testing.T) {
	s := NewSpaceSaving(1)
	s.AddOne(1)
	s.AddOne(2) // evicts 1
	if s.Count(1) != 0 || s.IsMonitored(1) {
		t.Fatal("evicted item count must be 0")
	}
}

func TestSpaceSavingOverflowSaturates(t *testing.T) {
	s := NewSpaceSaving(1)
	s.Add(7, math.MaxUint64)
	s.Add(7, math.MaxUint64)
	if s.Count(7) != math.MaxUint64 {
		t.Fatalf("count = %d, want MaxUint64", s.Count(7))
	}
}

func TestOrderDependence(t *testing.T) {
	a := NewSpaceSaving(2)
	for _, it := range []int32{1, 1, 2, 3} {
		a.AddOne(it)
	}
	// A: 1->1, 1->2, 2 admitted->1, 3 evicts min(count): 2 has count1 vs 1 has
	// count2 -> victim 2 (count1); 3-> count2 error1. Set {1:2, 3:2e1}.
	want := []SSEntry{tri(1, 2, 0), tri(3, 2, 1)}
	if got := a.MonitoredSet(); !reflect.DeepEqual(got, want) {
		t.Fatalf("monitored_set = %v, want %v", got, want)
	}
}

func TestErrorFloorUnchangedAcrossEvictions(t *testing.T) {
	s := NewSpaceSaving(2)
	s.AddOne(1) // {1:1}
	s.AddOne(2) // {1:1, 2:1} full
	s.AddOne(3) // evict 1 (tie, smallest signed): 3-> count2 error1
	if s.Error(3) != 1 {
		t.Fatal("3 error should be 1")
	}
	s.AddOne(3) // monitored re-add: count3 error1 (UNCHANGED)
	if s.Count(3) != 3 || s.Error(3) != 1 {
		t.Fatalf("count=%d error=%d, want 3/1", s.Count(3), s.Error(3))
	}
	s.AddOne(4) // full {2:1, 3:3}; evict 2 (count1): 4-> count2 error1
	if s.Error(4) != 1 || s.Count(4) != 2 {
		t.Fatalf("4 count=%d error=%d, want 2/1", s.Count(4), s.Error(4))
	}
	if s.Error(3) != 1 {
		t.Fatal("3 error must stay unchanged")
	}
}

func TestMZeroTraps(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("m=0 must panic")
		}
	}()
	NewSpaceSaving(0)
}
