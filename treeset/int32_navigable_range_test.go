package treeset

import (
	"slices"
	"testing"

	"github.com/mapdb/mapdb-golang/rangev"
)

func TestInt32_NavStrictness(t *testing.T) {
	s := Int32Of(10, 20, 30)
	if v, ok := s.Floor(25); !ok || v != 20 {
		t.Errorf("Floor(25) = (%d,%v), want 20", v, ok)
	}
	if v, ok := s.Ceiling(25); !ok || v != 30 {
		t.Errorf("Ceiling(25) = (%d,%v), want 30", v, ok)
	}
	if v, ok := s.Floor(10); !ok || v != 10 {
		t.Errorf("Floor(10) = (%d,%v), want 10", v, ok)
	}
	if _, ok := s.Lower(10); ok {
		t.Error("Lower(10) should be false")
	}
	if _, ok := s.Higher(30); ok {
		t.Error("Higher(30) should be false")
	}
	if v, ok := s.Ceiling(5); !ok || v != 10 {
		t.Errorf("Ceiling(5) = (%d,%v), want 10", v, ok)
	}
	if v, ok := s.First(); !ok || v != 10 {
		t.Errorf("First = (%d,%v), want 10", v, ok)
	}
	if v, ok := s.Last(); !ok || v != 30 {
		t.Errorf("Last = (%d,%v), want 30", v, ok)
	}
}

func TestInt32_NavEmpty(t *testing.T) {
	s := NewInt32()
	if _, ok := s.Floor(5); ok {
		t.Error("Floor on empty should be false")
	}
	if _, ok := s.Ceiling(5); ok {
		t.Error("Ceiling on empty should be false")
	}
	if _, ok := s.Lower(5); ok {
		t.Error("Lower on empty should be false")
	}
	if _, ok := s.Higher(5); ok {
		t.Error("Higher on empty should be false")
	}
	if _, ok := s.First(); ok {
		t.Error("First on empty should be false")
	}
	if _, ok := s.Last(); ok {
		t.Error("Last on empty should be false")
	}
}

func TestInt32_NavSignedExtremes(t *testing.T) {
	const minI32 = int32(-2147483648)
	const maxI32 = int32(2147483647)
	s := Int32Of(minI32, -1, 0, 1, maxI32)
	if v, ok := s.Floor(minI32); !ok || v != minI32 {
		t.Errorf("Floor(MIN) = (%d,%v)", v, ok)
	}
	if _, ok := s.Lower(minI32); ok {
		t.Error("Lower(MIN) should be false")
	}
	if v, ok := s.Higher(-1); !ok || v != 0 {
		t.Errorf("Higher(-1) = (%d,%v), want 0", v, ok)
	}
	if v, ok := s.Ceiling(maxI32); !ok || v != maxI32 {
		t.Errorf("Ceiling(MAX) = (%d,%v)", v, ok)
	}
	if _, ok := s.Higher(maxI32); ok {
		t.Error("Higher(MAX) should be false")
	}
	if got := s.Descending(); !slices.Equal(got, []int32{maxI32, 1, 0, -1, minI32}) {
		t.Errorf("Descending = %v", got)
	}
}

func TestInt32_PollFirstLastThenEmpty(t *testing.T) {
	s := Int32Of(10, 20, 30)
	if v, ok := s.PollFirst(); !ok || v != 10 {
		t.Errorf("PollFirst = (%d,%v), want 10", v, ok)
	}
	if v, ok := s.PollLast(); !ok || v != 30 {
		t.Errorf("PollLast = (%d,%v), want 30", v, ok)
	}
	if v, ok := s.PollFirst(); !ok || v != 20 {
		t.Errorf("PollFirst = (%d,%v), want 20", v, ok)
	}
	if _, ok := s.PollFirst(); ok {
		t.Error("PollFirst on empty should be false")
	}
	if _, ok := s.PollLast(); ok {
		t.Error("PollLast on empty should be false")
	}
}

func TestInt32_PollSingle(t *testing.T) {
	s := Int32Of(7)
	if v, ok := s.PollFirst(); !ok || v != 7 {
		t.Errorf("PollFirst = (%d,%v), want 7", v, ok)
	}
	if _, ok := s.PollFirst(); ok {
		t.Error("PollFirst on now-empty should be false")
	}
	if s.Len() != 0 {
		t.Errorf("Len = %d, want 0", s.Len())
	}
}

func TestInt32_RangeAndDescending(t *testing.T) {
	s := Int32Of(10, 20, 30, 40, 50, 60, 70, 80, 90, 100)
	if got := s.RangeElements(rangev.ClosedOpen(30, 70)); !slices.Equal(got, []int32{30, 40, 50, 60}) {
		t.Errorf("RangeElements(closed_open(30,70)) = %v", got)
	}
	if got := s.DescendingRangeElements(rangev.ClosedOpen(30, 70)); !slices.Equal(got, []int32{60, 50, 40, 30}) {
		t.Errorf("DescendingRangeElements = %v", got)
	}
	if got := s.RangeElements(rangev.OpenClosed(30, 70)); !slices.Equal(got, []int32{40, 50, 60, 70}) {
		t.Errorf("RangeElements(open_closed(30,70)) = %v", got)
	}
	if got := s.RangeElements(rangev.AtLeast(80)); !slices.Equal(got, []int32{80, 90, 100}) {
		t.Errorf("RangeElements(at_least(80)) = %v", got)
	}
}

func TestInt32_RangeOpenNoInteger(t *testing.T) {
	s := Int32Of(1, 2)
	if got := s.RangeElements(rangev.Open(1, 2)); len(got) != 0 {
		t.Errorf("RangeElements(open(1,2)) = %v, want []", got)
	}
	if n := s.RemoveRange(rangev.Open(1, 2)); n != 0 {
		t.Errorf("RemoveRange(open(1,2)) = %d, want 0", n)
	}
	if s.Len() != 2 {
		t.Errorf("Len = %d, want 2", s.Len())
	}
}

func TestInt32_RemoveRangeCountAndNoop(t *testing.T) {
	s := Int32Of(10, 20, 30, 40, 50, 60, 70, 80, 90, 100)
	if n := s.RemoveRange(rangev.ClosedOpen(30, 70)); n != 4 {
		t.Errorf("RemoveRange = %d, want 4", n)
	}
	if n := s.RemoveRange(rangev.ClosedOpen(30, 70)); n != 0 {
		t.Errorf("RemoveRange (repeat) = %d, want 0", n)
	}
	if got := s.ToSlice(); !slices.Equal(got, []int32{10, 20, 70, 80, 90, 100}) {
		t.Errorf("elements after remove = %v", got)
	}
}

func TestInt32_SubSetIndependence(t *testing.T) {
	s := Int32Of(10, 20, 30, 40, 50)
	snap := s.SubSet(rangev.Closed(20, 40))
	if got := snap.ToSlice(); !slices.Equal(got, []int32{20, 30, 40}) {
		t.Errorf("SubSet = %v, want [20 30 40]", got)
	}
	snap.Add(99)
	snap.Remove(20)
	if !s.Contains(20) {
		t.Error("original must still contain 20 after mutating snapshot")
	}
	if s.Contains(99) {
		t.Error("original must not gain 99 from snapshot")
	}
	s.Remove(30)
	if !snap.Contains(30) {
		t.Error("snapshot must still contain 30 after mutating original")
	}
}
