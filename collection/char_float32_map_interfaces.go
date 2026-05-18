
package collection

import "iter"

// CharFloat32MapIterable is the read-only interface for any map from uint16 to float32.
// Satisfied by: CharFloat32HashMap, CharFloat32TreeMap, CharFloat32SentinelHashMap,
// ImmutableCharFloat32HashMap.
type CharFloat32MapIterable interface {
	// Get returns the value for the key, or (zero, false) if not found.
	Get(key uint16) (float32, bool)

	// GetOrDefault returns the value for the key, or defaultValue if not found.
	GetOrDefault(key uint16, defaultValue float32) float32

	// ContainsKey returns true if the map contains the given key.
	ContainsKey(key uint16) bool

	// Size returns the number of key-value pairs.
	Size() int

	// IsEmpty returns true if the map contains no entries.
	IsEmpty() bool

	// All returns an iter.Seq2 that yields all key-value pairs.
	All() iter.Seq2[uint16, float32]

	// Keys returns an iter.Seq that yields all keys.
	Keys() iter.Seq[uint16]

	// Values returns an iter.Seq that yields all values.
	Values() iter.Seq[float32]

	// ForEach calls the given function for each key-value pair.
	ForEach(f func(uint16, float32))

	// AnySatisfy returns true if any entry satisfies the predicate.
	AnySatisfy(predicate func(uint16, float32) bool) bool

	// AllSatisfy returns true if all entries satisfy the predicate.
	AllSatisfy(predicate func(uint16, float32) bool) bool

	// NoneSatisfy returns true if no entry satisfies the predicate.
	NoneSatisfy(predicate func(uint16, float32) bool) bool

	// String returns a string representation.
	String() string
}

// CharFloat32MutableMap extends CharFloat32MapIterable with mutation operations.
// Satisfied by: CharFloat32HashMap, CharFloat32TreeMap, CharFloat32SentinelHashMap.
type CharFloat32MutableMap interface {
	CharFloat32MapIterable

	// Put adds or updates a key-value pair. Returns (old value, existed).
	Put(key uint16, value float32) (float32, bool)

	// Remove removes the key. Returns (old value, existed).
	Remove(key uint16) (float32, bool)

	// Clear removes all entries.
	Clear()
}
