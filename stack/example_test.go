package stack_test

import (
	"fmt"

	"github.com/mapdb/mapdb-golang/stack"
)

func ExampleInt32ArrayStack() {
	s := stack.NewInt32ArrayStack()
	s.Push(10)
	s.Push(20)
	s.Push(30)

	peek, _ := s.Peek()
	fmt.Println("Peek:", peek)
	pop1, _ := s.Pop()
	fmt.Println("Pop:", pop1)
	pop2, _ := s.Pop()
	fmt.Println("Pop:", pop2)
	fmt.Println("Size:", s.Size())
	// Output:
	// Peek: 30
	// Pop: 30
	// Pop: 20
	// Size: 1
}

func ExampleImmutableInt32ArrayStack() {
	// Immutable stacks use persistent operations —
	// Push/Pop return NEW stacks, originals are unchanged.
	s := stack.NewImmutableInt32ArrayStack(1, 2, 3)

	s2 := s.Push(4)
	fmt.Println("Original size:", s.Size())
	fmt.Println("After push size:", s2.Size())
	peek, _ := s2.Peek()
	fmt.Println("After push peek:", peek)

	s3, top, _ := s2.Pop()
	fmt.Println("Popped:", top)
	fmt.Println("After pop size:", s3.Size())
	// Output:
	// Original size: 3
	// After push size: 4
	// After push peek: 4
	// Popped: 4
	// After pop size: 3
}
