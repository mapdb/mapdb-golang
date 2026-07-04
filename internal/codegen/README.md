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
//go:generate go run ../internal/codegen <subcommand>
```

Regenerate everything, and run the drift gate, with:

```sh
go generate ./...
scripts/check-codegen.sh   # delete-then-regenerate, fail on any diff/stray/orphan
```

### Subcommands

Per-family generators (each emits per-primitive sources into its package):
`arraylist`, `stack`, `deque`, `priorityqueue`, `interval`, `hashset`,
`treeset`, `hashmap`, `sentinelhashmap`, `treemap`, `multimap`, `bag`,
`tuple`. The families that need IEEE float comparison (`arraylist`,
`priorityqueue`, `treeset`, `treemap`, `bag`, `tuple`) also emit a
`cmp_float.go` from the single shared `genCmpFloat` template, so the total-order
comparator has one source of truth.

Two non-family generators, both wired from `collection/doc.go`:

- `interfaces` — renders `collection/<prim>_interfaces.go` (the composable
  interface vocabulary) from one template.
- `matrix` — renders `collection/FAMILY_MATRIX.md` from the manifest.

## Architecture

- **`main.go`** — dispatch is a `generators` map (subcommand → generator
  function), not a switch.
- **`primitives.go`** — `Primitives()` is the canonical 7-primitive set
  (int8/16/32/64, char=uint16, float32/64) every generator iterates.
  Per-family deviations are deliberately local to the generator that needs
  them: `interval` skips char and stubs floats; `hashset` appends its own
  `bool` entry (adding `bool` to the shared set would make the other families
  drift). `MinStepExpr` is the sole helper on `Primitive` today.
- **`fragments.go`** — `parse(name, body)` prepends a shared `const fragments`
  library of `{{define}}` blocks, so a cross-cutting method body lives in one
  place and is invoked with `{{template "name" .}}`. Fragments are
  parameterized by a small opt-in contract (`.Recv`, `.Name`, `.GoType`, …); a
  `{{define}}` that no body invokes emits nothing, so adding one is
  output-neutral. Shape-suffixed names (`contains_slice`) keep a fragment to
  families of the matching storage shape.
- **`manifest.go`** — `Families` is the declarative table of the 13 families
  (storage, order, type coverage, which Immutable/Synchronized/extra variants
  exist). It renders `FAMILY_MATRIX.md` and is the single source of truth for
  the family set. `manifest_test.go` keeps it honest: the family set must match
  across the manifest, the `generators` registry, and the `//go:generate`
  directives, and the Immutable/Synchronized booleans are checked against the
  presence of `immutable_*.go` / `synchronized_*.go` files.

## How to add a new collection family

1. Add `internal/codegen/<family>.go` with a `gen<Family>()` function. Use
   `arraylist.go` (single-type) or `treemap.go` (K×V) as a model; call
   `genCmpFloat("<family>")` at the end if it needs float ordering. Build
   templates with `parse(...)` and reuse fragments where a body is shared.
2. Add it to the `generators` map in `main.go`.
3. Add a row to `Families` in `manifest.go` (drives the matrix + the guard).
4. Drop a `doc.go` into the target package with the `//go:generate` directive.
5. Extend `primitives.go` if a template needs new per-primitive metadata.

`manifest_test.go` fails until steps 2–4 agree, so none can be forgotten.

## Scope

This generator only emits production sources. Test files are
hand-maintained — they exist to validate the generated code from a user's
point of view, and embedding them in the template would force template
churn whenever a test is refined.
