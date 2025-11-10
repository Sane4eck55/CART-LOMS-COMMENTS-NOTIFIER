// Package service ...
package service

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/model"
)

// OrderCreate ...
func (s *Service) OrderCreate(ctx context.Context, order model.Order) (int64, error) {
	ctx, span := s.tracer.Start(
		ctx,
		"LomsService:OrderCreate",
	)
	defer span.End()
	items := sortItems(order.Items)
	order.Items = items

	var (
		orderID int64
		err     error
	)

	orderID, err = s.repository.CreateOrder(ctx, order)
	if err != nil {
		return model.ErrorOrderID, err
	}

	if err = s.repository.Reserve(ctx, order.Items); err != nil {
		if errors.Is(err, model.ErrNoStockForReserve) {
			if err = s.repository.SetStatusOrder(ctx, orderID, model.StatusOrderFailed); err != nil {
				return orderID, err
			}
			return orderID, model.ErrNoStockForReserve
		}

		if errors.Is(err, model.ErrStockInfoNotFound) {
			if err = s.repository.SetStatusOrder(ctx, orderID, model.StatusOrderFailed); err != nil {
				return orderID, err
			}
			return orderID, model.ErrStockInfoNotFound
		}

		if err = s.repository.SetStatusOrder(ctx, orderID, model.StatusOrderFailed); err != nil {
			return orderID, err
		}

		return orderID, fmt.Errorf("reserve : %v", err)
	}

	if err = s.repository.SetStatusOrder(ctx, orderID, model.StatusOrderAwaitingPayment); err != nil {
		return orderID, err
	}

	return orderID, err
}

func sortItems(items []model.Item) []model.Item {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Sku < items[j].Sku
	})

	return items
}
