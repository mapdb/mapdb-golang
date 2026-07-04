package seq

import (
	"context"
	"iter"
	"slices"
	"testing"
	"time"
)

func TestFromChannel(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	ch <- 3
	close(ch)
	if got := FromChannel(ch).ToSlice(); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("FromChannel = %v", got)
	}
}

func TestFromChannelEarlyBreakDoesNotClose(t *testing.T) {
	ch := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		ch <- i
	}
	// take 2, then stop; the channel must remain open (sender owns closing)
	got := FromChannel(ch).Take(2).ToSlice()
	if !slices.Equal(got, []int{1, 2}) {
		t.Errorf("FromChannel.Take(2) = %v", got)
	}
	// still readable — proof it was not closed
	if v := <-ch; v != 3 {
		t.Errorf("after early break, next = %d, want 3", v)
	}
}

func TestToChannel(t *testing.T) {
	out := ToChannel(context.Background(), Of(1, 2, 3, 4), 2)
	var got []int
	for v := range out {
		got = append(got, v)
	}
	if !slices.Equal(got, []int{1, 2, 3, 4}) {
		t.Errorf("ToChannel = %v", got)
	}
}

func TestToChannelContextCancelCloses(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	out := ToChannel(ctx, Iterate(0, func(n int) int { return n + 1 }), 0)
	<-out // consume one to prove it's flowing
	cancel()
	// the goroutine must observe cancellation and close out; draining terminates.
	drained := make(chan struct{})
	go func() {
		for range out {
		}
		close(drained)
	}()
	select {
	case <-drained:
	case <-time.After(2 * time.Second):
		t.Fatal("ToChannel did not close after ctx cancel")
	}
}

func TestBuffered(t *testing.T) {
	got := Buffered(context.Background(), Of(1, 2, 3, 4, 5), 2).ToSlice()
	if !slices.Equal(got, []int{1, 2, 3, 4, 5}) {
		t.Errorf("Buffered = %v", got)
	}
}

func TestBufferedReRunnable(t *testing.T) {
	b := Buffered(context.Background(), Of(1, 2, 3), 2)
	for r := 0; r < 2; r++ {
		if got := b.ToSlice(); !slices.Equal(got, []int{1, 2, 3}) {
			t.Errorf("Buffered re-run %d = %v", r, got)
		}
	}
}

func TestBufferedEarlyBreakTearsDownProducer(t *testing.T) {
	producerStopped := make(chan struct{})
	// an infinite source that signals when its iteration is abandoned
	src := Seq[int](func(yield func(int) bool) {
		defer close(producerStopped)
		for i := 0; ; i++ {
			if !yield(i) {
				return
			}
		}
	})
	var got []int
	for v := range Buffered(context.Background(), src, 4) {
		got = append(got, v)
		if len(got) == 3 {
			break
		}
	}
	if !slices.Equal(got, []int{0, 1, 2}) {
		t.Errorf("Buffered early = %v", got)
	}
	select {
	case <-producerStopped:
	case <-time.After(2 * time.Second):
		t.Fatal("producer goroutine did not stop after consumer break")
	}
}

func TestBufferedContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	nats := Iterate(0, func(n int) int { return n + 1 })
	next, stop := iter.Pull(Buffered(ctx, nats, 2).Std())
	defer stop()
	// pull a couple, then cancel; further pulls must eventually stop.
	next()
	next()
	cancel()
	// drain until the buffered seq ends (producer saw ctx.Done and closed ch)
	deadline := time.After(2 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("Buffered did not terminate after ctx cancel")
		default:
		}
		if _, ok := next(); !ok {
			return
		}
	}
}
