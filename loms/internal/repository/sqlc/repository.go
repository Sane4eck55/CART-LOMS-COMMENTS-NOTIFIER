// Package sqlc ...
package sqlc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/infra/metrics"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/model"
	repository_sqlc "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/repository/sqlc/generated"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/service"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	grpccode "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"

	pbKafka "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/pkg/api/kafka"
)

const (
	// batchItems ...
	batchItems = 5
)

// Repo ...
type Repo struct {
	Master            *repository_sqlc.Queries
	Replica           *repository_sqlc.Queries
	MasterPool        *pgxpool.Pool
	ReplicaPool       *pgxpool.Pool
	CountRequestStock int64
	CountRequestOrder int64
	tracer            service.Tracer
}

// NewRepo ...
func NewRepo(master *pgxpool.Pool, replica *pgxpool.Pool, tracer service.Tracer) *Repo {
	return &Repo{
		Master:      repository_sqlc.New(master),
		Replica:     repository_sqlc.New(replica),
		MasterPool:  master,
		ReplicaPool: replica,
		tracer:      tracer,
	}
}

// CreateOrder ...
func (r *Repo) CreateOrder(ctx context.Context, usersOrders model.Order) (int64, error) {
	metrics.IncRequestCount("repo_CreateOrder", model.TypeDB)
	start := time.Now()

	ctx, span := r.tracer.Start(
		ctx,
		"repo CreateOrder",
		trace.WithAttributes(
			attribute.Int64("UserID", usersOrders.UserID),
		),
	)
	defer span.End()
	tx, err := r.MasterPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}

	//nolint:errcheck
	defer func() {
		if err = tx.Rollback(ctx); err != nil {
			log.Printf("tx.Rollback: %v", err)
		}
	}()

	orderID, err := r.Master.WithTx(tx).AddOrderToOrders(ctx, usersOrders.UserID)
	if err != nil {
		_, span := r.tracer.Start(
			ctx,
			"repo CreateOrder",
			trace.WithAttributes(
				attribute.String("err_AddOrderToOrders", err.Error()),
			),
		)
		defer span.End()

		metrics.RequestDuration(model.OrderCreateHandler, grpcstatus.Code(err).String(), model.TypeDB, time.Since(start))

		return model.ErrorOrderID, errors.Wrap(err, "CreateOrder AddOrderToOrders")
	}

	for _, items := range lo.Chunk(usersOrders.Items, batchItems) {
		tmpItems := convetToAddOrderToOrdersItemsParams(orderID, items)

		if err = r.Master.WithTx(tx).AddOrderToOrdersItems(ctx,
			tmpItems,
		); err != nil {
			_, span := r.tracer.Start(
				ctx,
				"repo CreateOrder",
				trace.WithAttributes(
					attribute.String("err_AddOrderToOrdersItems", err.Error()),
				),
			)
			defer span.End()

			metrics.RequestDuration(model.OrderCreateHandler, grpcstatus.Code(err).String(), model.TypeDB, time.Since(start))

			return model.ErrorOrderID, errors.Wrap(err, "CreateOrder AddOrderToOrdersItems")
		}
	}

	event := &pbKafka.MsgProduce{
		OrderId: orderID,
		Status:  model.StatusOrderNew,
		Moment:  time.Now().Format(time.RFC3339),
	}

	if err = r.AddOutbox(ctx, tx, event); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return orderID, nil
}

// SetStatusOrder ...
func (r *Repo) SetStatusOrder(ctx context.Context, orderID int64, status string) error {
	metrics.IncRequestCount("repo_SetStatusOrder", model.TypeDB)
	start := time.Now()

	ctx, span := r.tracer.Start(
		ctx,
		"repo SetStatusOrder",
	)
	defer span.End()

	tx, err := r.MasterPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	//nolint:errcheck
	defer func() {
		if err = tx.Rollback(ctx); err != nil {
			log.Printf("tx.Rollback: %v", err)
		}
	}()

	if err = r.Master.WithTx(tx).SetStatusOrder(ctx,
		&repository_sqlc.SetStatusOrderParams{
			Status: status,
			ID:     orderID,
		},
	); err != nil {
		metrics.RequestDuration("repo_SetStatusOrder", grpcstatus.Code(err).String(), model.TypeDB, time.Since(start))
		_, span := r.tracer.Start(
			ctx,
			"repo SetStatusOrder",
			trace.WithAttributes(
				attribute.Int64("orderID", orderID),
				attribute.String("status", status),
				attribute.String("err", err.Error()),
			),
		)
		defer span.End()
		return errors.Wrap(err, "SetStatusOrder")
	}

	event := &pbKafka.MsgProduce{
		OrderId: orderID,
		Status:  status,
		Moment:  time.Now().Format(time.RFC3339),
	}

	if err = r.AddOutbox(ctx, tx, event); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// GetInfoByOrderIDMaster ...
func (r *Repo) GetInfoByOrderIDMaster(ctx context.Context, orderID int64) (*model.OrderInfo, error) {
	metrics.IncRequestCount("repo_GetInfoByOrderIDMaster", model.TypeDB)
	start := time.Now()

	ctx, span := r.tracer.Start(
		ctx,
		"repo GetInfoByOrderIDMaster",
	)
	defer span.End()
	orderInfo := model.OrderInfo{}

	infoOrdersRow, err := r.Master.GetInfoOrders(ctx, orderID)
	if err != nil {
		_, span := r.tracer.Start(
			ctx,
			"repo GetInfoOrders",
			trace.WithAttributes(
				attribute.Int64("orderID", orderID),
				attribute.String("err", err.Error()),
			),
		)
		defer span.End()

		metrics.RequestDuration("repo_GetInfoByOrderIDMaster", grpcstatus.Code(err).String(), model.TypeDB, time.Since(start))

		return nil, errors.Wrap(err, "GetInfoByOrderID GetInfoOrders")
	}

	if len(infoOrdersRow) < 1 {
		_, span := r.tracer.Start(
			ctx,
			"repo GetInfoOrders",
			trace.WithAttributes(
				attribute.Int64("orderID", orderID),
				attribute.String("err", model.ErrOrderPayNotFound.Error()),
			),
		)
		defer span.End()

		metrics.RequestDuration("repo_GetInfoByOrderIDMaster", grpcstatus.Code(err).String(), model.TypeDB, time.Since(start))

		return nil, model.ErrOrderPayNotFound
	}

	orderInfo.UserID = infoOrdersRow[0].UserID
	orderInfo.Status = infoOrdersRow[0].Status

	infoOrdersItemsRow, err := r.Master.GetInfoOrdersItems(ctx, orderID)
	if err != nil {
		_, span := r.tracer.Start(
			ctx,
			"repo GetInfoOrdersItems",
			trace.WithAttributes(
				attribute.Int64("orderID", orderID),
				attribute.String("err", err.Error()),
			),
		)
		span.End()

		metrics.RequestDuration("repo_GetInfoByOrderIDMaster", grpcstatus.Code(err).String(), model.TypeDB, time.Since(start))

		return nil, errors.Wrap(err, "GetInfoByOrderID GetInfoOrdersItems")
	}

	for _, items := range infoOrdersItemsRow {
		item := model.Item{
			Sku: items.Sku,
			//nolint:gosec
			Count: uint32(*items.Count),
		}
		orderInfo.Items = append(orderInfo.Items, item)
	}

	//r.CountRequestOrder++
	atomic.AddInt64(&r.CountRequestOrder, 1)

	return &orderInfo, nil
}

// GetInfoByOrderIDReplica ...
func (r *Repo) GetInfoByOrderIDReplica(ctx context.Context, orderID int64) (*model.OrderInfo, error) {
	metrics.IncRequestCount("repo_GetInfoByOrderIDReplica", model.TypeDB)
	start := time.Now()

	ctx, span := r.tracer.Start(
		ctx,
		"repo GetInfoByOrderIDReplica",
	)
	defer span.End()
	orderInfo := model.OrderInfo{}

	infoOrdersRow, err := r.Replica.GetInfoOrders(ctx, orderID)
	if err != nil {
		_, span := r.tracer.Start(
			ctx,
			"repo GetInfoOrders",
			trace.WithAttributes(
				attribute.Int64("orderID", orderID),
				attribute.String("err", err.Error()),
			),
		)
		defer span.End()

		metrics.RequestDuration("repo_GetInfoByOrderIDReplica", grpcstatus.Code(err).String(), model.TypeDB, time.Since(start))

		return nil, errors.Wrap(err, "GetInfoByOrderID GetInfoOrders")
	}

	if len(infoOrdersRow) < 1 {
		_, span := r.tracer.Start(
			ctx,
			"repo GetInfoOrders",
			trace.WithAttributes(
				attribute.Int64("orderID", orderID),
				attribute.String("err", model.ErrOrderPayNotFound.Error()),
			),
		)
		defer span.End()

		metrics.RequestDuration("repo_GetInfoByOrderIDReplica", grpccode.NotFound.String(), model.TypeDB, time.Since(start))

		return nil, model.ErrOrderPayNotFound
	}

	orderInfo.UserID = infoOrdersRow[0].UserID
	orderInfo.Status = infoOrdersRow[0].Status

	infoOrdersItemsRow, err := r.Replica.GetInfoOrdersItems(ctx, orderID)
	if err != nil {
		_, span := r.tracer.Start(
			ctx,
			"repo GetInfoOrdersItems",
			trace.WithAttributes(
				attribute.Int64("orderID", orderID),
				attribute.String("err", err.Error()),
			),
		)
		span.End()

		metrics.RequestDuration("repo_GetInfoByOrderIDReplica", grpcstatus.Code(err).String(), model.TypeDB, time.Since(start))

		return nil, errors.Wrap(err, "GetInfoByOrderID GetInfoOrdersItems")
	}

	for _, items := range infoOrdersItemsRow {
		item := model.Item{
			Sku: items.Sku,
			//nolint:gosec
			Count: uint32(*items.Count),
		}
		orderInfo.Items = append(orderInfo.Items, item)
	}

	//r.CountRequestOrder++
	atomic.AddInt64(&r.CountRequestOrder, 1)

	return &orderInfo, nil
}

// Reserve ...
func (r *Repo) Reserve(ctx context.Context, items []model.Item) error {
	metrics.IncRequestCount("repo_Reserve", model.TypeDB)
	start := time.Now()

	ctx, span := r.tracer.Start(
		ctx,
		"repo Reserve",
	)
	defer span.End()
	tx, err := r.MasterPool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return err
	}

	//nolint:errcheck
	defer func() {
		if err = tx.Rollback(ctx); err != nil {
			log.Printf("tx.Rollback: %v", err)
		}
	}()
	mapReserved := make(map[int64]uint32)

	for _, item := range items {
		infoStocks, err := r.Master.WithTx(tx).GetStocksBySkuForUpdate(ctx, item.Sku)
		if err != nil {
			_, span := r.tracer.Start(
				ctx,
				"repo GetStocksBySkuForUpdate",
				trace.WithAttributes(
					attribute.Int64("sku", item.Sku),
					attribute.String("err", err.Error()),
				),
			)
			defer span.End()

			metrics.RequestDuration("repo_Reserve", grpcstatus.Code(err).String(), model.TypeDB, time.Since(start))

			return errors.Wrap(err, "Reserve GetStocksBySku")
		}

		if len(infoStocks) < 1 {
			_, span := r.tracer.Start(
				ctx,
				"repo GetStocksBySkuForUpdate",
				trace.WithAttributes(
					attribute.Int64("sku", item.Sku),
					attribute.String("err", model.ErrStockInfoNotFound.Error()),
				),
			)
			defer span.End()

			metrics.RequestDuration("repo_Reserve", grpccode.NotFound.String(), model.TypeDB, time.Since(start))

			return model.ErrStockInfoNotFound
		}
		//nolint:gosec
		if !haveFreeStock(uint32(*infoStocks[0].TotalCount), uint32(*infoStocks[0].Reserved), item.Count) {
			_, span := r.tracer.Start(
				ctx,
				"repo haveFreeStock",
				trace.WithAttributes(
					attribute.Int64("sku", item.Sku),
					attribute.String("err", model.ErrNoStockForReserve.Error()),
				),
			)
			defer span.End()

			metrics.RequestDuration("repo_Reserve", grpccode.NotFound.String(), model.TypeDB, time.Since(start))

			//nolint:errcheck
			return model.ErrNoStockForReserve
		}
		//nolint:gosec
		mapReserved[item.Sku] = uint32(*infoStocks[0].Reserved)
	}

	for _, item := range items {
		newReserved := int64(mapReserved[item.Sku] + item.Count)

		if err := r.Master.WithTx(tx).ReserveStockBySku(ctx,
			&repository_sqlc.ReserveStockBySkuParams{
				Reserved: &newReserved,
				Sku:      item.Sku,
			},
		); err != nil {
			_, span := r.tracer.Start(
				ctx,
				"repo ReserveStockBySku",
				trace.WithAttributes(
					attribute.Int64("sku", item.Sku),
					attribute.Int64("newReserved", newReserved),
					attribute.String("err", err.Error()),
				),
			)
			defer span.End()

			metrics.RequestDuration("repo_Reserve", grpcstatus.Code(err).String(), model.TypeDB, time.Since(start))

			return errors.Wrap(err, "Reserve ReserveStockBySku")
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// GetFreeStocksBySkuMaster ...
func (r *Repo) GetFreeStocksBySkuMaster(ctx context.Context, sku int64) (uint32, error) {
	metrics.IncRequestCount("repo_GetFreeStocksBySkuMaster", model.TypeDB)
	start := time.Now()

	ctx, span := r.tracer.Start(
		ctx,
		"repo GetFreeStocksBySkuMaster",
	)
	defer span.End()
	tx, err := r.MasterPool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return model.ErrorStockCount, err
	}

	//nolint:errcheck
	defer func() {
		if err = tx.Rollback(ctx); err != nil {
			log.Printf("tx.Rollback: %v", err)
		}
	}()
	infoStocks, err := r.Master.WithTx(tx).GetStocksBySku(ctx, sku)
	if err != nil {
		_, span := r.tracer.Start(
			ctx,
			"repo GetStocksBySku",
			trace.WithAttributes(
				attribute.Int64("sku", sku),
				attribute.String("err", err.Error()),
			),
		)
		defer span.End()

		metrics.RequestDuration("repo_GetFreeStocksBySkuMaster", grpcstatus.Code(err).String(), model.TypeDB, time.Since(start))

		return model.ErrorStockCount, errors.Wrap(err, "GetStocksBySku")
	}

	if len(infoStocks) < 1 {
		_, span := r.tracer.Start(
			ctx,
			"repo GetStocksBySku",
			trace.WithAttributes(
				attribute.Int64("sku", sku),
				attribute.String("err", model.ErrStockSkuNotFound.Error()),
			),
		)
		defer span.End()

		metrics.RequestDuration("repo_GetFreeStocksBySkuMaster", grpccode.NotFound.String(), model.TypeDB, time.Since(start))

		return model.ErrorStockCount, model.ErrStockSkuNotFound
	}

	//nolint:gosec
	freeStock := uint32(*infoStocks[0].TotalCount) - uint32(*infoStocks[0].Reserved)

	if err := tx.Commit(ctx); err != nil {
		return model.ErrorStockCount, err
	}

	//r.CountRequestStock++
	atomic.AddInt64(&r.CountRequestStock, 1)

	return freeStock, nil
}

// GetFreeStocksBySkuReplica ...
func (r *Repo) GetFreeStocksBySkuReplica(ctx context.Context, sku int64) (uint32, error) {
	metrics.IncRequestCount("repo_GetFreeStocksBySkuReplica", model.TypeDB)
	start := time.Now()

	ctx, span := r.tracer.Start(
		ctx,
		"repo GetFreeStocksBySkuReplica",
	)
	defer span.End()
	tx, err := r.ReplicaPool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return model.ErrorStockCount, err
	}

	//nolint:errcheck
	defer func() {
		if err = tx.Rollback(ctx); err != nil {
			log.Printf("tx.Rollback: %v", err)
		}
	}()
	infoStocks, err := r.Replica.WithTx(tx).GetStocksBySku(ctx, sku)
	if err != nil {
		_, span := r.tracer.Start(
			ctx,
			"repo GetStocksBySku",
			trace.WithAttributes(
				attribute.Int64("sku", sku),
				attribute.String("err", err.Error()),
			),
		)
		defer span.End()

		metrics.RequestDuration("repo_GetFreeStocksBySkuReplica", grpcstatus.Code(err).String(), model.TypeDB, time.Since(start))

		return model.ErrorStockCount, errors.Wrap(err, "GetStocksBySku")
	}

	if len(infoStocks) < 1 {
		_, span := r.tracer.Start(
			ctx,
			"repo GetStocksBySku",
			trace.WithAttributes(
				attribute.Int64("sku", sku),
				attribute.String("err", model.ErrStockSkuNotFound.Error()),
			),
		)
		defer span.End()

		metrics.RequestDuration("repo_GetFreeStocksBySkuReplica", grpccode.NotFound.String(), model.TypeDB, time.Since(start))

		return model.ErrorStockCount, model.ErrStockSkuNotFound
	}

	//nolint:gosec
	freeStock := uint32(*infoStocks[0].TotalCount) - uint32(*infoStocks[0].Reserved)

	if err := tx.Commit(ctx); err != nil {
		return model.ErrorStockCount, err
	}

	//r.CountRequestStock++
	atomic.AddInt64(&r.CountRequestStock, 1)

	return freeStock, nil
}

// ReserveRemove ...
func (r *Repo) ReserveRemove(ctx context.Context, item model.Item) error {
	metrics.IncRequestCount("repo_ReserveRemove", model.TypeDB)
	start := time.Now()

	ctx, span := r.tracer.Start(
		ctx,
		"repo ReserveRemove",
	)
	defer span.End()

	tx, err := r.MasterPool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return err
	}

	//nolint:errcheck
	defer func() {
		if err = tx.Rollback(ctx); err != nil {
			log.Printf("tx.Rollback: %v", err)
		}
	}()
	infoStocks, err := r.Master.GetStocksBySkuForUpdate(ctx, item.Sku)
	if err != nil {
		_, span := r.tracer.Start(
			ctx,
			"repo GetStocksBySkuForUpdate",
			trace.WithAttributes(
				attribute.Int64("sku", item.Sku),
				attribute.String("err", err.Error()),
			),
		)
		defer span.End()

		metrics.RequestDuration("repo_ReserveRemove", grpcstatus.Code(err).String(), model.TypeDB, time.Since(start))

		return errors.Wrap(err, "ReserveRemove GetStocksBySku")
	}
	//nolint:gosec
	newTotal := int64(uint32(*infoStocks[0].TotalCount) - item.Count)
	//nolint:gosec
	newReserved := int64(uint32(*infoStocks[0].Reserved) - item.Count)

	if err := r.Master.WithTx(tx).ReserveRemove(ctx,
		&repository_sqlc.ReserveRemoveParams{
			TotalCount: &newTotal,
			Reserved:   &newReserved,
			Sku:        item.Sku,
		},
	); err != nil {
		_, span := r.tracer.Start(
			ctx,
			"repo ReserveRemove",
			trace.WithAttributes(
				attribute.Int64("sku", item.Sku),
				attribute.Int64("newReserved", newReserved),
				attribute.Int64("newTotal", newTotal),
				attribute.String("err", err.Error()),
			),
		)
		defer span.End()

		metrics.RequestDuration("repo_ReserveRemove", grpcstatus.Code(err).String(), model.TypeDB, time.Since(start))

		return errors.Wrap(err, "ReserveRemove")
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// ReserveCancel ...
func (r *Repo) ReserveCancel(ctx context.Context, item model.Item) error {
	metrics.IncRequestCount("repo_ReserveCancel", model.TypeDB)
	start := time.Now()

	ctx, span := r.tracer.Start(
		ctx,
		"repo ReserveCancel",
	)
	defer span.End()
	tx, err := r.MasterPool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return err
	}

	//nolint:errcheck
	defer func() {
		if err = tx.Rollback(ctx); err != nil {
			log.Printf("tx.Rollback: %v", err)
		}
	}()
	infoReserveStock, err := r.Master.WithTx(tx).GetReservedStocksBySku(ctx, item.Sku)
	if err != nil {
		_, span := r.tracer.Start(
			ctx,
			"repo GetReservedStocksBySku",
			trace.WithAttributes(
				attribute.Int64("sku", item.Sku),
				attribute.String("err", err.Error()),
			),
		)
		defer span.End()

		metrics.RequestDuration("repo_ReserveCancel", grpcstatus.Code(err).String(), model.TypeDB, time.Since(start))

		return errors.Wrap(err, "ReserveCancel GetReservedStocksBySku")
	}
	//nolint:gosec
	newReserved := int64(uint32(*infoReserveStock[0]) - item.Count)

	if err := r.Master.WithTx(tx).ReserveCancel(ctx,
		&repository_sqlc.ReserveCancelParams{
			Reserved: &newReserved,
			Sku:      item.Sku,
		},
	); err != nil {
		_, span := r.tracer.Start(
			ctx,
			"repo GetReservedStocksBySku",
			trace.WithAttributes(
				attribute.Int64("sku", item.Sku),
				attribute.Int64("newReserved", newReserved),
				attribute.String("err", err.Error()),
			),
		)
		defer span.End()

		metrics.RequestDuration("repo_ReserveCancel", grpcstatus.Code(err).String(), model.TypeDB, time.Since(start))

		return errors.Wrap(err, "ReserveCancel")
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// Delete ...
func (r *Repo) Delete(ctx context.Context, orderID int64) error {
	metrics.IncRequestCount("repo_Delete", model.TypeDB)
	start := time.Now()

	ctx, span := r.tracer.Start(
		ctx,
		"repo Delete",
	)
	defer span.End()

	tx, err := r.MasterPool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return err
	}

	//nolint:errcheck
	defer func() {
		if err = tx.Rollback(ctx); err != nil {
			log.Printf("tx.Rollback: %v", err)
		}
	}()
	masterTx := repository_sqlc.New(tx)

	if err = masterTx.DeleteOrders(ctx, orderID); err != nil {
		_, span := r.tracer.Start(
			ctx,
			"repo DeleteOrders",
			trace.WithAttributes(
				attribute.Int64("orderID", orderID),
				attribute.String("err", err.Error()),
			),
		)
		defer span.End()

		metrics.RequestDuration("repo_Delete", grpcstatus.Code(err).String(), model.TypeDB, time.Since(start))

		return err
	}

	if err = masterTx.DeleteOrdersItems(ctx, orderID); err != nil {
		_, span := r.tracer.Start(
			ctx,
			"repo DeleteOrdersItems",
			trace.WithAttributes(
				attribute.Int64("orderID", orderID),
				attribute.String("err", err.Error()),
			),
		)
		defer span.End()

		metrics.RequestDuration("repo_Delete", grpcstatus.Code(err).String(), model.TypeDB, time.Since(start))

		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// UpdateStocks ...
func (r *Repo) UpdateStocks(ctx context.Context, sku, total, reserved int64) error {
	metrics.IncRequestCount("repo_UpdateStocks", model.TypeDB)
	start := time.Now()

	ctx, span := r.tracer.Start(
		ctx,
		"repo UpdateStocks",
	)
	defer span.End()

	err := r.Master.ReserveRemove(ctx,
		&repository_sqlc.ReserveRemoveParams{
			TotalCount: &total,
			Reserved:   &reserved,
			Sku:        sku,
		},
	)

	metrics.RequestDuration("repo_UpdateStocks", grpcstatus.Code(err).String(), model.TypeDB, time.Since(start))

	return err
}

// GetStocksBySku ...
func (r *Repo) GetStocksBySku(ctx context.Context, sku int64) (*model.Stock, error) {
	metrics.IncRequestCount("repo_GetStocksBySku", model.TypeDB)
	start := time.Now()

	ctx, span := r.tracer.Start(
		ctx,
		"repo GetStocksBySku",
	)
	defer span.End()

	infoStocks, err := r.Replica.GetStocksBySku(ctx, sku)
	if err != nil {
		_, span := r.tracer.Start(
			ctx,
			"repo GetStocksBySku",
			trace.WithAttributes(
				attribute.Int64("sku", sku),
				attribute.String("err", err.Error()),
			),
		)
		defer span.End()

		metrics.RequestDuration("repo_GetStocksBySku", grpcstatus.Code(err).String(), model.TypeDB, time.Since(start))

		return nil, errors.Wrap(err, "GetStocksBySku")
	}

	if len(infoStocks) < 1 {
		_, span := r.tracer.Start(
			ctx,
			"repo GetStocksBySku",
			trace.WithAttributes(
				attribute.Int64("sku", sku),
				attribute.String("err", model.ErrStockSkuNotFound.Error()),
			),
		)
		defer span.End()

		metrics.RequestDuration("repo_GetStocksBySku", grpccode.NotFound.String(), model.TypeDB, time.Since(start))

		return nil, model.ErrStockSkuNotFound
	}

	return &model.Stock{
		Sku: sku,
		//nolint:gosec
		TotalCount: uint32(*infoStocks[0].TotalCount),
		//nolint:gosec
		Reserved: uint32(*infoStocks[0].Reserved),
	}, nil
}

// haveFreeStock ...
func haveFreeStock(totalCount uint32, reserved uint32, count uint32) bool {
	return totalCount-reserved >= count
}

// UseMaster ...
func (r *Repo) UseMaster(typeReq string) bool {
	switch typeReq {
	case model.RequestOrder:
		return r.CountRequestOrder%10 == 0
	case model.RequestStock:
		return r.CountRequestStock%10 == 0
	default:
		return false
	}
}

// convetToAddOrderToOrdersItemsParams ...
func convetToAddOrderToOrdersItemsParams(orderID int64, items []model.Item) *repository_sqlc.AddOrderToOrdersItemsParams {
	switch len(items) {
	case 1:
		count := int64(items[0].Count)
		return &repository_sqlc.AddOrderToOrdersItemsParams{
			OrderID: orderID,
			Sku:     items[0].Sku,
			Count:   &count,
		}
	case 2:
		count := int64(items[0].Count)
		count2 := int64(items[1].Count)
		return &repository_sqlc.AddOrderToOrdersItemsParams{
			OrderID:   orderID,
			Sku:       items[0].Sku,
			Count:     &count,
			OrderID_2: orderID,
			Sku_2:     items[1].Sku,
			Count_2:   &count2,
		}
	case 3:
		count := int64(items[0].Count)
		count2 := int64(items[1].Count)
		count3 := int64(items[2].Count)
		return &repository_sqlc.AddOrderToOrdersItemsParams{
			OrderID:   orderID,
			Sku:       items[0].Sku,
			Count:     &count,
			OrderID_2: orderID,
			Sku_2:     items[1].Sku,
			Count_2:   &count2,
			OrderID_3: orderID,
			Sku_3:     items[2].Sku,
			Count_3:   &count3,
		}
	case 4:
		count := int64(items[0].Count)
		count2 := int64(items[1].Count)
		count3 := int64(items[2].Count)
		count4 := int64(items[3].Count)
		return &repository_sqlc.AddOrderToOrdersItemsParams{
			OrderID:   orderID,
			Sku:       items[0].Sku,
			Count:     &count,
			OrderID_2: orderID,
			Sku_2:     items[1].Sku,
			Count_2:   &count2,
			OrderID_3: orderID,
			Sku_3:     items[2].Sku,
			Count_3:   &count3,
			OrderID_4: orderID,
			Sku_4:     items[3].Sku,
			Count_4:   &count4,
		}
	case 5:
		count := int64(items[0].Count)
		count2 := int64(items[1].Count)
		count3 := int64(items[2].Count)
		count4 := int64(items[3].Count)
		count5 := int64(items[4].Count)

		return &repository_sqlc.AddOrderToOrdersItemsParams{
			OrderID:   orderID,
			Sku:       items[0].Sku,
			Count:     &count,
			OrderID_2: orderID,
			Sku_2:     items[1].Sku,
			Count_2:   &count2,
			OrderID_3: orderID,
			Sku_3:     items[2].Sku,
			Count_3:   &count3,
			OrderID_4: orderID,
			Sku_4:     items[3].Sku,
			Count_4:   &count4,
			OrderID_5: orderID,
			Sku_5:     items[4].Sku,
			Count_5:   &count5,
		}
	default:
		return &repository_sqlc.AddOrderToOrdersItemsParams{}
	}
}

// AddOutbox ...
func (r *Repo) AddOutbox(ctx context.Context, tx pgx.Tx, event *pbKafka.MsgProduce) error {
	orderIDstr := fmt.Sprintf("%d", event.GetOrderId())

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if err = r.Master.WithTx(tx).AddOutbox(ctx, &repository_sqlc.AddOutboxParams{
		Topic:   model.TopicOrderEvents,
		Key:     &orderIDstr,
		Payload: payload,
	}); err != nil {
		return err
	}

	return nil
}

// GetNewMsgOutbox ...
func (r *Repo) GetNewMsgOutbox(ctx context.Context) ([]*repository_sqlc.GetNewMsgOutboxRow, error) {
	tx, err := r.MasterPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}

	//nolint:errcheck
	defer func() {
		if err = tx.Rollback(ctx); err != nil {
			log.Printf("tx.Rollback: %v", err)
		}
	}()

	newMsgs, err := r.Master.WithTx(tx).GetNewMsgOutbox(ctx)
	if err != nil {
		return nil, err
	}

	for _, msg := range newMsgs {
		arg := &repository_sqlc.UpdateStatusMsgOutboxParams{
			Status: model.StatusMsgProcess,
			ID:     msg.ID,
		}
		if err = r.Master.WithTx(tx).UpdateStatusMsgOutbox(ctx, arg); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return newMsgs, nil
}

// UpdateStatusMsgOutbox ...
func (r *Repo) UpdateStatusMsgOutbox(ctx context.Context, id int64, status string) error {
	args := &repository_sqlc.UpdateStatusMsgOutboxParams{
		Status: status,
		ID:     id,
	}
	if err := r.Master.UpdateStatusMsgOutbox(ctx, args); err != nil {
		return err
	}

	return nil
}
