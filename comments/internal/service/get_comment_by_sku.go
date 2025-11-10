package service

import (
	"context"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/model"
)

// CommentListBySku ...
func (s *Service) CommentListBySku(ctx context.Context, sku int64, cursor *model.Cursor) ([]model.Comment, *model.Cursor, error) {
	comments, cursor, err := s.Repository.CommentListBySku(ctx, sku, cursor)
	if err != nil {
		return nil, nil, err
	}
	return comments, cursor, nil
}
