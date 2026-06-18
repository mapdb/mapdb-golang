package hashmap

import (
	"encoding/binary"
	"hash/maphash"
	"math/bits"
)

// ---------------------------------------------------------------------------
// Swiss-table control-byte machinery (shared by every generated hash map).
//
// Each map stores entries in a flat slice alongside a control-byte slice.
// A control byte is one of:
//   - swissEmpty   (0xFF): slot never used (terminates a probe)
//   - swissDeleted (0x80): tombstone (probe continues past it)
//   - a 7-bit tag  (0x00..0x7F): FULL slot; the low 7 bits of the key hash.
//
// Lookups probe by GROUPS of swissGroupWidth buckets using a SWAR matcher over
// a little-endian uint64 word, then a triangular group sequence that visits
// every group exactly once for a power-of-two capacity. Only FULL slots hold
// live entries; EMPTY/DELETED slots hold the zero value and are never read.
//
// The control slice has length cap+swissGroupWidth; the trailing
// swissGroupWidth bytes mirror the first swissGroupWidth so a group load near
// the end can read 8 bytes without a bounds branch. Capacity is always a power
// of two and at least swissMinCapacity (>= swissGroupWidth), so a full group is
// always loadable and the group width divides the capacity.
// ---------------------------------------------------------------------------

const (
	swissGroupWidth  = 8
	swissMinCapacity = 16
	swissEmpty       = 0xFF
	swissDeleted     = 0x80

	swissOnes  uint64 = 0x0101010101010101
	swissHighs uint64 = 0x8080808080808080
)

// swissMaxLoad returns cap*7/8, the maximum number of live entries permitted.
func swissMaxLoad(capacity int) int {
	return capacity / 8 * 7
}

// swissCapacityFor returns the smallest power-of-two cap >= swissMinCapacity
// such that n <= swissMaxLoad(cap).
func swissCapacityFor(n int) int {
	c := swissMinCapacity
	for swissMaxLoad(c) < n {
		c *= 2
	}
	return c
}

// newSwissCtrl allocates a control slice of length cap+swissGroupWidth with
// every byte set to EMPTY (the mirror suffix is therefore already correct).
func newSwissCtrl(capacity int) []uint8 {
	ctrl := make([]uint8, capacity+swissGroupWidth)
	for i := range ctrl {
		ctrl[i] = swissEmpty
	}
	return ctrl
}

// swissSetCtrl writes control byte idx and mirrors it into the suffix when
// idx < swissGroupWidth, keeping the trailing mirror bytes consistent.
func swissSetCtrl(ctrl []uint8, idx int, b uint8, capacity int) {
	ctrl[idx] = b
	if idx < swissGroupWidth {
		ctrl[capacity+idx] = b
	}
}

// swissLoadGroup loads the eight control bytes starting at i as a
// little-endian uint64. ctrl must have at least i+swissGroupWidth bytes, which
// the mirror suffix guarantees for any i < cap.
func swissLoadGroup(ctrl []uint8, i int) uint64 {
	return binary.LittleEndian.Uint64(ctrl[i : i+swissGroupWidth])
}

// swissHasZero is the classic SWAR zero-byte detector: a lane's high bit is set
// in the result iff the corresponding byte of x is 0x00.
func swissHasZero(x uint64) uint64 {
	return (x - swissOnes) &^ x & swissHighs
}

// swissMatchByte returns a lane-bitmask of bytes in g equal to b. Lane i is
// "set" iff bit 8*i+7 is 1.
func swissMatchByte(g uint64, b uint8) uint64 {
	return swissHasZero(g ^ (swissOnes * uint64(b)))
}

// swissMatchEmpty returns a lane-bitmask of EMPTY (0xFF) bytes in g.
func swissMatchEmpty(g uint64) uint64 {
	return swissMatchByte(g, swissEmpty)
}

// swissMatchFull returns a lane-bitmask of FULL bytes (high bit clear) in g.
// EMPTY (0xFF) and DELETED (0x80) both have the high bit set, so ^g & HIGHS
// marks exactly the FULL lanes — letting iterators skip empty/deleted runs a
// whole group at a time.
func swissMatchFull(g uint64) uint64 {
	return ^g & swissHighs
}

// swissLowestLane returns the index (0..swissGroupWidth) of the lowest set lane
// in a non-zero bitmask.
func swissLowestLane(mask uint64) int {
	return bits.TrailingZeros64(mask) >> 3
}

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
