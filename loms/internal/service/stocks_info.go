// Package service ...
package service

import (
	"context"
	"errors"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/model"
)

// GetStocksBySku ...
func (s *Service) GetStocksBySku(ctx context.Context, sku int64) (uint32, error) {
	ctx, span := s.tracer.Start(
		ctx,
		"LomsService:GetStocksBySku",
	)
	defer span.End()

	if s.repository.UseMaster(model.RequestStock) {
		freeStock, err := s.repository.GetFreeStocksBySkuMaster(ctx, sku)
		if err != nil {
			if errors.Is(err, model.ErrStockSkuNotFound) {
				return freeStock, model.ErrStockSkuNotFound
			}
			return freeStock, err
		}
		return freeStock, nil
	}

	freeStock, err := s.repository.GetFreeStocksBySkuReplica(ctx, sku)
	if err != nil {
		if errors.Is(err, model.ErrStockSkuNotFound) {
			return freeStock, model.ErrStockSkuNotFound
		}
		return freeStock, err
	}

	return freeStock, nil
}
