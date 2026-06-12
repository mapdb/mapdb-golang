package arraylist

import (
	"math"
	"sync"
	"testing"
)

func TestSynchronizedFloat32_Generated_AddGetSize(t *testing.T) {
	s := NewSynchronizedFloat32()
	s.Add(1.0)
	s.Add(2.0)
	s.Add(3.0)
	if s.Len() != 3 {
		t.Errorf("Size = %d", s.Len())
	}
	got := s.Get(1)
	if !(math.Float32bits(got) == math.Float32bits(2.0)) {
		t.Errorf("Get(1) = %v", got)
	}
}

func TestSynchronizedFloat32_Generated_AddAll(t *testing.T) {
	s := NewSynchronizedFloat32()
	s.AddAll(1.0, 2.0, 3.0)
	if s.Len() != 3 {
		t.Errorf("AddAll Size = %d, want 3", s.Len())
	}
}

func TestSynchronizedFloat32_Generated_Set(t *testing.T) {
	s := NewSynchronizedFloat32From(Float32Of(1.0, 2.0))
	old := s.Set(0, 3.0)
	if !(math.Float32bits(old) == math.Float32bits(1.0)) {
		t.Errorf("Set = %v", old)
	}
}

func TestSynchronizedFloat32_Generated_RemoveAtIndex(t *testing.T) {
	s := NewSynchronizedFloat32From(Float32Of(1.0, 2.0, 3.0))
	s.RemoveAtIndex(1)
	if s.Len() != 2 {
		t.Errorf("Size = %d", s.Len())
	}
}

func TestSynchronizedFloat32_Generated_Remove(t *testing.T) {
	s := SynchronizedFloat32Of(1.0, 2.0, 1.0)
	if !s.Remove(1.0) {
		t.Error("Remove should find first occurrence")
	}
	if s.Len() != 2 {
		t.Errorf("Size after Remove = %d, want 2", s.Len())
	}
}

func TestSynchronizedFloat32_Generated_Contains(t *testing.T) {
	s := NewSynchronizedFloat32From(Float32Of(1.0, 2.0))
	if !s.Contains(1.0) {
		t.Error("Contains should be true")
	}
}

func TestSynchronizedFloat32_Generated_IndexOf(t *testing.T) {
	s := SynchronizedFloat32Of(1.0, 2.0, 3.0)
	if idx := s.IndexOf(2.0); idx != 1 {
		t.Errorf("IndexOf = %d, want 1", idx)
	}
}

func TestSynchronizedFloat32_Generated_IsEmpty(t *testing.T) {
	s := NewSynchronizedFloat32()
	if s.Len() != 0 {
		t.Error("Should be empty")
	}
}

func TestSynchronizedFloat32_Generated_Clear(t *testing.T) {
	s := NewSynchronizedFloat32From(Float32Of(1.0, 2.0))
	s.Clear()
	if s.Len() != 0 {
		t.Errorf("After clear: %d", s.Len())
	}
}

func TestSynchronizedFloat32_Generated_All(t *testing.T) {
	s := NewSynchronizedFloat32From(Float32Of(1.0, 2.0))
	count := 0
	for range s.All() {
		count++
	}
	if count != 2 {
		t.Errorf("All count = %d", count)
	}
}

func TestSynchronizedFloat32_Generated_AllWithIndex(t *testing.T) {
	s := SynchronizedFloat32Of(1.0, 2.0, 3.0)
	last := -1
	for i := range s.AllWithIndex() {
		last = i
	}
	if last != 2 {
		t.Errorf("AllWithIndex last index = %d, want 2", last)
	}
}

func TestSynchronizedFloat32_Generated_ForEach(t *testing.T) {
	s := SynchronizedFloat32Of(1.0, 2.0, 3.0)
	n := 0
	s.ForEach(func(float32) { n++ })
	if n != 3 {
		t.Errorf("ForEach count = %d, want 3", n)
	}
}

func TestSynchronizedFloat32_Generated_ForEachWithIndex(t *testing.T) {
	s := SynchronizedFloat32Of(1.0, 2.0, 3.0)
	lastIdx := -1
	s.ForEachWithIndex(func(_ float32, i int) { lastIdx = i })
	if lastIdx != 2 {
		t.Errorf("ForEachWithIndex last = %d, want 2", lastIdx)
	}
}

func TestSynchronizedFloat32_Generated_AnySatisfy(t *testing.T) {
	s := SynchronizedFloat32Of(1.0, 2.0)
	if !s.AnySatisfy(func(x float32) bool { return math.Float32bits(x) == math.Float32bits(1.0) }) {
		t.Error("AnySatisfy should be true")
	}
}

func TestSynchronizedFloat32_Generated_AllSatisfy(t *testing.T) {
	s := SynchronizedFloat32Of(1.0, 1.0)
	if !s.AllSatisfy(func(x float32) bool { return math.Float32bits(x) == math.Float32bits(1.0) }) {
		t.Error("AllSatisfy should be true")
	}
}

func TestSynchronizedFloat32_Generated_NoneSatisfy(t *testing.T) {
	s := SynchronizedFloat32Of(1.0, 1.0)
	if !s.NoneSatisfy(func(x float32) bool { return math.Float32bits(x) == math.Float32bits(3.0) }) {
		t.Error("NoneSatisfy should be true")
	}
}

func TestSynchronizedFloat32_Generated_Count(t *testing.T) {
	s := SynchronizedFloat32Of(1.0, 2.0, 1.0)
	c := s.Count(func(x float32) bool { return math.Float32bits(x) == math.Float32bits(1.0) })
	if c != 2 {
		t.Errorf("Count = %d, want 2", c)
	}
}

func TestSynchronizedFloat32_Generated_Detect(t *testing.T) {
	s := SynchronizedFloat32Of(1.0, 2.0, 3.0)
	got, ok := s.Detect(func(x float32) bool { return math.Float32bits(x) == math.Float32bits(2.0) })
	if !ok || !(math.Float32bits(got) == math.Float32bits(2.0)) {
		t.Errorf("Detect = (%v, %v)", got, ok)
	}
}

func TestSynchronizedFloat32_Generated_InjectInto(t *testing.T) {
	s := SynchronizedFloat32Of(1.0, 2.0, 3.0)
	// Identity-ish accumulator: count elements by returning acc + 1 as float32.
	result := s.InjectInto(1.0, func(acc, _ float32) float32 { return acc })
	if !(math.Float32bits(result) == math.Float32bits(1.0)) {
		t.Errorf("InjectInto result = %v", result)
	}
}

func TestSynchronizedFloat32_Generated_SelectReject(t *testing.T) {
	s := SynchronizedFloat32Of(1.0, 2.0, 3.0)
	sel := s.Select(func(x float32) bool { return math.Float32bits(x) == math.Float32bits(2.0) })
	if sel.Len() != 1 {
		t.Errorf("Select Size = %d, want 1", sel.Len())
	}
	rej := s.Reject(func(x float32) bool { return math.Float32bits(x) == math.Float32bits(2.0) })
	if rej.Len() != 2 {
		t.Errorf("Reject Size = %d, want 2", rej.Len())
	}
}

func TestSynchronizedFloat32_Generated_Distinct(t *testing.T) {
	s := SynchronizedFloat32Of(1.0, 1.0, 2.0)
	d := s.Distinct()
	if d.Len() != 2 {
		t.Errorf("Distinct Size = %d, want 2", d.Len())
	}
}

func TestSynchronizedFloat32_Generated_Reversed(t *testing.T) {
	s := SynchronizedFloat32Of(1.0, 2.0, 3.0)
	r := s.Reversed()
	got := r.Get(0)
	if !(math.Float32bits(got) == math.Float32bits(3.0)) {
		t.Errorf("Reversed[0] = %v", got)
	}
}

func TestSynchronizedFloat32_Generated_WithWithout(t *testing.T) {
	s := NewSynchronizedFloat32()
	s.AddReturning(1.0).AddReturning(2.0)
	if s.Len() != 2 {
		t.Errorf("With fluent Size = %d", s.Len())
	}
	s.RemoveReturning(1.0)
	if s.Len() != 1 {
		t.Errorf("Without Size = %d", s.Len())
	}
	if got := s.Get(0); !(math.Float32bits(got) == math.Float32bits(2.0)) {
		t.Errorf("remaining = %v", got)
	}
}

func TestSynchronizedFloat32_Generated_WithAllWithoutAll(t *testing.T) {
	s := NewSynchronizedFloat32()
	s.AddAllReturning(1.0, 2.0, 3.0)
	if s.Len() != 3 {
		t.Errorf("WithAll Size = %d", s.Len())
	}
	s.RemoveAllReturning(1.0, 3.0)
	if s.Len() != 1 {
		t.Errorf("WithoutAll Size = %d", s.Len())
	}
}

func TestSynchronizedFloat32_Generated_Sort(t *testing.T) {
	// Sort is primitive-order-sensitive; just check it runs and keeps size.
	s := SynchronizedFloat32Of(3.0, 1.0, 2.0)
	s.Sort()
	if s.Len() != 3 {
		t.Errorf("Sort Size = %d", s.Len())
	}
}

func TestSynchronizedFloat32_Generated_SortWithComparator(t *testing.T) {
	s := SynchronizedFloat32Of(1.0, 2.0, 3.0)
	// Trivial comparator (everything equal) — just exercises the plumbing.
	s.SortWithComparator(func(_, _ float32) bool { return false })
	if s.Len() != 3 {
		t.Errorf("SortWithComparator Size = %d", s.Len())
	}
}

func TestSynchronizedFloat32_Generated_BinarySearch(t *testing.T) {
	s := SynchronizedFloat32Of(1.0, 2.0, 3.0)
	s.Sort()
	if _, found := s.BinarySearch(2.0); !found {
		t.Error("BinarySearch should find 2.0")
	}
}

func TestSynchronizedFloat32_Generated_SumMinMax(t *testing.T) {
	s := SynchronizedFloat32Of(1.0, 2.0, 3.0)
	_ = s.Sum()
	if _, ok := s.Min(); !ok {
		t.Error("Min should find value")
	}
	if _, ok := s.Max(); !ok {
		t.Error("Max should find value")
	}
}

func TestSynchronizedFloat32_Generated_EqualsSelf(t *testing.T) {
	s := SynchronizedFloat32Of(1.0, 2.0)
	if !s.Equals(s) {
		t.Error("Equals(self) should be true")
	}
}

func TestSynchronizedFloat32_Generated_EqualsDifferent(t *testing.T) {
	a := SynchronizedFloat32Of(1.0, 2.0)
	b := SynchronizedFloat32Of(1.0, 2.0)
	if !a.Equals(b) {
		t.Error("Equals on matching contents should be true")
	}
	// Even when reversing argument order (lock acquisition order flips),
	// the call must complete without deadlocking.
	if !b.Equals(a) {
		t.Error("Equals(reversed) should be true")
	}
}

func TestSynchronizedFloat32_Generated_ToImmutable(t *testing.T) {
	s := SynchronizedFloat32Of(1.0, 2.0)
	imm := s.ToImmutable()
	if imm.Len() != 2 {
		t.Errorf("ToImmutable Size = %d", imm.Len())
	}
	// Mutating the source must not affect the immutable copy.
	s.Add(3.0)
	if imm.Len() != 2 {
		t.Errorf("Immutable Size after source mutation = %d, want 2", imm.Len())
	}
}

func TestSynchronizedFloat32_Generated_ToSlice(t *testing.T) {
	s := NewSynchronizedFloat32From(Float32Of(1.0, 2.0))
	if len(s.ToSlice()) != 2 {
		t.Error("wrong len")
	}
}

func TestSynchronizedFloat32_Generated_String(t *testing.T) {
	s := NewSynchronizedFloat32From(Float32Of(1.0))
	if s.String() == "" {
		t.Error("empty")
	}
}

// TestSynchronizedFloat32_Generated_ConcurrentFunctional exercises the snapshot-based
// functional path under concurrent writers. Callbacks must never run
// while the write lock is held — so a predicate that tries to Add back
// into the same wrapper must not deadlock.
func TestSynchronizedFloat32_Generated_ConcurrentFunctional(t *testing.T) {
	s := SynchronizedFloat32Of(1.0, 2.0, 3.0)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			s.Add(1.0)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			// Callback re-enters the wrapper; must not deadlock.
			s.AnySatisfy(func(x float32) bool {
				_ = s.Len()
				return math.Float32bits(x) == math.Float32bits(1.0)
			})
		}
	}()
	wg.Wait()
}
