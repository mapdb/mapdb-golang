package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
)

// ifaceData is the per-primitive view the collection-interface template
// iterates over. Only the identifier stem and the element Go type vary between
// the seven files.
type ifaceData struct {
	Name   string // Int32, Float32, Char (interface identifier stem)
	GoType string // int32, float32, uint16 (element Go type)
}

// genInterfaces writes collection/<prim>_interfaces.go for every primitive from
// one shared template. These files define the composable interface vocabulary
// (Sized/Iterable/Searchable/… and the List/Set/Bag/Stack categories) and were
// previously seven hand-maintained near-duplicates; generating them keeps the
// vocabulary defined once. Invoked from collection/ via go:generate.
func genInterfaces() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	tmpl := parse("interfaces", collectionInterfacesTmpl)
	for _, p := range Primitives() {
		data := ifaceData{Name: p.Name, GoType: p.GoType}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("execute %s: %w", p.SnakeName, err)
		}
		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			return fmt.Errorf("format %s_interfaces.go: %w\n---\n%s", p.SnakeName, err, buf.String())
		}
		out := filepath.Join(cwd, p.SnakeName+"_interfaces.go")
		if err := os.WriteFile(out, formatted, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", out, err)
		}
	}
	return nil
}

// collectionInterfacesTmpl is the source of the per-primitive interface files.
const collectionInterfacesTmpl = genHeader + `package collection

import (
	"fmt"
	"iter"
)

// ── Composable sub-interfaces ─────────────────────────────────────────
//
// Following Go's io.Reader / io.Writer / io.ReadWriter pattern, the
// primitive-collection API is built from small, single-concern
// interfaces that can be composed as needed.

// {{.Name}}Sized exposes the element count of a collection.
type {{.Name}}Sized interface {
	// Len returns the number of elements. Use x.Len() == 0 to test for
	// emptiness.
	Len() int
}

// {{.Name}}Iterable provides element-by-element traversal.
type {{.Name}}Iterable interface {
	// All returns an iter.Seq that yields all elements.
	All() iter.Seq[{{.GoType}}]

	// ForEach calls the given function for each element.
	ForEach(f func({{.GoType}}))
}

// {{.Name}}Searchable supports membership and predicate queries.
type {{.Name}}Searchable interface {
	// Contains returns true if the collection contains the given value.
	Contains(value {{.GoType}}) bool

	// AnySatisfy returns true if any element satisfies the predicate.
	AnySatisfy(predicate func({{.GoType}}) bool) bool

	// AllSatisfy returns true if all elements satisfy the predicate.
	AllSatisfy(predicate func({{.GoType}}) bool) bool

	// NoneSatisfy returns true if no element satisfies the predicate.
	NoneSatisfy(predicate func({{.GoType}}) bool) bool
}

// {{.Name}}Convertible supports bulk conversion to a slice.
type {{.Name}}Convertible interface {
	// ToSlice returns all elements as a slice.
	ToSlice() []{{.GoType}}
}

// ── Composed collection interfaces ────────────────────────────────────

// {{.Name}}Collection is the full read-only interface for any collection of
// {{.GoType}} values.  It composes the smaller sub-interfaces above, so a
// caller that only needs iteration can accept {{.Name}}Iterable while code
// that needs everything can accept {{.Name}}Collection.
//
// Satisfied by: arraylist.{{.Name}}, hashset.{{.Name}}, bag.Hash{{.Name}},
// bag.Tree{{.Name}}, stack.{{.Name}}, treeset.{{.Name}}, and their immutable
// variants.
type {{.Name}}Collection interface {
	{{.Name}}Sized
	{{.Name}}Iterable
	{{.Name}}Searchable
	{{.Name}}Convertible
	fmt.Stringer
}

// {{.Name}}MutableCollection extends {{.Name}}Collection with mutation operations.
// Satisfied by: arraylist.{{.Name}}, hashset.{{.Name}}, bag.Hash{{.Name}}, stack.{{.Name}}.
type {{.Name}}MutableCollection interface {
	{{.Name}}Collection

	// Clear removes all elements.
	Clear()
}

// ── Category interfaces ────────────────────────────────────────────────
//
// These distinguish *what kind of collection* is required without naming
// a concrete type. They mirror Java's IntList / IntSet / IntBag / IntStack
// hierarchy and the matching trait/comptime layers in Rust and Zig.
//
// Note: every mutable non-stack category exposes the unified Add(value) bool
// (the {{.Name}}Adder capability): lists and bags always return true (they always
// accept the element), sets return whether the value was newly inserted. Stacks
// use Push instead. The bool is collection-specific insertion info that bulk
// loaders (seq.Into) ignore.

// {{.Name}}List is the read-only interface for ordered lists with positional
// access. Satisfied by: arraylist.{{.Name}}, arraylist.Immutable{{.Name}}.
type {{.Name}}List interface {
	{{.Name}}Collection

	// Get returns the element at the given index. It panics on out-of-range index.
	Get(index int) {{.GoType}}

	// IndexOf returns the index of the first occurrence of value, or -1 if absent.
	IndexOf(value {{.GoType}}) int
}

// {{.Name}}MutableList extends {{.Name}}List + {{.Name}}MutableCollection.
// Satisfied by: arraylist.{{.Name}}.
type {{.Name}}MutableList interface {
	{{.Name}}List
	{{.Name}}MutableCollection

	// Add appends a value to the end of the list. Returns true always (a list
	// always accepts the element); the bool is the {{.Name}}Adder contract.
	Add(value {{.GoType}}) bool

	// Set sets the value at the given index, returning the previous value.
	// It panics on out-of-range index.
	Set(index int, value {{.GoType}}) {{.GoType}}
}

// {{.Name}}Set marker interface for set-like collections (uniqueness implied).
// Satisfied by: hashset.{{.Name}}, treeset.{{.Name}}, hashset.Immutable{{.Name}}.
type {{.Name}}Set interface {
	{{.Name}}Collection
}

// {{.Name}}MutableSet adds insertion. Add returns true if the value was newly inserted.
// Satisfied by: hashset.{{.Name}}, treeset.{{.Name}}.
type {{.Name}}MutableSet interface {
	{{.Name}}Set
	{{.Name}}MutableCollection

	// Add inserts a value. Returns true if the value was not already present.
	Add(value {{.GoType}}) bool
}

// {{.Name}}Bag read-only multiset interface with occurrence counts.
// Satisfied by: bag.Hash{{.Name}}, bag.Tree{{.Name}}, bag.ImmutableHash{{.Name}}.
type {{.Name}}Bag interface {
	{{.Name}}Collection

	// OccurrencesOf returns the number of times value occurs in the bag.
	OccurrencesOf(value {{.GoType}}) int

	// SizeDistinct returns the number of *distinct* values (ignoring multiplicity).
	SizeDistinct() int
}

// {{.Name}}MutableBag adds insertion.
// Satisfied by: bag.Hash{{.Name}}, bag.Tree{{.Name}}.
type {{.Name}}MutableBag interface {
	{{.Name}}Bag
	{{.Name}}MutableCollection

	// Add adds one occurrence of value. Returns true always (a bag always accepts
	// the element); the bool is the {{.Name}}Adder contract.
	Add(value {{.GoType}}) bool
}

// {{.Name}}Stack read-only LIFO stack. Peek returns the top element and false
// if the stack is empty (no error is returned).
// Satisfied by: stack.{{.Name}}, stack.Immutable{{.Name}}.
type {{.Name}}Stack interface {
	{{.Name}}Collection

	// Peek returns the top element without removing it. The bool is false if empty.
	Peek() ({{.GoType}}, bool)
}

// {{.Name}}MutableStack mutable LIFO stack with Push and Pop.
// Satisfied by: stack.{{.Name}}.
type {{.Name}}MutableStack interface {
	{{.Name}}Stack
	{{.Name}}MutableCollection

	// Push pushes a value onto the top of the stack.
	Push(value {{.GoType}})

	// Pop removes and returns the top element. The bool is false if empty.
	Pop() ({{.GoType}}, bool)
}

// ── Framework capability interfaces ────────────────────────────────────

// {{.Name}}Adder is the single-element insertion capability shared by every
// mutable non-stack collection ({{.Name}}ArrayList, {{.Name}}HashSet,
// {{.Name}}HashBag, {{.Name}}TreeSet, {{.Name}}TreeBag, …). It is the sink a bulk
// loader targets: the bool reports collection-specific insertion info (lists/bags
// always true, sets whether newly inserted) and bulk loaders ignore it.
type {{.Name}}Adder interface {
	Add(value {{.GoType}}) bool
}
`
