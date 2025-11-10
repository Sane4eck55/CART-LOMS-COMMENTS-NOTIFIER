package server

import (
	"context"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/model"
	pb "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/pkg/api/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CommentGetByID ...
func (s *Server) CommentGetByID(ctx context.Context, in *pb.CommentGetByIDRequest) (*pb.CommentGetByIDResponse, error) {
	comment, err := s.impl.CommentGetByID(ctx, in.GetId())
	if err != nil {
		return nil, err
	}

	return convertToPbComment(comment), nil
}

func convertToPbComment(comment *model.Comment) *pb.CommentGetByIDResponse {
	return &pb.CommentGetByIDResponse{
		Id:        comment.ID,
		UserId:    comment.UserID,
		Sku:       comment.Sku,
		Comment:   comment.Comment,
		CreatedAt: timestamppb.New(comment.CreatedAt),
	}
}
