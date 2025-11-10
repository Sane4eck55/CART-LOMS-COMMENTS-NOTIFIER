package server

import (
	"context"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/model"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/service"
)

//go:generate minimock -i Service -o ./mocks/service_mock.go -n ProductMock -p ServiceMock

// Service ...
type Service interface {
	AddItem(ctx context.Context, datas model.RequestData) error
	DeleteItem(ctx context.Context, data model.RequestData) error
	DeleteItemsByUserID(ctx context.Context, data model.RequestData) error
	GetItemsFromCart(ctx context.Context, data model.RequestData) (*model.GetItemsFromCartResponce, error)
	OrderCreate(ctx context.Context, UserID int64, items *model.GetItemsFromCartResponce) (int64, error)
}

// Server ...
type Server struct {
	cartService Service
	tracer      service.Tracer
}

// NewServer ...
func NewServer(service Service, traser service.Tracer) *Server {
	return &Server{
		cartService: service,
		tracer:      traser,
	}
}
