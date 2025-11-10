// Package model ...
package model

// RequestData ...
type RequestData struct {
	UserID int64  `json:"user_id" validate:"min=1"`
	Sku    int64  `json:"sku" validate:"min=1"`
	Count  uint32 `json:"count" validate:"min=1"`
}

// GetItemsFromCartResponce ...
type GetItemsFromCartResponce struct {
	Items      []Item `json:"items"`
	TotalPrice uint32 `json:"total_price"`
}

// Item ...
type Item struct {
	Sku   int64  `json:"sku"`
	Name  string `json:"name"`
	Count uint32 `json:"count"`
	Price uint32 `json:"price"`
}

// OrderID ...
type OrderID struct {
	OrderID int64 `json:"order_id"`
}

// ValidateTypeFull ...
type ValidateTypeFull int //user + sku + count
// ValidateTypeBySku ...
type ValidateTypeBySku int //user + sku
// ValidateTypeByUserID ...
type ValidateTypeByUserID int //user

const (
	// ValidateFull ...
	ValidateFull ValidateTypeFull = 1
	// ValidateBySku ...
	ValidateBySku ValidateTypeBySku = 2
	// ValidateByUserID ...
	ValidateByUserID ValidateTypeByUserID = 3
)

var (
	// AddItemURL ...
	AddItemURL = "POST /user/{user_id}/cart/{sku_id}"
	// DeleteItemURL ...
	DeleteItemURL = "DELETE /user/{user_id}/cart/{sku_id}"
	// DeleteItemsByUserIDURL ...
	DeleteItemsByUserIDURL = "DELETE /user/{user_id}/cart"
	// GetItemsByUserIDURL ...
	GetItemsByUserIDURL = "GET /user/{user_id}/cart"
	// OrderFullCartURL ...
	OrderFullCartURL = "POST /checkout/{user_id}"
	// GetMetricsURL ...
	GetMetricsURL = "GET /metrics"
)

// ручки Product-service ...
var (
	// GetProductBySkuURL ...
	GetProductBySkuURL = "GET /product/{sku_id}"
)

// ручки Loms GRPC ...
var (
	// OrderCreateGRPC ...
	OrderCreateGRPC = "OrderCreate"
	// StocksInfoGRPC ...
	StocksInfoGRPC = "StocksInfo"
)
var (
	// DebugPprof ...
	DebugPprof = "/debug/pprof/"
)
