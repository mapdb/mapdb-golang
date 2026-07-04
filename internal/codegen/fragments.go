package main

import "text/template"

// fragments is the shared library of reusable template bodies, prepended to
// every family template at parse time by parse(). Each fragment is a
// {{define "name"}}…{{end}} block a family body invokes with
// {{template "name" .}}. A fragment emits nothing unless invoked, so adding a
// fragment here cannot change the output of a template that does not reference
// it — the codegen drift gate proves this on every change.
//
// Fragments are the mechanism that turns a cross-cutting method from N per-family
// hand-pastes into a one-place edit (see todo/fable-golang/14 §1a). They are
// parameterized by the shared data contract every per-family struct exposes:
//
//	.Recv    receiver variable (l, s, q, m, …)
//	.Name    concrete type identifier stem (Int32, Float32, Char, …)
//	.GoType  element Go type (int32, float32, uint16, …)
//	.IsFloat true for float elements (selects bit-pattern equality)
//	.BitsFn  math.FloatNNbits for the element width (floats only)
//
// A fragment must only reference fields present on every struct that invokes it;
// storage-shape-specific fragments (slice vs ring vs tree) are named with their
// shape suffix so they are only invoked by families of that shape.
const fragments = `
{{- define "contains_slice" -}}
func ({{.Recv}} *{{.Name}}) Contains(value {{.GoType}}) bool {
	for _, v := range {{.Recv}}.items {
		if {{if .IsFloat}}{{.BitsFn}}(v) == {{.BitsFn}}(value){{else}}v == value{{end}} {
			return true
		}
	}
	return false
}
{{- end -}}
`

// parse builds a template named name whose body is body, with the shared
// fragment library available for {{template "…"}} invocation. It replaces the
// bare template.Must(template.New(name).Parse(body)) idiom the generators used
// so every family sees the same fragment set. Because fragments contains only
// {{define}} blocks, prepending it never adds output to a body that invokes no
// fragment.
func parse(name, body string) *template.Template {
	return template.Must(template.New(name).Parse(fragments + body))
}
