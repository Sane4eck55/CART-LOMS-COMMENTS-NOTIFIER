// Package postgres ...
package postgres

import (
	"context"
	"fmt"
	"sync"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

const (
	tableOrders      = "orders"
	tableOrdersItems = "orders_items"
	tableStocks      = "stocks"
)

// RepoNoBuilder ...
type RepoNoBuilder struct {
	Master  *pgxpool.Pool
	Replica *pgxpool.Pool
	mx      sync.RWMutex
}

// NewRepoNoBuilder ...
func NewRepoNoBuilder(master *pgxpool.Pool, replica *pgxpool.Pool) *RepoNoBuilder {
	return &RepoNoBuilder{
		Master:  master,
		Replica: replica,
	}
}

// CreateOrder ...
func (r *RepoNoBuilder) CreateOrder(ctx context.Context, usersOrders model.Order) (int64, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	tx, err := r.Master.Begin(ctx)
	if err != nil {
		return model.ErrorOrderID, errors.Wrap(err, "CreateOrder Begin")
	}

	defer func() {
		if p := recover(); p != nil {
			//nolint:errcheck, gosec
			tx.Rollback(ctx)
			panic(p)
		}
	}()

	var orderID int64
	queryOrders := fmt.Sprintf("INSERT INTO %s (user_id) VALUES ($1) returning id;", tableOrders)

	if err = tx.QueryRow(ctx, queryOrders, usersOrders.UserID).
		Scan(&orderID); err != nil {
		//nolint:errcheck, gosec
		tx.Rollback(ctx)
		return model.ErrorOrderID, errors.Wrap(err, "CreateOrder Scan")
	}

	queryItems := fmt.Sprintf("INSERT INTO %s (order_id, sku, count) VALUES ($1, $2, $3);", tableOrdersItems)

	for _, items := range usersOrders.Items {
		_, err = tx.Exec(ctx, queryItems, orderID, items.Sku, items.Count)
		if err != nil {
			//nolint:errcheck, gosec
			tx.Rollback(ctx)
			return model.ErrorOrderID, errors.Wrap(err, "CreateOrder Exec")
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return orderID, err
	}

	return orderID, nil
}

// SetStatusOrder ...
func (r *RepoNoBuilder) SetStatusOrder(ctx context.Context, orderID int64, status string) error {
	r.mx.Lock()
	defer r.mx.Unlock()
	query := fmt.Sprintf("UPDATE %s SET status = $1 WHERE id = $2;", tableOrders)

	_, err := r.Master.Exec(ctx, query, status, orderID)
	if err != nil {
		return errors.Wrap(err, "SetStatusOrder Scan")
	}

	return nil
}

// GetInfoByOrderID ...
func (r *RepoNoBuilder) GetInfoByOrderID(ctx context.Context, orderID int64) (*model.OrderInfo, error) {
	r.mx.RLock()
	defer r.mx.RUnlock()
	orderInfo := model.OrderInfo{}

	queryOrders := fmt.Sprintf("SELECT user_id, status FROM %s WHERE id = $1;", tableOrders)

	if err := r.Replica.QueryRow(ctx, queryOrders, orderID).
		Scan(&orderInfo.UserID, &orderInfo.Status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrOrderPayNotFound
		}
		return nil, errors.Wrap(err, "GetInfoByOrderID Scan")
	}

	queryItems := fmt.Sprintf("SELECT sku, count FROM %s WHERE order_id = $1;", tableOrdersItems)
	rows, err := r.Replica.Query(ctx, queryItems, orderID)
	if err != nil {
		return nil, errors.Wrap(err, "GetInfoByOrderID Query")
	}
	defer rows.Close()

	for rows.Next() {
		var item model.Item
		if err := rows.Scan(&item.Sku, &item.Count); err != nil {
			return nil, errors.Wrap(err, "GetInfoByOrderID Next")
		}
		orderInfo.Items = append(orderInfo.Items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, " GetInfoByOrderID Err")
	}

	return &orderInfo, nil
}

// Reserve ...
func (r *RepoNoBuilder) Reserve(ctx context.Context, items []model.Item) error {
	r.mx.Lock()
	defer r.mx.Unlock()
	tx, err := r.Master.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "Reserve Begin")
	}

	defer func() {
		if p := recover(); p != nil {
			//nolint:errcheck, gosec
			tx.Rollback(ctx)
			panic(p)
		}
	}()

	var (
		totalCount uint32
		reserved   uint32
	)

	mapReserved := make(map[int64]uint32)
	query := fmt.Sprintf("SELECT total_count, reserved FROM %s WHERE sku = $1;", tableStocks)

	for _, item := range items {
		if err = tx.QueryRow(ctx, query, item.Sku).
			Scan(&totalCount, &reserved); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				//nolint:errcheck, gosec
				tx.Rollback(ctx)
				return model.ErrStockInfoNotFound
			}
			//nolint:errcheck, gosec
			tx.Rollback(ctx)
			return errors.Wrap(err, "Reserve Scan")
		}

		if !haveFreeStock(totalCount, reserved, item.Count) {
			//nolint:errcheck, gosec
			tx.Rollback(ctx)
			return model.ErrNoStockForReserve
		}

		mapReserved[item.Sku] = reserved
	}

	queryExec := fmt.Sprintf("UPDATE %s SET reserved = $1 WHERE sku = $2;", tableStocks)

	for _, item := range items {
		newReserved := mapReserved[item.Sku] + item.Count

		_, err = tx.Exec(ctx, queryExec, newReserved, item.Sku)
		if err != nil {
			//nolint:errcheck, gosec
			tx.Rollback(ctx)
			return errors.Wrap(err, "Reserve Exec")
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

// GetStocksBySku ...
func (r *RepoNoBuilder) GetStocksBySku(ctx context.Context, sku int64) (uint32, error) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	var (
		totalCount uint32
		reserved   uint32
	)

	query := fmt.Sprintf("SELECT total_count, reserved FROM %s WHERE sku = $1;", tableStocks)

	if err := r.Replica.QueryRow(ctx, query, sku).
		Scan(&totalCount, &reserved); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrorStockCount, model.ErrStockSkuNotFound
		}
		return model.ErrorStockCount, errors.Wrap(err, "GetStocksBySku Scan")
	}

	freeStock := totalCount - reserved

	return freeStock, nil
}

// ReserveRemove ...
func (r *RepoNoBuilder) ReserveRemove(ctx context.Context, item model.Item) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	var (
		totalCount uint32
		reserved   uint32
	)

	query := fmt.Sprintf("SELECT total_count, reserved FROM %s WHERE sku = $1;", tableStocks)

	if err := r.Master.QueryRow(ctx, query, item.Sku).
		Scan(&totalCount, &reserved); err != nil {
		return errors.Wrap(err, "ReserveRemove Scan")
	}

	query = fmt.Sprintf("UPDATE %s SET total_count = $1, reserved = $2 WHERE sku = $3;", tableStocks)

	newTotal := totalCount - item.Count
	newReserved := reserved - item.Count

	_, err := r.Master.Exec(ctx, query, newTotal, newReserved, item.Sku)
	if err != nil {
		return errors.Wrap(err, "ReserveRemove Exec")
	}

	return nil
}

// ReserveCancel ...
func (r *RepoNoBuilder) ReserveCancel(ctx context.Context, item model.Item) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	var (
		reserved uint32
	)

	query := fmt.Sprintf("SELECT reserved FROM %s WHERE sku = $1;", tableStocks)

	if err := r.Master.QueryRow(ctx, query, item.Sku).
		Scan(&reserved); err != nil {
		return errors.Wrap(err, "ReserveCancel Scan")
	}

	query = fmt.Sprintf("UPDATE %s SET reserved = $1 WHERE sku = $2;", tableStocks)

	newReserved := reserved - item.Count

	_, err := r.Master.Exec(ctx, query, newReserved, item.Sku)
	if err != nil {
		return errors.Wrap(err, "ReserveCancel Exec")
	}

	return nil
}

// isFreeStock ...
func haveFreeStock(totalCount uint32, reserved uint32, count uint32) bool {
	return totalCount-reserved > count
}
