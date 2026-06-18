package rangev

import "testing"

func assertPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Error("expected panic, got none")
		}
	}()
	fn()
}

func TestContainsClosed(t *testing.T) {
	r := Closed(10, 20)
	for _, x := range []int32{10, 15, 20} {
		if !r.Contains(x) {
			t.Errorf("Closed(10,20).Contains(%d) = false, want true", x)
		}
	}
	for _, x := range []int32{9, 21} {
		if r.Contains(x) {
			t.Errorf("Closed(10,20).Contains(%d) = true, want false", x)
		}
	}
}

func TestContainsOpen(t *testing.T) {
	r := Open(10, 20)
	if r.Contains(10) || r.Contains(20) {
		t.Error("Open(10,20) must not contain its endpoints")
	}
	if !r.Contains(11) || !r.Contains(19) {
		t.Error("Open(10,20) must contain interior points")
	}
}

func TestContainsHalfOpen(t *testing.T) {
	co := ClosedOpen(10, 20)
	if !co.Contains(10) || co.Contains(20) {
		t.Error("ClosedOpen(10,20): want 10 in, 20 out")
	}
	oc := OpenClosed(10, 20)
	if oc.Contains(10) || !oc.Contains(20) {
		t.Error("OpenClosed(10,20): want 10 out, 20 in")
	}
}

func TestContainsUnbounded(t *testing.T) {
	all := All()
	for _, x := range []int32{-2147483648, 0, 2147483647} {
		if !all.Contains(x) {
			t.Errorf("All().Contains(%d) = false, want true", x)
		}
	}

	al := AtLeast(10)
	if !al.Contains(10) || al.Contains(9) || !al.Contains(2147483647) {
		t.Error("AtLeast(10) contains check failed")
	}

	gt := GreaterThan(10)
	if gt.Contains(10) || !gt.Contains(11) {
		t.Error("GreaterThan(10) contains check failed")
	}

	lt := LessThan(5)
	if !lt.Contains(4) || lt.Contains(5) {
		t.Error("LessThan(5) contains check failed")
	}

	am := AtMost(5)
	if !am.Contains(5) || am.Contains(6) {
		t.Error("AtMost(5) contains check failed")
	}
}

func assertBound(t *testing.T, label string, gotBT BoundType, gotOK bool, wantBT BoundType, wantOK bool) {
	t.Helper()
	if gotOK != wantOK || (wantOK && gotBT != wantBT) {
		t.Errorf("%s = (%v, %v), want (%v, %v)", label, gotBT, gotOK, wantBT, wantOK)
	}
}

func assertEndpoint(t *testing.T, label string, gotV int32, gotOK bool, wantV int32, wantOK bool) {
	t.Helper()
	if gotOK != wantOK || (wantOK && gotV != wantV) {
		t.Errorf("%s = (%d, %v), want (%d, %v)", label, gotV, gotOK, wantV, wantOK)
	}
}

func TestBoundTypesAndEndpoints(t *testing.T) {
	r := ClosedOpen(10, 20)
	bt, ok := r.LowerBoundType()
	assertBound(t, "ClosedOpen.LowerBoundType", bt, ok, BoundClosed, true)
	bt, ok = r.UpperBoundType()
	assertBound(t, "ClosedOpen.UpperBoundType", bt, ok, BoundOpen, true)
	v, ok := r.LowerEndpoint()
	assertEndpoint(t, "ClosedOpen.LowerEndpoint", v, ok, 10, true)
	v, ok = r.UpperEndpoint()
	assertEndpoint(t, "ClosedOpen.UpperEndpoint", v, ok, 20, true)
	if !r.HasLowerBound() || !r.HasUpperBound() {
		t.Error("ClosedOpen(10,20) must have both bounds")
	}

	all := All()
	if _, ok := all.LowerBoundType(); ok {
		t.Error("All().LowerBoundType should be absent")
	}
	if _, ok := all.UpperBoundType(); ok {
		t.Error("All().UpperBoundType should be absent")
	}
	if _, ok := all.LowerEndpoint(); ok {
		t.Error("All().LowerEndpoint should be absent")
	}
	if _, ok := all.UpperEndpoint(); ok {
		t.Error("All().UpperEndpoint should be absent")
	}
	if all.HasLowerBound() || all.HasUpperBound() {
		t.Error("All() must have no finite bounds")
	}
}

func TestEmptyCutSemantics(t *testing.T) {
	// Open(1,2) is NOT cut-empty (no DiscreteDomain), even though no int32 is
	// contained.
	o := Open(1, 2)
	if o.IsEmpty() {
		t.Error("Open(1,2) must not be empty (no DiscreteDomain)")
	}
	if o.Contains(1) || o.Contains(2) {
		t.Error("Open(1,2) must not contain 1 or 2")
	}

	// ClosedOpen(v,v) and OpenClosed(v,v) are both empty but DISTINCT.
	co := ClosedOpen(5, 5)
	oc := OpenClosed(5, 5)
	if !co.IsEmpty() || !oc.IsEmpty() {
		t.Error("ClosedOpen(5,5) and OpenClosed(5,5) must both be empty")
	}
	if co == oc {
		t.Error("ClosedOpen(5,5) and OpenClosed(5,5) must be distinct (no canonicalization)")
	}
	if co.Contains(5) || oc.Contains(5) {
		t.Error("empty ranges contain nothing")
	}
	// bound types are preserved per the cuts.
	bt, ok := co.LowerBoundType()
	assertBound(t, "ClosedOpen(5,5).LowerBoundType", bt, ok, BoundClosed, true)
	bt, ok = co.UpperBoundType()
	assertBound(t, "ClosedOpen(5,5).UpperBoundType", bt, ok, BoundOpen, true)
	bt, ok = oc.LowerBoundType()
	assertBound(t, "OpenClosed(5,5).LowerBoundType", bt, ok, BoundOpen, true)
	bt, ok = oc.UpperBoundType()
	assertBound(t, "OpenClosed(5,5).UpperBoundType", bt, ok, BoundClosed, true)
	// empties at different positions are unequal.
	if co == ClosedOpen(6, 6) {
		t.Error("ClosedOpen(5,5) and ClosedOpen(6,6) must be distinct")
	}
}

func TestSingletonNotEmpty(t *testing.T) {
	s := Singleton(5)
	if s.IsEmpty() {
		t.Error("Singleton(5) must not be empty")
	}
	if !s.Contains(5) || s.Contains(4) || s.Contains(6) {
		t.Error("Singleton(5) contains check failed")
	}
}

func TestEnclosesCutDefined(t *testing.T) {
	big := Closed(10, 30)
	if !big.Encloses(Closed(15, 25)) {
		t.Error("[10,30] should enclose [15,25]")
	}
	if big.Encloses(Closed(5, 25)) {
		t.Error("[10,30] should NOT enclose [5,25]")
	}
	// [10,30] encloses empty@20.
	if !big.Encloses(ClosedOpen(20, 20)) {
		t.Error("[10,30] should enclose empty@20")
	}
	// [1,5) encloses empty@5 (cut-defined; 5 NOT contained).
	half := ClosedOpen(1, 5)
	if !half.Encloses(ClosedOpen(5, 5)) {
		t.Error("[1,5) should enclose empty@5 (cut-defined)")
	}
	if half.Contains(5) {
		t.Error("[1,5) must not contain 5")
	}
}

func TestConnectedOverlap(t *testing.T) {
	a := Closed(10, 20)
	b := Closed(15, 25)
	if !a.IsConnected(b) {
		t.Error("[10,20] and [15,25] should be connected")
	}
	i, ok := a.Intersection(b)
	if !ok {
		t.Fatal("connected -> intersection present")
	}
	if i.IsEmpty() {
		t.Error("overlap intersection should not be empty")
	}
	if i != Closed(15, 20) {
		t.Errorf("intersection = %v, want [15,20]", i)
	}
}

func TestConnectedAbutPresentEmpty(t *testing.T) {
	// [10,20) & [20,30) -> connected, present cut-empty at (Below20,Below20).
	a := ClosedOpen(10, 20)
	b := ClosedOpen(20, 30)
	if !a.IsConnected(b) {
		t.Error("abutting ranges should be connected")
	}
	i, ok := a.Intersection(b)
	if !ok {
		t.Fatal("abut -> intersection present")
	}
	if !i.IsEmpty() {
		t.Error("abut intersection should be empty")
	}
	if i != ClosedOpen(20, 20) {
		t.Errorf("intersection = %v, want ClosedOpen(20,20)", i)
	}
	bt, _ := i.LowerBoundType()
	if bt != BoundClosed {
		t.Error("abut intersection lower type should be closed")
	}
	bt, _ = i.UpperBoundType()
	if bt != BoundOpen {
		t.Error("abut intersection upper type should be open")
	}
}

func TestAbutOpenClosedPresentEmpty(t *testing.T) {
	// [10,20] & (20,30) -> connected, present cut-empty at (Above20,Above20).
	a := Closed(10, 20)
	b := Open(20, 30)
	if !a.IsConnected(b) {
		t.Error("[10,20] and (20,30) should be connected")
	}
	i, ok := a.Intersection(b)
	if !ok {
		t.Fatal("abut -> intersection present")
	}
	if !i.IsEmpty() {
		t.Error("abut intersection should be empty")
	}
	if i != OpenClosed(20, 20) {
		t.Errorf("intersection = %v, want OpenClosed(20,20)", i)
	}
	bt, _ := i.LowerBoundType()
	if bt != BoundOpen {
		t.Error("abut intersection lower type should be open")
	}
	bt, _ = i.UpperBoundType()
	if bt != BoundClosed {
		t.Error("abut intersection upper type should be closed")
	}
}

func TestDisjointIsNone(t *testing.T) {
	a := ClosedOpen(10, 15)
	b := ClosedOpen(20, 25)
	if a.IsConnected(b) {
		t.Error("[10,15) and [20,25) should be disconnected")
	}
	if _, ok := a.Intersection(b); ok {
		t.Error("disjoint intersection should be absent")
	}
}

func TestConnectedUnboundedAbut(t *testing.T) {
	// lessThan(5) & atLeast(5) -> connected, present empty (Below5,Below5).
	a := LessThan(5)
	b := AtLeast(5)
	if !a.IsConnected(b) {
		t.Error("LessThan(5) and AtLeast(5) should be connected")
	}
	i, ok := a.Intersection(b)
	if !ok {
		t.Fatal("abut -> intersection present")
	}
	if !i.IsEmpty() {
		t.Error("intersection should be empty")
	}
	if i != ClosedOpen(5, 5) {
		t.Errorf("intersection = %v, want ClosedOpen(5,5)", i)
	}
	bt, _ := i.LowerBoundType()
	if bt != BoundClosed {
		t.Error("lower type should be closed")
	}
	bt, _ = i.UpperBoundType()
	if bt != BoundOpen {
		t.Error("upper type should be open")
	}
}

func TestDisjointUnboundedIsNone(t *testing.T) {
	// lessThan(5) & greaterThan(5) -> DISCONNECTED (5 is the gap).
	a := LessThan(5)
	b := GreaterThan(5)
	if a.IsConnected(b) {
		t.Error("LessThan(5) and GreaterThan(5) should be disconnected")
	}
	if _, ok := a.Intersection(b); ok {
		t.Error("disjoint intersection should be absent")
	}
}

func TestSpanBasic(t *testing.T) {
	a := Closed(10, 15)
	b := Closed(20, 25)
	s := a.Span(b)
	if s != Closed(10, 25) {
		t.Errorf("span = %v, want [10,25]", s)
	}
	v, _ := s.LowerEndpoint()
	if v != 10 {
		t.Error("span lower endpoint should be 10")
	}
	v, _ = s.UpperEndpoint()
	if v != 25 {
		t.Error("span upper endpoint should be 25")
	}
	bt, _ := s.LowerBoundType()
	if bt != BoundClosed {
		t.Error("span lower type should be closed")
	}
	bt, _ = s.UpperBoundType()
	if bt != BoundClosed {
		t.Error("span upper type should be closed")
	}
}

func TestSpanUnbounded(t *testing.T) {
	a := AtLeast(10)
	b := Closed(0, 5)
	s := a.Span(b)
	if s != AtLeast(0) {
		t.Errorf("span = %v, want [0,+inf)", s)
	}
	v, ok := s.LowerEndpoint()
	if !ok || v != 0 {
		t.Error("span lower endpoint should be 0")
	}
	if _, ok := s.UpperEndpoint(); ok {
		t.Error("span upper endpoint should be absent")
	}
	bt, _ := s.LowerBoundType()
	if bt != BoundClosed {
		t.Error("span lower type should be closed")
	}
	if _, ok := s.UpperBoundType(); ok {
		t.Error("span upper type should be absent")
	}
}

func TestClosedBadOrderPanics(t *testing.T) {
	assertPanics(t, func() { Closed(5, 1) })
}

func TestOpenEqualPanics(t *testing.T) {
	// Open(3,3) = (Above(3), Below(3)) is lower > upper.
	assertPanics(t, func() { Open(3, 3) })
}

func TestOpenClosedBadOrderPanics(t *testing.T) {
	assertPanics(t, func() { OpenClosed(5, 1) })
}

func TestString(t *testing.T) {
	cases := []struct {
		r    Int32Range
		want string
	}{
		{Closed(1, 5), "[1, 5]"},
		{Open(1, 5), "(1, 5)"},
		{ClosedOpen(1, 5), "[1, 5)"},
		{AtLeast(1), "[1, +∞)"},
		{LessThan(5), "(-∞, 5)"},
		{All(), "(-∞, +∞)"},
	}
	for _, c := range cases {
		if got := c.r.String(); got != c.want {
			t.Errorf("String() = %q, want %q", got, c.want)
		}
	}
}
