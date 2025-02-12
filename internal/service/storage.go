package service

import (
	"context"
	"github.com/Quizert/PostCommentService/internal/models"
)

type Storage struct {
	PostProvider
	CommentProvider
	UserProvider
}

func NewStorage(postProvider PostProvider, commentProvider CommentProvider, userProvider UserProvider) *Storage {
	return &Storage{
		postProvider,
		commentProvider,
		userProvider,
	}
}

//go:generate mockgen -source=storage.go -destination=mocks/providers-mock.go -package=mocks PostProvider
type PostProvider interface {
	CreatePost(ctx context.Context, input models.NewPost) (*models.Post, error)
	GetAllPosts(ctx context.Context, limit int, offset int) ([]*models.Post, error)
	GetPostByID(ctx context.Context, id int) (*models.Post, error)
}

type CommentProvider interface {
	CreateComment(ctx context.Context, input models.NewComment) (*models.Comment, error)
	GetCommentsByPostID(ctx context.Context, limit int, offset int, postID int) ([]*models.Comment, error)
	Replies(ctx context.Context, commentID int, limit int, offset int) ([]*models.Comment, error)
}

type UserProvider interface {
	GetUserByID(ctx context.Context, userID int) (*models.User, error)
}
