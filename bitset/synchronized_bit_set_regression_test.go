package bitset

import (
	"sync"
	"testing"
	"time"
)

func TestSynchronizedBitSetSelfAlgebra(t *testing.T) {
	tests := []struct {
		name string
		op   func(*SynchronizedBitSet, *SynchronizedBitSet)
		want []int
	}{
		{"and", (*SynchronizedBitSet).AndInPlace, []int{1, 130}},
		{"or", (*SynchronizedBitSet).OrInPlace, []int{1, 130}},
		{"xor", (*SynchronizedBitSet).XorInPlace, nil},
		{"and not", (*SynchronizedBitSet).AndNotInPlace, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewSynchronizedBitSet()
			b.Set(1)
			b.Set(130)
			tt.op(b, b)
			got := b.ToSlice()
			if len(got) != len(tt.want) {
				t.Fatalf("bits = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("bits = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestSynchronizedBitSetOpposingAlgebra(t *testing.T) {
	ops := []struct {
		name string
		op   func(*SynchronizedBitSet, *SynchronizedBitSet)
	}{
		{"and", (*SynchronizedBitSet).AndInPlace},
		{"or", (*SynchronizedBitSet).OrInPlace},
		{"xor", (*SynchronizedBitSet).XorInPlace},
		{"and not", (*SynchronizedBitSet).AndNotInPlace},
	}
	for _, tt := range ops {
		t.Run(tt.name, func(t *testing.T) {
			a, b := NewSynchronizedBitSet(), NewSynchronizedBitSet()
			a.Set(1)
			b.Set(130)
			var wg sync.WaitGroup
			for _, pair := range [][2]*SynchronizedBitSet{{a, b}, {b, a}} {
				wg.Add(1)
				go func(dst, src *SynchronizedBitSet) {
					defer wg.Done()
					for range 500 {
						tt.op(dst, src)
					}
				}(pair[0], pair[1])
			}
			done := make(chan struct{})
			go func() { wg.Wait(); close(done) }()
			select {
			case <-done:
			case <-time.After(3 * time.Second):
				t.Fatal("opposing operations deadlocked")
			}
		})
	}
}

func TestSynchronizedBitSetAlgebraConcurrentSourceMutation(t *testing.T) {
	ops := []func(*SynchronizedBitSet, *SynchronizedBitSet){
		(*SynchronizedBitSet).AndInPlace,
		(*SynchronizedBitSet).OrInPlace,
		(*SynchronizedBitSet).XorInPlace,
		(*SynchronizedBitSet).AndNotInPlace,
	}
	dst, src := NewSynchronizedBitSet(), NewSynchronizedBitSet()
	dst.Set(1)
	src.Set(130)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 2000; i++ {
			src.Set(i*64 + 3)
			src.Flip(i % 128)
		}
	}()
	for i := 0; i < 2000; i++ {
		ops[i%len(ops)](dst, src)
	}
	wg.Wait()
}
