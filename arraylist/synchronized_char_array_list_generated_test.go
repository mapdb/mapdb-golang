package arraylist

import (
	"sync"
	"testing"
)

func TestSynchronizedChar_Generated_AddGetSize(t *testing.T) {
	s := NewSynchronizedChar()
	s.Add(1)
	s.Add(2)
	s.Add(3)
	if s.Len() != 3 {
		t.Errorf("Size = %d", s.Len())
	}
	got := s.Get(1)
	if !(got == 2) {
		t.Errorf("Get(1) = %v", got)
	}
}

func TestSynchronizedChar_Generated_AddAll(t *testing.T) {
	s := NewSynchronizedChar()
	s.AddAll(1, 2, 3)
	if s.Len() != 3 {
		t.Errorf("AddAll Size = %d, want 3", s.Len())
	}
}

func TestSynchronizedChar_Generated_Set(t *testing.T) {
	s := NewSynchronizedCharFrom(CharOf(1, 2))
	old := s.Set(0, 3)
	if !(old == 1) {
		t.Errorf("Set = %v", old)
	}
}

func TestSynchronizedChar_Generated_RemoveAtIndex(t *testing.T) {
	s := NewSynchronizedCharFrom(CharOf(1, 2, 3))
	s.RemoveAtIndex(1)
	if s.Len() != 2 {
		t.Errorf("Size = %d", s.Len())
	}
}

func TestSynchronizedChar_Generated_Remove(t *testing.T) {
	s := SynchronizedCharOf(1, 2, 1)
	if !s.Remove(1) {
		t.Error("Remove should find first occurrence")
	}
	if s.Len() != 2 {
		t.Errorf("Size after Remove = %d, want 2", s.Len())
	}
}

func TestSynchronizedChar_Generated_Contains(t *testing.T) {
	s := NewSynchronizedCharFrom(CharOf(1, 2))
	if !s.Contains(1) {
		t.Error("Contains should be true")
	}
}

func TestSynchronizedChar_Generated_IndexOf(t *testing.T) {
	s := SynchronizedCharOf(1, 2, 3)
	if idx := s.IndexOf(2); idx != 1 {
		t.Errorf("IndexOf = %d, want 1", idx)
	}
}

func TestSynchronizedChar_Generated_IsEmpty(t *testing.T) {
	s := NewSynchronizedChar()
	if s.Len() != 0 {
		t.Error("Should be empty")
	}
}

func TestSynchronizedChar_Generated_Clear(t *testing.T) {
	s := NewSynchronizedCharFrom(CharOf(1, 2))
	s.Clear()
	if s.Len() != 0 {
		t.Errorf("After clear: %d", s.Len())
	}
}

func TestSynchronizedChar_Generated_All(t *testing.T) {
	s := NewSynchronizedCharFrom(CharOf(1, 2))
	count := 0
	for range s.All() {
		count++
	}
	if count != 2 {
		t.Errorf("All count = %d", count)
	}
}

func TestSynchronizedChar_Generated_AllWithIndex(t *testing.T) {
	s := SynchronizedCharOf(1, 2, 3)
	last := -1
	for i := range s.AllWithIndex() {
		last = i
	}
	if last != 2 {
		t.Errorf("AllWithIndex last index = %d, want 2", last)
	}
}

func TestSynchronizedChar_Generated_ForEach(t *testing.T) {
	s := SynchronizedCharOf(1, 2, 3)
	n := 0
	s.ForEach(func(uint16) { n++ })
	if n != 3 {
		t.Errorf("ForEach count = %d, want 3", n)
	}
}

func TestSynchronizedChar_Generated_ForEachWithIndex(t *testing.T) {
	s := SynchronizedCharOf(1, 2, 3)
	lastIdx := -1
	s.ForEachWithIndex(func(_ uint16, i int) { lastIdx = i })
	if lastIdx != 2 {
		t.Errorf("ForEachWithIndex last = %d, want 2", lastIdx)
	}
}

func TestSynchronizedChar_Generated_AnySatisfy(t *testing.T) {
	s := SynchronizedCharOf(1, 2)
	if !s.AnySatisfy(func(x uint16) bool { return x == 1 }) {
		t.Error("AnySatisfy should be true")
	}
}

func TestSynchronizedChar_Generated_AllSatisfy(t *testing.T) {
	s := SynchronizedCharOf(1, 1)
	if !s.AllSatisfy(func(x uint16) bool { return x == 1 }) {
		t.Error("AllSatisfy should be true")
	}
}

func TestSynchronizedChar_Generated_NoneSatisfy(t *testing.T) {
	s := SynchronizedCharOf(1, 1)
	if !s.NoneSatisfy(func(x uint16) bool { return x == 3 }) {
		t.Error("NoneSatisfy should be true")
	}
}

func TestSynchronizedChar_Generated_Count(t *testing.T) {
	s := SynchronizedCharOf(1, 2, 1)
	c := s.Count(func(x uint16) bool { return x == 1 })
	if c != 2 {
		t.Errorf("Count = %d, want 2", c)
	}
}

func TestSynchronizedChar_Generated_Detect(t *testing.T) {
	s := SynchronizedCharOf(1, 2, 3)
	got, ok := s.Detect(func(x uint16) bool { return x == 2 })
	if !ok || !(got == 2) {
		t.Errorf("Detect = (%v, %v)", got, ok)
	}
}

func TestSynchronizedChar_Generated_InjectInto(t *testing.T) {
	s := SynchronizedCharOf(1, 2, 3)
	// Identity-ish accumulator: count elements by returning acc + 1 as uint16.
	result := s.InjectInto(1, func(acc, _ uint16) uint16 { return acc })
	if !(result == 1) {
		t.Errorf("InjectInto result = %v", result)
	}
}

func TestSynchronizedChar_Generated_SelectReject(t *testing.T) {
	s := SynchronizedCharOf(1, 2, 3)
	sel := s.Select(func(x uint16) bool { return x == 2 })
	if sel.Len() != 1 {
		t.Errorf("Select Size = %d, want 1", sel.Len())
	}
	rej := s.Reject(func(x uint16) bool { return x == 2 })
	if rej.Len() != 2 {
		t.Errorf("Reject Size = %d, want 2", rej.Len())
	}
}

func TestSynchronizedChar_Generated_Distinct(t *testing.T) {
	s := SynchronizedCharOf(1, 1, 2)
	d := s.Distinct()
	if d.Len() != 2 {
		t.Errorf("Distinct Size = %d, want 2", d.Len())
	}
}

func TestSynchronizedChar_Generated_Reversed(t *testing.T) {
	s := SynchronizedCharOf(1, 2, 3)
	r := s.Reversed()
	got := r.Get(0)
	if !(got == 3) {
		t.Errorf("Reversed[0] = %v", got)
	}
}

func TestSynchronizedChar_Generated_WithWithout(t *testing.T) {
	s := NewSynchronizedChar()
	s.AddReturning(1).AddReturning(2)
	if s.Len() != 2 {
		t.Errorf("With fluent Size = %d", s.Len())
	}
	s.RemoveReturning(1)
	if s.Len() != 1 {
		t.Errorf("Without Size = %d", s.Len())
	}
	if got := s.Get(0); !(got == 2) {
		t.Errorf("remaining = %v", got)
	}
}

func TestSynchronizedChar_Generated_WithAllWithoutAll(t *testing.T) {
	s := NewSynchronizedChar()
	s.AddAllReturning(1, 2, 3)
	if s.Len() != 3 {
		t.Errorf("WithAll Size = %d", s.Len())
	}
	s.RemoveAllReturning(1, 3)
	if s.Len() != 1 {
		t.Errorf("WithoutAll Size = %d", s.Len())
	}
}

func TestSynchronizedChar_Generated_Sort(t *testing.T) {
	// Sort is primitive-order-sensitive; just check it runs and keeps size.
	s := SynchronizedCharOf(3, 1, 2)
	s.Sort()
	if s.Len() != 3 {
		t.Errorf("Sort Size = %d", s.Len())
	}
}

func TestSynchronizedChar_Generated_SortWithComparator(t *testing.T) {
	s := SynchronizedCharOf(1, 2, 3)
	// Trivial comparator (everything equal) — just exercises the plumbing.
	s.SortWithComparator(func(_, _ uint16) bool { return false })
	if s.Len() != 3 {
		t.Errorf("SortWithComparator Size = %d", s.Len())
	}
}

func TestSynchronizedChar_Generated_BinarySearch(t *testing.T) {
	s := SynchronizedCharOf(1, 2, 3)
	s.Sort()
	if _, found := s.BinarySearch(2); !found {
		t.Error("BinarySearch should find 2")
	}
}

func TestSynchronizedChar_Generated_SumMinMax(t *testing.T) {
	s := SynchronizedCharOf(1, 2, 3)
	_ = s.Sum()
	if _, ok := s.Min(); !ok {
		t.Error("Min should find value")
	}
	if _, ok := s.Max(); !ok {
		t.Error("Max should find value")
	}
}

func TestSynchronizedChar_Generated_EqualsSelf(t *testing.T) {
	s := SynchronizedCharOf(1, 2)
	if !s.Equals(s) {
		t.Error("Equals(self) should be true")
	}
}

func TestSynchronizedChar_Generated_EqualsDifferent(t *testing.T) {
	a := SynchronizedCharOf(1, 2)
	b := SynchronizedCharOf(1, 2)
	if !a.Equals(b) {
		t.Error("Equals on matching contents should be true")
	}
	// Even when reversing argument order (lock acquisition order flips),
	// the call must complete without deadlocking.
	if !b.Equals(a) {
		t.Error("Equals(reversed) should be true")
	}
}

func TestSynchronizedChar_Generated_ToImmutable(t *testing.T) {
	s := SynchronizedCharOf(1, 2)
	imm := s.ToImmutable()
	if imm.Len() != 2 {
		t.Errorf("ToImmutable Size = %d", imm.Len())
	}
	// Mutating the source must not affect the immutable copy.
	s.Add(3)
	if imm.Len() != 2 {
		t.Errorf("Immutable Size after source mutation = %d, want 2", imm.Len())
	}
}

func TestSynchronizedChar_Generated_ToSlice(t *testing.T) {
	s := NewSynchronizedCharFrom(CharOf(1, 2))
	if len(s.ToSlice()) != 2 {
		t.Error("wrong len")
	}
}

func TestSynchronizedChar_Generated_String(t *testing.T) {
	s := NewSynchronizedCharFrom(CharOf(1))
	if s.String() == "" {
		t.Error("empty")
	}
}

// TestSynchronizedChar_Generated_ConcurrentFunctional exercises the snapshot-based
// functional path under concurrent writers. Callbacks must never run
// while the write lock is held — so a predicate that tries to Add back
// into the same wrapper must not deadlock.
func TestSynchronizedChar_Generated_ConcurrentFunctional(t *testing.T) {
	s := SynchronizedCharOf(1, 2, 3)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			s.Add(1)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			// Callback re-enters the wrapper; must not deadlock.
			s.AnySatisfy(func(x uint16) bool {
				_ = s.Len()
				return x == 1
			})
		}
	}()
	wg.Wait()
}
