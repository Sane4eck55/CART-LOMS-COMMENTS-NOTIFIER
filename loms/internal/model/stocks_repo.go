// Package model ...
package model

// StocksStorage ...
type StocksStorage = map[int64]Stock

// Stock ...
type Stock struct {
	Sku        int64  `json:"sku"`
	TotalCount uint32 `json:"total_count"`
	Reserved   uint32 `json:"reserved"`
}

var (
	// ErrorStockCount ...
	ErrorStockCount uint32
)
