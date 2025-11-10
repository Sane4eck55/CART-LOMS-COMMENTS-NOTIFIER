// Package model ...
package model

// Cart ...
type Cart struct {
	SkuID int64
	Count uint32
}

// Storage ...
type Storage = map[int64][]Cart
