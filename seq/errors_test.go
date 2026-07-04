package seq

import (
	"errors"
	"iter"
	"slices"
	"testing"
)

// fallible builds an iter.Seq2[int, error] from parallel value/error slices.
func fallible(vals []int, errs []error) iter.Seq2[int, error] {
	return func(yield func(int, error) bool) {
		for i, v := range vals {
			var e error
			if i < len(errs) {
				e = errs[i]
			}
			if !yield(v, e) {
				return
			}
		}
	}
}

func TestStopOnErr(t *testing.T) {
	boom := errors.New("boom")
	s, errf := StopOnErr(fallible([]int{1, 2, 3, 4}, []error{nil, nil, boom, nil}))
	got := s.ToSlice()
	if !slices.Equal(got, []int{1, 2}) {
		t.Errorf("StopOnErr values = %v, want [1 2]", got)
	}
	if errf() != boom {
		t.Errorf("StopOnErr err = %v, want boom", errf())
	}
}

func TestStopOnErrClean(t *testing.T) {
	s, errf := StopOnErr(fallible([]int{1, 2, 3}, nil))
	if got := s.ToSlice(); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("StopOnErr clean = %v", got)
	}
	if errf() != nil {
		t.Errorf("StopOnErr clean err = %v, want nil", errf())
	}
	// re-range resets the error state
	s2, errf2 := StopOnErr(fallible([]int{1}, nil))
	_ = s2.ToSlice()
	_ = s2.ToSlice()
	if errf2() != nil {
		t.Errorf("re-range err = %v, want nil", errf2())
	}
}

func TestSkipErr(t *testing.T) {
	boom := errors.New("boom")
	var seen []error
	s := SkipErr(fallible([]int{1, 2, 3, 4}, []error{nil, boom, nil, boom}),
		func(e error) { seen = append(seen, e) })
	if got := s.ToSlice(); !slices.Equal(got, []int{1, 3}) {
		t.Errorf("SkipErr = %v, want [1 3]", got)
	}
	if len(seen) != 2 {
		t.Errorf("SkipErr onErr called %d times, want 2", len(seen))
	}
	// re-runnable: no cross-call state
	if got := s.ToSlice(); !slices.Equal(got, []int{1, 3}) {
		t.Errorf("SkipErr re-run = %v", got)
	}
}

func TestMustAll(t *testing.T) {
	if got := MustAll(fallible([]int{1, 2, 3}, nil)).ToSlice(); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("MustAll clean = %v", got)
	}
	defer func() {
		if recover() == nil {
			t.Error("MustAll did not panic on error")
		}
	}()
	MustAll(fallible([]int{1, 2}, []error{nil, errors.New("x")})).ToSlice()
}

func TestWithErrRoundTrip(t *testing.T) {
	// lift then stop-on-err should recover the originals with no error
	s, errf := StopOnErr(WithErr(Of(1, 2, 3)).Std2())
	if got := s.ToSlice(); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("WithErr round-trip = %v", got)
	}
	if errf() != nil {
		t.Errorf("WithErr round-trip err = %v", errf())
	}
}
