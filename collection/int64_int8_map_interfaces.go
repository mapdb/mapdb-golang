
package collection

import "iter"

// Int64Int8MapIterable is the read-only interface for any map from int64 to int8.
// Satisfied by: Int64Int8HashMap, Int64Int8TreeMap, Int64Int8SentinelHashMap,
// ImmutableInt64Int8HashMap.
type Int64Int8MapIterable interface {
	// Get returns the value for the key, or (zero, false) if not found.
	Get(key int64) (int8, bool)

	// GetOrDefault returns the value for the key, or defaultValue if not found.
	GetOrDefault(key int64, defaultValue int8) int8

	// ContainsKey returns true if the map contains the given key.
	ContainsKey(key int64) bool

	// Size returns the number of key-value pairs.
	Size() int

	// IsEmpty returns true if the map contains no entries.
	IsEmpty() bool

	// All returns an iter.Seq2 that yields all key-value pairs.
	All() iter.Seq2[int64, int8]

	// Keys returns an iter.Seq that yields all keys.
	Keys() iter.Seq[int64]

	// Values returns an iter.Seq that yields all values.
	Values() iter.Seq[int8]

	// ForEach calls the given function for each key-value pair.
	ForEach(f func(int64, int8))

	// AnySatisfy returns true if any entry satisfies the predicate.
	AnySatisfy(predicate func(int64, int8) bool) bool

	// AllSatisfy returns true if all entries satisfy the predicate.
	AllSatisfy(predicate func(int64, int8) bool) bool

	// NoneSatisfy returns true if no entry satisfies the predicate.
	NoneSatisfy(predicate func(int64, int8) bool) bool

	// String returns a string representation.
	String() string
}

// Int64Int8MutableMap extends Int64Int8MapIterable with mutation operations.
// Satisfied by: Int64Int8HashMap, Int64Int8TreeMap, Int64Int8SentinelHashMap.
type Int64Int8MutableMap interface {
	Int64Int8MapIterable

	// Put adds or updates a key-value pair. Returns (old value, existed).
	Put(key int64, value int8) (int8, bool)

	// Remove removes the key. Returns (old value, existed).
	Remove(key int64) (int8, bool)

	// Clear removes all entries.
	Clear()
}
