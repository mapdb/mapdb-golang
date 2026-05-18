
package collection

import "iter"

// Int16Int64MapIterable is the read-only interface for any map from int16 to int64.
// Satisfied by: Int16Int64HashMap, Int16Int64TreeMap, Int16Int64SentinelHashMap,
// ImmutableInt16Int64HashMap.
type Int16Int64MapIterable interface {
	// Get returns the value for the key, or (zero, false) if not found.
	Get(key int16) (int64, bool)

	// GetOrDefault returns the value for the key, or defaultValue if not found.
	GetOrDefault(key int16, defaultValue int64) int64

	// ContainsKey returns true if the map contains the given key.
	ContainsKey(key int16) bool

	// Size returns the number of key-value pairs.
	Size() int

	// IsEmpty returns true if the map contains no entries.
	IsEmpty() bool

	// All returns an iter.Seq2 that yields all key-value pairs.
	All() iter.Seq2[int16, int64]

	// Keys returns an iter.Seq that yields all keys.
	Keys() iter.Seq[int16]

	// Values returns an iter.Seq that yields all values.
	Values() iter.Seq[int64]

	// ForEach calls the given function for each key-value pair.
	ForEach(f func(int16, int64))

	// AnySatisfy returns true if any entry satisfies the predicate.
	AnySatisfy(predicate func(int16, int64) bool) bool

	// AllSatisfy returns true if all entries satisfy the predicate.
	AllSatisfy(predicate func(int16, int64) bool) bool

	// NoneSatisfy returns true if no entry satisfies the predicate.
	NoneSatisfy(predicate func(int16, int64) bool) bool

	// String returns a string representation.
	String() string
}

// Int16Int64MutableMap extends Int16Int64MapIterable with mutation operations.
// Satisfied by: Int16Int64HashMap, Int16Int64TreeMap, Int16Int64SentinelHashMap.
type Int16Int64MutableMap interface {
	Int16Int64MapIterable

	// Put adds or updates a key-value pair. Returns (old value, existed).
	Put(key int16, value int64) (int64, bool)

	// Remove removes the key. Returns (old value, existed).
	Remove(key int16) (int64, bool)

	// Clear removes all entries.
	Clear()
}
