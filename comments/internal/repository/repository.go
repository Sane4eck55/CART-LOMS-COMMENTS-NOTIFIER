// Package repository ...
package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// limit ...
const limit = 10

// shardManager ...
type shardManager interface {
	AddShard(shard *pgxpool.Pool, shardID int)
	GetShard(key string) (*pgxpool.Pool, int, error)
	GetShardByCommentID(id int64) *pgxpool.Pool
	GetShardPool() []*pgxpool.Pool
	Close()
}

// Repo ...
type Repo struct {
	sm shardManager
}

// NewRepo ...
func NewRepo(sm shardManager) *Repo {
	return &Repo{
		sm: sm,
	}
}

// AddComment ...
func (r *Repo) AddComment(ctx context.Context, comment model.Comment) (int64, error) {
	db, shardID, err := r.sm.GetShard(strconv.FormatInt(comment.Sku, 10))
	if err != nil {
		return model.ErrCommentID, err
	}
	var commentID int64

	const query = `INSERT INTO comments (id, user_id, sku, comment) 
				   VALUES (nextval('comment_id_manual_seq') + $1, $2, $3, $4) 
				   returning id;`

	if err := db.QueryRow(ctx, query, shardID, comment.UserID, comment.Sku, comment.Comment).Scan(&commentID); err != nil {
		return model.ErrCommentID, err
	}

	return commentID, nil
}

// GetCommentByID ...
func (r *Repo) GetCommentByID(ctx context.Context, commentID int64) (*model.Comment, error) {
	db := r.sm.GetShardByCommentID(commentID)

	const query = `SELECT id, user_id, sku, comment, created_at FROM comments WHERE id = $1;`

	var comment model.Comment

	err := db.QueryRow(ctx, query, commentID).
		Scan(
			&comment.ID,
			&comment.UserID,
			&comment.Sku,
			&comment.Comment,
			&comment.CreatedAt,
		)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrCommentNotFoundByID
		}
		return nil, err
	}
	return &comment, nil
}

// EditComment ...
func (r *Repo) EditComment(ctx context.Context, comment model.Comment) error {
	db := r.sm.GetShardByCommentID(comment.ID)

	const query = `UPDATE comments SET comment = $1, updated_at = $2 WHERE id = $3;`

	_, err := db.Exec(ctx, query, comment.Comment, time.Now(), comment.ID)
	if err != nil {
		return err
	}

	return nil
}

// CommentListBySku ...
func (r *Repo) CommentListBySku(ctx context.Context, sku int64, cursor *model.Cursor) ([]model.Comment, *model.Cursor, error) {
	db, _, err := r.sm.GetShard(strconv.FormatInt(sku, 10))
	if err != nil {
		return nil, nil, err
	}

	query := `SELECT id, user_id, comment, created_at FROM comments WHERE sku = $1 ORDER BY created_at DESC, id DESC LIMIT $2;`
	args := []interface{}{sku, limit}

	if cursor != nil {
		query = `SELECT id, user_id, comment, created_at FROM comments WHERE sku = $1 and id < $2 ORDER BY created_at DESC, id DESC LIMIT $3;`
		args = []interface{}{sku, cursor.ID, limit}
	}

	var (
		comment  model.Comment
		comments []model.Comment
	)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&comment.ID, &comment.UserID, &comment.Comment, &comment.CreatedAt)
		if err != nil {
			return nil, nil, err
		}
		comments = append(comments, comment)
	}

	if err != nil {
		return nil, nil, err
	}

	var nextCursor *model.Cursor

	if len(comments) > 0 {
		lastComment := comments[len(comments)-1]
		nextCursor = &model.Cursor{
			ID: lastComment.ID,
		}

		return comments, nextCursor, nil
	}

	return comments, nil, nil
}

// CommentListByUser ...
func (r *Repo) CommentListByUser(ctx context.Context, userID int64) ([]model.Comment, error) {
	const query = `SELECT id, sku, comment, created_at FROM comments WHERE user_id = $1 ORDER BY created_at DESC, id DESC  LIMIT $2;`

	var comments []model.Comment

	for _, shard := range r.sm.GetShardPool() {
		rows, err := shard.Query(ctx, query, userID, limit)
		if err != nil {
			return nil, err
		}

		defer rows.Close()

		for rows.Next() {
			var comment model.Comment
			err = rows.Scan(&comment.ID, &comment.Sku, &comment.Comment, &comment.CreatedAt)
			if err != nil {
				return nil, err
			}
			comments = append(comments, comment)
		}

	}

	return comments, nil
}

// Close ...
func (r *Repo) Close() {
	r.sm.Close()
}
