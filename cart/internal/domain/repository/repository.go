// Package repository ...
package repository

import (
	"context"
	"sync"
	"time"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/model"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/service"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/infra/metrics"
)

const (
	// timeUpdateMetricRepoSize ...
	timeUpdateMetricRepoSize = 10
)

// InMemoryRepository ...
type InMemoryRepository struct {
	storage model.Storage
	mx      sync.RWMutex
	done    chan struct{}
	tracer  service.Tracer
}

// NewInMemoryRepository ...
func NewInMemoryRepository(tracer service.Tracer) *InMemoryRepository {
	repo := InMemoryRepository{
		storage: make(model.Storage),
		done:    make(chan struct{}),
		tracer:  tracer,
	}

	go func() {
		t := time.NewTicker(timeUpdateMetricRepoSize * time.Second)
		for {
			select {
			case <-t.C:
				repo.mx.RLock()
				metrics.StoreRepoSize(float64(len(repo.storage)))
				repo.mx.RUnlock()
			case <-repo.done:
				t.Stop()
				return
			}
		}
	}()

	return &repo
}

// Add ...
func (r *InMemoryRepository) Add(ctx context.Context, cartItems model.RequestData) error {
	_, span := r.tracer.Start(ctx, "CartRepo:Add")
	defer span.End()

	r.mx.Lock()
	defer r.mx.Unlock()

	if items, ok := r.storage[cartItems.UserID]; ok {
		for i, item := range items {
			if item.SkuID == cartItems.Sku {
				items[i].Count += cartItems.Count
				return nil
			}
		}

		items = append(items, model.Cart{
			SkuID: cartItems.Sku,
			Count: cartItems.Count,
		})

		r.storage[cartItems.UserID] = items
	} else {
		items := model.Cart{
			SkuID: cartItems.Sku,
			Count: cartItems.Count,
		}
		r.storage[cartItems.UserID] = []model.Cart{items}
	}

	return nil
}

// GetItemsByUserID ...
func (r *InMemoryRepository) GetItemsByUserID(ctx context.Context, cartItems model.RequestData) ([]model.Cart, error) {
	_, span := r.tracer.Start(ctx, "CartRepo:GetItemsByUserID")
	defer span.End()

	r.mx.RLock()
	defer r.mx.RUnlock()

	if items, ok := r.storage[cartItems.UserID]; ok {
		if len(items) > 0 {
			newItems := make([]model.Cart, len(items))
			copy(newItems, items)
			return newItems, nil
		}
	}

	return nil, model.ErrNotFound
}

// DeleteItemsBySku ...
func (r *InMemoryRepository) DeleteItemsBySku(ctx context.Context, cartItems model.RequestData) error {
	_, span := r.tracer.Start(ctx, "CartRepo:DeleteItemsBySku")
	defer span.End()

	r.mx.Lock()
	defer r.mx.Unlock()

	if value, ok := r.storage[cartItems.UserID]; ok {
		for i, items := range value {
			if items.SkuID == cartItems.Sku {
				value = deleteFromMemory(value, i)
				r.storage[cartItems.UserID] = value
				break
			}
		}
	}

	return model.ErrNoContent
}

// DeleteAllItemsFromCart ...
func (r *InMemoryRepository) DeleteAllItemsFromCart(ctx context.Context, cartItems model.RequestData) error {
	_, span := r.tracer.Start(ctx, "CartRepo:DeleteAllItemsFromCart")
	defer span.End()

	r.mx.Lock()
	defer r.mx.Unlock()

	if _, ok := r.storage[cartItems.UserID]; ok {
		r.storage[cartItems.UserID] = nil
	}

	return model.ErrNoContent
}

func deleteFromMemory(items []model.Cart, i int) []model.Cart {
	copy(items[i:], items[i+1:]) //i=2 1 2 3 4 5 -> 1 2 4 5 5
	items = items[:len(items)-1]

	return items
}

// Close ...
func (r *InMemoryRepository) Close() {
	r.done <- struct{}{}
	close(r.done)
}
