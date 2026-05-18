package main

import "fmt"

// Primitive is the metadata each per-primitive template iterates over.
// Templates access fields by name and call MinStepExpr where needed.
type Primitive struct {
	Name       string // capitalised, used in identifiers (Int8, Float32, …)
	GoType     string // Go type expression (int8, float32, …)
	SnakeName  string // lowercase, used in file names (int8, float32, …)
	ByteSize   int    // size in bytes of the primitive
	IsFloating bool
	IsSigned   bool
	IsChar     bool // mapdb's CHAR primitive (uint16 with text semantics)
}

// MinStepExpr returns a Go expression for the minimum value of a signed
// integer primitive (used by intervals to detect the unreversible step).
// Returns "" for primitives where the concept does not apply.
func (p Primitive) MinStepExpr() string {
	if !p.IsSigned || p.IsFloating {
		return ""
	}
	return fmt.Sprintf("%s(-1<<%d)", p.GoType, p.ByteSize*8-1)
}

// Primitives returns the full primitive set supported by mapdb-golang.
// Order is the canonical iteration order for code generation.
func Primitives() []Primitive {
	return []Primitive{
		{Name: "Int8", GoType: "int8", SnakeName: "int8", ByteSize: 1, IsSigned: true},
		{Name: "Int16", GoType: "int16", SnakeName: "int16", ByteSize: 2, IsSigned: true},
		{Name: "Int32", GoType: "int32", SnakeName: "int32", ByteSize: 4, IsSigned: true},
		{Name: "Int64", GoType: "int64", SnakeName: "int64", ByteSize: 8, IsSigned: true},
		{Name: "Char", GoType: "uint16", SnakeName: "char", ByteSize: 2, IsChar: true},
		{Name: "Float32", GoType: "float32", SnakeName: "float32", ByteSize: 4, IsFloating: true},
		{Name: "Float64", GoType: "float64", SnakeName: "float64", ByteSize: 8, IsFloating: true},
	}
}
