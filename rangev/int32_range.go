// Package rangev provides the Bound/Range value model: a pure in-memory
// value type describing a region [lo, hi), (-inf, hi], (lo, +inf), ... with
// each endpoint independently unbounded / open / closed.
//
// This is NOT interval.Int32 (which materialises an arithmetic progression
// and enumerates elements). An Int32Range holds no elements; it only
// describes a region (Contains(x)) and supports the open/unbounded endpoints
// an Interval cannot.
//
// The design follows Google Guava's Range / BoundType / Cut. The algebra
// (Intersection, Span, IsConnected, Encloses) is total and unambiguous
// because endpoints are modelled as cuts BETWEEN values rather than
// (value, inclusive) pairs. See spec/features/bound-range.md for the
// normative algorithms; every operation here reduces to a side-aware cut
// comparison, never to a (value, inclusive) boolean.
//
// Side-aware cut ordering: an unbounded cut is contextual (as a lower cut it
// is -inf, as an upper cut it is +inf). We avoid that ambiguity by splitting
// the unbounded state into two distinct sentinels — cutBelowAll (-inf) and
// cutAboveAll (+inf). With those two sentinels the four-variant cut has a
// single total order (BelowAll < Below(v) < Above(v) < AboveAll, finite cuts
// breaking ties by value then Below < Above), and the three spec comparators
// all collapse onto it.
//
// v1 ships the int32 specialisation only (matching the cross-language
// validation universe); the wider matrix widens later exactly as Interval
// did. Go has no method-level generics, so this is a per-primitive type.
package rangev

import (
	"fmt"
	"strings"
)

// BoundType is the kind of a finite endpoint: Open (exclusive) or Closed
// (inclusive).
type BoundType int

// The BoundType constants are prefixed (BoundOpen / BoundClosed) because the
// bare names Open / Closed are taken by the Guava-parity factory functions —
// the factory names are the hard cross-language constraint, so the enum gets
// the prefix.
const (
	// BoundOpen is an exclusive endpoint.
	BoundOpen BoundType = iota
	// BoundClosed is an inclusive endpoint.
	BoundClosed
)

// String renders the bound type as "open"/"closed".
func (b BoundType) String() string {
	switch b {
	case BoundOpen:
		return "open"
	case BoundClosed:
		return "closed"
	default:
		return fmt.Sprintf("BoundType(%d)", int(b))
	}
}

// cutKind is one of the four cut variants. A cut sits BETWEEN values
// (Guava's Cut). The four-variant form carries two distinct unbounded
// sentinels (BelowAll = -inf, AboveAll = +inf) so the cut has a single,
// total, context-free order — there is no lone Unbounded value with an
// ambiguous position.
type cutKind int8

const (
	cutBelowAll cutKind = iota // -inf; only ever a lower cut.
	cutBelow                   // the cut immediately below value.
	cutAbove                   // the cut immediately above value.
	cutAboveAll                // +inf; only ever an upper cut.
)

// cut is a cut between values. value is meaningful only when kind is
// cutBelow or cutAbove.
//
// Endpoint meaning:
//   - Below(v): closed lower [v, or open upper v).
//   - Above(v): open lower (v, or closed upper v].
type cut struct {
	kind  cutKind
	value int32
}

// rank collapses the finite variants so the sentinels order around them.
func (c cut) rank() int8 {
	switch c.kind {
	case cutBelowAll:
		return 0
	case cutBelow, cutAbove:
		return 1
	default: // cutAboveAll
		return 2
	}
}

// cmp is the total order on cuts (the single source of truth for the
// algebra). The three side-aware spec comparators all reduce to this because
// the two unbounded states are distinct sentinels rather than one ambiguous
// Unbounded. Returns -1, 0, or +1.
//
// Total order: BelowAll < Below(v) < Above(v) < AboveAll. Finite cuts at
// different values order by value; at the same value Below(v) < Above(v).
func (c cut) cmp(o cut) int {
	cFinite := c.kind == cutBelow || c.kind == cutAbove
	oFinite := o.kind == cutBelow || o.kind == cutAbove
	if cFinite && oFinite {
		switch {
		case c.value < o.value:
			return -1
		case c.value > o.value:
			return 1
		}
		// Same value: Below(v) < Above(v).
		cAbove := c.kind == cutAbove
		oAbove := o.kind == cutAbove
		switch {
		case !cAbove && oAbove:
			return -1
		case cAbove && !oAbove:
			return 1
		default:
			return 0
		}
	}
	switch {
	case c.rank() < o.rank():
		return -1
	case c.rank() > o.rank():
		return 1
	default:
		return 0
	}
}

func maxCut(a, b cut) cut {
	if a.cmp(b) < 0 {
		return b
	}
	return a
}

func minCut(a, b cut) cut {
	if a.cmp(b) > 0 {
		return b
	}
	return a
}

// Int32Range is an ordered region (lowerCut, upperCut) over int32, with the
// invariant lowerCut <= upperCut. Equality is structural on the two cuts —
// ClosedOpen(v, v) and OpenClosed(v, v) are distinct (both empty) values, so
// Int32Range is comparable with == in Go.
type Int32Range struct {
	lower cut
	upper cut
}

// fromCuts constructs from raw cuts after validating lower <= upper. Panics
// if the cuts are out of order (a programming error, like interval.Reversed
// at the minimum step).
func fromCuts(lower, upper cut) Int32Range {
	if lower.cmp(upper) > 0 {
		panic("rangev: lower cut must not exceed upper cut")
	}
	return Int32Range{lower: lower, upper: upper}
}

// ---- factories (Guava-parity names) --------------------------------------

// Open returns (a, b) — both endpoints open. Panics if a >= b (including
// Open(v, v), which is empty-but-invalid-as-open).
func Open(a, b int32) Int32Range {
	return fromCuts(cut{cutAbove, a}, cut{cutBelow, b})
}

// Closed returns [a, b] — both endpoints closed. Panics if a > b.
func Closed(a, b int32) Int32Range {
	return fromCuts(cut{cutBelow, a}, cut{cutAbove, b})
}

// OpenClosed returns (a, b]. Panics if a > b.
func OpenClosed(a, b int32) Int32Range {
	return fromCuts(cut{cutAbove, a}, cut{cutAbove, b})
}

// ClosedOpen returns [a, b). Panics if a > b. ClosedOpen(v, v) is the valid
// empty range (Below(v), Below(v)).
func ClosedOpen(a, b int32) Int32Range {
	return fromCuts(cut{cutBelow, a}, cut{cutBelow, b})
}

// GreaterThan returns (a, +inf).
func GreaterThan(a int32) Int32Range {
	return fromCuts(cut{cutAbove, a}, cut{kind: cutAboveAll})
}

// AtLeast returns [a, +inf).
func AtLeast(a int32) Int32Range {
	return fromCuts(cut{cutBelow, a}, cut{kind: cutAboveAll})
}

// LessThan returns (-inf, b).
func LessThan(b int32) Int32Range {
	return fromCuts(cut{kind: cutBelowAll}, cut{cutBelow, b})
}

// AtMost returns (-inf, b].
func AtMost(b int32) Int32Range {
	return fromCuts(cut{kind: cutBelowAll}, cut{cutAbove, b})
}

// All returns (-inf, +inf).
func All() Int32Range {
	return Int32Range{lower: cut{kind: cutBelowAll}, upper: cut{kind: cutAboveAll}}
}

// Singleton returns [v, v].
func Singleton(v int32) Int32Range {
	return fromCuts(cut{cutBelow, v}, cut{cutAbove, v})
}

// ---- queries -------------------------------------------------------------

// Contains reports whether x falls within the range (normative contains).
func (r Int32Range) Contains(x int32) bool {
	var lowerOK bool
	switch r.lower.kind {
	case cutBelowAll:
		lowerOK = true
	case cutBelow:
		lowerOK = r.lower.value <= x
	case cutAbove:
		lowerOK = r.lower.value < x
	default: // cutAboveAll
		lowerOK = false
	}
	var upperOK bool
	switch r.upper.kind {
	case cutAboveAll:
		upperOK = true
	case cutBelow:
		upperOK = x < r.upper.value
	case cutAbove:
		upperOK = x <= r.upper.value
	default: // cutBelowAll
		upperOK = false
	}
	return lowerOK && upperOK
}

// IsEmpty reports cut-emptiness: lowerCut == upperCut. This is NOT discrete
// cardinality — Open(1, 2) over int32 is not empty (no DiscreteDomain in
// Phase 0).
func (r Int32Range) IsEmpty() bool {
	return r.lower.cmp(r.upper) == 0
}

// LowerBoundType returns the bound type of the lower endpoint and true, or
// (0, false) when unbounded below.
func (r Int32Range) LowerBoundType() (BoundType, bool) {
	switch r.lower.kind {
	case cutBelow:
		return BoundClosed, true
	case cutAbove:
		return BoundOpen, true
	default:
		return 0, false
	}
}

// UpperBoundType returns the bound type of the upper endpoint and true, or
// (0, false) when unbounded above.
func (r Int32Range) UpperBoundType() (BoundType, bool) {
	switch r.upper.kind {
	case cutBelow:
		return BoundOpen, true
	case cutAbove:
		return BoundClosed, true
	default:
		return 0, false
	}
}

// LowerEndpoint returns the lower endpoint value and true, or (0, false) when
// unbounded below.
func (r Int32Range) LowerEndpoint() (int32, bool) {
	if r.lower.kind == cutBelow || r.lower.kind == cutAbove {
		return r.lower.value, true
	}
	return 0, false
}

// UpperEndpoint returns the upper endpoint value and true, or (0, false) when
// unbounded above.
func (r Int32Range) UpperEndpoint() (int32, bool) {
	if r.upper.kind == cutBelow || r.upper.kind == cutAbove {
		return r.upper.value, true
	}
	return 0, false
}

// HasLowerBound reports whether the lower endpoint is finite.
func (r Int32Range) HasLowerBound() bool {
	return r.lower.kind == cutBelow || r.lower.kind == cutAbove
}

// HasUpperBound reports whether the upper endpoint is finite.
func (r Int32Range) HasUpperBound() bool {
	return r.upper.kind == cutBelow || r.upper.kind == cutAbove
}

// ---- algebra (all via cut comparison) ------------------------------------

// Encloses reports cut-defined containment: self.lower <= other.lower and
// self.upper >= other.upper. This is NOT "every value in other is contained"
// — [1, 5) encloses the empty [5, 5) though 5 is not in [1, 5).
func (r Int32Range) Encloses(other Int32Range) bool {
	return r.lower.cmp(other.lower) <= 0 && r.upper.cmp(other.upper) >= 0
}

// IsConnected reports whether there is a (possibly empty) range enclosed by
// both. Cut-equal endpoints count as connected (empty overlap).
func (r Int32Range) IsConnected(other Int32Range) bool {
	return r.lower.cmp(other.upper) <= 0 && other.lower.cmp(r.upper) <= 0
}

// Intersection returns the overlap and true. The second result is false ONLY
// when the ranges are disconnected; abutting operands return a present
// cut-empty range at the touch point.
func (r Int32Range) Intersection(other Int32Range) (Int32Range, bool) {
	if !r.IsConnected(other) {
		return Int32Range{}, false
	}
	lower := maxCut(r.lower, other.lower)
	upper := minCut(r.upper, other.upper)
	return Int32Range{lower: lower, upper: upper}, true
}

// Span returns the smallest range enclosing both. No cross-shape
// canonicalisation.
func (r Int32Range) Span(other Int32Range) Int32Range {
	lower := minCut(r.lower, other.lower)
	upper := maxCut(r.upper, other.upper)
	return Int32Range{lower: lower, upper: upper}
}

// String renders the range in interval notation, e.g. "[1, 5)", "(-inf, 5)".
func (r Int32Range) String() string {
	var b strings.Builder
	switch r.lower.kind {
	case cutBelowAll:
		b.WriteString("(-∞")
	case cutBelow:
		fmt.Fprintf(&b, "[%d", r.lower.value)
	case cutAbove:
		fmt.Fprintf(&b, "(%d", r.lower.value)
	default: // cutAboveAll
		b.WriteString("(+∞")
	}
	b.WriteString(", ")
	switch r.upper.kind {
	case cutAboveAll:
		b.WriteString("+∞)")
	case cutBelow:
		fmt.Fprintf(&b, "%d)", r.upper.value)
	case cutAbove:
		fmt.Fprintf(&b, "%d]", r.upper.value)
	default: // cutBelowAll
		b.WriteString("-∞)")
	}
	return b.String()
}
