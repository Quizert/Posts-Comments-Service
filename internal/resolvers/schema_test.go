package graphql

import (
	"context"
	"errors"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Quizert/PostCommentService/internal/models"
	"github.com/Quizert/PostCommentService/internal/resolvers/mocks"
	"go.uber.org/zap"
)

func TestCommentResolver_Replies(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	commentServiceMock := mocks.NewMockCommentService(ctl)
	postServiceMock := mocks.NewMockPostService(ctl)
	subscriptionServiceMock := mocks.NewMockSubscriptionService(ctl)

	logger := zap.NewNop()
	res := NewResolver(logger, postServiceMock, commentServiceMock, subscriptionServiceMock)
	commentResolver := res.Comment()

	ctx := context.Background()
	comment := &models.Comment{ID: 123}
	limit := 10
	offset := 0

	t.Run("success", func(t *testing.T) {
		expectedComments := []*models.Comment{
			{ID: 1001, Payload: "reply #1"},
			{ID: 1002, Payload: "reply #2"},
		}

		commentServiceMock.
			EXPECT().
			Replies(gomock.Any(), comment.ID, &limit, &offset).
			Return(expectedComments, nil).
			Times(1)

		got, err := commentResolver.Replies(ctx, comment, &limit, &offset)
		require.NoError(t, err)
		require.Equal(t, expectedComments, got)
	})

	t.Run("service error", func(t *testing.T) {
		commentServiceMock.
			EXPECT().
			Replies(gomock.Any(), comment.ID, &limit, &offset).
			Return(nil, errors.New("some error")).
			Times(1)

		got, err := commentResolver.Replies(ctx, comment, &limit, &offset)
		assert.Nil(t, got)

		var appErr *gqlerror.Error
		require.Error(t, err)
		assert.ErrorAs(t, err, &appErr)
	})
}

func TestMutationResolver_CreatePost(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	commentServiceMock := mocks.NewMockCommentService(ctl)
	postServiceMock := mocks.NewMockPostService(ctl)
	subscriptionServiceMock := mocks.NewMockSubscriptionService(ctl)

	logger := zap.NewNop()
	res := NewResolver(logger, postServiceMock, commentServiceMock, subscriptionServiceMock)
	mutationResolver := res.Mutation()

	ctx := context.Background()
	input := models.NewPost{Title: "New post", Payload: "Payload content", AuthorID: 1}

	t.Run("success", func(t *testing.T) {
		createdPost := &models.Post{ID: 10, Title: "New post", Payload: "Payload content"}

		postServiceMock.
			EXPECT().
			CreatePost(gomock.Any(), input).
			Return(createdPost, nil).
			Times(1)

		got, err := mutationResolver.CreatePost(ctx, input)
		require.NoError(t, err)
		require.Equal(t, createdPost, got)
	})

	t.Run("service error", func(t *testing.T) {
		postServiceMock.
			EXPECT().
			CreatePost(gomock.Any(), input).
			Return(nil, errors.New("failed")).
			Times(1)

		got, err := mutationResolver.CreatePost(ctx, input)
		assert.Nil(t, got)

		var appErr *gqlerror.Error
		require.Error(t, err)
		assert.ErrorAs(t, err, &appErr)
	})
}

func TestMutationResolver_CreateComment(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	commentServiceMock := mocks.NewMockCommentService(ctl)
	postServiceMock := mocks.NewMockPostService(ctl)
	subscriptionServiceMock := mocks.NewMockSubscriptionService(ctl)

	logger := zap.NewNop()
	res := NewResolver(logger, postServiceMock, commentServiceMock, subscriptionServiceMock)
	mutationResolver := res.Mutation()

	ctx := context.Background()
	input := models.NewComment{PostID: 1, AuthorID: 1, Payload: "hello"}

	t.Run("success", func(t *testing.T) {
		createdComment := &models.Comment{ID: 11, Payload: "hello", PostID: 1}

		commentServiceMock.
			EXPECT().
			CreateComment(gomock.Any(), input).
			Return(createdComment, nil).
			Times(1)

		subscriptionServiceMock.
			EXPECT().
			Notify(gomock.Any(), createdComment).
			Return(nil).
			Times(1)

		got, err := mutationResolver.CreateComment(ctx, input)
		require.NoError(t, err)
		require.Equal(t, createdComment, got)
	})

	t.Run("create comment error", func(t *testing.T) {
		commentServiceMock.
			EXPECT().
			CreateComment(gomock.Any(), input).
			Return(nil, errors.New("db error")).
			Times(1)

		got, err := mutationResolver.CreateComment(ctx, input)
		assert.Nil(t, got)

		var appErr *gqlerror.Error
		require.Error(t, err)
		assert.ErrorAs(t, err, &appErr)
	})

	t.Run("notify error", func(t *testing.T) {
		createdComment := &models.Comment{ID: 11, Payload: "hello"}

		commentServiceMock.
			EXPECT().
			CreateComment(gomock.Any(), input).
			Return(createdComment, nil).
			Times(1)

		subscriptionServiceMock.
			EXPECT().
			Notify(gomock.Any(), createdComment).
			Return(errors.New("notify failed")).
			Times(1)

		got, err := mutationResolver.CreateComment(ctx, input)
		assert.Nil(t, got)

		var appErr *gqlerror.Error
		require.Error(t, err)
		assert.ErrorAs(t, err, &appErr)
	})
}

func TestPostResolver_Comments(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	commentServiceMock := mocks.NewMockCommentService(ctl)
	postServiceMock := mocks.NewMockPostService(ctl)
	subscriptionServiceMock := mocks.NewMockSubscriptionService(ctl)

	logger := zap.NewNop()
	res := NewResolver(logger, postServiceMock, commentServiceMock, subscriptionServiceMock)
	postResolver := res.Post()

	ctx := context.Background()
	post := &models.Post{ID: 99}
	limit := 5
	offset := 10

	t.Run("success", func(t *testing.T) {
		expectedComments := []*models.Comment{{ID: 1}, {ID: 2}}
		commentServiceMock.
			EXPECT().
			GetCommentsByPostID(gomock.Any(), &limit, &offset, post.ID).
			Return(expectedComments, nil).
			Times(1)

		got, err := postResolver.Comments(ctx, post, &limit, &offset)
		require.NoError(t, err)
		require.Equal(t, expectedComments, got)
	})

	t.Run("service error", func(t *testing.T) {
		commentServiceMock.
			EXPECT().
			GetCommentsByPostID(gomock.Any(), &limit, &offset, post.ID).
			Return(nil, errors.New("db error")).
			Times(1)

		got, err := postResolver.Comments(ctx, post, &limit, &offset)
		assert.Nil(t, got)
		var appErr *gqlerror.Error
		require.Error(t, err)
		assert.ErrorAs(t, err, &appErr)
	})
}

func TestQueryResolver_GetPostByID(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	postServiceMock := mocks.NewMockPostService(ctl)
	commentServiceMock := mocks.NewMockCommentService(ctl)
	subscriptionServiceMock := mocks.NewMockSubscriptionService(ctl)

	logger := zap.NewNop()
	res := NewResolver(logger, postServiceMock, commentServiceMock, subscriptionServiceMock)
	queryResolver := res.Query()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expectedPost := &models.Post{ID: 12, Title: "the post"}
		postServiceMock.
			EXPECT().
			GetPostByID(gomock.Any(), 12).
			Return(expectedPost, nil).
			Times(1)

		got, err := queryResolver.GetPostByID(ctx, 12)
		require.NoError(t, err)
		assert.Equal(t, expectedPost, got)
	})

	t.Run("service error", func(t *testing.T) {
		postServiceMock.
			EXPECT().
			GetPostByID(gomock.Any(), 12).
			Return(nil, errors.New("not found")).
			Times(1)

		got, err := queryResolver.GetPostByID(ctx, 12)
		assert.Nil(t, got)
		var appErr *gqlerror.Error
		require.Error(t, err)
		assert.ErrorAs(t, err, &appErr)
	})
}

func TestQueryResolver_GetAllPosts(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	postServiceMock := mocks.NewMockPostService(ctl)
	commentServiceMock := mocks.NewMockCommentService(ctl)
	subscriptionServiceMock := mocks.NewMockSubscriptionService(ctl)

	logger := zap.NewNop()
	res := NewResolver(logger, postServiceMock, commentServiceMock, subscriptionServiceMock)
	queryResolver := res.Query()

	ctx := context.Background()
	limit := 10
	offset := 0

	t.Run("success", func(t *testing.T) {
		posts := []*models.Post{{ID: 1}, {ID: 2}}
		postServiceMock.
			EXPECT().
			GetAllPosts(gomock.Any(), &limit, &offset).
			Return(posts, nil).
			Times(1)

		got, err := queryResolver.GetAllPosts(ctx, &limit, &offset)
		require.NoError(t, err)
		require.Equal(t, posts, got)
	})

	t.Run("service error", func(t *testing.T) {
		postServiceMock.
			EXPECT().
			GetAllPosts(gomock.Any(), &limit, &offset).
			Return(nil, errors.New("some error")).
			Times(1)

		got, err := queryResolver.GetAllPosts(ctx, &limit, &offset)
		assert.Nil(t, got)
		var appErr *gqlerror.Error
		require.Error(t, err)
		assert.ErrorAs(t, err, &appErr)
	})
}

func TestSubscriptionResolver_CommentsSubscription(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	postServiceMock := mocks.NewMockPostService(ctl)
	commentServiceMock := mocks.NewMockCommentService(ctl)
	subscriptionServiceMock := mocks.NewMockSubscriptionService(ctl)

	logger := zap.NewNop()
	res := NewResolver(logger, postServiceMock, commentServiceMock, subscriptionServiceMock)
	subscriptionResolver := res.Subscription()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("success", func(t *testing.T) {
		ch := make(chan *models.Comment)

		subscriptionServiceMock.
			EXPECT().
			CreateSubscription(gomock.Any(), 123).
			Return(ch, nil).
			Times(1)

		got, err := subscriptionResolver.CommentsSubscription(ctx, 123)
		require.NoError(t, err)
		assert.Equal(t, (<-chan *models.Comment)(ch), got)

		done := make(chan struct{})

		subscriptionServiceMock.
			EXPECT().
			DeleteSubscription(gomock.Any(), 123, ch).
			DoAndReturn(func(_ context.Context, _ int, _ chan *models.Comment) error {
				close(done)
				return nil
			}).
			Times(1)

		cancel()

		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Fatal("timeout: DeleteSubscription was not called")
		}
	})
}
