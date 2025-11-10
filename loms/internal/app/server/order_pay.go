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

// OrderPay ...
func (s *Server) OrderPay(ctx context.Context, in *pb.OrderPayRequest) (*pb.OrderPayResponse, error) {
	ctx, span := s.tracer.Start(
		ctx,
		model.OrderPayHandler,
		trace.WithAttributes(
			attribute.Int64("OrderID", in.GetOrderID()),
		),
	)
	defer span.End()

	if err := s.impl.OrderPay(ctx, in.GetOrderID()); err != nil {
		defer func() {
			if err != nil {
				_, span := s.tracer.Start(
					ctx,
					model.OrderPayHandler,
					trace.WithAttributes(
						attribute.Int64("OrderID", in.GetOrderID()),
						attribute.String("err", err.Error()),
					),
				)
				defer span.End()
				logger.Errorw(fmt.Sprintf("OrderPay : %v", err), "span", span)
			}
		}()
		if errors.Is(err, model.ErrOrderPayNotFound) {
			return nil, status.Error(codes.NotFound, model.ErrOrderPayNotFound.Error())
		}
		if errors.Is(err, model.ErrOrderAlreadyPay) {
			return &pb.OrderPayResponse{}, status.Error(codes.OK, model.ErrOrderAlreadyPay.Error())
		}
		if errors.Is(err, model.ErrOrderStatusNotAwaitingPayment) {
			return nil, status.Error(codes.FailedPrecondition, model.ErrOrderStatusNotAwaitingPayment.Error())
		}
		return nil, err
	}

	return &pb.OrderPayResponse{}, nil
}
