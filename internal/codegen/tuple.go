package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

// pairData is the per-combination view the prim×prim pair template iterates
// over.
//
// The tuple family is an immutable two-element pair. The prim×prim shape
// (7×7 = 49 files) exposes the full surface: One/Two/Equals/String/CompareTo/
// Swap. Only two bits of logic are type-dependent:
//
//   - Equals compares each field independently. A float field is compared by
//     IEEE 754 bit pattern (OneBitsFn / TwoBitsFn, math.Float{32,64}bits); an
//     int/char field is compared with ==. The two fields branch independently
//     on their own OneIsFloat / TwoIsFloat.
//   - The "math" import is needed iff (OneIsFloat || TwoIsFloat).
//
// CompareTo orders the two fields lexicographically (first, then second). A
// float field is compared via the shared IEEE total-order helper
// (OneCmpFn / TwoCmpFn, cmpFloat32/64) so the ordering is a true total order
// (NaN sorts to the top, ±0 distinguished, transitive); an int/char field uses
// raw < / >. Each field branches independently on its own OneIsFloat /
// TwoIsFloat. The shared cmp_float.go is emitted in the tuple package.
//
// Swap returns the TRANSPOSED pair: an <One><Two>Pair.Swap() returns a
// <Two><One>Pair via New<Two><One>Pair(p.two, p.one). The transposed
// identifiers are SwapName (the <Two><One> stem).
type pairData struct {
	OneName string // Int32, Float32, Char (first element identifier stem)
	OneType string // int32, float32, uint16 (Go type of first element)
	TwoName string // Int32, Float32, Char (second element identifier stem)
	TwoType string // int32, float32, uint16 (Go type of second element)

	// PairName / PairSnake are the combined identifiers, e.g. Int32Float32 /
	// int32_float32 (used for Int32Float32Pair and the file name).
	PairName  string
	PairSnake string

	// SwapName is the transposed (Two,One) identifier stem, e.g. Float32Int32
	// for an Int32Float32Pair. Swap() returns a <SwapName>Pair.
	SwapName string

	// OneIsFloat / TwoIsFloat select bit-pattern equality per field and
	// (together) gate the math import.
	OneIsFloat bool
	TwoIsFloat bool
	OneBitsFn  string // math.Float32bits / math.Float64bits (float first field)
	TwoBitsFn  string // math.Float32bits / math.Float64bits (float second field)
	OneCmpFn   string // cmpFloat32 / cmpFloat64 (float first field, CompareTo)
	TwoCmpFn   string // cmpFloat32 / cmpFloat64 (float second field, CompareTo)

	// NeedsMath drives the import block: true iff either field is float.
	NeedsMath bool
}

// objPairData is the per-type view the object pair templates iterate over.
//
// Object pairs carry a generic element T that is neither comparable nor
// ordered, so they expose a REDUCED surface: One/Two/String only — no Equals,
// CompareTo, or Swap. There are two shapes, each 7 files:
//
//   - Shape A — Object<Two>Pair[T any] (object first, prim second): the prim is
//     the second element (the payload).
//   - Shape B — <One>ObjectPair[T any] (prim first, object second): the prim is
//     the first element.
//
// The prim half is a pure payload — Equals is absent, so there is no float
// bit-pattern branch; object_int32 and object_float32 differ only by the prim
// Go type and the doc-comment text.
type objPairData struct {
	PrimName  string // Int32, Float32, Char
	PrimType  string // int32, float32, uint16
	PairName  string // ObjectInt32 / Int32Object
	PairSnake string // object_int32 / int32_object
}

// genTuple writes the 63 pair sources into the current working directory:
// 49 prim×prim <One><Two>Pair, 7 Object<Two>Pair[T any], and 7
// <One>ObjectPair[T any] (object/object is excluded). Invoked from the
// tuple package's go:generate directive.
func genTuple() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	pairTmpl := template.Must(template.New("pair").Parse(pairTmpl))
	objKeyTmpl := template.Must(template.New("objkey-pair").Parse(objectKeyPairTmpl))
	objValTmpl := template.Must(template.New("objval-pair").Parse(objectValuePairTmpl))

	write := func(name string, tmpl *template.Template, data any) error {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("execute %s: %w", name, err)
		}
		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			return fmt.Errorf("format %s: %w\n---\n%s", name, err, buf.String())
		}
		return os.WriteFile(filepath.Join(cwd, name), formatted, 0o644)
	}

	bitsFn := func(p Primitive) string {
		if p.ByteSize == 8 {
			return "math.Float64bits"
		}
		return "math.Float32bits"
	}

	cmpFn := func(p Primitive) string {
		if p.ByteSize == 8 {
			return "cmpFloat64"
		}
		return "cmpFloat32"
	}

	prims := Primitives()

	// prim×prim pairs (49).
	for _, one := range prims {
		for _, two := range prims {
			data := pairData{
				OneName:    one.Name,
				OneType:    one.GoType,
				TwoName:    two.Name,
				TwoType:    two.GoType,
				PairName:   one.Name + two.Name,
				PairSnake:  one.SnakeName + "_" + two.SnakeName,
				SwapName:   two.Name + one.Name,
				OneIsFloat: one.IsFloating,
				TwoIsFloat: two.IsFloating,
				NeedsMath:  one.IsFloating || two.IsFloating,
			}
			if one.IsFloating {
				data.OneBitsFn = bitsFn(one)
				data.OneCmpFn = cmpFn(one)
			}
			if two.IsFloating {
				data.TwoBitsFn = bitsFn(two)
				data.TwoCmpFn = cmpFn(two)
			}
			if err := write(data.PairSnake+"_pair.go", pairTmpl, data); err != nil {
				return err
			}
		}
	}

	// Object pairs (7 + 7); object/object excluded.
	for _, p := range prims {
		a := objPairData{
			PrimName:  p.Name,
			PrimType:  p.GoType,
			PairName:  "Object" + p.Name,
			PairSnake: "object_" + p.SnakeName,
		}
		if err := write(a.PairSnake+"_pair.go", objKeyTmpl, a); err != nil {
			return err
		}

		b := objPairData{
			PrimName:  p.Name,
			PrimType:  p.GoType,
			PairName:  p.Name + "Object",
			PairSnake: p.SnakeName + "_object",
		}
		if err := write(b.PairSnake+"_pair.go", objValTmpl, b); err != nil {
			return err
		}
	}

	// Emit the shared cmp_float.go so float pairs' CompareTo can route
	// through cmpFloat32/64. It is generated unconditionally for the tuple
	// package; pairs with no float field simply never reference it (a
	// package-level func may be unused by some files in the package).
	return genCmpFloat("tuple")
}

const pairTmpl = genHeader + `package tuple

import (
	"fmt"
{{- if .NeedsMath}}
	"math"
{{- end}}
)

// {{.PairName}}Pair is an immutable pair of ({{.OneType}}, {{.TwoType}}).
type {{.PairName}}Pair struct {
	one {{.OneType}}
	two {{.TwoType}}
}

// New{{.PairName}}Pair creates a new {{.PairName}}Pair.
func New{{.PairName}}Pair(one {{.OneType}}, two {{.TwoType}}) {{.PairName}}Pair {
	return {{.PairName}}Pair{one: one, two: two}
}

// One returns the first element.
func (p {{.PairName}}Pair) One() {{.OneType}} {
	return p.one
}

// Two returns the second element.
func (p {{.PairName}}Pair) Two() {{.TwoType}} {
	return p.two
}

// Equals returns true if both elements are equal.
func (p {{.PairName}}Pair) Equals(other {{.PairName}}Pair) bool {
	return {{if .OneIsFloat}}{{.OneBitsFn}}(p.one) == {{.OneBitsFn}}(other.one){{else}}p.one == other.one{{end}} && {{if .TwoIsFloat}}{{.TwoBitsFn}}(p.two) == {{.TwoBitsFn}}(other.two){{else}}p.two == other.two{{end}}
}

// String returns a string representation.
func (p {{.PairName}}Pair) String() string {
	return fmt.Sprintf("(%v, %v)", p.one, p.two)
}

// CompareTo compares this pair to another. Compares first element first,
// then second element if first elements are equal.
// Returns negative if p < other, zero if equal, positive if p > other.
func (p {{.PairName}}Pair) CompareTo(other {{.PairName}}Pair) int {
{{- if .OneIsFloat}}
	if c := {{.OneCmpFn}}(p.one, other.one); c != 0 {
		return c
	}
{{- else}}
	if p.one < other.one {
		return -1
	}
	if p.one > other.one {
		return 1
	}
{{- end}}
{{- if .TwoIsFloat}}
	if c := {{.TwoCmpFn}}(p.two, other.two); c != 0 {
		return c
	}
{{- else}}
	if p.two < other.two {
		return -1
	}
	if p.two > other.two {
		return 1
	}
{{- end}}
	return 0
}

// Swap returns a new pair with elements swapped: (two, one).
func (p {{.PairName}}Pair) Swap() {{.SwapName}}Pair {
	return New{{.SwapName}}Pair(p.two, p.one)
}
`

const objectKeyPairTmpl = genHeader + `package tuple

import (
	"fmt"
)

// {{.PairName}}Pair is an immutable pair of (T, {{.PrimType}}) where T is any type.
type {{.PairName}}Pair[T any] struct {
	one T
	two {{.PrimType}}
}

// New{{.PairName}}Pair creates a new {{.PairName}}Pair.
func New{{.PairName}}Pair[T any](one T, two {{.PrimType}}) {{.PairName}}Pair[T] {
	return {{.PairName}}Pair[T]{one: one, two: two}
}

// One returns the first element (object).
func (p {{.PairName}}Pair[T]) One() T {
	return p.one
}

// Two returns the second element (primitive).
func (p {{.PairName}}Pair[T]) Two() {{.PrimType}} {
	return p.two
}

// String returns a string representation.
func (p {{.PairName}}Pair[T]) String() string {
	return fmt.Sprintf("(%v, %v)", p.one, p.two)
}
`

const objectValuePairTmpl = genHeader + `package tuple

import (
	"fmt"
)

// {{.PairName}}Pair is an immutable pair of ({{.PrimType}}, T) where T is any type.
type {{.PairName}}Pair[T any] struct {
	one {{.PrimType}}
	two T
}

// New{{.PairName}}Pair creates a new {{.PairName}}Pair.
func New{{.PairName}}Pair[T any](one {{.PrimType}}, two T) {{.PairName}}Pair[T] {
	return {{.PairName}}Pair[T]{one: one, two: two}
}

// One returns the first element (primitive).
func (p {{.PairName}}Pair[T]) One() {{.PrimType}} {
	return p.one
}

// Two returns the second element (object).
func (p {{.PairName}}Pair[T]) Two() T {
	return p.two
}

// String returns a string representation.
func (p {{.PairName}}Pair[T]) String() string {
	return fmt.Sprintf("(%v, %v)", p.one, p.two)
}
`
