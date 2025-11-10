// Package service ...
package service

import (
	"context"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/model"
)

// OrderPay ...
func (s *Service) OrderPay(ctx context.Context, orderID int64) error {
	ctx, span := s.tracer.Start(
		ctx,
		"LomsService:OrderPay",
	)
	defer span.End()

	info, err := s.repository.GetInfoByOrderIDMaster(ctx, orderID)
	if err != nil {
		return err
	}

	if info == nil {
		return model.ErrOrderPayNotFound
	}

	if info.Status == model.StatusOrderPaid {
		return model.ErrOrderAlreadyPay
	}

	if info.Status != model.StatusOrderAwaitingPayment {
		return model.ErrOrderStatusNotAwaitingPayment
	}

	for _, item := range info.Items {
		if err := s.repository.ReserveRemove(ctx, item); err != nil {
			return err
		}
	}

	if err := s.repository.SetStatusOrder(ctx, orderID, model.StatusOrderPaid); err != nil {
		return err
	}

	return nil
}
