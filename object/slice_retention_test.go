package object

import "testing"

func TestArrayListClearsRemovedBackingSlots(t *testing.T) {
	a, b, c := new(int), new(int), new(int)
	list := NewArrayListFrom(a, b, c)
	if !list.Remove(b) || len(list.items) != 2 || list.items[0] != a || list.items[1] != c {
		t.Fatal("Remove did not preserve the surviving elements")
	}
	if list.items[:cap(list.items)][2] != nil {
		t.Fatal("Remove retained a pointer in the hidden tail")
	}
	list.Clear()
	for _, v := range list.items[:cap(list.items)] {
		if v != nil {
			t.Fatal("Clear retained a pointer in the backing array")
		}
	}
}

func TestArrayStackClearsRemovedBackingSlots(t *testing.T) {
	a, b := new(int), new(int)
	stack := NewArrayStackFrom(a, b)
	if top, ok := stack.Pop(); !ok || top != b || stack.Len() != 1 {
		t.Fatal("Pop did not return the top element")
	}
	if stack.items[:cap(stack.items)][1] != nil {
		t.Fatal("Pop retained a pointer in the hidden tail")
	}
	stack.Clear()
	for _, v := range stack.items[:cap(stack.items)] {
		if v != nil {
			t.Fatal("Clear retained a pointer in the backing array")
		}
	}
}

func TestMultimapRemoveMatchingClearsCompactedTail(t *testing.T) {
	a, b, c := new(int), new(int), new(int)
	equal := func(x, y *int) bool { return x == y }

	hash := NewHashMultimap[string, *int]()
	hash.PutAll("k", a, b, c)
	hashVals, _ := hash.m.get("k")
	if hash.RemoveMatching("k", b, equal) != 1 || len(hash.Get("k")) != 2 {
		t.Fatal("hash multimap failed to remove the matching value")
	}
	if hashVals[2] != nil {
		t.Fatal("hash multimap retained a pointer in the compacted tail")
	}

	tree := NewTreeMultimap[string, *int](NaturalComparator[string]())
	tree.PutAll("k", a, b, c)
	treeVals, _ := tree.tm.Get("k")
	if tree.RemoveMatching("k", b, equal) != 1 || len(tree.Get("k")) != 2 {
		t.Fatal("tree multimap failed to remove the matching value")
	}
	if treeVals[2] != nil {
		t.Fatal("tree multimap retained a pointer in the compacted tail")
	}
}
