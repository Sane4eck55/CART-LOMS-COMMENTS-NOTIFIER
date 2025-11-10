package server

import (
	"context"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/model"
	pb "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/pkg/api/v1"
)

// CommentsService ...
type CommentsService interface {
	CommentAdd(ctx context.Context, comment model.Comment) (int64, error)
	CommentGetByID(ctx context.Context, commentID int64) (*model.Comment, error)
	CommentEdit(ctx context.Context, comment model.Comment) error
	CommentListBySku(ctx context.Context, sku int64, cursor *model.Cursor) ([]model.Comment, *model.Cursor, error)
	CommentListByUser(ctx context.Context, userID int64) ([]model.Comment, error)
}

// Server ...
type Server struct {
	pb.UnimplementedCommentsServer
	impl CommentsService
}

// NewServer ...
func NewServer(impl CommentsService) *Server {
	return &Server{
		impl: impl,
	}
}
