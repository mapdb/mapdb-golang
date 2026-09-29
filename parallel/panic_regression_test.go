package parallel

import (
	"sync/atomic"
	"testing"
)

func TestCallbackPanicsReachCaller(t *testing.T) {
	data := []int{1, 2, 3, 4}
	marker := new(int)
	visit := func(v int) {
		if v == 2 {
			panic(marker)
		}
	}
	predicate := func(v int) bool {
		visit(v)
		return false
	}
	allPredicate := func(v int) bool {
		visit(v)
		return true
	}
	transform := func(v int) int {
		visit(v)
		return v
	}
	view := AsParallelWith(data, 1, 2)
	for _, tc := range []struct {
		name string
		run  func()
	}{
		{"ForEach", func() { ForEachWith(data, visit, 1, 2) }},
		{"Select", func() { SelectWith(data, predicate, 1, 2) }},
		{"Reject", func() { RejectWith(data, predicate, 1, 2) }},
		{"Collect", func() { CollectWith(data, transform, 1, 2) }},
		{"Count", func() { CountWith(data, predicate, 1, 2) }},
		{"AnySatisfy", func() { AnySatisfyWith(data, predicate, 1, 2) }},
		{"view AllSatisfy", func() { view.AllSatisfy(allPredicate) }},
		{"view NoneSatisfy", func() { view.NoneSatisfy(predicate) }},
		{"view Detect", func() { view.Detect(predicate) }},
		{"view ParallelCollect", func() { ParallelCollect(view, transform) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			func() {
				defer func() {
					if got := recover(); got != marker {
						t.Errorf("recovered %v, want original callback panic", got)
					}
				}()
				tc.run()
			}()
		})
	}
}

func TestCallbackPanicWaitsForWorkers(t *testing.T) {
	var completed atomic.Int32
	defer func() {
		if got := recover(); got != "first batch" {
			t.Errorf("recovered %v, want first batch panic", got)
		}
		if got := completed.Load(); got != 1 {
			t.Errorf("completed workers = %d, want 1", got)
		}
	}()
	ForEachWith([]int{0, 1}, func(v int) {
		if v == 0 {
			panic("first batch")
		}
		completed.Add(1)
	}, 1, 2)
}

func TestMultipleCallbackPanicsChooseEarliestBatch(t *testing.T) {
	defer func() {
		if got := recover(); got != "batch 0" {
			t.Errorf("recovered %v, want batch 0 panic", got)
		}
	}()
	ForEachWith([]int{0, 1}, func(v int) {
		if v == 0 {
			panic("batch 0")
		}
		panic("batch 1")
	}, 1, 2)
}
