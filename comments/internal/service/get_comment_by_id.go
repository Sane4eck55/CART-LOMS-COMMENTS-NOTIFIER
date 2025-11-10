package service

import (
	"context"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/model"
)

// CommentGetByID ...
func (s *Service) CommentGetByID(ctx context.Context, commentID int64) (*model.Comment, error) {
	comment, err := s.Repository.GetCommentByID(ctx, commentID)
	if err != nil {
		return nil, err
	}

	return comment, nil
}
