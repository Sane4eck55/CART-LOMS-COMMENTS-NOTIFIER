// Package service ...
package service

import (
	"context"
	"errors"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/model"
)

// OrderCancel ...
func (s *Service) OrderCancel(ctx context.Context, orderID int64) error {
	ctx, span := s.tracer.Start(
		ctx,
		"LomsService:OrderCancel",
	)
	defer span.End()

	orderInfo, err := s.repository.GetInfoByOrderIDMaster(ctx, orderID)
	if err != nil {
		if errors.Is(err, model.ErrOrderPayNotFound) {
			return model.ErrOrderCancelNotFound
		}
		return model.ErrDefault
	}

	if orderInfo == nil {
		return model.ErrOrderCancelNotFound
	}

	if orderInfo.Status == model.StatusOrderCancelled {
		return model.ErrOrderAlreadyCanceled
	}

	if orderInfo.Status == model.StatusOrderPaid || orderInfo.Status == model.StatusOrderFailed {
		return model.ErrOrderStatusFailedOrPaid
	}

	for _, item := range orderInfo.Items {
		// nolint:govet
		if err := s.repository.ReserveCancel(ctx, item); err != nil {
			return err
		}
	}

	if err = s.repository.SetStatusOrder(ctx, orderID, model.StatusOrderCancelled); err != nil {
		return model.ErrDefault
	}

	return nil
}
