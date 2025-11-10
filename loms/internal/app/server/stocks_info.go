// Package server ...
package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/pkg/logger"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/model"
	pb "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/pkg/api/v1"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StocksInfo ...
func (s *Server) StocksInfo(ctx context.Context, in *pb.StocksInfoRequest) (*pb.StocksInfoResponse, error) {
	ctx, span := s.tracer.Start(
		ctx,
		model.StocksInfoHandler,
		trace.WithAttributes(
			attribute.Int64("Sku", in.GetSku()),
		),
	)
	defer span.End()

	countStock, err := s.impl.GetStocksBySku(ctx, in.GetSku())
	if err != nil {
		defer func() {
			if err != nil {
				_, span := s.tracer.Start(
					ctx,
					model.StocksInfoHandler,
					trace.WithAttributes(
						attribute.Int64("Sku", in.GetSku()),
						attribute.String("err", err.Error()),
					),
				)
				defer span.End()
				logger.Errorw(fmt.Sprintf("GetStocksBySku : %v", err), "span", span)
			}
		}()
		if errors.Is(err, model.ErrStockSkuNotFound) {
			return &pb.StocksInfoResponse{
					Count: countStock,
				},
				status.Error(codes.NotFound, model.ErrStockSkuNotFound.Error())
		}
		return nil, err
	}

	return &pb.StocksInfoResponse{
		Count: countStock,
	}, nil
}
