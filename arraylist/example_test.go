package arraylist_test

import (
	"fmt"

	"github.com/mapdb/mapdb-golang/arraylist"
)

func ExampleInt32() {
	list := arraylist.Int32Of(5, 3, 1, 4, 2)
	fmt.Println("Size:", list.Len())
	v := list.Get(0)
	fmt.Println("Get(0):", v)

	list.Sort()
	fmt.Println("After sort:", list.ToSlice())
	fmt.Println("Sum:", list.Sum())
	// Output:
	// Size: 5
	// Get(0): 5
	// After sort: [1 2 3 4 5]
	// Sum: 15
}

func ExampleInt32_Select() {
	list := arraylist.Int32Of(1, 2, 3, 4, 5, 6, 7, 8)
	evens := list.Select(func(v int32) bool { return v%2 == 0 })
	fmt.Println("Evens:", evens.ToSlice())
	// Output:
	// Evens: [2 4 6 8]
}

func ExampleInt32_All() {
	list := arraylist.Int32Of(10, 20, 30)
	for v := range list.All() {
		fmt.Println(v)
	}
	// Output:
	// 10
	// 20
	// 30
}

func ExampleInt32_BinarySearch() {
	list := arraylist.Int32Of(10, 20, 30, 40, 50)
	idx, found := list.BinarySearch(30)
	fmt.Printf("BinarySearch(30): index=%d, found=%v\n", idx, found)

	idx, found = list.BinarySearch(25)
	fmt.Printf("BinarySearch(25): index=%d, found=%v\n", idx, found)
	// Output:
	// BinarySearch(30): index=2, found=true
	// BinarySearch(25): index=2, found=false
}

func ExampleInt32_Min() {
	list := arraylist.Int32Of(30, 10, 50, 20, 40)
	min, _ := list.Min()
	max, _ := list.Max()
	fmt.Printf("Min: %d, Max: %d\n", min, max)
	// Output:
	// Min: 10, Max: 50
}
