package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
)

// genCmpFloat writes cmp_float.go into the current working directory for the
// given package. The cmpFloat32/cmpFloat64 implementations are byte-identical
// across every collection package that needs a float total order
// (arraylist/bag/priorityqueue/treemap/treeset); this shared template is the
// single canonical source so each package's copy is generated, not
// hand-maintained.
func genCmpFloat(pkg string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	tmpl := parse("cmpfloat", cmpFloatTmpl)
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct{ Package string }{Package: pkg}); err != nil {
		return fmt.Errorf("execute cmp_float.go: %w", err)
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("format cmp_float.go: %w\n---\n%s", err, buf.String())
	}
	out := filepath.Join(cwd, "cmp_float.go")
	return os.WriteFile(out, formatted, 0o644)
}

const cmpFloatTmpl = genHeader + `package {{.Package}}

import "math"

// cmpFloat32 / cmpFloat64 implement the IEEE 754 totalOrder construction
// (bit-identical to Rust's f32::total_cmp / f64::total_cmp), as mandated by
// the collection spec (algorithms.md "Float ordering"):
//
//	-NaN < -Inf < negative finite < -0.0 < +0.0 < positive finite < +Inf < +NaN
//
// A naive ` + "`<`" + ` returns false for any NaN comparison (so NaN never moves during a
// sort and min/max ignore it), and a raw unsigned bit compare is intransitive
// because a negative float's sign bit makes its bit pattern sort above a
// positive NaN. The sign-flip-then-signed-compare trick below is a true total
// order: it keeps NaN sortable, places NaN above +Inf, and distinguishes -0
// from +0.
func cmpFloat32(a, b float32) int {
	ai := int32(math.Float32bits(a))
	bi := int32(math.Float32bits(b))
	// Flip all bits except the sign bit for negatives; flip only the sign bit
	// for non-negatives. Equivalent to ai ^= (uint32(ai>>31) >> 1).
	ai ^= int32(uint32(ai>>31) >> 1)
	bi ^= int32(uint32(bi>>31) >> 1)
	switch {
	case ai < bi:
		return -1
	case ai > bi:
		return 1
	default:
		return 0
	}
}

func cmpFloat64(a, b float64) int {
	ai := int64(math.Float64bits(a))
	bi := int64(math.Float64bits(b))
	ai ^= int64(uint64(ai>>63) >> 1)
	bi ^= int64(uint64(bi>>63) >> 1)
	switch {
	case ai < bi:
		return -1
	case ai > bi:
		return 1
	default:
		return 0
	}
}
`
