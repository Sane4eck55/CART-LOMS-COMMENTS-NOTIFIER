package server

import (
	"context"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/model"
	pb "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/pkg/api/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CommentListByUser ...
func (s *Server) CommentListByUser(ctx context.Context, in *pb.CommentListByUserRequest) (*pb.CommentListByUserResponse, error) {
	comments, err := s.impl.CommentListByUser(ctx, in.GetUserId())
	if err != nil {
		return nil, err
	}

	return &pb.CommentListByUserResponse{
		Comments: converToCommentByUserID(comments),
	}, nil
}

// converToCommentByUserID ...
func converToCommentByUserID(comments []model.Comment) []*pb.CommentByUserID {
	res := make([]*pb.CommentByUserID, 0, len(comments))

	for _, comment := range comments {
		res = append(res, &pb.CommentByUserID{
			Id:        comment.ID,
			Sku:       comment.Sku,
			Comment:   comment.Comment,
			CreatedAt: timestamppb.New(comment.CreatedAt),
		})
	}

	return res
}
