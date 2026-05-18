
package collection

import "iter"

// Float32Float64MapIterable is the read-only interface for any map from float32 to float64.
// Satisfied by: Float32Float64HashMap, Float32Float64TreeMap, Float32Float64SentinelHashMap,
// ImmutableFloat32Float64HashMap.
type Float32Float64MapIterable interface {
	// Get returns the value for the key, or (zero, false) if not found.
	Get(key float32) (float64, bool)

	// GetOrDefault returns the value for the key, or defaultValue if not found.
	GetOrDefault(key float32, defaultValue float64) float64

	// ContainsKey returns true if the map contains the given key.
	ContainsKey(key float32) bool

	// Size returns the number of key-value pairs.
	Size() int

	// IsEmpty returns true if the map contains no entries.
	IsEmpty() bool

	// All returns an iter.Seq2 that yields all key-value pairs.
	All() iter.Seq2[float32, float64]

	// Keys returns an iter.Seq that yields all keys.
	Keys() iter.Seq[float32]

	// Values returns an iter.Seq that yields all values.
	Values() iter.Seq[float64]

	// ForEach calls the given function for each key-value pair.
	ForEach(f func(float32, float64))

	// AnySatisfy returns true if any entry satisfies the predicate.
	AnySatisfy(predicate func(float32, float64) bool) bool

	// AllSatisfy returns true if all entries satisfy the predicate.
	AllSatisfy(predicate func(float32, float64) bool) bool

	// NoneSatisfy returns true if no entry satisfies the predicate.
	NoneSatisfy(predicate func(float32, float64) bool) bool

	// String returns a string representation.
	String() string
}

// Float32Float64MutableMap extends Float32Float64MapIterable with mutation operations.
// Satisfied by: Float32Float64HashMap, Float32Float64TreeMap, Float32Float64SentinelHashMap.
type Float32Float64MutableMap interface {
	Float32Float64MapIterable

	// Put adds or updates a key-value pair. Returns (old value, existed).
	Put(key float32, value float64) (float64, bool)

	// Remove removes the key. Returns (old value, existed).
	Remove(key float32) (float64, bool)

	// Clear removes all entries.
	Clear()
}
