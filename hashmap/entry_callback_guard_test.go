package hashmap

import (
	"strings"
	"testing"
)

// Entry callbacks must not mutate the map (owner ruling 2026-10-02, plan row
// T2). AndModify hands its callback a pointer into the live slot: a Remove
// inside the callback used to backward-shift another key into that slot, so
// the callback's write landed on the wrong key with no panic. Every mutator
// now panics before changing anything while an Entry callback runs.

type guardNum interface {
	~int8 | ~int16 | ~int32 | ~int64 | ~uint16 | ~float32 | ~float64
}

type guardEntry[V any, E any] interface {
	AndModify(func(*V)) E
	OrInsert(V) V
	OrInsertWith(func() V) V
}

type guardMap[K, V any, M any, E any] interface {
	Put(K, V) (V, bool)
	Get(K) (V, bool)
	Remove(K) (V, bool)
	Len() int
	Clear()
	ContainsKey(K) bool
	AddToValue(K, V) V
	UpdateValue(K, V, func(V) V) V
	PutReturning(K, V) M
	RemoveKeyReturning(K) M
	WithoutAllKeys([]K) M
	Entry(K) E
}

// expectGuardPanic runs f and requires the Entry-callback guard's panic.
func expectGuardPanic(t *testing.T, what string, f func()) {
	t.Helper()
	defer func() {
		t.Helper()
		r := recover()
		if r == nil {
			// Errorf, not Fatalf: let the caller's state checks report the damage.
			t.Errorf("%s: expected a panic", what)
			return
		}
		if msg, ok := r.(string); !ok || !strings.Contains(msg, "Entry callback") {
			t.Fatalf("%s: unexpected panic %v", what, r)
		}
	}()
	f()
}

func runEntryGuardTests[K, V guardNum, E guardEntry[V, E], M guardMap[K, V, M, E]](
	t *testing.T, newMap func() M, home func(M, K) int,
) {
	const n = 20
	fill := func() M {
		m := newMap()
		for i := 1; i <= n; i++ {
			m.Put(K(i), V(10+i))
		}
		return m
	}
	// intact checks that m still holds exactly keys 1..n with their values.
	intact := func(t *testing.T, what string, m M) {
		t.Helper()
		if m.Len() != n {
			t.Fatalf("%s: Len = %d, want %d", what, m.Len(), n)
		}
		for i := 1; i <= n; i++ {
			if v, ok := m.Get(K(i)); !ok || v != V(10+i) {
				t.Fatalf("%s: Get(%d) = (%v, %v), want (%v, true)", what, i, v, ok, V(10+i))
			}
		}
	}
	// usable checks the map still accepts mutations after a recovered panic.
	usable := func(t *testing.T, what string, m M) {
		t.Helper()
		m.Put(K(n+1), V(1))
		if _, ok := m.Remove(K(n + 1)); !ok {
			t.Fatalf("%s: map not usable after recovered panic", what)
		}
	}

	mutators := map[string]func(m M){
		"Put new":            func(m M) { m.Put(K(n+5), V(1)) },
		"Put existing":       func(m M) { m.Put(K(2), V(99)) },
		"Remove":             func(m M) { m.Remove(K(3)) },
		"Remove absent":      func(m M) { m.Remove(K(n + 7)) },
		"Clear":              func(m M) { m.Clear() },
		"AddToValue":         func(m M) { m.AddToValue(K(4), V(1)) },
		"UpdateValue":        func(m M) { m.UpdateValue(K(5), V(0), func(v V) V { return v + 1 }) },
		"PutReturning":       func(m M) { m.PutReturning(K(n+6), V(1)) },
		"RemoveKeyReturning": func(m M) { m.RemoveKeyReturning(K(6)) },
		"WithoutAllKeys":     func(m M) { m.WithoutAllKeys([]K{K(7), K(8)}) },
		"Entry.OrInsert":     func(m M) { m.Entry(K(n + 8)).OrInsert(V(1)) },
		"Entry.OrInsertWith": func(m M) { m.Entry(K(n + 9)).OrInsertWith(func() V { return 1 }) },
	}

	t.Run("RemoveShiftRepro", func(t *testing.T) {
		// Two keys sharing a home slot: b sits right after a, so removing a
		// backward-shifts b into a's slot, where AndModify's pointer points.
		m := newMap()
		a := K(1)
		h := home(m, a)
		b := a
		for i := 2; i < 128; i++ {
			if home(m, K(i)) == h {
				b = K(i)
				break
			}
		}
		if b == a {
			t.Fatal("no colliding key found")
		}
		m.Put(a, V(20))
		m.Put(b, V(30))
		expectGuardPanic(t, "Remove inside AndModify", func() {
			m.Entry(a).AndModify(func(p *V) {
				m.Remove(a)
				*p = V(99)
			})
		})
		if vb, ok := m.Get(b); !ok || vb != V(30) {
			t.Fatalf("after: Get(b) = (%v, %v), want (30, true) — the callback's write hit b", vb, ok)
		}
		if va, ok := m.Get(a); !ok || va != V(20) {
			t.Fatalf("after: Get(a) = (%v, %v), want (20, true)", va, ok)
		}
		if m.Len() != 2 {
			t.Fatalf("after: Len = %d, want 2", m.Len())
		}
	})

	for name, mutate := range mutators {
		t.Run("AndModify/"+name, func(t *testing.T) {
			m := fill()
			expectGuardPanic(t, name, func() {
				m.Entry(K(1)).AndModify(func(p *V) { mutate(m) })
			})
			intact(t, name, m)
			usable(t, name, m)
		})
		t.Run("OrInsertWith/"+name, func(t *testing.T) {
			m := fill()
			expectGuardPanic(t, name, func() {
				m.Entry(K(n + 3)).OrInsertWith(func() V { mutate(m); return V(1) })
			})
			intact(t, name, m)
			if m.ContainsKey(K(n + 3)) {
				t.Fatalf("%s: OrInsertWith inserted despite the panic", name)
			}
			usable(t, name, m)
		})
	}

	t.Run("ReadsAndWriteThroughPointer", func(t *testing.T) {
		m := fill()
		m.Entry(K(1)).AndModify(func(p *V) {
			if v, ok := m.Get(K(2)); !ok || v != V(12) {
				t.Fatalf("Get inside callback = (%v, %v)", v, ok)
			}
			if !m.ContainsKey(K(3)) || m.Len() != n {
				t.Fatal("ContainsKey/Len inside callback")
			}
			*p = V(77)
		})
		if v, _ := m.Get(K(1)); v != V(77) {
			t.Fatalf("AndModify write = %v, want 77", v)
		}
		got := m.Entry(K(n + 4)).OrInsertWith(func() V {
			v, _ := m.Get(K(2))
			return v + 1
		})
		if got != V(13) || m.Len() != n+1 {
			t.Fatalf("OrInsertWith = %v, Len %d", got, m.Len())
		}
	})

	t.Run("CallbackPanicUnlocks", func(t *testing.T) {
		m := fill()
		func() {
			defer func() {
				if r := recover(); r != "boom" {
					t.Fatalf("recovered %v, want boom", r)
				}
			}()
			m.Entry(K(1)).AndModify(func(p *V) { panic("boom") })
		}()
		func() {
			defer func() { _ = recover() }()
			m.Entry(K(n + 2)).OrInsertWith(func() V { panic("boom") })
		}()
		intact(t, "after callback panic", m)
		usable(t, "after callback panic", m)
	})

	t.Run("NestedCallbackKeepsOuterLocked", func(t *testing.T) {
		m := fill()
		expectGuardPanic(t, "Remove after nested AndModify", func() {
			m.Entry(K(1)).AndModify(func(p *V) {
				m.Entry(K(2)).AndModify(func(q *V) { *q = V(55) })
				m.Remove(K(3))
			})
		})
		if v, _ := m.Get(K(2)); v != V(55) {
			t.Fatalf("nested AndModify write = %v, want 55", v)
		}
		if m.Len() != n || !m.ContainsKey(K(3)) {
			t.Fatal("outer callback was unlocked by the nested one")
		}
		usable(t, "after nested", m)
	})
}

// One family per key primitive; all families share the generator template.
func TestEntryCallbackGuard(t *testing.T) {
	t.Run("Int8Int64", func(t *testing.T) {
		runEntryGuardTests[int8, int64, Int8Int64Entry](t, NewInt8Int64,
			func(m *Int8Int64, k int8) int { return m.home(k) })
	})
	t.Run("Int16Float32", func(t *testing.T) {
		runEntryGuardTests[int16, float32, Int16Float32Entry](t, NewInt16Float32,
			func(m *Int16Float32, k int16) int { return m.home(k) })
	})
	t.Run("Int32Int32", func(t *testing.T) {
		runEntryGuardTests[int32, int32, Int32Int32Entry](t, NewInt32Int32,
			func(m *Int32Int32, k int32) int { return m.home(k) })
	})
	t.Run("Int64Char", func(t *testing.T) {
		runEntryGuardTests[int64, uint16, Int64CharEntry](t, NewInt64Char,
			func(m *Int64Char, k int64) int { return m.home(k) })
	})
	t.Run("CharInt16", func(t *testing.T) {
		runEntryGuardTests[uint16, int16, CharInt16Entry](t, NewCharInt16,
			func(m *CharInt16, k uint16) int { return m.home(k) })
	})
	t.Run("Float32Float64", func(t *testing.T) {
		runEntryGuardTests[float32, float64, Float32Float64Entry](t, NewFloat32Float64,
			func(m *Float32Float64, k float32) int { return m.home(k) })
	})
	t.Run("Float64Int8", func(t *testing.T) {
		runEntryGuardTests[float64, int8, Float64Int8Entry](t, NewFloat64Int8,
			func(m *Float64Int8, k float64) int { return m.home(k) })
	})
}
