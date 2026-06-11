package hashmap

import (
	"unsafe"
)

// hashComparable computes a hash for any comparable value using its memory representation.
// This is a general-purpose hash that works with any comparable type (string, int, struct, etc.)
func hashComparable[K comparable](key K) uint64 {
	size := unsafe.Sizeof(key)
	p := unsafe.Pointer(&key)

	var h uint64
	switch {
	case size == 1:
		h = uint64(*(*uint8)(p))
	case size == 2:
		h = uint64(*(*uint16)(p))
	case size == 4:
		h = uint64(*(*uint32)(p))
	case size == 8:
		h = *(*uint64)(p)
	default:
		// FNV-1a for larger types (strings, structs)
		b := unsafe.Slice((*byte)(p), size)
		h = 14695981039346656037
		for _, c := range b {
			h ^= uint64(c)
			h *= 1099511628211
		}
	}
	// Mix bits (splitmix64 finalizer)
	h ^= h >> 30
	h *= 0xbf58476d1ce4e5b9
	h ^= h >> 27
	h *= 0x94d049bb133111eb
	h ^= h >> 31
	return h
}
