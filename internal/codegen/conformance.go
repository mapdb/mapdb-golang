package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
)

// confType is one primitive row in a family's stamped conformance test: the
// concrete monomorphized type stem (Int32 ⇒ <pkg>.Int32, built via <pkg>.Int32Of)
// and the element literal type used for the fixture.
type confType struct {
	TypeName string // Int32, Float64, Char — the <pkg>.<TypeName> stem
	GoType   string // int32, float64, uint16 — the element literal type
}

// confData drives the conformance-test template for one family.
type confData struct {
	Package string     // package directory / clause stem, e.g. "arraylist"
	Import  string     // full import path of the family package
	Ordered bool       // law-1 order sensitivity (true ⇒ All() order == ToSlice() order)
	Types   []confType // primitives to stamp, in canonical order
}

// genConformanceTest stamps conformance_generated_test.go into the current
// working directory (the family package, when run from its go:generate
// directive). It is the first internal/codegen output that is a _test.go file:
// the manifest-driven conformance laws of todo 14 §4, one stamped test per
// family × primitive, with the law logic itself living once in
// internal/conformance. ordered selects the family's law-1 comparison mode
// (element-for-element vs multiset).
//
// Applicable to single-value families exposing the variadic <TypeName>Of
// constructor plus All() iter.Seq[T] and ToSlice() []T. Bool and other
// non-Of-constructed variants are excluded by the caller.
func genConformanceTest(pkg string, ordered bool, types []confType) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	tmpl := parse("conformance-"+pkg, conformanceTestTmpl)
	data := confData{
		Package: pkg,
		Import:  "github.com/mapdb/mapdb-golang/" + pkg,
		Ordered: ordered,
		Types:   types,
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("execute conformance %s: %w", pkg, err)
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("format conformance %s: %w\n---\n%s", pkg, err, buf.String())
	}
	out := filepath.Join(cwd, "conformance_generated_test.go")
	return os.WriteFile(out, formatted, 0o644)
}

// genConformanceForPrimitives stamps law-1 conformance tests for a family whose
// element types are exactly Primitives() (the 7 numeric/char types, no bool) and
// which exposes the variadic <TypeName>Of constructor plus All()/ToSlice().
// ordered selects the law-1 comparison mode (element-for-element vs multiset).
func genConformanceForPrimitives(pkg string, ordered bool) error {
	ps := Primitives()
	names := make([]string, len(ps))
	goTypes := make([]string, len(ps))
	for i, p := range ps {
		names[i] = p.Name
		goTypes[i] = p.GoType
	}
	return genConformanceTest(pkg, ordered, confTypes(names, goTypes, nil))
}

// confTypes projects a primitive slice onto the conformance rows, dropping any
// element type (e.g. bool) whose two-value domain makes the law-1 fixture
// degenerate — skip is matched on GoType.
func confTypes(names, goTypes []string, skip map[string]bool) []confType {
	out := make([]confType, 0, len(names))
	for i := range names {
		if skip[goTypes[i]] {
			continue
		}
		out = append(out, confType{TypeName: names[i], GoType: goTypes[i]})
	}
	return out
}

// The fixture deliberately carries duplicates (1 appears twice) and stays within
// int8 range (all values < 100) so the SAME literal list is valid for every
// numeric and char element type. Sets collapse the duplicate; lists/bags keep
// it — law 1 compares All() against ToSlice() of the same instance, so it holds
// either way. Values are distinct enough that a reordered or dropped element in
// an ordered family fails the element-for-element comparison.
const conformanceTestTmpl = genHeader + `package {{.Package}}_test

import (
	"testing"

	"{{.Import}}"
	"github.com/mapdb/mapdb-golang/internal/conformance"
)
{{$pkg := .Package}}{{$ordered := .Ordered}}
{{- range .Types}}

// TestConformanceAllMatchesToSlice{{.TypeName}} pins law 1 (todo 14 §4):
// iterating All() yields the same elements as ToSlice(){{if $ordered}}, in the
// family's documented iteration order{{else}} as a multiset (unordered family){{end}}.
func TestConformanceAllMatchesToSlice{{.TypeName}}(t *testing.T) {
	c := {{$pkg}}.{{.TypeName}}Of({{.GoType}}(3), {{.GoType}}(1), {{.GoType}}(4), {{.GoType}}(1), {{.GoType}}(5), {{.GoType}}(9), {{.GoType}}(2))
	conformance.AllMatchesToSlice(t, c.All(), c.ToSlice(), {{$ordered}})
}
{{- end}}
`
