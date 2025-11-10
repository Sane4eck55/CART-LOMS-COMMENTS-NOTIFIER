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

// OrderCreate ...
func (s *Server) OrderCreate(ctx context.Context, in *pb.OrderCreateRequest) (*pb.OrderCreateResponse, error) {
	var attrs []attribute.KeyValue
	for _, item := range in.GetItems() {
		attrs = append(attrs,
			attribute.Int64("UserID", in.GetUserID()),
			attribute.Int64("Sku", item.GetSku()),
			attribute.Int64("Count", int64(item.GetCount())),
		)
	}

	ctx, span := s.tracer.Start(
		ctx,
		model.OrderCreateHandler,
		trace.WithAttributes(
			attrs...,
		),
	)
	defer span.End()

	order := preparedInputOrderCreate(in)
	orderID, err := s.impl.OrderCreate(ctx, order)
	if err != nil {
		defer func() {
			if err != nil {
				_, span := s.tracer.Start(
					ctx,
					model.OrderCreateHandler,
					trace.WithAttributes(
						attribute.Int64("UserID", order.UserID),
						attribute.String("err", err.Error()),
					),
				)
				defer span.End()
				logger.Errorw(fmt.Sprintf("OrderCreate : %v", err), "span", span)
			}
		}()
		if errors.Is(err, model.ErrNoStockForReserve) {
			return &pb.OrderCreateResponse{
					OrderID: orderID,
				},
				status.Error(codes.FailedPrecondition, model.ErrNoStockForReserve.Error())
		}
		if errors.Is(err, model.ErrStockInfoNotFound) {
			return &pb.OrderCreateResponse{
					OrderID: orderID,
				},
				status.Error(codes.FailedPrecondition, model.ErrStockInfoNotFound.Error())
		}
		return &pb.OrderCreateResponse{
			OrderID: orderID,
		}, err
	}
	return &pb.OrderCreateResponse{
		OrderID: orderID,
	}, nil
}

func preparedInputOrderCreate(in *pb.OrderCreateRequest) model.Order {

	items := make([]model.Item, 0, len(in.Items))

	for _, item := range in.GetItems() {
		items = append(items,
			model.Item{
				Sku:   item.Sku,
				Count: item.Count,
			},
		)
	}

	return model.Order{
		UserID: in.UserID,
		Items:  items,
	}
}
