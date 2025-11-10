package server

import (
	"context"
	"errors"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/model"
	pb "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/pkg/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CommentEdit ...
func (s *Server) CommentEdit(ctx context.Context, in *pb.CommentEditRequest) (*pb.CommentEditResponse, error) {
	comment := convertToEditComment(in)
	if err := s.impl.CommentEdit(ctx, comment); err != nil {
		if errors.Is(err, model.ErrCommentNotBelongUser) {
			return nil, status.Error(codes.PermissionDenied, model.ErrCommentNotBelongUser.Error())
		}

		if errors.Is(err, model.ErrTimeForEditOver) {
			return nil, status.Error(codes.InvalidArgument, model.ErrTimeForEditOver.Error())
		}

		if errors.Is(err, model.ErrCommentNotFoundByID) {
			return nil, status.Error(codes.NotFound, model.ErrCommentNotFoundByID.Error())
		}

		return nil, err
	}

	return &pb.CommentEditResponse{}, nil
}

func convertToEditComment(in *pb.CommentEditRequest) model.Comment {
	return model.Comment{
		ID:      in.GetCommentId(),
		UserID:  in.GetUserId(),
		Comment: in.GetNewComment(),
	}

}
