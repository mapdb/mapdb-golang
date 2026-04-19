# mapdb-collections (Go)

High-performance primitive-specialized and generic collections for Go, inspired by [Eclipse Collections](https://eclipse.dev/collections/).

## Why?

Go's standard library provides `map` and `slice` — great general-purpose containers, but they box all values and lack the rich iteration API that Eclipse Collections is known for. This library provides:

- **Primitive-specialized types** (`I32ArrayList`, `I32HashSet`, `I32I64HashMap`, etc.) with no interface boxing and contiguous memory layout
- **Generic object collections** (`ArrayList[T]`, `HashSet[T]`, `HashMap[K,V]`, etc.) with the full Eclipse Collections API
- Rich functional methods: `Select`, `Reject`, `Detect`, `AnySatisfy`, `AllSatisfy`, `InjectInto`, `CollectInt`, and more

## Primitive Collections

| Type | Mutable | Variants |
|------|---------|----------|
| **ArrayList** | `I32ArrayList` | 8 types |
| **HashSet** | `I32HashSet` | 8 types |
| **HashBag** | `I32HashBag` | 8 types |
| **ArrayStack** | `I32ArrayStack` | 8 types |
| **HashMap** | `I32I64HashMap` | 64 pairs (8x8) |
| **TreeSet** | `I32TreeSet` | 8 types |
| **TreeMap** | `I32I64TreeMap` | 64 pairs |
| **Pair** | `I32I64Pair` | 64 pairs |
| **Interval** | `I32Interval` | range type |

## Object Collections

Generic collections using Go 1.23 type parameters:

| Type | Description |
|------|-------------|
| `ArrayList[T]` | Ordered list backed by `[]T` |
| `HashSet[T]` | Unordered set backed by `map[T]struct{}` |
| `HashMap[K, V]` | Key-value map backed by `map[K]V` |
| `HashBag[T]` | Counting bag backed by `map[T]int` |
| `ArrayStack[T]` | LIFO stack backed by `[]T` |
| `HashBiMap[K, V]` | Bidirectional map with unique keys and values |

All generic types implement composable interfaces: `Collection[T]`, `MutableList[T]`, `MutableSet[T]`, `MutableBag[T]`, `MutableStack[T]`, `MutableMap[K,V]`, `MutableBiMap[K,V]`.

## Quick Start

```go
import (
    "github.com/mapdb/mapdb-golang/arraylist"
    "github.com/mapdb/mapdb-golang/object"
)

// Primitive ArrayList
list := arraylist.I32ArrayListOf([]int32{3, 1, 4, 1, 5})
list.Sort()
selected := list.Select(func(v int32) bool { return v > 2 })

// Generic ArrayList
names := object.ArrayListOf([]string{"Alice", "Bob", "Charlie"})
found := names.Detect(func(s string) bool { return s[0] == 'B' })
// found == &"Bob"

// Generic HashBag
bag := object.NewHashBag[string]()
bag.Add("apple")
bag.AddOccurrences("apple", 3)
// bag.OccurrencesOf("apple") == 4
```

## Stats

- **471 source files**, **370 test files**, **4,743 tests** passing
- All 8 Go primitive types + generic object types
- Zero external dependencies
- Requires Go 1.23+
