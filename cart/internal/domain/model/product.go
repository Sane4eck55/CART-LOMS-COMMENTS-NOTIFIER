// Package model ...
package model

// GetProductResponse ...
type GetProductResponse struct {
	Name  string `json:"name"`
	Price int64  `json:"price"`
	Sku   int64  `json:"sku"`
}
