package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
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

	applies := template.Must(template.New("interval").Parse(intervalTmpl))
	stub := template.Must(template.New("interval-stub").Parse(intervalStubTmpl))

	for _, p := range Primitives() {
		if p.IsChar {
			continue
		}
		data := struct {
			StructName  string
			GoType      string
			MinStepExpr string
		}{
			StructName:  p.Name + "Interval",
			GoType:      p.GoType,
			MinStepExpr: p.MinStepExpr(),
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
	return nil
}

const intervalStubTmpl = `package interval

// Interval is not applicable to {{.GoType}}.
`

const intervalTmpl = `package interval

import (
	"fmt"
	"iter"
	"strings"
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
		panic("{{.StructName}}: step must not be zero")
	}
	if from < to && step < 0 {
		panic("{{.StructName}}: step must be positive when from < to")
	}
	if from > to && step > 0 {
		panic("{{.StructName}}: step must be negative when from > to")
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

// Size returns the number of elements in the interval.
func (iv *{{.StructName}}) Size() int {
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

// IsEmpty returns true if the interval contains no elements.
func (iv *{{.StructName}}) IsEmpty() bool { return iv.Size() == 0 }

// Contains returns true if the interval contains the given value.
func (iv *{{.StructName}}) Contains(value {{.GoType}}) bool {
	if iv.step > 0 {
		return value >= iv.from && value <= iv.to && (uint64(int64(value))-uint64(int64(iv.from)))%iv.absStep() == 0
	}
	return value <= iv.from && value >= iv.to && (uint64(int64(iv.from))-uint64(int64(value)))%iv.absStep() == 0
}

// Get returns the element at the given index, or an error if out of bounds.
func (iv *{{.StructName}}) Get(index int) ({{.GoType}}, error) {
	if index < 0 || index >= iv.Size() {
		return 0, fmt.Errorf("{{.StructName}}: index out of bounds: %d (size %d)", index, iv.Size())
	}
	return {{.GoType}}(int64(iv.from) + int64(iv.step)*int64(index)), nil
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
		size := iv.Size()
		for i := 0; i < size; i++ {
			value, err := iv.Get(i)
			if err != nil {
				return
			}
			if !yield(value) {
				return
			}
		}
	}
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
	n := iv.Size()
	result := make([]{{.GoType}}, 0, n)
	for v := range iv.All() {
		result = append(result, v)
	}
	return result
}

// Reversed returns a new interval with elements in reverse order.
func (iv *{{.StructName}}) Reversed() *{{.StructName}} {
	if iv.step == {{.MinStepExpr}} {
		panic("{{.StructName}}: cannot reverse interval with minimum step")
	}
	return &{{.StructName}}{from: iv.to, to: iv.from, step: -iv.step}
}

// String returns a string representation of the interval.
func (iv *{{.StructName}}) String() string {
	n := iv.Size()
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
