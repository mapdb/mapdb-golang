// Command codegen generates per-primitive collection sources for
// mapdb-golang. It is intended to be invoked from a //go:generate
// directive in the collection package being generated.
//
// Go has no method-level generics, so per-primitive specialization is the
// only way to expose unboxed APIs (Int32ArrayList.Sum(), etc.) without
// boxing through interface{}. Rather than hand-maintain ~1300 nearly
// identical files, this tool re-emits them from a small set of
// text/template sources kept next to each collection generator.
//
// Usage (from inside the target collection directory):
//
//	//go:generate go run ../internal/codegen <collection>
//
// where <collection> is one of: arraylist, interval, hashset, stack, deque,
// treeset, treemap, hashmap, sentinelhashmap, multimap, priorityqueue, bag,
// tuple.
//
// Drift guard: `go generate ./... && git diff --exit-code` is sufficient.
package main

import (
	"fmt"
	"os"
)

// generators maps each codegen subcommand to its generator function. The
// per-family keys — every key except the two non-family subcommands "matrix"
// (renders FAMILY_MATRIX.md) and "interfaces" (renders the collection interface
// vocabulary) — are kept in lockstep with the manifest (Families) and the
// per-package go:generate directives by TestManifestMatchesGenerators, which
// applies the same exclusion. Adding a family means updating all three.
var generators = map[string]func() error{
	"arraylist":       genArrayList,
	"interval":        genInterval,
	"hashset":         genHashSet,
	"stack":           genStack,
	"deque":           genDeque,
	"treeset":         genTreeSet,
	"treemap":         genTreeMap,
	"hashmap":         genHashMap,
	"sentinelhashmap": genSentinelHashMap,
	"multimap":        genMultimap,
	"priorityqueue":   genPriorityQueue,
	"bag":             genBag,
	"tuple":           genTuple,
	"matrix":          genFamilyMatrix,
	"interfaces":      genInterfaces,
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: codegen <collection>")
		os.Exit(2)
	}
	gen, ok := generators[os.Args[1]]
	if !ok {
		fmt.Fprintln(os.Stderr, "codegen:", fmt.Errorf("unknown collection %q", os.Args[1]))
		os.Exit(1)
	}
	if err := gen(); err != nil {
		fmt.Fprintln(os.Stderr, "codegen:", err)
		os.Exit(1)
	}
}
