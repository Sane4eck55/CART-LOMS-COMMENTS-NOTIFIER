// Package service ...
package service

import (
	"context"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/model"
	repository_sqlc "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/repository/sqlc/generated"
	pbKafka "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/pkg/api/kafka"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/trace"
)

// IOrderRepo ...
type IOrderRepo interface {
	CreateOrder(ctx context.Context, usersOrders model.Order) (int64, error)
	SetStatusOrder(ctx context.Context, orderID int64, status string) error
	GetInfoByOrderID(ctx context.Context, orderID int64) (*model.OrderInfo, error)
}

// IStocksRepo ...
type IStocksRepo interface {
	Reserve(ctx context.Context, items []model.Item) error
	GetStocksBySku(ctx context.Context, sku int64) (uint32, error)
	ReserveRemove(ctx context.Context, item model.Item) error
	ReserveCancel(ctx context.Context, item model.Item) error
}

// IRepository ...
type IRepository interface {
	CreateOrder(ctx context.Context, usersOrders model.Order) (int64, error)
	SetStatusOrder(ctx context.Context, orderID int64, status string) error
	GetInfoByOrderIDMaster(ctx context.Context, orderID int64) (*model.OrderInfo, error)
	GetInfoByOrderIDReplica(ctx context.Context, orderID int64) (*model.OrderInfo, error)
	Reserve(ctx context.Context, items []model.Item) error
	GetFreeStocksBySkuMaster(ctx context.Context, sku int64) (uint32, error)
	GetFreeStocksBySkuReplica(ctx context.Context, sku int64) (uint32, error)
	ReserveRemove(ctx context.Context, item model.Item) error
	ReserveCancel(ctx context.Context, item model.Item) error
	Delete(ctx context.Context, orderID int64) error
	UseMaster(typeReq string) bool
	AddOutbox(ctx context.Context, tx pgx.Tx, event *pbKafka.MsgProduce) error
	GetNewMsgOutbox(ctx context.Context) ([]*repository_sqlc.GetNewMsgOutboxRow, error)
	UpdateStatusMsgOutbox(ctx context.Context, id int64, status string) error
}

// Tracer ...
type Tracer interface {
	Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
}

// IProducerOrderEvent ...
type IProducerOrderEvent interface {
	SendMsg(msg *model.OrderEvent) (int32, int64, error)
}

// Service ...
type Service struct {
	repository IRepository
	tracer     Tracer
	producer   IProducerOrderEvent
}

// NewService ...
func NewService(repo IRepository, tracer Tracer, producer IProducerOrderEvent) Service {
	return Service{
		repository: repo,
		tracer:     tracer,
		producer:   producer,
	}
}
