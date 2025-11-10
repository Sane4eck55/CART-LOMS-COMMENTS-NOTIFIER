// Package stocks ...
package stocks

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/model"
)

//go:embed stock-data.json
var stockData []byte

// Repo ...
type Repo struct {
	storage model.StocksStorage
	mx      sync.RWMutex
}

// NewStocksRepo ...
func NewStocksRepo() (*Repo, error) {
	storage, err := makeStocksStorage()
	if err != nil {
		return nil, err
	}
	return &Repo{
		storage: storage,
	}, nil
}

// Reserve ...
func (sr *Repo) Reserve(_ context.Context, items []model.Item) error {
	sr.mx.Lock()
	defer sr.mx.Unlock()
	for _, item := range items {
		if stocksInfo, ok := sr.storage[item.Sku]; ok {
			if stocksInfo.TotalCount-stocksInfo.Reserved < item.Count {
				return model.ErrNoStockForReserve
			}
		}
	}

	for _, item := range items {
		if stocksInfo, ok := sr.storage[item.Sku]; ok {
			stocksInfo.Reserved += item.Count
			sr.storage[item.Sku] = stocksInfo
			continue
		}

		return model.ErrStockInfoNotFound
	}

	return nil
}

// GetStocksBySku ...
func (sr *Repo) GetStocksBySku(_ context.Context, sku int64) (uint32, error) {
	sr.mx.RLock()
	defer sr.mx.RUnlock()

	if info, ok := sr.storage[sku]; ok {
		freeStock := info.TotalCount - info.Reserved
		return freeStock, nil
	}

	return model.ErrorStockCount, model.ErrStockSkuNotFound
}

// ReserveRemove ...
func (sr *Repo) ReserveRemove(_ context.Context, item model.Item) error {
	sr.mx.Lock()
	defer sr.mx.Unlock()

	if info, ok := sr.storage[item.Sku]; ok {
		info.Reserved -= item.Count
		info.TotalCount -= item.Count
		sr.storage[item.Sku] = info
		return nil
	}

	return model.ErrStockSkuNotFound
}

// ReserveCancel ...
func (sr *Repo) ReserveCancel(_ context.Context, item model.Item) error {
	sr.mx.Lock()
	defer sr.mx.Unlock()

	if info, ok := sr.storage[item.Sku]; ok {
		info.Reserved -= item.Count
		sr.storage[item.Sku] = info
		return nil
	}

	return model.ErrStockSkuNotFound
}

// makeStocksStorage ...
func makeStocksStorage() (model.StocksStorage, error) {
	var stocks []model.Stock

	if err := json.Unmarshal(stockData, &stocks); err != nil {
		return nil, fmt.Errorf("unmarshal stockData: %v", err)
	}

	storage := make(model.StocksStorage)
	for _, stock := range stocks {
		storage[stock.Sku] = stock
	}

	return storage, nil
}
