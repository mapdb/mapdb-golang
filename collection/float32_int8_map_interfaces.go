
package collection

import "iter"

// Float32Int8MapIterable is the read-only interface for any map from float32 to int8.
// Satisfied by: Float32Int8HashMap, Float32Int8TreeMap, Float32Int8SentinelHashMap,
// ImmutableFloat32Int8HashMap.
type Float32Int8MapIterable interface {
	// Get returns the value for the key, or (zero, false) if not found.
	Get(key float32) (int8, bool)

	// GetOrDefault returns the value for the key, or defaultValue if not found.
	GetOrDefault(key float32, defaultValue int8) int8

	// ContainsKey returns true if the map contains the given key.
	ContainsKey(key float32) bool

	// Size returns the number of key-value pairs.
	Size() int

	// IsEmpty returns true if the map contains no entries.
	IsEmpty() bool

	// All returns an iter.Seq2 that yields all key-value pairs.
	All() iter.Seq2[float32, int8]

	// Keys returns an iter.Seq that yields all keys.
	Keys() iter.Seq[float32]

	// Values returns an iter.Seq that yields all values.
	Values() iter.Seq[int8]

	// ForEach calls the given function for each key-value pair.
	ForEach(f func(float32, int8))

	// AnySatisfy returns true if any entry satisfies the predicate.
	AnySatisfy(predicate func(float32, int8) bool) bool

	// AllSatisfy returns true if all entries satisfy the predicate.
	AllSatisfy(predicate func(float32, int8) bool) bool

	// NoneSatisfy returns true if no entry satisfies the predicate.
	NoneSatisfy(predicate func(float32, int8) bool) bool

	// String returns a string representation.
	String() string
}

// Float32Int8MutableMap extends Float32Int8MapIterable with mutation operations.
// Satisfied by: Float32Int8HashMap, Float32Int8TreeMap, Float32Int8SentinelHashMap.
type Float32Int8MutableMap interface {
	Float32Int8MapIterable

	// Put adds or updates a key-value pair. Returns (old value, existed).
	Put(key float32, value int8) (int8, bool)

	// Remove removes the key. Returns (old value, existed).
	Remove(key float32) (int8, bool)

	// Clear removes all entries.
	Clear()
}
