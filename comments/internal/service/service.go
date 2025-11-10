// Package service ...
package service

import (
	"context"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/model"
)

// IRepository ...
type IRepository interface {
	AddComment(ctx context.Context, comment model.Comment) (int64, error)
	GetCommentByID(ctx context.Context, commentID int64) (*model.Comment, error)
	EditComment(ctx context.Context, comment model.Comment) error
	CommentListBySku(ctx context.Context, sku int64, cursor *model.Cursor) ([]model.Comment, *model.Cursor, error)
	CommentListByUser(ctx context.Context, userID int64) ([]model.Comment, error)
	Close()
}

// Service ...
type Service struct {
	Repository IRepository
}

// NewService ...
func NewService(repository IRepository) *Service {
	return &Service{
		Repository: repository,
	}
}
