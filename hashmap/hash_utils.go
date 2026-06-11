package hashmap

import (
	"hash/maphash"
)

// objectKeySeed is a single process-wide seed used for all object/comparable-key
// hashing. Reusing one seed keeps every map internally consistent for the
// lifetime of the process. Map iteration order is not part of the contract, so a
// per-process random seed is fine and adds hash-flooding resistance.
var objectKeySeed = maphash.MakeSeed()

// hashComparable computes a hash for any comparable value.
//
// It delegates to hash/maphash.Comparable, which is guaranteed to be
// ==-consistent: Comparable(seed, v1) == Comparable(seed, v2) whenever
// v1 == v2, for any comparable type (strings, named string types, interfaces,
// and pointer-bearing structs). This matters because the object-keyed maps use
// hashComparable to pick a bucket and Go's == to confirm key equality in the
// probe loop, so the two must agree.
//
// The previous implementation hashed the key's raw memory via unsafe, which
// keyed on backing-array addresses. Two ==-equal values with distinct backings
// — the motivating gaps being a NAMED string type (type S string) and a
// pointer-bearing struct (e.g. struct{ Name string; Age int }) — would hash
// apart, making the second value unfindable and allowing duplicate logical keys.
// maphash.Comparable closes that gap for all comparable types.
func hashComparable[K comparable](key K) uint64 {
	return maphash.Comparable(objectKeySeed, key)
}
