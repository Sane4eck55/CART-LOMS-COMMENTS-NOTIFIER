package server

import (
	"context"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/model"
	pb "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/pkg/api/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CommentListBySku ...
func (s *Server) CommentListBySku(ctx context.Context, in *pb.CommentListBySKURequest) (*pb.CommentListBySKUResponse, error) {
	comments, cursor, err := s.impl.CommentListBySku(ctx, in.GetSku(), convertInCursor(in.GetCursor()))
	if err != nil {
		return nil, err
	}

	return &pb.CommentListBySKUResponse{
		Comments: converToCommentBySku(comments),
		Cursor:   converOutCursor(cursor),
	}, nil
}

// convertInputCursor ...
func convertInCursor(in *pb.Cursor) *model.Cursor {
	if in.GetId() == 0 {
		return nil
	}

	return &model.Cursor{
		ID: in.GetId(),
	}
}

// converOutCursor ...
func converOutCursor(newCursor *model.Cursor) *pb.Cursor {
	if newCursor == nil {
		return &pb.Cursor{}
	}

	return &pb.Cursor{
		Id: newCursor.ID,
	}
}

// converToCommentBySku ...
func converToCommentBySku(comments []model.Comment) []*pb.CommentBySku {
	res := make([]*pb.CommentBySku, 0, len(comments))

	for _, comment := range comments {
		res = append(res, &pb.CommentBySku{
			Id:        comment.ID,
			UserId:    comment.UserID,
			Comment:   comment.Comment,
			CreatedAt: timestamppb.New(comment.CreatedAt),
		})
	}

	return res
}
