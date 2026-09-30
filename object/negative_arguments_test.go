package object

import (
	"slices"
	"testing"
)

func TestObjectNegativeArgumentsHaveNamedPanics(t *testing.T) {
	bag := NewHashBagFrom(2, 2, 2, 1, 1, 3)
	for _, tc := range []struct {
		name string
		call func()
		want string
	}{
		{"capacity", func() { NewArrayListWithCapacity[int](-1) }, "mapdb: NewArrayListWithCapacity: negative capacity"},
		{"top-empty", func() { NewHashBag[int]().TopOccurrences(-1) }, "mapdb: TopOccurrences: negative count"},
		{"bottom-empty", func() { NewHashBag[int]().BottomOccurrences(-1) }, "mapdb: BottomOccurrences: negative count"},
		{"top-populated", func() { bag.TopOccurrences(-1) }, "mapdb: TopOccurrences: negative count"},
		{"bottom-populated", func() { bag.BottomOccurrences(-1) }, "mapdb: BottomOccurrences: negative count"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != tc.want {
					t.Errorf("panic=%v want=%q", got, tc.want)
				}
			}()
			tc.call()
		})
	}
	if bag.Len() != 6 || bag.OccurrencesOf(2) != 3 || bag.OccurrencesOf(1) != 2 || bag.OccurrencesOf(3) != 1 {
		t.Fatal("negative call changed bag")
	}
}

func TestObjectNonnegativeArguments(t *testing.T) {
	for _, capacity := range []int{0, 1, 16} {
		list := NewArrayListWithCapacity[int](capacity)
		if list.Len() != 0 {
			t.Fatal("new list nonempty")
		}
		list.Add(5)
		if !slices.Equal(slices.Collect(list.All()), []int{5}) {
			t.Fatal("list insertion")
		}
	}
	bag := NewHashBagFrom(2, 2, 2, 1, 1, 3)
	for _, n := range []int{0, 1, 3, 100} {
		want := min(n, 3)
		top, bottom := bag.TopOccurrences(n), bag.BottomOccurrences(n)
		if len(top) != want || len(bottom) != want {
			t.Fatalf("n=%d lengths %d %d", n, len(top), len(bottom))
		}
		if n > 0 && (top[0] != (ValueCount[int]{2, 3}) || bottom[0] != (ValueCount[int]{3, 1})) {
			t.Fatal("ordering")
		}
	}
	if len(NewHashBag[int]().TopOccurrences(0)) != 0 || len(NewHashBag[int]().BottomOccurrences(10)) != 0 {
		t.Fatal("empty bag")
	}
	if bag.Len() != 6 {
		t.Fatal("selection changed bag")
	}
}
