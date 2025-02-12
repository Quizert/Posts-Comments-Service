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

type PostService struct {
	log     *zap.Logger
	storage *Storage
}

func NewPostService(log *zap.Logger, storage *Storage) *PostService {
	return &PostService{
		log:     log,
		storage: storage,
	}
}

func (p *PostService) CreatePost(ctx context.Context, input models.NewPost) (*models.Post, error) {
	author, err := p.storage.GetUserByID(ctx, input.AuthorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errdefs.UserDoesNotExistError(input.AuthorID)
		}
		return nil, errdefs.InternalServerError()
	}

	if len(input.Payload) > consts.MaxPayloadSize {
		return nil, errdefs.CommentTooLongError(consts.MaxPayloadSize, len(input.Payload))
	}

	post, err := p.storage.CreatePost(ctx, input)
	if err != nil {
		return nil, errdefs.InternalServerError()
	}
	post.Author = author
	return post, nil
}

func (p *PostService) GetPostByID(ctx context.Context, id int) (*models.Post, error) {
	post, err := p.storage.GetPostByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errdefs.PostDoesNotExistError(id)
		}
		return nil, errdefs.InternalServerError()
	}
	return post, nil
}

func (p *PostService) GetAllPosts(ctx context.Context, limit *int, offset *int) ([]*models.Post, error) {
	limitValue, offsetValue := utils.ParseLimitOffset(limit, offset)

	posts, err := p.storage.GetAllPosts(ctx, limitValue, offsetValue)
	if err != nil {
		return nil, errdefs.InternalServerError()
	}

	return posts, nil
}
