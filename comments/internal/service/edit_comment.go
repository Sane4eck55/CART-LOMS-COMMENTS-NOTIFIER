package service

import (
	"context"
	"time"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/model"
)

const (
	// timeForEdit ...
	timeForEdit = 1 * time.Second
)

// CommentEdit ...
func (s *Service) CommentEdit(ctx context.Context, comment model.Comment) error {
	oldComment, err := s.Repository.GetCommentByID(ctx, comment.ID)
	if err != nil {
		return err
	}

	if !isSameUser(comment.UserID, oldComment.UserID) {
		return model.ErrCommentNotBelongUser
	}

	if !isEditTimeout(oldComment.CreatedAt) {
		return model.ErrTimeForEditOver
	}

	if err := s.Repository.EditComment(ctx, comment); err != nil {
		return err
	}

	return nil
}

// isSameUser ...
func isSameUser(inputUserID, oldUserID int64) bool {
	return inputUserID == oldUserID
}

// isEditTimeout ...
func isEditTimeout(oldCreatedAt time.Time) bool {
	now := time.Now()
	duration := now.Sub(oldCreatedAt)
	return duration < timeForEdit
}
