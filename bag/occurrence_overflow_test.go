package bag

import (
	"iter"
	"math"
	"testing"
)

// overflowBag is the mutable surface every primitive bag family shares.
type overflowBag[T comparable] interface {
	Add(value T)
	AddOccurrences(value T, occurrences int) int
	OccurrencesOf(value T) int
	Contains(value T) bool
	Len() int
	SizeDistinct() int
	AllWithOccurrences() iter.Seq2[T, int]
}

type overflowFamily[T comparable] struct {
	name   string
	mk     func() overflowBag[T]
	addAll func(overflowBag[T], ...T)
}

// snapshot captures everything a refused add must leave untouched. It reads
// distinct (value, count) pairs, never expanding MaxInt occurrences.
func snapshot[T comparable](b overflowBag[T]) (int, int, map[T]int) {
	pairs := map[T]int{}
	for v, n := range b.AllWithOccurrences() {
		pairs[v] = n
	}
	return b.Len(), b.SizeDistinct(), pairs
}

func requireRefused[T comparable](t *testing.T, b overflowBag[T], what string, op func()) {
	t.Helper()
	size, distinct, pairs := snapshot(b)
	panicked := func() (p bool) {
		defer func() { p = recover() != nil }()
		op()
		return false
	}()
	if !panicked {
		t.Fatalf("%s: no overflow panic", what)
	}
	size2, distinct2, pairs2 := snapshot(b)
	if size2 != size || distinct2 != distinct || len(pairs2) != len(pairs) {
		t.Fatalf("%s: bag changed: Len %d->%d SizeDistinct %d->%d", what, size, size2, distinct, distinct2)
	}
	for v, n := range pairs {
		if pairs2[v] != n || b.OccurrencesOf(v) != n || !b.Contains(v) {
			t.Fatalf("%s: count of %v changed: %d -> %d", what, v, n, pairs2[v])
		}
	}
}

// checkOccurrenceOverflow drives the per-value count and the total size to
// exactly MaxInt (accepted) and then one past it (refused, bag unchanged),
// without allocating an occurrence per count.
func checkOccurrenceOverflow[T comparable](t *testing.T, f overflowFamily[T], a, c T) {
	t.Run(f.name, func(t *testing.T) {
		// Per-value count boundary: one value holds every occurrence.
		b := f.mk()
		b.AddOccurrences(a, math.MaxInt-1)
		b.Add(a)
		if b.OccurrencesOf(a) != math.MaxInt || b.Len() != math.MaxInt {
			t.Fatalf("exact max: count %d Len %d", b.OccurrencesOf(a), b.Len())
		}
		requireRefused(t, b, "Add(a) at max", func() { b.Add(a) })
		requireRefused(t, b, "Add(new) at max", func() { b.Add(c) })
		requireRefused(t, b, "AddOccurrences(a, 1)", func() { b.AddOccurrences(a, 1) })
		requireRefused(t, b, "AddOccurrences(new, 1)", func() { b.AddOccurrences(c, 1) })
		requireRefused(t, b, "AddOccurrences(a, MaxInt)", func() { b.AddOccurrences(a, math.MaxInt) })
		if b.Contains(c) {
			t.Fatal("refused add of a new value left it present")
		}
		if f.addAll != nil {
			requireRefused(t, b, "AddAllReturning(new)", func() { f.addAll(b, c) })
		}
		if got := b.AddOccurrences(a, 0); got != math.MaxInt {
			t.Fatalf("AddOccurrences(a, 0) at max = %d", got)
		}

		// Total-size boundary: two values whose counts sum to exactly MaxInt.
		b = f.mk()
		b.AddOccurrences(a, math.MaxInt/2+1)
		if got := b.AddOccurrences(c, math.MaxInt/2); got != math.MaxInt/2 || b.Len() != math.MaxInt {
			t.Fatalf("exact total max: count %d Len %d", got, b.Len())
		}
		requireRefused(t, b, "AddOccurrences(c, 1) past total", func() { b.AddOccurrences(c, 1) })
		requireRefused(t, b, "Add(a) past total", func() { b.Add(a) })
		if f.addAll != nil {
			b = f.mk()
			b.AddOccurrences(a, math.MaxInt-1)
			// One of two values fits; all or nothing refuses both.
			requireRefused(t, b, "AddAllReturning(c, c) partial fit", func() { f.addAll(b, c, c) })
		}

		// After a refusal the bag keeps working.
		b.AddOccurrences(a, 0)
		_ = b.AddOccurrences(c, 0)
	})
}

func families[T comparable, H overflowBag[T], R overflowBag[T], S overflowBag[T]](
	name string, newHash func() H, newTree func() R, newSync func() S,
	hashAll func(H, ...T), treeAll func(R, ...T), syncAll func(S, ...T),
) []overflowFamily[T] {
	return []overflowFamily[T]{
		{"Hash" + name, func() overflowBag[T] { return newHash() }, func(b overflowBag[T], v ...T) { hashAll(b.(H), v...) }},
		{"Tree" + name, func() overflowBag[T] { return newTree() }, func(b overflowBag[T], v ...T) { treeAll(b.(R), v...) }},
		{"SynchronizedHash" + name, func() overflowBag[T] { return newSync() }, func(b overflowBag[T], v ...T) { syncAll(b.(S), v...) }},
	}
}

func TestOccurrenceOverflowPanicsAndLeavesBagUnchanged(t *testing.T) {
	for _, f := range families("Int8", NewHashInt8, NewTreeInt8, NewSynchronizedHashInt8,
		func(b *HashInt8, v ...int8) { b.AddAllReturning(v...) },
		func(b *TreeInt8, v ...int8) { b.AddAllReturning(v...) },
		func(b *SynchronizedHashInt8, v ...int8) { b.AddAllReturning(v...) }) {
		checkOccurrenceOverflow(t, f, int8(-3), int8(7))
	}
	for _, f := range families("Int16", NewHashInt16, NewTreeInt16, NewSynchronizedHashInt16,
		func(b *HashInt16, v ...int16) { b.AddAllReturning(v...) },
		func(b *TreeInt16, v ...int16) { b.AddAllReturning(v...) },
		func(b *SynchronizedHashInt16, v ...int16) { b.AddAllReturning(v...) }) {
		checkOccurrenceOverflow(t, f, int16(-3), int16(7))
	}
	for _, f := range families("Int32", NewHashInt32, NewTreeInt32, NewSynchronizedHashInt32,
		func(b *HashInt32, v ...int32) { b.AddAllReturning(v...) },
		func(b *TreeInt32, v ...int32) { b.AddAllReturning(v...) },
		func(b *SynchronizedHashInt32, v ...int32) { b.AddAllReturning(v...) }) {
		checkOccurrenceOverflow(t, f, int32(-3), int32(7))
	}
	for _, f := range families("Int64", NewHashInt64, NewTreeInt64, NewSynchronizedHashInt64,
		func(b *HashInt64, v ...int64) { b.AddAllReturning(v...) },
		func(b *TreeInt64, v ...int64) { b.AddAllReturning(v...) },
		func(b *SynchronizedHashInt64, v ...int64) { b.AddAllReturning(v...) }) {
		checkOccurrenceOverflow(t, f, int64(math.MinInt64), int64(7))
	}
	for _, f := range families("Char", NewHashChar, NewTreeChar, NewSynchronizedHashChar,
		func(b *HashChar, v ...uint16) { b.AddAllReturning(v...) },
		func(b *TreeChar, v ...uint16) { b.AddAllReturning(v...) },
		func(b *SynchronizedHashChar, v ...uint16) { b.AddAllReturning(v...) }) {
		checkOccurrenceOverflow(t, f, uint16('a'), uint16('z'))
	}
	for _, f := range families("Float32", NewHashFloat32, NewTreeFloat32, NewSynchronizedHashFloat32,
		func(b *HashFloat32, v ...float32) { b.AddAllReturning(v...) },
		func(b *TreeFloat32, v ...float32) { b.AddAllReturning(v...) },
		func(b *SynchronizedHashFloat32, v ...float32) { b.AddAllReturning(v...) }) {
		checkOccurrenceOverflow(t, f, float32(-0.5), float32(2.5))
	}
	for _, f := range families("Float64", NewHashFloat64, NewTreeFloat64, NewSynchronizedHashFloat64,
		func(b *HashFloat64, v ...float64) { b.AddAllReturning(v...) },
		func(b *TreeFloat64, v ...float64) { b.AddAllReturning(v...) },
		func(b *SynchronizedHashFloat64, v ...float64) { b.AddAllReturning(v...) }) {
		checkOccurrenceOverflow(t, f, -0.5, 2.5)
	}
}
