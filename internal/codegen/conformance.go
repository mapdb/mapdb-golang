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
	// WantExpr is an independent oracle: a Go slice literal spelling out the
	// fixture's expected All() output. Set it for families that implement
	// ToSlice() by ranging All() — there the All ≡ ToSlice law is tautological,
	// so the literal is stamped instead. Empty means the ToSlice cross-check is
	// meaningful and is stamped as usual.
	WantExpr string
}

// confData drives the conformance-test template for one family.
type confData struct {
	Package     string     // package directory / clause stem, e.g. "arraylist"
	Import      string     // full import path of the family package
	Types       []confType // instances to stamp, in canonical order
	HasSegments bool       // family exposes Segments(n) ⇒ also stamp the partition law
}

// genConformanceTest stamps conformance_generated_test.go into the current
// working directory (the family package, when run from its go:generate
// directive). It is the internal/codegen output for the manifest-driven
// conformance laws of todo 14 §4: one stamped test per family × instance, with
// the law logic itself living once in internal/conformance.
//
// Each row's fixture instance must expose All() iter.Seq[T] and ToSlice() []T;
// when hasSegments is set it must also expose Segments(n) []iter.Seq[T].
func genConformanceTest(pkg string, types []confType, hasSegments bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	tmpl := parse("conformance-"+pkg, conformanceTestTmpl)
	data := confData{
		Package:     pkg,
		Import:      "github.com/mapdb/mapdb-golang/" + pkg,
		Types:       types,
		HasSegments: hasSegments,
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
func genConformanceForOfTypes(pkg string, ordered bool, names, goTypes []string, skip map[string]bool, hasSegments bool) error {
	rows := make([]confType, 0, len(names))
	for i := range names {
		if skip[goTypes[i]] {
			continue
		}
		rows = append(rows, ofRow(pkg, names[i], goTypes[i], ordered))
	}
	return genConformanceTest(pkg, rows, hasSegments)
}

// genConformanceForPrimitives stamps law-1 tests for a family whose element
// types are exactly Primitives() (7 numeric/char types, no bool) and which
// exposes the variadic <TypeName>Of constructor. ordered selects the law-1 mode.
func genConformanceForPrimitives(pkg string, ordered, hasSegments bool) error {
	return genConformanceForPrimitivesOracle(pkg, ordered, hasSegments, nil)
}

// genConformanceForPrimitivesOracle is genConformanceForPrimitives with an
// independent expected-output oracle. Pass a non-nil wantFn for a family that
// implements ToSlice() by ranging All(): the All ≡ ToSlice law is tautological
// there, so wantFn(goType) supplies the slice literal the fixture must produce.
func genConformanceForPrimitivesOracle(pkg string, ordered, hasSegments bool, wantFn func(goType string) string) error {
	ps := Primitives()
	rows := make([]confType, 0, len(ps))
	for _, p := range ps {
		row := ofRow(pkg, p.Name, p.GoType, ordered)
		if wantFn != nil {
			row.WantExpr = wantFn(p.GoType)
		}
		rows = append(rows, row)
	}
	return genConformanceTest(pkg, rows, hasSegments)
}

// sortedSetFixtureExpr renders the shared law-1 fixture as the slice a SORTED,
// duplicate-collapsing family must yield: {3,1,4,1,5,9,2} → {1,2,3,4,5,9}. It is
// written out independently of any collection code, which is the whole point —
// it cannot be corrupted by the same bug as the implementation.
func sortedSetFixtureExpr(goType string) string {
	vals := []string{"1", "2", "3", "4", "5", "9"}
	for i, v := range vals {
		vals[i] = goType + "(" + v + ")"
	}
	return "[]" + goType + "{" + join(vals, ", ") + "}"
}

// mapConfType is one row in a key/value family's stamped conformance test: a
// concrete monomorphized map to build and check. MapName is the <pkg>.<MapName>
// stem (also the test-func suffix); Puts are the fixture's constructor arguments
// (already typed); Ordered adds the KeysAscending law for sorted maps.
type mapConfType struct {
	MapName string   // Int32Int32 — the <pkg>.<MapName> stem
	Puts    []string // fixture put args, e.g. "int32(3), int32(0)"
	Ordered bool     // sorted map ⇒ also stamp the KeysAscending law
}

// mapConfData drives the map conformance-test template for one family.
type mapConfData struct {
	Package     string
	Import      string
	Types       []mapConfType
	HasSegments bool // family exposes Segments2(n) ⇒ also stamp the pair partition law
}

// genMapConformanceTest stamps conformance_generated_test.go for a key/value
// family exposing New<MapName>(), Put(K,V), Len() and All() iter.Seq2[K,V]. It
// stamps the size-accounting law (Len ≡ |All|) for every map and, for ordered
// maps, the KeysAscending law.
func genMapConformanceTest(pkg string, types []mapConfType, hasSegments bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	tmpl := parse("conformance-map-"+pkg, mapConformanceTestTmpl)
	data := mapConfData{
		Package:     pkg,
		Import:      "github.com/mapdb/mapdb-golang/" + pkg,
		Types:       types,
		HasSegments: hasSegments,
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("execute map conformance %s: %w", pkg, err)
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("format map conformance %s: %w\n---\n%s", pkg, err, buf.String())
	}
	out := filepath.Join(cwd, "conformance_generated_test.go")
	return os.WriteFile(out, formatted, 0o644)
}

// mapFixturePuts renders the shared map fixture as typed Put arguments: seven
// DISTINCT keys (so none overwrite → Len is 7 and the ascending check is
// non-trivial) with positional values. Keys stay < 100 so they are valid and
// distinct for every numeric/char key type.
func mapFixturePuts(keyType, valType string) []string {
	keys := []string{"3", "1", "4", "5", "9", "2", "6"}
	vals := []string{"0", "1", "2", "3", "4", "5", "6"}
	puts := make([]string, len(keys))
	for i := range keys {
		puts[i] = keyType + "(" + keys[i] + "), " + valType + "(" + vals[i] + ")"
	}
	return puts
}

// primitivePairRows builds the 49 prim×prim map conformance rows whose type stem
// is <Key><Val> and whose constructor is New<Key><Val>(). suffix is appended to
// the stem (empty for the plain maps, "BiMap" for the bidirectional variants,
// which share the same distinct-key/distinct-value fixture — a valid bijection).
func primitivePairRows(ordered bool, suffix string) []mapConfType {
	ps := Primitives()
	rows := make([]mapConfType, 0, len(ps)*len(ps))
	for _, k := range ps {
		for _, v := range ps {
			rows = append(rows, mapConfType{
				MapName: k.Name + v.Name + suffix,
				Puts:    mapFixturePuts(k.GoType, v.GoType),
				Ordered: ordered,
			})
		}
	}
	return rows
}

// genMapConformanceForPairs stamps map conformance for every (key, value) pair
// over Primitives() × Primitives() — the 49 monomorphized map variants. ordered
// marks a sorted map family (adds the KeysAscending law); hasSegments marks a
// family exposing Segments2(n) (adds the pair partition law).
func genMapConformanceForPairs(pkg string, ordered, hasSegments bool) error {
	return genMapConformanceTest(pkg, primitivePairRows(ordered, ""), hasSegments)
}

// multimapFixturePuts renders the multimap size-accounting fixture: six pairs
// over three DISTINCT keys (1×2, 3×3) so Len (6) exceeds the distinct-key count
// (3) and the law genuinely counts every value. All six (key,value) pairs are
// distinct, so the Set variant keeps all six (none collapses). Keys 1..3 and
// values 10..60 are valid and distinct for every numeric/char primitive.
func multimapFixturePuts(keyType, valType string) []string {
	keys := []string{"1", "1", "2", "3", "3", "3"}
	vals := []string{"10", "20", "30", "40", "50", "60"}
	puts := make([]string, len(keys))
	for i := range keys {
		puts[i] = keyType + "(" + keys[i] + "), " + valType + "(" + vals[i] + ")"
	}
	return puts
}

// genMultimapConformance stamps the size-accounting law (Len ≡ |All|) for the 49
// prim×prim List multimaps AND the 49 Set multimaps. Multimaps are unordered
// (map-backed keys) and their keys repeat, so only the size-accounting law
// applies — but on a repeated-key fixture it meaningfully counts every value.
func genMultimapConformance() error {
	ps := Primitives()
	rows := make([]mapConfType, 0, len(ps)*len(ps)*2)
	for _, k := range ps {
		for _, v := range ps {
			puts := multimapFixturePuts(k.GoType, v.GoType)
			rows = append(rows, mapConfType{MapName: k.Name + v.Name + "List", Puts: puts, Ordered: false})
			rows = append(rows, mapConfType{MapName: k.Name + v.Name + "Set", Puts: puts, Ordered: false})
		}
	}
	return genMapConformanceTest("multimap", rows, false)
}

const mapConformanceTestTmpl = genHeader + `package {{.Package}}_test

import (
	"testing"

	"{{.Import}}"
	"github.com/mapdb/mapdb-golang/internal/conformance"
)
{{$pkg := .Package}}
{{- range .Types}}

func buildConformance{{.MapName}}() *{{$pkg}}.{{.MapName}} {
	m := {{$pkg}}.New{{.MapName}}()
{{- range .Puts}}
	m.Put({{.}})
{{- end}}
	return m
}

// TestConformanceLen2{{.MapName}} pins the size-accounting law (todo 14 §4):
// Len() equals the number of pairs All() yields.
func TestConformanceLen2{{.MapName}}(t *testing.T) {
	m := buildConformance{{.MapName}}()
	conformance.Len2MatchesAll(t, m.Len(), m.All())
}
{{- if .Ordered}}

// TestConformanceKeysAscending{{.MapName}} pins the sorted-map ordering law
// (todo 14 §4): All() yields keys in strictly ascending order.
func TestConformanceKeysAscending{{.MapName}}(t *testing.T) {
	m := buildConformance{{.MapName}}()
	conformance.KeysAscending(t, m.All())
}
{{- end}}
{{- if $.HasSegments}}

// TestConformanceSegments2{{.MapName}} pins the pair partition law (todo 14 §4):
// concat(Segments2(n)) reproduces the All() key→value map with no key split
// across segments, and each segment is re-runnable, for n ∈ {1, 2, 7, len+1}.
func TestConformanceSegments2{{.MapName}}(t *testing.T) {
	m := buildConformance{{.MapName}}()
	conformance.Segments2CoverAll(t, m.All(), m.Segments2)
}
{{- end}}
{{- end}}
`

const conformanceTestTmpl = genHeader + `package {{.Package}}_test

import (
	"testing"

	"{{.Import}}"
	"github.com/mapdb/mapdb-golang/internal/conformance"
)
{{- range .Types}}

{{- if .WantExpr}}
// TestConformanceAllMatchesToSlice{{.TypeName}} pins law 1 (todo 14 §4) against
// an INDEPENDENT oracle. This family builds ToSlice() by ranging All(), so the
// All ≡ ToSlice form would hold no matter how All() broke; the expected
// elements are spelled out instead.
func TestConformanceAllMatchesToSlice{{.TypeName}}(t *testing.T) {
	c := {{.CtorExpr}}
	conformance.AllMatchesLiteral(t, c.All(), {{.WantExpr}})
	// ToSlice() is held to the same oracle, so if it ever becomes an
	// independent implementation it cannot drift from All() unnoticed.
	conformance.SliceMatchesLiteral(t, c.ToSlice(), {{.WantExpr}})
}
{{- else}}
// TestConformanceAllMatchesToSlice{{.TypeName}} pins law 1 (todo 14 §4):
// iterating All() yields the same elements as ToSlice(){{if .Ordered}}, in the
// family's documented iteration order{{else}} as a multiset (unordered family){{end}}.
func TestConformanceAllMatchesToSlice{{.TypeName}}(t *testing.T) {
	c := {{.CtorExpr}}
	conformance.AllMatchesToSlice(t, c.All(), c.ToSlice(), {{.Ordered}})
}
{{- end}}
{{- if $.HasSegments}}

// TestConformanceSegments{{.TypeName}} pins the Segments partition law (todo
// 14 §4): concat(Segments(n)) ≡ All() as a multiset and each segment is
// re-runnable, for n ∈ {1, 2, 7, len+1}.
func TestConformanceSegments{{.TypeName}}(t *testing.T) {
	c := {{.CtorExpr}}
	conformance.SegmentsCoverAll(t, c.All(), c.Segments)
}
{{- end}}
{{- end}}
`
