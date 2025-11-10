// Package server ...
package server

import (
	"context"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/model"
	pb "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/pkg/api/v1"
)

// CommentAdd ...
func (s *Server) CommentAdd(ctx context.Context, in *pb.CommentAddRequest) (*pb.CommentAddResponse, error) {
	comment := convertToNewComment(in)
	commentID, err := s.impl.CommentAdd(ctx, comment)
	if err != nil {
		return nil, err
	}

	return &pb.CommentAddResponse{
		Id: commentID,
	}, nil
}

// convertToNewComment ...
func convertToNewComment(in *pb.CommentAddRequest) model.Comment {
	return model.Comment{
		UserID:  in.GetUserId(),
		Sku:     in.GetSku(),
		Comment: in.GetComment(),
	}
}
