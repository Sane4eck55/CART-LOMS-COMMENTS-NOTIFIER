// Package order ...
package order

import (
	"context"
	"sync"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/model"
)

// Repo ...
type Repo struct {
	storage model.OrderStorage
	mx      sync.RWMutex
}

// NewOrderRepo ...
func NewOrderRepo() *Repo {
	return &Repo{
		storage: make(model.OrderStorage),
	}
}

// CreateOrder ...
func (or *Repo) CreateOrder(_ context.Context, usersOrders model.Order) (int64, error) {
	or.mx.Lock()
	defer or.mx.Unlock()

	items := make([]model.Item, 0, len(usersOrders.Items))

	for _, item := range usersOrders.Items {
		items = append(items, model.Item{
			Sku:   item.Sku,
			Count: item.Count,
		})
	}

	orderID := int64(len(or.storage) + 1)

	order := model.OrderInfo{
		UserID: usersOrders.UserID,
		Status: model.StatusOrderNew,
		Items:  items,
	}
	or.storage[orderID] = order

	return orderID, nil
}

// SetStatusOrder ...
func (or *Repo) SetStatusOrder(_ context.Context, orderID int64, status string) error {
	or.mx.Lock()
	defer or.mx.Unlock()

	if order, ok := or.storage[orderID]; ok {
		newOrder := model.OrderInfo{
			UserID: order.UserID,
			Status: status,
			Items:  order.Items,
		}
		or.storage[orderID] = newOrder
	}

	return nil
}

// GetInfoByOrderID ...
func (or *Repo) GetInfoByOrderID(_ context.Context, orderID int64) (*model.OrderInfo, error) {
	or.mx.RLock()
	defer or.mx.RUnlock()

	if order, ok := or.storage[orderID]; ok {
		return &order, nil
	}

	return nil, nil
}
