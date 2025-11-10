package server

import (
	"context"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/model"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/service"
	pb "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/pkg/api/v1"
)

// LomsService ...
type LomsService interface {
	OrderCreate(ctx context.Context, order model.Order) (int64, error)
	OrderInfo(ctx context.Context, orderID int64) (*model.OrderInfo, error)
	OrderPay(ctx context.Context, orderID int64) error
	OrderCancel(ctx context.Context, orderID int64) error
	GetStocksBySku(ctx context.Context, sku int64) (uint32, error)
	ProduceFromOutbox(ctx context.Context)
}

// Server ...
type Server struct {
	pb.UnimplementedLomsServer
	impl   LomsService
	tracer service.Tracer
}

// NewServer ...
func NewServer(impl LomsService, tracer service.Tracer) *Server {
	return &Server{
		impl:   impl,
		tracer: tracer,
	}
}
