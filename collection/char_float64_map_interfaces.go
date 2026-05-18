
package collection

import "iter"

// CharFloat64MapIterable is the read-only interface for any map from uint16 to float64.
// Satisfied by: CharFloat64HashMap, CharFloat64TreeMap, CharFloat64SentinelHashMap,
// ImmutableCharFloat64HashMap.
type CharFloat64MapIterable interface {
	// Get returns the value for the key, or (zero, false) if not found.
	Get(key uint16) (float64, bool)

	// GetOrDefault returns the value for the key, or defaultValue if not found.
	GetOrDefault(key uint16, defaultValue float64) float64

	// ContainsKey returns true if the map contains the given key.
	ContainsKey(key uint16) bool

	// Size returns the number of key-value pairs.
	Size() int

	// IsEmpty returns true if the map contains no entries.
	IsEmpty() bool

	// All returns an iter.Seq2 that yields all key-value pairs.
	All() iter.Seq2[uint16, float64]

	// Keys returns an iter.Seq that yields all keys.
	Keys() iter.Seq[uint16]

	// Values returns an iter.Seq that yields all values.
	Values() iter.Seq[float64]

	// ForEach calls the given function for each key-value pair.
	ForEach(f func(uint16, float64))

	// AnySatisfy returns true if any entry satisfies the predicate.
	AnySatisfy(predicate func(uint16, float64) bool) bool

	// AllSatisfy returns true if all entries satisfy the predicate.
	AllSatisfy(predicate func(uint16, float64) bool) bool

	// NoneSatisfy returns true if no entry satisfies the predicate.
	NoneSatisfy(predicate func(uint16, float64) bool) bool

	// String returns a string representation.
	String() string
}

// CharFloat64MutableMap extends CharFloat64MapIterable with mutation operations.
// Satisfied by: CharFloat64HashMap, CharFloat64TreeMap, CharFloat64SentinelHashMap.
type CharFloat64MutableMap interface {
	CharFloat64MapIterable

	// Put adds or updates a key-value pair. Returns (old value, existed).
	Put(key uint16, value float64) (float64, bool)

	// Remove removes the key. Returns (old value, existed).
	Remove(key uint16) (float64, bool)

	// Clear removes all entries.
	Clear()
}
