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
// parameterized by a shared data contract that a per-family struct opts into by
// adding the fields a fragment it invokes needs (today alData/stData/pqData carry
// .Recv, used by contains_slice and the predicate-query fragments; pqData invokes
// only contains_slice):
//
//	.Recv    receiver variable (l, s, q, m, …)
//	.Name    concrete type identifier stem (Int32, Float32, Char, …)
//	.GoType  element Go type (int32, float32, uint16, …)
//	.IsFloat true for float elements (selects bit-pattern equality)
//	.BitsFn  math.FloatNNbits for the element width (floats only)
//
// A fragment must only reference fields present on every struct that invokes it;
// storage-shape-specific fragments (slice vs ring vs tree) are named with their
// shape suffix so they are only invoked by families of that shape. The
// segments_* fragments emit the Segmenter capability (par.From on-ramp) for
// slice-backed families in their three variants — segments_slice over the live
// backing array, segments_delegate forwarding to an immutable's store, and
// segments_snapshot over a synchronized snapshot; each calls internal/segment,
// so a template invoking one imports "…/internal/segment".
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

{{- define "count_slice" -}}
func ({{.Recv}} *{{.Name}}) Count(predicate func({{.GoType}}) bool) int {
	count := 0
	for _, v := range {{.Recv}}.items {
		if predicate(v) {
			count++
		}
	}
	return count
}
{{- end -}}

{{- define "any_satisfy_slice" -}}
func ({{.Recv}} *{{.Name}}) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range {{.Recv}}.items {
		if predicate(v) {
			return true
		}
	}
	return false
}
{{- end -}}

{{- define "all_satisfy_slice" -}}
func ({{.Recv}} *{{.Name}}) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range {{.Recv}}.items {
		if !predicate(v) {
			return false
		}
	}
	return true
}
{{- end -}}

{{- define "none_satisfy_slice" -}}
func ({{.Recv}} *{{.Name}}) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range {{.Recv}}.items {
		if predicate(v) {
			return false
		}
	}
	return true
}
{{- end -}}

{{- define "segments_slice" -}}
// Segments cuts the elements into up to n balanced, contiguous, non-overlapping
// views (k = min(n, Len), or 1 when empty) whose concatenation covers every
// element exactly once, so a *{{.Name}} satisfies par.Segmenter[{{.GoType}}] and
// feeds par.From directly. Segment order follows the backing array and is not
// guaranteed to match All (the Segmenter contract is unordered). The views are
// live over the backing array: mutating it while a view is consumed is undefined
// behavior. O(1) memory per view.
func ({{.Recv}} *{{.Name}}) Segments(n int) []iter.Seq[{{.GoType}}] {
	return segment.Split({{.Recv}}.items, n)
}
{{- end -}}

{{- define "segments_delegate" -}}
// Segments cuts the elements into up to n balanced, contiguous, non-overlapping
// views covering every element exactly once, satisfying par.Segmenter[{{.GoType}}].
// It delegates to the immutable backing store; the views are stable because an
// immutable value never changes.
func ({{.Recv}} *Immutable{{.Name}}) Segments(n int) []iter.Seq[{{.GoType}}] {
	return {{.Recv}}.delegate.Segments(n)
}
{{- end -}}

{{- define "segments_snapshot" -}}
// Segments cuts a point-in-time snapshot into up to n balanced, contiguous,
// non-overlapping views covering it exactly once, satisfying par.Segmenter[{{.GoType}}].
// The snapshot is taken once under lock; the views iterate it lock-free — the
// same snapshot contract as All.
func ({{.Recv}} *Synchronized{{.Name}}) Segments(n int) []iter.Seq[{{.GoType}}] {
	return segment.Split({{.Recv}}.snapshot(), n)
}
{{- end -}}
`

// parse builds a template named name whose body is body, with the shared
// fragment library available for {{template "…"}} invocation. It replaces the
// bare template.Must(template.New(name).Parse(body)) idiom the generators used
// so every family sees the same fragment set. Because fragments contains only
// {{define}} blocks, prepending it never adds output to a body that invokes no
// fragment.
//
// Concatenating fragments ahead of body in a single Parse (rather than two
// successive Parse calls) is deliberate: it makes a duplicate {{define}} name a
// loud parse-time panic instead of a silent redefinition. The cost is that a
// parse error in a family body reports a line number offset by the fragment
// prefix's length, not the body-relative line.
func parse(name, body string) *template.Template {
	return template.Must(template.New(name).Parse(fragments + body))
}
