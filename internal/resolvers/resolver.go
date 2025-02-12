package graphql

import (
	"context"
	"github.com/Quizert/PostCommentService/internal/models"
	"go.uber.org/zap"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

//go:generate mockgen -source=resolver.go -destination=mocks/services-mock.go -package=mocks
type PostService interface {
	CreatePost(ctx context.Context, input models.NewPost) (*models.Post, error)
	GetPostByID(ctx context.Context, id int) (*models.Post, error)
	GetAllPosts(ctx context.Context, limit *int, offset *int) ([]*models.Post, error)
}

type CommentService interface {
	CreateComment(ctx context.Context, input models.NewComment) (*models.Comment, error)
	GetCommentsByPostID(ctx context.Context, limit *int, offset *int, postID int) ([]*models.Comment, error)
	Replies(ctx context.Context, commentID int, limit *int, offset *int) ([]*models.Comment, error)
}

type SubscriptionService interface {
	CreateSubscription(ctx context.Context, postID int) (chan *models.Comment, error)
	DeleteSubscription(ctx context.Context, postID int, ch chan *models.Comment) error
	Notify(ctx context.Context, comment *models.Comment) error
}

type Resolver struct {
	log                 *zap.Logger
	postService         PostService
	commentService      CommentService
	subscriptionManager SubscriptionService
}

func NewResolver(log *zap.Logger, postService PostService, commentService CommentService, subscriptionManager SubscriptionService) *Resolver {
	return &Resolver{
		log:                 log,
		postService:         postService,
		commentService:      commentService,
		subscriptionManager: subscriptionManager,
	}
}
