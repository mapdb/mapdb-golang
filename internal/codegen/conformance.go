package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
)

// confType is one row in a family's stamped conformance test: a concrete
// monomorphized instance to build and check. TypeName is the test-function
// suffix (must be unique within the family's file); CtorExpr is the full Go
// expression that builds a fixture instance; Ordered selects law 1's comparison
// mode for this instance (element-for-element vs multiset). Carrying the order
// per row lets a single family mix ordered and unordered instances — e.g. bag,
// whose hash-backed variant is unordered and tree-backed variant is sorted.
type confType struct {
	TypeName string // Int32, HashInt32, TreeInt32 — the test-func suffix
	CtorExpr string // full expression building the fixture, e.g. arraylist.Int32Of(int32(3), …)
	Ordered  bool   // law-1 mode: All() order == ToSlice() order
}

// confData drives the conformance-test template for one family.
type confData struct {
	Package string     // package directory / clause stem, e.g. "arraylist"
	Import  string     // full import path of the family package
	Types   []confType // instances to stamp, in canonical order
}

// genConformanceTest stamps conformance_generated_test.go into the current
// working directory (the family package, when run from its go:generate
// directive). It is the internal/codegen output for the manifest-driven
// conformance laws of todo 14 §4: one stamped test per family × instance, with
// the law logic itself living once in internal/conformance.
//
// Each row's fixture instance must expose All() iter.Seq[T] and ToSlice() []T.
func genConformanceTest(pkg string, types []confType) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	tmpl := parse("conformance-"+pkg, conformanceTestTmpl)
	data := confData{
		Package: pkg,
		Import:  "github.com/mapdb/mapdb-golang/" + pkg,
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

// fixtureArgs renders the shared law-1 fixture as typed literals for goType. The
// values carry a duplicate (1 appears twice) and stay within int8 range (all <
// 100), so the SAME list is valid for every numeric and char element type. A set
// collapses the duplicate; a list/bag keeps it — law 1 compares All() against
// ToSlice() of the same instance, so it holds either way, and the distinct
// values still catch a reordered or dropped element in an ordered family.
func fixtureArgs(goType string) string {
	vals := []string{"3", "1", "4", "1", "5", "9", "2"}
	for i, v := range vals {
		vals[i] = goType + "(" + v + ")"
	}
	return join(vals, ", ")
}

// join concatenates parts with sep (avoids importing strings for one call).
func join(parts []string, sep string) string {
	var b bytes.Buffer
	for i, p := range parts {
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(p)
	}
	return b.String()
}

// ofCtorExpr builds the constructor expression for a family exposing the
// variadic <TypeName>Of constructor loaded with the shared fixture.
func ofCtorExpr(pkg, typeName, goType string) string {
	return pkg + "." + typeName + "Of(" + fixtureArgs(goType) + ")"
}

// ofRow builds a conformance row for a <pkg>.<TypeName>Of-constructed instance.
func ofRow(pkg, typeName, goType string, ordered bool) confType {
	return confType{TypeName: typeName, CtorExpr: ofCtorExpr(pkg, typeName, goType), Ordered: ordered}
}

// genConformanceForOfTypes stamps law-1 tests for a family whose instances are
// built via <pkg>.<TypeName>Of, over the given (name, goType) rows sharing a
// single order class. Element types whose GoType is in skip (e.g. bool, whose
// two-value domain makes the numeric fixture degenerate) are dropped.
func genConformanceForOfTypes(pkg string, ordered bool, names, goTypes []string, skip map[string]bool) error {
	rows := make([]confType, 0, len(names))
	for i := range names {
		if skip[goTypes[i]] {
			continue
		}
		rows = append(rows, ofRow(pkg, names[i], goTypes[i], ordered))
	}
	return genConformanceTest(pkg, rows)
}

// genConformanceForPrimitives stamps law-1 tests for a family whose element
// types are exactly Primitives() (7 numeric/char types, no bool) and which
// exposes the variadic <TypeName>Of constructor. ordered selects the law-1 mode.
func genConformanceForPrimitives(pkg string, ordered bool) error {
	ps := Primitives()
	names := make([]string, len(ps))
	goTypes := make([]string, len(ps))
	for i, p := range ps {
		names[i] = p.Name
		goTypes[i] = p.GoType
	}
	return genConformanceForOfTypes(pkg, ordered, names, goTypes, nil)
}

const conformanceTestTmpl = genHeader + `package {{.Package}}_test

import (
	"testing"

	"{{.Import}}"
	"github.com/mapdb/mapdb-golang/internal/conformance"
)
{{- range .Types}}

// TestConformanceAllMatchesToSlice{{.TypeName}} pins law 1 (todo 14 §4):
// iterating All() yields the same elements as ToSlice(){{if .Ordered}}, in the
// family's documented iteration order{{else}} as a multiset (unordered family){{end}}.
func TestConformanceAllMatchesToSlice{{.TypeName}}(t *testing.T) {
	c := {{.CtorExpr}}
	conformance.AllMatchesToSlice(t, c.All(), c.ToSlice(), {{.Ordered}})
}
{{- end}}
`
