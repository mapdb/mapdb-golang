// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package rangev

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func entry(r Int32Range, v int32) Int32Int32Entry {
	return Int32Int32Entry{Range: r, Value: v}
}

func assertEntries(t *testing.T, m *Int32Int32RangeMap, want ...Int32Int32Entry) {
	t.Helper()
	got := m.AsMapOfRanges()
	if want == nil {
		want = []Int32Int32Entry{}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("entries = %v, want %v", got, want)
	}
}

func getOr(m *Int32Int32RangeMap, k int32) (int32, bool) { return m.Get(k) }

func TestRangeMapPutBasic(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 5), 100)
	m.Put(Closed(8, 9), 200)
	assertEntries(t, m, entry(ClosedOpen(1, 5), 100), entry(Closed(8, 9), 200))
	if v, ok := getOr(m, 3); !ok || v != 100 {
		t.Errorf("get(3) = %v %v", v, ok)
	}
	if _, ok := getOr(m, 6); ok {
		t.Error("get(6) should be absent")
	}
	if v, ok := getOr(m, 8); !ok || v != 200 {
		t.Errorf("get(8) = %v %v", v, ok)
	}
}

func TestRangeMapPutOverwriteClips(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 5), 100)
	m.Put(ClosedOpen(3, 9), 200)
	assertEntries(t, m, entry(ClosedOpen(1, 3), 100), entry(ClosedOpen(3, 9), 200))
	if v, _ := getOr(m, 2); v != 100 {
		t.Error("get(2) want 100")
	}
	if v, _ := getOr(m, 4); v != 200 {
		t.Error("get(4) want 200")
	}
	if v, _ := getOr(m, 8); v != 200 {
		t.Error("get(8) want 200")
	}
}

func TestRangeMapPutSplitStraddle(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 9), 100)
	m.Put(ClosedOpen(3, 5), 200)
	assertEntries(t, m,
		entry(ClosedOpen(1, 3), 100),
		entry(ClosedOpen(3, 5), 200),
		entry(ClosedOpen(5, 9), 100),
	)
	if v, _ := getOr(m, 2); v != 100 {
		t.Error("get(2) want 100")
	}
	if v, _ := getOr(m, 4); v != 200 {
		t.Error("get(4) want 200")
	}
	if v, _ := getOr(m, 6); v != 100 {
		t.Error("get(6) want 100")
	}
}

func TestRangeMapPutCoalescesEqualValueAbut(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 5), 100)
	m.Put(ClosedOpen(5, 9), 100)
	// ONE entry: equal value and abutting, so plain Put merges them.
	// Guava's TreeRangeMap leaves two here; this is the divergence.
	assertEntries(t, m, entry(ClosedOpen(1, 9), 100))
	if v, _ := getOr(m, 5); v != 100 {
		t.Error("get(5) want 100")
	}
}

func TestRangeMapPutDifferentValueNoMerge(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 5), 100)
	m.Put(ClosedOpen(5, 9), 200)
	assertEntries(t, m, entry(ClosedOpen(1, 5), 100), entry(ClosedOpen(5, 9), 200))
}

func TestRangeMapPutCoalescesBothSides(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 5), 100)
	m.Put(ClosedOpen(9, 12), 100)
	m.Put(ClosedOpen(5, 9), 100)
	assertEntries(t, m, entry(ClosedOpen(1, 12), 100))
}

// The chain that the old Put/PutCoalescing split allowed to accumulate is built
// here in ASCENDING order; it never forms, because each Put merges as it lands.
func TestRangeMapPutCoalescesChainAscendingOrder(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 2), 7)
	m.Put(ClosedOpen(2, 3), 7)
	assertEntries(t, m, entry(ClosedOpen(1, 3), 7))
	m.Put(ClosedOpen(3, 4), 7)
	assertEntries(t, m, entry(ClosedOpen(1, 4), 7))
}

// Mirror of the above: the same three Puts, inserted so the existing entries lie
// to the RIGHT of the last one. Identical result - this is the pair that pins
// order-independence.
func TestRangeMapPutCoalescesOrderIndependent(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(2, 3), 7)
	m.Put(ClosedOpen(3, 4), 7)
	m.Put(ClosedOpen(1, 2), 7)
	assertEntries(t, m, entry(ClosedOpen(1, 4), 7))
}

func TestRangeMapPutDifferentValueIsAHardBarrier(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 2), 7)
	m.Put(ClosedOpen(2, 3), 8)
	m.Put(ClosedOpen(3, 4), 7)
	// The 8 entry is neither absorbed nor crossed, so the far [1,2) -> 7 is
	// unreachable even though both hold 7.
	assertEntries(t, m,
		entry(ClosedOpen(1, 2), 7),
		entry(ClosedOpen(2, 3), 8),
		entry(ClosedOpen(3, 4), 7))
}

func TestRangeMapPutSplitFragmentsDoNotRejoinAcrossTheInsert(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 9), 100)
	m.Put(ClosedOpen(3, 5), 200)
	// The two 100 fragments are separated by the 200 entry, so they are not
	// connected and must not be re-merged by the coalescing step.
	assertEntries(t, m,
		entry(ClosedOpen(1, 3), 100),
		entry(ClosedOpen(3, 5), 200),
		entry(ClosedOpen(5, 9), 100))
}

// The global invariant that the old Put/PutCoalescing split could not state:
// after every operation, no two connected entries hold an equal value.
func TestRangeMapNormalFormHasNoConnectedEqualValuedPair(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 2), 7)
	m.Put(ClosedOpen(2, 3), 7)
	m.Put(ClosedOpen(3, 4), 8)
	m.Put(ClosedOpen(4, 5), 8)
	m.Put(ClosedOpen(5, 6), 7)
	v := m.AsMapOfRanges()
	for i := 0; i+1 < len(v); i++ {
		if v[i].Range.IsConnected(v[i+1].Range) && v[i].Value == v[i+1].Value {
			t.Errorf("connected entries %v and %v hold an equal value", v[i], v[i+1])
		}
	}
	assertEntries(t, m,
		entry(ClosedOpen(1, 3), 7),
		entry(ClosedOpen(3, 5), 8),
		entry(ClosedOpen(5, 6), 7))
}

func TestRangeMapRemoveSplits(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 9), 100)
	m.Remove(ClosedOpen(4, 7))
	assertEntries(t, m, entry(ClosedOpen(1, 4), 100), entry(ClosedOpen(7, 9), 100))
	if _, ok := getOr(m, 5); ok {
		t.Error("get(5) should be absent after remove")
	}
}

func TestRangeMapGetEntryLookup(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 5), 100)
	if r, v, ok := m.GetEntry(3); !ok || r != ClosedOpen(1, 5) || v != 100 {
		t.Errorf("getEntry(3) = %v %v %v", r, v, ok)
	}
	if _, _, ok := m.GetEntry(6); ok {
		t.Error("getEntry(6) should be absent")
	}
}

func TestRangeMapSpanOverEntries(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 5), 100)
	m.Put(Closed(8, 9), 200)
	if sp, ok := m.Span(); !ok || sp != Closed(1, 9) {
		t.Errorf("span = %v %v", sp, ok)
	}
}

func TestRangeMapEmptyPutIsNoop(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(5, 5), 100)
	if !m.IsEmpty() {
		t.Error("cut-empty put should be no-op")
	}
	assertEntries(t, m)
}

func TestRangeMapSubRangeMapClipsSnapshot(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 5), 100)
	m.Put(Closed(8, 9), 200)
	sub := m.SubRangeMap(ClosedOpen(3, 6))
	assertEntries(t, sub, entry(ClosedOpen(3, 5), 100))
	// snapshot independence: mutate the parent, sub unchanged.
	m.Put(Closed(3, 3), 999)
	assertEntries(t, sub, entry(ClosedOpen(3, 5), 100))
	// mutating the snapshot does not touch the parent.
	sub.Put(Closed(50, 60), 7)
	if _, ok := m.Get(55); ok {
		t.Error("parent must not see snapshot mutation")
	}
}

func TestRangeMapSignedExtremesNoPlusMinusOne(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(math.MinInt32, 0), 1)
	m.Put(Closed(0, math.MaxInt32), 2)
	if v, _ := getOr(m, math.MinInt32); v != 1 {
		t.Error("get(MIN) want 1")
	}
	if v, _ := getOr(m, 0); v != 2 {
		t.Error("get(0) want 2")
	}
	if v, _ := getOr(m, math.MaxInt32); v != 2 {
		t.Error("get(MAX) want 2")
	}
}

func TestRangeMapNormalFormDisjointAfterSequence(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 10), 1)
	m.Put(ClosedOpen(3, 5), 2)
	m.Put(ClosedOpen(7, 20), 3)
	m.Put(ClosedOpen(20, 25), 3)
	v := m.AsMapOfRanges()
	for i := 0; i+1 < len(v); i++ {
		if v[i].Range.lower.cmp(v[i+1].Range.lower) >= 0 {
			t.Errorf("not ascending at %d: %v", i, v)
		}
		inter, ok := v[i].Range.Intersection(v[i+1].Range)
		if ok && !inter.IsEmpty() {
			t.Errorf("overlapping entries at %d: %v", i, v)
		}
	}
	for _, e := range v {
		if e.Range.IsEmpty() {
			t.Errorf("empty entry range: %v", v)
		}
	}
}

func TestRangeMapClear(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 9), 100)
	m.Clear()
	if !m.IsEmpty() {
		t.Error("clear should empty")
	}
	assertEntries(t, m)
}

// Unbounded entries on both sides: each pair merges as it lands, then the
// bridging Put collapses everything to a single all() entry. No endpoint
// arithmetic is involved; the sentinels alone carry the connectivity.
func TestRangeMapPutUnboundedChainsOnBothSidesCollapseToAll(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(LessThan(-5), 7)
	m.Put(ClosedOpen(-5, 0), 7)
	m.Put(ClosedOpen(5, 10), 7)
	m.Put(AtLeast(10), 7)
	assertEntries(t, m, entry(LessThan(0), 7), entry(AtLeast(5), 7))
	m.Put(ClosedOpen(0, 5), 7)
	assertEntries(t, m, entry(All(), 7))
	for _, k := range []int32{math.MinInt32, 0, math.MaxInt32} {
		if v, ok := getOr(m, k); !ok || v != 7 {
			t.Errorf("get(%d) = %v %v, want 7 true", k, v, ok)
		}
	}
}

// The Put overlaps the tail of an equal-valued entry and the head of a
// different-valued one: clipOut leaves [0,6) -> 7 and [11,12) -> 9, and the
// equal fragment must rejoin the inserted range.
func TestRangeMapPutRejoinsClippedFragmentAndChainBeyondIt(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(0, 10), 7)
	m.Put(ClosedOpen(10, 12), 9)
	m.Put(ClosedOpen(6, 11), 7)
	assertEntries(t, m, entry(ClosedOpen(0, 11), 7), entry(ClosedOpen(11, 12), 9))
	if v, _ := getOr(m, 0); v != 7 {
		t.Error("get(0) want 7")
	}
	if v, _ := getOr(m, 10); v != 7 {
		t.Error("get(10) want 7")
	}
	if v, _ := getOr(m, 11); v != 9 {
		t.Error("get(11) want 9")
	}
}

// A Put strictly inside an equal-valued entry splits it into two fragments,
// both of which must rejoin: the map is unchanged. The second half pins that
// the flanking different-valued entries are untouched too.
func TestRangeMapPutRejoinsBothClipFragmentsOfStraddledEqualEntry(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(0, 20), 7)
	m.Put(ClosedOpen(6, 14), 7)
	assertEntries(t, m, entry(ClosedOpen(0, 20), 7))

	m2 := NewInt32Int32RangeMap()
	m2.Put(ClosedOpen(0, 2), 1)
	m2.Put(ClosedOpen(2, 18), 7)
	m2.Put(ClosedOpen(18, 20), 1)
	m2.Put(ClosedOpen(6, 14), 7)
	assertEntries(t, m2,
		entry(ClosedOpen(0, 2), 1),
		entry(ClosedOpen(2, 18), 7),
		entry(ClosedOpen(18, 20), 1))
}

// A cut-empty Put is a no-op against a non-empty map — it must return before
// clipOut, so nothing is clipped and nothing moves. Both cut-empty forms are
// dropped on the abutment seam, and one lands inside an entry.
func TestRangeMapPutCutEmptyAtAbutmentIsNoop(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 5), 100)
	m.Put(ClosedOpen(5, 9), 200)
	m.Put(ClosedOpen(5, 5), 100)
	assertEntries(t, m, entry(ClosedOpen(1, 5), 100), entry(ClosedOpen(5, 9), 200))
	m.Put(OpenClosed(5, 5), 200)
	assertEntries(t, m, entry(ClosedOpen(1, 5), 100), entry(ClosedOpen(5, 9), 200))
	m.Put(ClosedOpen(3, 3), 999)
	assertEntries(t, m, entry(ClosedOpen(1, 5), 100), entry(ClosedOpen(5, 9), 200))
}

// Open(1,2) is cut-non-empty but holds no int32, so no point lookup can see
// it. It must still be stored, split the surrounding entry, and act as a
// barrier between the two equal-valued halves; removing it leaves the halves
// disconnected (the gap is still there) so they must NOT rejoin.
func TestRangeMapNoIntegerRangeIsAStoredBarrier(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(All(), 1)
	m.Put(Open(1, 2), 2)
	assertEntries(t, m, entry(AtMost(1), 1), entry(Open(1, 2), 2), entry(AtLeast(2), 1))
	m.Remove(Open(1, 2))
	assertEntries(t, m, entry(AtMost(1), 1), entry(AtLeast(2), 1))
}

// Removing an unbounded range from all() must clip to the exact cut of the
// removed range's finite end: LessThan(0) leaves Below(0) as the new lower
// cut, AtMost(0) leaves Above(0). No +-1 endpoint arithmetic.
func TestRangeMapRemoveUnboundedClipsToExactSentinelCut(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(All(), 1)
	m.Remove(LessThan(0))
	assertEntries(t, m, entry(AtLeast(0), 1))

	m2 := NewInt32Int32RangeMap()
	m2.Put(All(), 1)
	m2.Remove(AtMost(0))
	assertEntries(t, m2, entry(GreaterThan(0), 1))
}

// ---- oracle test -----------------------------------------------------------

// oracleDomainLo..oracleDomainHi is the dense point universe the oracle tracks;
// random endpoints are drawn from the narrower oracleEndLo..oracleEndHi so the
// unbounded forms reach points no bounded range can.
const (
	oracleDomainLo = -12
	oracleDomainHi = 12
	oracleEndLo    = -8
	oracleEndHi    = 8
)

// oracleRand is a tiny xorshift64* so the sequence is deterministic without a
// dependency on math/rand's generator stability.
type oracleRand struct{ s uint64 }

func (r *oracleRand) next() uint64 {
	r.s ^= r.s >> 12
	r.s ^= r.s << 25
	r.s ^= r.s >> 27
	return r.s * 2685821657736338717
}

func (r *oracleRand) intn(n int) int { return int(r.next() % uint64(n)) }

// oracleRandRange draws one of the four bounded bound types or one of the five
// unbounded forms. Cut-empty draws ([a,a) / (a,a]) are let through: Put/Remove
// must treat them as no-ops, and the oracle (which contains no point for them)
// agrees. Open(a,a) is inverted rather than cut-empty, so it is diverted.
func oracleRandRange(r *oracleRand) Int32Range {
	span := oracleEndHi - oracleEndLo + 1
	a := int32(oracleEndLo + r.intn(span))
	b := int32(oracleEndLo + r.intn(span))
	if a > b {
		a, b = b, a
	}
	switch r.intn(9) {
	case 0:
		return LessThan(b)
	case 1:
		return AtMost(b)
	case 2:
		return GreaterThan(a)
	case 3:
		return AtLeast(a)
	case 4:
		return All()
	case 5:
		return Closed(a, b)
	case 6:
		if a == b {
			return Closed(a, a)
		}
		return Open(a, b)
	case 7:
		return OpenClosed(a, b)
	default:
		return ClosedOpen(a, b)
	}
}

// oracleCell is one point of the dense oracle: the value held there, if any.
type oracleCell struct {
	value int32
	ok    bool
}

// assertRangeMapMatchesOracle checks Get, GetEntry, normal form (ascending,
// cut-non-empty, pairwise disjoint, no connected equal-valued pair) and that
// the entries reconstruct the dense oracle exactly.
func assertRangeMapMatchesOracle(t *testing.T, m *Int32Int32RangeMap, oracle []oracleCell, ctx string) {
	t.Helper()
	es := m.AsMapOfRanges()
	for i := range es {
		if es[i].Range.IsEmpty() {
			t.Fatalf("%s: cut-empty entry %d in %v", ctx, i, es)
		}
		if i == 0 {
			continue
		}
		if es[i-1].Range.lower.cmp(es[i].Range.lower) >= 0 {
			t.Fatalf("%s: not ascending at %d in %v", ctx, i, es)
		}
		if inter, ok := es[i-1].Range.Intersection(es[i].Range); ok && !inter.IsEmpty() {
			t.Fatalf("%s: overlap at %d in %v", ctx, i, es)
		}
		if es[i-1].Range.IsConnected(es[i].Range) && es[i-1].Value == es[i].Value {
			t.Fatalf("%s: connected equal-valued pair at %d in %v", ctx, i, es)
		}
	}
	rebuilt := make([]oracleCell, len(oracle))
	for _, e := range es {
		for p := oracleDomainLo; p <= oracleDomainHi; p++ {
			if e.Range.Contains(int32(p)) {
				rebuilt[p-oracleDomainLo] = oracleCell{e.Value, true}
			}
		}
	}
	for p := oracleDomainLo; p <= oracleDomainHi; p++ {
		want := oracle[p-oracleDomainLo]
		if v, ok := m.Get(int32(p)); v != want.value || ok != want.ok {
			t.Fatalf("%s: get(%d) = (%d,%v), oracle (%d,%v); entries %v",
				ctx, p, v, ok, want.value, want.ok, es)
		}
		r, v, ok := m.GetEntry(int32(p))
		if ok != want.ok || (ok && (v != want.value || !r.Contains(int32(p)))) {
			t.Fatalf("%s: getEntry(%d) = (%v,%d,%v), oracle (%d,%v); entries %v",
				ctx, p, r, v, ok, want.value, want.ok, es)
		}
		if rebuilt[p-oracleDomainLo] != want {
			t.Fatalf("%s: entries rebuild point %d as %v, oracle %v; entries %v",
				ctx, p, rebuilt[p-oracleDomainLo], want, es)
		}
	}
}

// Seeded random Put/Remove sequence checked after EVERY op against a naive
// dense oracle (one cell per point of -12..12). Values come from {1,2,3} so
// equal-value coalescing happens constantly.
func TestRangeMapRandomPutRemoveMatchesDenseOracle(t *testing.T) {
	for _, seed := range []uint64{20260731, 4242, 0x9E3779B97F4A7C15} {
		rnd := &oracleRand{s: seed}
		m := NewInt32Int32RangeMap()
		oracle := make([]oracleCell, oracleDomainHi-oracleDomainLo+1)
		for op := 0; op < 400; op++ {
			r := oracleRandRange(rnd)
			var ctx string
			if rnd.intn(10) < 7 {
				v := int32(1 + rnd.intn(3))
				m.Put(r, v)
				for p := oracleDomainLo; p <= oracleDomainHi; p++ {
					if r.Contains(int32(p)) {
						oracle[p-oracleDomainLo] = oracleCell{v, true}
					}
				}
				ctx = fmt.Sprintf("seed %d op %d put %v -> %d", seed, op, r, v)
			} else {
				m.Remove(r)
				for p := oracleDomainLo; p <= oracleDomainHi; p++ {
					if r.Contains(int32(p)) {
						oracle[p-oracleDomainLo] = oracleCell{}
					}
				}
				ctx = fmt.Sprintf("seed %d op %d remove %v", seed, op, r)
			}
			assertRangeMapMatchesOracle(t, m, oracle, ctx)
		}
	}
}
