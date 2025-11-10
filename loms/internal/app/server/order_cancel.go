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

// OrderCancel ...
func (s *Server) OrderCancel(ctx context.Context, in *pb.OrderCancelRequest) (*pb.OrderCancelResponse, error) {
	ctx, span := s.tracer.Start(
		ctx,
		model.OrderCancelHandler,
		trace.WithAttributes(
			attribute.Int64("OrderID", in.GetOrderID()),
		),
	)
	defer span.End()

	if err := s.impl.OrderCancel(ctx, in.GetOrderID()); err != nil {
		defer func() {
			if err != nil {
				_, span := s.tracer.Start(
					ctx,
					model.OrderCancelHandler,
					trace.WithAttributes(
						attribute.Int64("OrderID", in.GetOrderID()),
						attribute.String("err", err.Error()),
					),
				)
				defer span.End()

				logger.Errorw(fmt.Sprintf("OrderCancel : %v", err), "span", span)
			}
		}()

		if errors.Is(err, model.ErrOrderCancelNotFound) {
			return nil, status.Error(codes.NotFound, model.ErrOrderCancelNotFound.Error())
		}
		if errors.Is(err, model.ErrOrderStatusFailedOrPaid) {
			return nil, status.Error(codes.FailedPrecondition, model.ErrOrderStatusFailedOrPaid.Error())
		}
		if errors.Is(err, model.ErrOrderAlreadyCanceled) {
			return &pb.OrderCancelResponse{}, status.Error(codes.OK, model.ErrOrderAlreadyCanceled.Error())
		}

		return nil, err
	}

	return &pb.OrderCancelResponse{}, nil
}
