// Package model ...
package model

import "time"

var (
	// CommentAddHandler ...
	CommentAddHandler = "CommentAdd"
	// CommentGetByIDHandler ...
	CommentGetByIDHandler = "CommentGetByID"
	// CommentEditHandler ...
	CommentEditHandler = "CommentEdit"
	// CommentListBySKUHandler ...
	CommentListBySKUHandler = "CommentListBySKU"
	// CommentListByUserHandler ...
	CommentListByUserHandler = "CommentListByUser"
)

// Comment ...
type Comment struct {
	ID        int64     `json:"id,string"`
	UserID    int64     `json:"userId,string"`
	Sku       int64     `json:"sku,string"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"createdAt"`
}
