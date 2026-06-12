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

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: codegen <collection>")
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "arraylist":
		err = genArrayList()
	case "interval":
		err = genInterval()
	case "hashset":
		err = genHashSet()
	case "stack":
		err = genStack()
	case "deque":
		err = genDeque()
	case "treeset":
		err = genTreeSet()
	case "treemap":
		err = genTreeMap()
	case "hashmap":
		err = genHashMap()
	case "sentinelhashmap":
		err = genSentinelHashMap()
	case "multimap":
		err = genMultimap()
	case "priorityqueue":
		err = genPriorityQueue()
	case "bag":
		err = genBag()
	case "tuple":
		err = genTuple()
	default:
		err = fmt.Errorf("unknown collection %q", os.Args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "codegen:", err)
		os.Exit(1)
	}
}
