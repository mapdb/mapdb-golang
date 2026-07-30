package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
)

// genInterval is invoked from interval/ via go:generate. It writes one
// <primitive>_interval.go per supported primitive into the current working
// directory. Floats get a not-applicable stub (intervals require a signed
// step domain that floats do not naturally provide). Char is skipped.
func genInterval() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	applies := parse("interval", intervalTmpl)
	stub := parse("interval-stub", intervalStubTmpl)

	for _, p := range Primitives() {
		if p.IsChar {
			continue
		}
		data := struct {
			StructName  string
			GoType      string
			MinStepExpr string
			IsWidest    bool
		}{
			StructName:  p.Name,
			GoType:      p.GoType,
			MinStepExpr: p.MinStepExpr(),
			// int64 is the widest value type: it has no wider native type to
			// widen into for the from + step*index product, so it computes the
			// wrapping arithmetic directly at int64 width via an explicit uint64
			// round-trip. Narrower types widen into int64 instead.
			IsWidest: p.GoType == "int64",
		}
		tmpl := applies
		if p.IsFloating {
			tmpl = stub
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("execute %s: %w", p.SnakeName, err)
		}
		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			return fmt.Errorf("format %s_interval.go: %w\n---\n%s", p.SnakeName, err, buf.String())
		}
		out := filepath.Join(cwd, p.SnakeName+"_interval.go")
		if err := os.WriteFile(out, formatted, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", out, err)
		}
	}

	// Stamp the conformance laws (todo 14 §4). Only the signed-int intervals have
	// a real All()/ToSlice() (char is skipped, floats are "not applicable"
	// stubs). An interval is computed from/to/step, not Of-constructed, and
	// enumerates ascending in step direction → law 1 is order-sensitive.
	//
	// ToSlice() is built by ranging All(), so an All ≡ ToSlice check would pass
	// however All() broke. New*(1, 8, 1) is inclusive of both ends, so the law is
	// stamped against the literal 1..8 the interval must enumerate.
	rows := make([]confType, 0, 4)
	for _, p := range Primitives() {
		if !p.IsSigned {
			continue
		}
		want := make([]string, 0, 8)
		for v := 1; v <= 8; v++ {
			want = append(want, fmt.Sprintf("%s(%d)", p.GoType, v))
		}
		rows = append(rows, confType{
			TypeName: p.Name,
			CtorExpr: "interval.New" + p.Name + "(1, 8, 1)",
			Ordered:  true,
			WantExpr: "[]" + p.GoType + "{" + join(want, ", ") + "}",
		})
	}
	return genConformanceTest("interval", rows, true)
}

const intervalStubTmpl = genHeader + `package interval

// Interval is not applicable to {{.GoType}}.
`

const intervalTmpl = genHeader + `package interval

import (
	"fmt"
	"iter"
	"strings"

	"github.com/mapdb/mapdb-golang/internal/segment"
)

// {{.StructName}} is a virtual collection representing a range of {{.GoType}} values
// [from, to] with a given step. No elements are materialised in memory.
type {{.StructName}} struct {
	from {{.GoType}}
	to   {{.GoType}}
	step {{.GoType}}
}

// New{{.StructName}} creates an interval from ` + "`from`" + ` to ` + "`to`" + ` (inclusive) with the
// given step. Panics if step is zero or if the step direction doesn't match
// the from/to direction.
func New{{.StructName}}(from, to, step {{.GoType}}) *{{.StructName}} {
	if step == 0 {
		panic("interval.{{.StructName}}: step must not be zero")
	}
	if from < to && step < 0 {
		panic("interval.{{.StructName}}: step must be positive when from < to")
	}
	if from > to && step > 0 {
		panic("interval.{{.StructName}}: step must be negative when from > to")
	}
	return &{{.StructName}}{from: from, to: to, step: step}
}

// {{.StructName}}FromTo creates an interval from ` + "`from`" + ` to ` + "`to`" + ` (inclusive) with
// step 1 (ascending) or -1 (descending).
func {{.StructName}}FromTo(from, to {{.GoType}}) *{{.StructName}} {
	var step {{.GoType}} = 1
	if from > to {
		step = -1
	}
	return &{{.StructName}}{from: from, to: to, step: step}
}

// {{.StructName}}OneTo creates an interval from 1 to ` + "`to`" + ` (inclusive).
func {{.StructName}}OneTo(to {{.GoType}}) *{{.StructName}} {
	return {{.StructName}}FromTo(1, to)
}

// {{.StructName}}ZeroTo creates an interval from 0 to ` + "`to`" + ` (inclusive).
func {{.StructName}}ZeroTo(to {{.GoType}}) *{{.StructName}} {
	return {{.StructName}}FromTo(0, to)
}

// From returns the start of the interval.
func (iv *{{.StructName}}) From() {{.GoType}} { return iv.from }

// To returns the end of the interval (inclusive).
func (iv *{{.StructName}}) To() {{.GoType}} { return iv.to }

// Step returns the step.
func (iv *{{.StructName}}) Step() {{.GoType}} { return iv.step }

// Len returns the number of elements in the interval. Use iv.Len() == 0 to
// test for emptiness.
func (iv *{{.StructName}}) Len() int {
	if (iv.step > 0 && iv.from > iv.to) || (iv.step < 0 && iv.from < iv.to) {
		return 0
	}
	count := iv.distance()/iv.absStep() + 1
	maxInt := uint64(^uint(0) >> 1)
	if count > maxInt {
		return int(maxInt)
	}
	return int(count)
}


// Contains returns true if the interval contains the given value.
func (iv *{{.StructName}}) Contains(value {{.GoType}}) bool {
	if iv.step > 0 {
		return value >= iv.from && value <= iv.to && (uint64(int64(value))-uint64(int64(iv.from)))%iv.absStep() == 0
	}
	return value <= iv.from && value >= iv.to && (uint64(int64(iv.from))-uint64(int64(value)))%iv.absStep() == 0
}

// Get returns the element at the given index. It panics if the index is out
// of bounds, matching the semantics of a native Go slice.
{{- if .IsWidest}}
//
// Narrower intervals (int8/16/32) widen into int64 before computing
// from + step*index so the product never overflows the value width. int64 has
// no wider native type, so this computes directly at int64 width. Per the spec
// integer-overflow contract (algorithms.md "Integer overflow contract"), the
// arithmetic is wrapping two's-complement at the value width — Go's native
// int64 arithmetic wraps, which is exactly the required semantics. The product
// is built in uint64 and reinterpreted so the wrap is explicit and free of
// implementation-defined signed-overflow assumptions.
{{- end}}
func (iv *{{.StructName}}) Get(index int) {{.GoType}} {
	if index < 0 || index >= iv.Len() {
		panic(fmt.Sprintf("interval.{{.StructName}}: index out of range [%d] with length %d", index, iv.Len()))
	}
	return {{if .IsWidest}}int64(uint64(iv.from) + uint64(iv.step)*uint64(int64(index))){{else}}{{.GoType}}(int64(iv.from) + int64(iv.step)*int64(index)){{end}}
}

func (iv *{{.StructName}}) absStep() uint64 {
	step := int64(iv.step)
	if step < 0 {
		return uint64(^step) + 1
	}
	return uint64(step)
}

func (iv *{{.StructName}}) distance() uint64 {
	if iv.step > 0 {
		return uint64(int64(iv.to)) - uint64(int64(iv.from))
	}
	return uint64(int64(iv.from)) - uint64(int64(iv.to))
}

// All returns an iter.Seq that yields elements in order.
func (iv *{{.StructName}}) All() iter.Seq[{{.GoType}}] {
	return func(yield func({{.GoType}}) bool) {
		size := iv.Len()
		for i := 0; i < size; i++ {
			if !yield(iv.Get(i)) {
				return
			}
		}
	}
}

// Segments cuts the interval's index space into up to n balanced, contiguous,
// non-overlapping views (k = min(n, Len), or 1 when empty) whose concatenation
// reproduces All in order, so a *{{.StructName}} satisfies par.Segmenter[{{.GoType}}]
// and feeds par.From directly. Each view computes its elements on the fly via Get
// (no materialization), matching the interval's virtual nature.
func (iv *{{.StructName}}) Segments(n int) []iter.Seq[{{.GoType}}] {
	return segment.SplitIndex(iv.Len(), n, iv.Get)
}

// ForEach calls the given function for each element in order.
func (iv *{{.StructName}}) ForEach(f func({{.GoType}})) {
	for v := range iv.All() {
		f(v)
	}
}

// AnySatisfy returns true if any element satisfies the predicate.
func (iv *{{.StructName}}) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	for v := range iv.All() {
		if predicate(v) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (iv *{{.StructName}}) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	for v := range iv.All() {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (iv *{{.StructName}}) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	for v := range iv.All() {
		if predicate(v) {
			return false
		}
	}
	return true
}

// ToSlice returns all elements as a slice.
func (iv *{{.StructName}}) ToSlice() []{{.GoType}} {
	n := iv.Len()
	result := make([]{{.GoType}}, 0, n)
	for v := range iv.All() {
		result = append(result, v)
	}
	return result
}

// Reversed returns a new interval with elements in reverse order.
func (iv *{{.StructName}}) Reversed() *{{.StructName}} {
	if iv.step == {{.MinStepExpr}} {
		panic("interval.{{.StructName}}: cannot reverse interval with minimum step")
	}
	return &{{.StructName}}{from: iv.to, to: iv.from, step: -iv.step}
}

// String returns a string representation of the interval.
func (iv *{{.StructName}}) String() string {
	n := iv.Len()
	if n == 0 {
		return "[]"
	}
	var sb strings.Builder
	sb.WriteString("[")
	first := true
	for v := range iv.All() {
		if !first {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v", v)
		first = false
	}
	sb.WriteString("]")
	return sb.String()
}
`
