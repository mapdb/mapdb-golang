package hashmap

import (
	"hash/maphash"
	"reflect"

	"github.com/mapdb/mapdb-golang/internal/floatid"
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
// probe loop, so the two must agree. Float keys never reach it: see
// hashObjectKey below.
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

// objectKeyKind reports whether K's underlying type is float32 or float64
// (reflect.Float32/Float64) or anything else (reflect.Invalid). The
// object-keyed maps (Object<Value>[K]) evaluate it once per operation and pass
// it to hashObjectKey/sameObjectKey.
//
// Float keys use bit-pattern identity (spec algorithms.md "NaN must hash and
// compare by bit pattern"): NaN finds itself, distinct NaN payloads are
// distinct keys and -0.0 != +0.0. Go's == (and maphash.Comparable, which is
// ==-consistent and hashes NaN randomly) would make a NaN key unreachable and
// merge the two zeros. Named float types are included; composite keys that
// merely contain floats keep == semantics.
func objectKeyKind[K comparable]() reflect.Kind { return floatid.Kind[K]() }

// hashObjectKey hashes an object-map key: the mapdb Fibonacci bit-pattern hash
// for float K, hashComparable otherwise. fk must be objectKeyKind[K]().
func hashObjectKey[K comparable](fk reflect.Kind, key K) uint64 {
	if fk != reflect.Invalid {
		return floatid.Hash(floatid.Bits(fk, key))
	}
	return hashComparable(key)
}

// sameObjectKey is the object-map key equality matching hashObjectKey:
// bit-pattern equality for float K, == otherwise. fk must be
// objectKeyKind[K]().
func sameObjectKey[K comparable](fk reflect.Kind, a, b K) bool {
	return floatid.SameKind(fk, a, b)
}
