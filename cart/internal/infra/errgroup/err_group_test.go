package errgroup

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func Test_ErrGroup(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		countGoroutine := 3

		g, _ := WithContext(context.Background())

		results := make([]int, countGoroutine)
		var id int64

		for i := 0; i < countGoroutine; i++ {
			val := i
			g.Go(func() error {
				time.Sleep(10 * time.Millisecond)
				pos := int(atomic.AddInt64(&id, 1)) - 1
				results[pos] = val
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			t.Errorf("g.Wait() %v", err)
		}

		if len(results) != countGoroutine {
			t.Errorf("len(results) != %d , got %d", countGoroutine, len(results))
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		g, ctx := WithContext(context.Background())

		g.Go(func() error {
			time.Sleep(100 * time.Millisecond)
			return errors.New("first error")
		})

		testCh := make(chan struct{})
		g.Go(func() error {
			defer close(testCh)
			<-ctx.Done()
			return ctx.Err()
		})

		if err := g.Wait(); err == nil || err.Error() != "first error" {
			t.Errorf("err nil or not 'first error' : %v", err)
		}

		select {
		case <-testCh:
		case <-time.After(200 * time.Millisecond):
			t.Error("Second goroutine did not finish in time")
		}
	})

	t.Run("cancel", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		g, _ := WithContext(ctx)

		g.Go(func() error {
			time.Sleep(100 * time.Millisecond)
			cancel()
			return nil
		})

		g.Go(func() error {
			<-ctx.Done()
			return ctx.Err()

		})

		err := g.Wait()
		if err != nil && err != context.Canceled {
			t.Errorf("err not context.Canceled : %v", err)
		}

	})
}
