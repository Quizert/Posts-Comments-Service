package service

import (
	"context"
	"errors"
	"github.com/Quizert/PostCommentService/internal/consts"
	"github.com/Quizert/PostCommentService/internal/errdefs"
	"github.com/Quizert/PostCommentService/internal/models"
	"github.com/Quizert/PostCommentService/internal/utils"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

type CommentService struct {
	log     *zap.Logger
	storage *Storage
}

func NewCommentService(log *zap.Logger, storage *Storage) *CommentService {
	return &CommentService{
		log,
		storage,
	}
}

func (c *CommentService) CreateComment(ctx context.Context, input models.NewComment) (*models.Comment, error) {
	author, err := c.storage.GetUserByID(ctx, input.AuthorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errdefs.UserDoesNotExistError(input.AuthorID)
		}
		return nil, errdefs.InternalServerError()
	}
	if len(input.Payload) > consts.MaxPayloadSize {
		return nil, errdefs.CommentTooLongError(consts.MaxPayloadSize, len(input.Payload))
	}

	post, err := c.storage.GetPostByID(ctx, input.PostID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errdefs.PostDoesNotExistError(input.PostID)
		}
		return nil, errdefs.InternalServerError()
	}
	if !post.IsCommentsAllowed {
		return nil, errdefs.CommentsNotAllowed(post.ID)
	}
	comment, err := c.storage.CreateComment(ctx, input)
	if err != nil {
		return nil, errdefs.InternalServerError()
	}
	comment.Author = author

	return comment, nil
}

func (c *CommentService) GetCommentsByPostID(ctx context.Context, limit *int, offset *int, postID int) ([]*models.Comment, error) {
	limitValue, offsetValue := utils.ParseLimitOffset(limit, offset)

	comments, err := c.storage.GetCommentsByPostID(ctx, limitValue, offsetValue, postID)
	if err != nil {
		return nil, errdefs.InternalServerError()
	}
	return comments, nil
}

func (c *CommentService) Replies(ctx context.Context, commentID int, limit *int, offset *int) ([]*models.Comment, error) {
	limitValue, offsetValue := utils.ParseLimitOffset(limit, offset)

	comments, err := c.storage.Replies(ctx, commentID, limitValue, offsetValue)
	if err != nil {
		return nil, errdefs.InternalServerError()
	}
	return comments, nil
}
