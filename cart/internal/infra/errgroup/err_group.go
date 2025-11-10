// Package errgroup ...
package errgroup

import (
	"context"
	"sync"
)

// Group ...
type Group struct {
	ctx    context.Context
	cancel context.CancelFunc

	wg sync.WaitGroup

	errOnce sync.Once
	mu      sync.Mutex
	first   error
}

// WithContext ...
func WithContext(ctx context.Context) (*Group, context.Context) {
	newCtx, cancel := context.WithCancel(ctx)
	return &Group{
		ctx:    newCtx,
		cancel: cancel,
	}, newCtx
}

// GoWithMutex реализация с мьютексами
func (g *Group) GoWithMutex(f func() error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if err := f(); err != nil {
			g.mu.Lock()
			defer g.mu.Unlock()
			if g.first == nil {
				g.first = err
				g.cancel()
			}
		}
	}()
}

// Go ...
func (g *Group) Go(f func() error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		select {
		case <-g.ctx.Done():
			return
		default:
			if err := f(); err != nil {
				g.errOnce.Do(func() {
					g.first = err
					g.cancel()
				})
			}
		}
	}()
}

// Wait ...
func (g *Group) Wait() error {
	g.wg.Wait()
	return g.first
}
