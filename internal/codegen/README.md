# mapdb-golang code generator

Minimal in-repo tool that re-emits per-primitive collection sources from
small `text/template`-based templates.

## Why this exists

Go has no method-level generics. Exposing unboxed APIs (e.g.
`Int32ArrayList.Sum() int64`) requires per-primitive types. Rather than
hand-maintain ~1300 near-identical files, this tool re-emits them from a
handful of templates kept right next to the corresponding collection
package.

The tool intentionally has no dependencies beyond the standard library
(`text/template`, `go/format`) and is small enough to be read and modified
without learning a separate DSL.

## How to use

From the target collection's directory:

```go
//go:generate go run ../internal/codegen <collection>
```

Currently supported: `arraylist`, `interval`, `hashset`, `stack`, `deque`,
`treeset`, `priorityqueue`, `bag`, `treemap`. Float-ordered collections
(`arraylist`, `treeset`, `priorityqueue`, `bag`, `treemap`) also emit a
`cmp_float.go` from the single shared `genCmpFloat` template, so the IEEE
total-order comparator has exactly one source of truth.

Run a regeneration with:

```sh
go generate ./...
```

A drift check for CI is just:

```sh
go generate ./... && git diff --exit-code
```

## How to add a new collection

1. Add a new file `internal/codegen/<collection>.go` exporting a
   `gen<Collection>()` function. Use `arraylist.go` (single-type) or
   `treemap.go` (K×V) as a model; call `genCmpFloat("<collection>")` at the
   end if the collection needs float ordering.
2. Add a `case` for it in `main.go`'s switch.
3. Drop a `doc.go` into the target package with a `//go:generate`
   directive.
4. Update `primitives.go` if you need new metadata fields.

The primitive set lives in `primitives.go`. `MinStepExpr` is the only
helper currently exposed on `Primitive`; add more as templates need them.

## Scope

This generator only emits production sources. Test files are
hand-maintained — they exist to validate the generated code from a user's
point of view, and embedding them in the template would force template
churn whenever a test is refined.
