package service

import (
	"context"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/model"
)

// CommentAdd ...
func (s *Service) CommentAdd(ctx context.Context, comment model.Comment) (int64, error) {
	commentID, err := s.Repository.AddComment(ctx, comment)
	if err != nil {
		return model.ErrCommentID, err
	}
	return commentID, nil
}
