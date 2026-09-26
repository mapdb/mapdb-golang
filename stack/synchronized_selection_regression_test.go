package stack

import (
	"reflect"
	"testing"
)

type selectableStack[T comparable, R any] interface {
	Push(T)
	ToSlice() []T
	Select(func(T) bool) R
	Reject(func(T) bool) R
}

type stackResult[T any] interface{ ToSlice() []T }

func checkSynchronizedSelection[T comparable, R stackResult[T], S selectableStack[T, R]](t *testing.T, newStack func() S, one, two, three, extra T) {
	t.Helper()
	for _, tc := range []struct {
		name         string
		selectValues bool
	}{
		{"Select", true},
		{"Reject", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newStack()
			s.Push(one)
			s.Push(two)
			s.Push(three)
			var visited []T
			predicate := func(v T) bool {
				visited = append(visited, v)
				s.Push(extra) // callback may re-enter and change the source
				return v == one || v == three
			}
			var result R
			if tc.selectValues {
				result = s.Select(predicate)
			} else {
				result = s.Reject(predicate)
			}
			if !reflect.DeepEqual(visited, []T{one, two, three}) {
				t.Fatalf("callback order = %v", visited)
			}
			want := []T{three, one}
			if !tc.selectValues {
				want = []T{two}
			}
			if got := result.ToSlice(); !reflect.DeepEqual(got, want) {
				t.Fatalf("result = %v, want %v", got, want)
			}
			if got := s.ToSlice(); !reflect.DeepEqual(got, []T{extra, extra, extra, three, two, one}) {
				t.Fatalf("source after callback = %v", got)
			}
		})
	}

	s := newStack()
	s.Push(one)
	func() {
		defer func() {
			if recover() == nil {
				t.Error("callback panic was lost")
			}
		}()
		s.Select(func(T) bool { panic("predicate panic") })
	}()
	s.Push(two) // panic must leave the wrapper unlocked
}

func TestSynchronizedSelectionCallbacks(t *testing.T) {
	t.Run("Int8", func(t *testing.T) {
		checkSynchronizedSelection(t, NewSynchronizedInt8, int8(1), int8(2), int8(3), int8(99))
	})
	t.Run("Int16", func(t *testing.T) {
		checkSynchronizedSelection(t, NewSynchronizedInt16, int16(1), int16(2), int16(3), int16(99))
	})
	t.Run("Int32", func(t *testing.T) {
		checkSynchronizedSelection(t, NewSynchronizedInt32, int32(1), int32(2), int32(3), int32(99))
	})
	t.Run("Int64", func(t *testing.T) {
		checkSynchronizedSelection(t, NewSynchronizedInt64, int64(1), int64(2), int64(3), int64(99))
	})
	t.Run("Float32", func(t *testing.T) {
		checkSynchronizedSelection(t, NewSynchronizedFloat32, float32(1), float32(2), float32(3), float32(99))
	})
	t.Run("Float64", func(t *testing.T) {
		checkSynchronizedSelection(t, NewSynchronizedFloat64, float64(1), float64(2), float64(3), float64(99))
	})
	t.Run("Char", func(t *testing.T) {
		checkSynchronizedSelection(t, NewSynchronizedChar, uint16(1), uint16(2), uint16(3), uint16(99))
	})
}
