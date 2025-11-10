// Package model ...
package model

// Order ...
type Order struct {
	UserID int64
	Items  []Item
}

// Item ...
type Item struct {
	Sku   int64
	Count uint32
}

// OrderIDTest ...
type OrderIDTest struct {
	OrderID string
}

// OrderCreateResponseTest ...
type OrderCreateResponseTest struct {
	Status string      `protobuf:"bytes,1,opt,name=Status,proto3" json:"Status,omitempty"`
	UserID string      `protobuf:"varint,2,opt,name=UserID,proto3" json:"UserID,omitempty"`
	Items  []*ItemTest `protobuf:"bytes,3,rep,name=Items,proto3" json:"Items,omitempty"`
}

// ItemTest ...
type ItemTest struct {
	Sku   string `protobuf:"varint,1,opt,name=Sku,json=sku,proto3" json:"Sku,omitempty"`
	Count uint32 `protobuf:"varint,2,opt,name=Count,json=count,proto3" json:"Count,omitempty"`
}

var (
	// OrderCancelHandler ...
	OrderCancelHandler = "OrderCancel"
	// OrderCreateHandler ...
	OrderCreateHandler = "OrderCreate"
	// OrderInfoHandler ...
	OrderInfoHandler = "OrderInfo"
	// OrderPayHandler ...
	OrderPayHandler = "OrderPay"
	// StocksInfoHandler ...
	StocksInfoHandler = "StocksInfo"
)

var (
	// TypeInternal ...
	TypeInternal = "internal"
	// TypeExternal ...
	TypeExternal = "external"
	// TypeDB ...
	TypeDB = "db"
)
