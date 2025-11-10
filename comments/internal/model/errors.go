package model

import "errors"

var (
	// ErrCommentNotBelongUser ...
	ErrCommentNotBelongUser = errors.New("комментарий не принадлежит этому пользователю")
	// ErrTimeForEditOver ...
	ErrTimeForEditOver = errors.New("время для редактирования прошло")
	// ErrCommentNotFoundByID ...
	ErrCommentNotFoundByID = errors.New("комментарий не найден")
)

// ErrCommentID ...
var ErrCommentID int64 = -1
