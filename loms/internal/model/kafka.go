package model

var (
	// TopicOrderEvents ...
	TopicOrderEvents = "loms.order-events"
)

// OrderEvent ...
type OrderEvent struct {
	OrderID int64  `json:"order_id"`
	Status  string `json:"status"`
	Moment  string `json:"moment"` // в формате RFC 3339
}

var (
	// StatusMsgNew ...
	StatusMsgNew = "new"
	// StatusMsgProcess ...
	StatusMsgProcess = "process"
	// StatusMsgSent ...
	StatusMsgSent = "sent"
)
