package hashmap

import (
	"math"
	"math/rand/v2"
	"testing"
)

// Randomized Put/Remove/Get regressions against a reference Go map for the
// <Key>Object[V] maps (objectValueHashMapTmpl). Their backward-shift deletion
// (rehashFrom) used to move an entry into the gap exactly when it must NOT
// (the condition was inverted), leaving keys unreachable after a Remove. A
// small key range keeps the table at 16..64 slots so probe runs wrap around
// the end of the table and deletions hit long clusters.

const randomOps = 5000

func TestBackwardShiftRandom_Int32Object(t *testing.T) {
	for seed := uint64(1); seed <= 8; seed++ {
		r := rand.New(rand.NewPCG(seed, 0x5eed))
		m := NewInt32ObjectWithCapacity[int](16)
		ref := map[int32]int{}
		for op := 0; op < randomOps; op++ {
			k := int32(r.IntN(40)) * 16 // multiples of 16: many share a home bucket
			switch r.IntN(3) {
			case 0:
				v := r.Int()
				old, had := m.Put(k, v)
				rold, rhad := ref[k]
				if had != rhad || (had && old != rold) {
					t.Fatalf("seed %d op %d Put(%d): got (%d,%v) want (%d,%v)", seed, op, k, old, had, rold, rhad)
				}
				ref[k] = v
			case 1:
				old, had := m.Remove(k)
				rold, rhad := ref[k]
				if had != rhad || (had && old != rold) {
					t.Fatalf("seed %d op %d Remove(%d): got (%d,%v) want (%d,%v)", seed, op, k, old, had, rold, rhad)
				}
				delete(ref, k)
			default:
				v, ok := m.Get(k)
				rv, rok := ref[k]
				if ok != rok || (ok && v != rv) {
					t.Fatalf("seed %d op %d Get(%d): got (%d,%v) want (%d,%v)", seed, op, k, v, ok, rv, rok)
				}
			}
			if m.Len() != len(ref) {
				t.Fatalf("seed %d op %d: Len %d want %d", seed, op, m.Len(), len(ref))
			}
		}
		for k, rv := range ref {
			if v, ok := m.Get(k); !ok || v != rv {
				t.Fatalf("seed %d final Get(%d): got (%d,%v) want (%d,true)", seed, k, v, ok, rv)
			}
		}
	}
}

func TestBackwardShiftRandom_Float64Object(t *testing.T) {
	for seed := uint64(1); seed <= 8; seed++ {
		r := rand.New(rand.NewPCG(seed, 0xf10a7))
		m := NewFloat64ObjectWithCapacity[int](16)
		ref := map[uint64]int{} // keyed by bit pattern (the map's identity)
		for op := 0; op < randomOps; op++ {
			k := float64(r.IntN(40)) / 2 // 0, 0.5, 1, ...: long zero mantissa tails
			if r.IntN(8) == 0 {
				k = math.Copysign(k, -1) // -0.0 is a distinct key
			}
			kb := math.Float64bits(k)
			switch r.IntN(3) {
			case 0:
				v := r.Int()
				old, had := m.Put(k, v)
				rold, rhad := ref[kb]
				if had != rhad || (had && old != rold) {
					t.Fatalf("seed %d op %d Put(%v): got (%d,%v) want (%d,%v)", seed, op, k, old, had, rold, rhad)
				}
				ref[kb] = v
			case 1:
				old, had := m.Remove(k)
				rold, rhad := ref[kb]
				if had != rhad || (had && old != rold) {
					t.Fatalf("seed %d op %d Remove(%v): got (%d,%v) want (%d,%v)", seed, op, k, old, had, rold, rhad)
				}
				delete(ref, kb)
			default:
				v, ok := m.Get(k)
				rv, rok := ref[kb]
				if ok != rok || (ok && v != rv) {
					t.Fatalf("seed %d op %d Get(%v): got (%d,%v) want (%d,%v)", seed, op, k, v, ok, rv, rok)
				}
			}
			if m.Len() != len(ref) {
				t.Fatalf("seed %d op %d: Len %d want %d", seed, op, m.Len(), len(ref))
			}
		}
	}
}

// The base primitive maps (hashMapTmpl) use a distance-based backward shift
// that was already correct; this guards it (and the top-bit index) the same way.
func TestBackwardShiftRandom_Int32Int32(t *testing.T) {
	for seed := uint64(1); seed <= 8; seed++ {
		r := rand.New(rand.NewPCG(seed, 0xba5e))
		m := NewInt32Int32WithCapacity(16)
		ref := map[int32]int32{}
		for op := 0; op < randomOps; op++ {
			k := int32(r.IntN(40)) * 16
			switch r.IntN(3) {
			case 0:
				v := r.Int32()
				old, had := m.Put(k, v)
				rold, rhad := ref[k]
				if had != rhad || (had && old != rold) {
					t.Fatalf("seed %d op %d Put(%d): got (%d,%v) want (%d,%v)", seed, op, k, old, had, rold, rhad)
				}
				ref[k] = v
			case 1:
				old, had := m.Remove(k)
				rold, rhad := ref[k]
				if had != rhad || (had && old != rold) {
					t.Fatalf("seed %d op %d Remove(%d): got (%d,%v) want (%d,%v)", seed, op, k, old, had, rold, rhad)
				}
				delete(ref, k)
			default:
				v, ok := m.Get(k)
				rv, rok := ref[k]
				if ok != rok || (ok && v != rv) {
					t.Fatalf("seed %d op %d Get(%d): got (%d,%v) want (%d,%v)", seed, op, k, v, ok, rv, rok)
				}
			}
			if m.Len() != len(ref) {
				t.Fatalf("seed %d op %d: Len %d want %d", seed, op, m.Len(), len(ref))
			}
		}
	}
}
