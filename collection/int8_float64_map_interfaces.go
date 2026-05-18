
package collection

import "iter"

// Int8Float64MapIterable is the read-only interface for any map from int8 to float64.
// Satisfied by: Int8Float64HashMap, Int8Float64TreeMap, Int8Float64SentinelHashMap,
// ImmutableInt8Float64HashMap.
type Int8Float64MapIterable interface {
	// Get returns the value for the key, or (zero, false) if not found.
	Get(key int8) (float64, bool)

	// GetOrDefault returns the value for the key, or defaultValue if not found.
	GetOrDefault(key int8, defaultValue float64) float64

	// ContainsKey returns true if the map contains the given key.
	ContainsKey(key int8) bool

	// Size returns the number of key-value pairs.
	Size() int

	// IsEmpty returns true if the map contains no entries.
	IsEmpty() bool

	// All returns an iter.Seq2 that yields all key-value pairs.
	All() iter.Seq2[int8, float64]

	// Keys returns an iter.Seq that yields all keys.
	Keys() iter.Seq[int8]

	// Values returns an iter.Seq that yields all values.
	Values() iter.Seq[float64]

	// ForEach calls the given function for each key-value pair.
	ForEach(f func(int8, float64))

	// AnySatisfy returns true if any entry satisfies the predicate.
	AnySatisfy(predicate func(int8, float64) bool) bool

	// AllSatisfy returns true if all entries satisfy the predicate.
	AllSatisfy(predicate func(int8, float64) bool) bool

	// NoneSatisfy returns true if no entry satisfies the predicate.
	NoneSatisfy(predicate func(int8, float64) bool) bool

	// String returns a string representation.
	String() string
}

// Int8Float64MutableMap extends Int8Float64MapIterable with mutation operations.
// Satisfied by: Int8Float64HashMap, Int8Float64TreeMap, Int8Float64SentinelHashMap.
type Int8Float64MutableMap interface {
	Int8Float64MapIterable

	// Put adds or updates a key-value pair. Returns (old value, existed).
	Put(key int8, value float64) (float64, bool)

	// Remove removes the key. Returns (old value, existed).
	Remove(key int8) (float64, bool)

	// Clear removes all entries.
	Clear()
}
