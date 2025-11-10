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

// OrderInfo ...
func (s *Server) OrderInfo(ctx context.Context, in *pb.OrderInfoRequest) (*pb.OrderInfoResponse, error) {
	ctx, span := s.tracer.Start(
		ctx,
		model.OrderInfoHandler,
		trace.WithAttributes(
			attribute.Int64("OrderID", in.GetOrderID()),
		),
	)
	defer span.End()

	orderInfo, err := s.impl.OrderInfo(ctx, in.GetOrderID())
	if err != nil {
		defer func() {
			if err != nil {
				_, span := s.tracer.Start(
					ctx,
					model.OrderInfoHandler,
					trace.WithAttributes(
						attribute.Int64("OrderID", in.GetOrderID()),
						attribute.String("err", err.Error()),
					),
				)
				defer span.End()
				logger.Errorw(fmt.Sprintf("OrderInfo : %v", err), "span", span)
			}
		}()
		if errors.Is(err, model.ErrOrderIDNotFound) {
			return nil, status.Error(codes.NotFound, model.ErrOrderIDNotFound.Error())
		}
		return nil, err
	}
	return orderInfoToOrderInfoResponse(*orderInfo), nil
}

func orderInfoToOrderInfoResponse(orderInfo model.OrderInfo) *pb.OrderInfoResponse {
	pbItem := make([]*pb.Item, 0, len(orderInfo.Items))
	for _, item := range orderInfo.Items {
		pbItem = append(pbItem, &pb.Item{
			Sku:   item.Sku,
			Count: item.Count,
		})
	}

	return &pb.OrderInfoResponse{
		Status: orderInfo.Status,
		UserID: orderInfo.UserID,
		Items:  pbItem,
	}
}
