
package collection

import "iter"

// Int8Int32MapIterable is the read-only interface for any map from int8 to int32.
// Satisfied by: Int8Int32HashMap, Int8Int32TreeMap, Int8Int32SentinelHashMap,
// ImmutableInt8Int32HashMap.
type Int8Int32MapIterable interface {
	// Get returns the value for the key, or (zero, false) if not found.
	Get(key int8) (int32, bool)

	// GetOrDefault returns the value for the key, or defaultValue if not found.
	GetOrDefault(key int8, defaultValue int32) int32

	// ContainsKey returns true if the map contains the given key.
	ContainsKey(key int8) bool

	// Size returns the number of key-value pairs.
	Size() int

	// IsEmpty returns true if the map contains no entries.
	IsEmpty() bool

	// All returns an iter.Seq2 that yields all key-value pairs.
	All() iter.Seq2[int8, int32]

	// Keys returns an iter.Seq that yields all keys.
	Keys() iter.Seq[int8]

	// Values returns an iter.Seq that yields all values.
	Values() iter.Seq[int32]

	// ForEach calls the given function for each key-value pair.
	ForEach(f func(int8, int32))

	// AnySatisfy returns true if any entry satisfies the predicate.
	AnySatisfy(predicate func(int8, int32) bool) bool

	// AllSatisfy returns true if all entries satisfy the predicate.
	AllSatisfy(predicate func(int8, int32) bool) bool

	// NoneSatisfy returns true if no entry satisfies the predicate.
	NoneSatisfy(predicate func(int8, int32) bool) bool

	// String returns a string representation.
	String() string
}

// Int8Int32MutableMap extends Int8Int32MapIterable with mutation operations.
// Satisfied by: Int8Int32HashMap, Int8Int32TreeMap, Int8Int32SentinelHashMap.
type Int8Int32MutableMap interface {
	Int8Int32MapIterable

	// Put adds or updates a key-value pair. Returns (old value, existed).
	Put(key int8, value int32) (int32, bool)

	// Remove removes the key. Returns (old value, existed).
	Remove(key int8) (int32, bool)

	// Clear removes all entries.
	Clear()
}
