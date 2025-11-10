package service

import (
	"context"
	"sort"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/model"
)

// CommentListByUser ...
func (s *Service) CommentListByUser(ctx context.Context, userID int64) ([]model.Comment, error) {
	comments, err := s.Repository.CommentListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return sortCommentsByCreatedAt(comments), nil
}

// sortCommentsByCreatedAt ...
func sortCommentsByCreatedAt(comments []model.Comment) []model.Comment {
	sort.Slice(comments, func(i, j int) bool {
		c1, c2 := comments[i], comments[j]

		if c1.CreatedAt.After(c2.CreatedAt) {
			return true
		}
		if c1.CreatedAt.Before(c2.CreatedAt) {
			return false
		}
		return c1.ID < c2.ID
	})

	return comments
}
