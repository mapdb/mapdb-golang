package hashmap

import (
	"unsafe"
)

// hashComparable computes a hash for any comparable value using its memory representation.
// This is a general-purpose hash that works with any comparable type (string, int, struct, etc.)
func hashComparable[K comparable](key K) uint64 {
	var h uint64
	if s, ok := any(key).(string); ok {
		// Hash the string's bytes, not its header. A string value is a
		// {data pointer, length} header; hashing that raw memory keys on the
		// backing-array address, so two equal-content strings with distinct
		// backings (e.g. a literal vs. string([]byte(...))) would hash apart
		// even though Go's == — which the probe loop uses — reports them equal,
		// making the second unfindable. FNV-1a over the content fixes that.
		h = 14695981039346656037
		for i := 0; i < len(s); i++ {
			h ^= uint64(s[i])
			h *= 1099511628211
		}
	} else {
		size := unsafe.Sizeof(key)
		p := unsafe.Pointer(&key)
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
			// FNV-1a over the raw memory. Correct for fixed-layout value types
			// (e.g. structs of integers/floats); types whose memory holds
			// pointers (named string types, pointer-bearing structs) share the
			// header-vs-content caveat above and are out of scope here.
			b := unsafe.Slice((*byte)(p), size)
			h = 14695981039346656037
			for _, c := range b {
				h ^= uint64(c)
				h *= 1099511628211
			}
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
