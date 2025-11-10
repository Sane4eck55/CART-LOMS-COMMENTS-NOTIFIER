// Package outbox ...
package outbox

import (
	"context"
	"sync"
	"time"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/app/server"
)

// tickerTime ...
const tickerTime = 3 * time.Second

// Outbox ...
type Outbox struct {
	ctx       context.Context
	cancel    context.CancelFunc
	waitGroup sync.WaitGroup
	stopChan  chan struct{}
}

// NewOutbox ...
func NewOutbox(ctx context.Context) *Outbox {
	ctx, cancel := context.WithCancel(ctx)
	return &Outbox{
		ctx:      ctx,
		cancel:   cancel,
		stopChan: make(chan struct{}),
	}
}

// Start ...
func (o *Outbox) Start(server server.LomsService) {
	o.waitGroup.Add(1)
	go func() {
		defer o.waitGroup.Done()
		ticker := time.NewTicker(tickerTime)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				server.ProduceFromOutbox(o.ctx)
			case <-o.ctx.Done():
				return
			}
		}
	}()
}

// Stop ...
func (o *Outbox) Stop() {
	o.cancel()
	o.waitGroup.Wait()
}
