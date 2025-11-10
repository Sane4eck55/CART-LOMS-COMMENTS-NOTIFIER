// Package service ...
package service

import (
	"context"
	"errors"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/model"
)

// OrderInfo ...
func (s *Service) OrderInfo(ctx context.Context, orderID int64) (*model.OrderInfo, error) {
	ctx, span := s.tracer.Start(
		ctx,
		"LomsService:OrderInfo",
	)
	defer span.End()

	var (
		info *model.OrderInfo
		err  error
	)

	if s.repository.UseMaster(model.RequestOrder) {
		info, err = s.repository.GetInfoByOrderIDMaster(ctx, orderID)
		if err != nil {
			if errors.Is(err, model.ErrOrderPayNotFound) {
				return nil, model.ErrOrderIDNotFound
			}
			return nil, err
		}
	} else {
		info, err = s.repository.GetInfoByOrderIDReplica(ctx, orderID)
		if err != nil {
			if errors.Is(err, model.ErrOrderPayNotFound) {
				return nil, model.ErrOrderIDNotFound
			}
			return nil, err
		}
	}

	if info == nil {
		return nil, model.ErrOrderIDNotFound
	}

	return info, nil
}
