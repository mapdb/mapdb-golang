package treemap

import (
	"slices"
	"testing"

	"github.com/mapdb/mapdb-golang/rangev"
)

func navMapOf(keys ...int32) *Int32Int32 {
	m := NewInt32Int32()
	for _, k := range keys {
		m.Put(k, k*10)
	}
	return m
}

func TestInt32Int32_NavStrictness(t *testing.T) {
	m := navMapOf(10, 20, 30)
	if k, ok := m.FloorKey(25); !ok || k != 20 {
		t.Errorf("FloorKey(25) = (%d, %v), want 20", k, ok)
	}
	if k, ok := m.CeilingKey(25); !ok || k != 30 {
		t.Errorf("CeilingKey(25) = (%d, %v), want 30", k, ok)
	}
	if k, ok := m.FloorKey(10); !ok || k != 10 { // inclusive
		t.Errorf("FloorKey(10) = (%d, %v), want 10", k, ok)
	}
	if _, ok := m.LowerKey(10); ok { // strict, nothing below
		t.Error("LowerKey(10) should be false")
	}
	if _, ok := m.HigherKey(30); ok { // strict, nothing above
		t.Error("HigherKey(30) should be false")
	}
	if k, ok := m.CeilingKey(5); !ok || k != 10 {
		t.Errorf("CeilingKey(5) = (%d, %v), want 10", k, ok)
	}
	if k, ok := m.LowerKey(25); !ok || k != 20 {
		t.Errorf("LowerKey(25) = (%d, %v), want 20", k, ok)
	}
	if k, ok := m.HigherKey(25); !ok || k != 30 {
		t.Errorf("HigherKey(25) = (%d, %v), want 30", k, ok)
	}
	// Entry forms carry the value (= key*10).
	if k, v, ok := m.FloorEntry(25); !ok || k != 20 || v != 200 {
		t.Errorf("FloorEntry(25) = (%d,%d,%v), want (20,200,true)", k, v, ok)
	}
	if fk, ok := m.FirstKey(); !ok || fk != 10 {
		t.Errorf("FirstKey = (%d,%v), want 10", fk, ok)
	}
	if lk, ok := m.LastKey(); !ok || lk != 30 {
		t.Errorf("LastKey = (%d,%v), want 30", lk, ok)
	}
}

func TestInt32Int32_NavEmpty(t *testing.T) {
	m := NewInt32Int32()
	for _, probe := range []func(int32) (int32, bool){m.FloorKey, m.CeilingKey, m.LowerKey, m.HigherKey} {
		if _, ok := probe(5); ok {
			t.Error("nav on empty map should be false")
		}
	}
	if _, ok := m.FirstKey(); ok {
		t.Error("FirstKey on empty should be false")
	}
	if _, ok := m.LastKey(); ok {
		t.Error("LastKey on empty should be false")
	}
}

func TestInt32Int32_NavSignedExtremes(t *testing.T) {
	const minI32 = int32(-2147483648)
	const maxI32 = int32(2147483647)
	m := navMapOf(minI32, -1, 0, 1, maxI32)
	if k, ok := m.FloorKey(minI32); !ok || k != minI32 {
		t.Errorf("FloorKey(MIN) = (%d,%v)", k, ok)
	}
	if _, ok := m.LowerKey(minI32); ok {
		t.Error("LowerKey(MIN) should be false")
	}
	if k, ok := m.HigherKey(-1); !ok || k != 0 {
		t.Errorf("HigherKey(-1) = (%d,%v), want 0", k, ok)
	}
	if k, ok := m.CeilingKey(maxI32); !ok || k != maxI32 {
		t.Errorf("CeilingKey(MAX) = (%d,%v)", k, ok)
	}
	if _, ok := m.HigherKey(maxI32); ok {
		t.Error("HigherKey(MAX) should be false")
	}
	got := []int32{}
	for k := range m.DescendingKeys() {
		got = append(got, k)
	}
	if !slices.Equal(got, []int32{maxI32, 1, 0, -1, minI32}) {
		t.Errorf("DescendingKeys = %v", got)
	}
}

func TestInt32Int32_PollFirstLastThenEmpty(t *testing.T) {
	m := navMapOf(10, 20, 30)
	if k, v, ok := m.PollFirstEntry(); !ok || k != 10 || v != 100 {
		t.Errorf("PollFirstEntry = (%d,%d,%v), want (10,100,true)", k, v, ok)
	}
	if k, v, ok := m.PollLastEntry(); !ok || k != 30 || v != 300 {
		t.Errorf("PollLastEntry = (%d,%d,%v), want (30,300,true)", k, v, ok)
	}
	if m.Len() != 1 {
		t.Errorf("Len = %d, want 1", m.Len())
	}
	if k, v, ok := m.PollFirstEntry(); !ok || k != 20 || v != 200 {
		t.Errorf("PollFirstEntry = (%d,%d,%v), want (20,200,true)", k, v, ok)
	}
	// Now empty: returns false, does not trap.
	if _, _, ok := m.PollFirstEntry(); ok {
		t.Error("PollFirstEntry on empty should be false")
	}
	if _, _, ok := m.PollLastEntry(); ok {
		t.Error("PollLastEntry on empty should be false")
	}
}

func TestInt32Int32_PollSingle(t *testing.T) {
	m := navMapOf(7)
	if k, v, ok := m.PollFirstEntry(); !ok || k != 7 || v != 70 {
		t.Errorf("PollFirstEntry = (%d,%d,%v)", k, v, ok)
	}
	if _, _, ok := m.PollFirstEntry(); ok {
		t.Error("PollFirstEntry on now-empty should be false")
	}
	if m.Len() != 0 {
		t.Errorf("Len = %d, want 0", m.Len())
	}
}

func TestInt32Int32_RangeKeysAndDescending(t *testing.T) {
	m := navMapOf(10, 20, 30, 40, 50, 60, 70, 80, 90, 100)
	if got := m.RangeKeysIn(rangev.ClosedOpen(30, 70)); !slices.Equal(got, []int32{30, 40, 50, 60}) {
		t.Errorf("RangeKeysIn(closed_open(30,70)) = %v", got)
	}
	if got := m.DescendingRangeKeys(rangev.ClosedOpen(30, 70)); !slices.Equal(got, []int32{60, 50, 40, 30}) {
		t.Errorf("DescendingRangeKeys = %v", got)
	}
	if got := m.RangeKeysIn(rangev.OpenClosed(30, 70)); !slices.Equal(got, []int32{40, 50, 60, 70}) {
		t.Errorf("RangeKeysIn(open_closed(30,70)) = %v", got)
	}
	if got := m.RangeKeysIn(rangev.AtLeast(80)); !slices.Equal(got, []int32{80, 90, 100}) {
		t.Errorf("RangeKeysIn(at_least(80)) = %v", got)
	}
	entries := m.RangeEntriesIn(rangev.ClosedOpen(30, 50))
	want := []Int32Int32Entry{{30, 300}, {40, 400}}
	if !slices.Equal(entries, want) {
		t.Errorf("RangeEntriesIn = %v, want %v", entries, want)
	}
}

func TestInt32Int32_RangeOpenNoInteger(t *testing.T) {
	// open(1, 2) over int32 matches NOTHING (membership = Contains), but is not
	// cut-empty.
	m := navMapOf(1, 2)
	if got := m.RangeKeysIn(rangev.Open(1, 2)); len(got) != 0 {
		t.Errorf("RangeKeysIn(open(1,2)) = %v, want []", got)
	}
	if n := m.RemoveRange(rangev.Open(1, 2)); n != 0 {
		t.Errorf("RemoveRange(open(1,2)) = %d, want 0", n)
	}
	if m.Len() != 2 {
		t.Errorf("Len = %d, want 2", m.Len())
	}
}

func TestInt32Int32_RemoveRangeCountAndNoop(t *testing.T) {
	m := navMapOf(10, 20, 30, 40, 50, 60, 70, 80, 90, 100)
	if n := m.RemoveRange(rangev.ClosedOpen(30, 70)); n != 4 {
		t.Errorf("RemoveRange = %d, want 4", n)
	}
	if n := m.RemoveRange(rangev.ClosedOpen(30, 70)); n != 0 { // no-op
		t.Errorf("RemoveRange (repeat) = %d, want 0", n)
	}
	var keys []int32
	for k := range m.Keys() {
		keys = append(keys, k)
	}
	if !slices.Equal(keys, []int32{10, 20, 70, 80, 90, 100}) {
		t.Errorf("keys after remove = %v", keys)
	}
}

func TestInt32Int32_SubMapRangeIndependence(t *testing.T) {
	m := navMapOf(10, 20, 30, 40, 50)
	snap := m.SubMapRange(rangev.Closed(20, 40))
	var snapKeys []int32
	for k := range snap.Keys() {
		snapKeys = append(snapKeys, k)
	}
	if !slices.Equal(snapKeys, []int32{20, 30, 40}) {
		t.Errorf("SubMapRange keys = %v", snapKeys)
	}
	// Mutate snapshot — original unchanged.
	snap.Put(99, 990)
	snap.Remove(20)
	if !m.ContainsKey(20) {
		t.Error("original must still contain 20 after mutating snapshot")
	}
	if m.ContainsKey(99) {
		t.Error("original must not gain 99 from snapshot")
	}
	// Mutate original — snapshot unchanged.
	m.Remove(30)
	if !snap.ContainsKey(30) {
		t.Error("snapshot must still contain 30 after mutating original")
	}
}
