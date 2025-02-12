package service

import (
	"context"
	"errors"
	"github.com/Quizert/PostCommentService/internal/consts"
	"github.com/Quizert/PostCommentService/internal/errdefs"
	"github.com/Quizert/PostCommentService/internal/models"
	"github.com/Quizert/PostCommentService/internal/service/mocks"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestPostService_CreatePost(t *testing.T) {
	now := time.Now().UTC()
	mockUser := &models.User{
		ID:       1,
		Username: "testuser",
	}

	tests := []struct {
		name          string
		input         models.NewPost
		mockUser      *models.User
		mockUserErr   error
		mockPost      *models.Post
		mockPostErr   error
		expectedPost  *models.Post
		expectedError error
		expectDBCalls bool
	}{
		{
			name: "successful post creation",
			input: models.NewPost{
				Title:             "Test Post",
				Payload:           "Valid content",
				AuthorID:          1,
				IsCommentsAllowed: true,
			},
			mockUser: mockUser,
			mockPost: &models.Post{
				ID:                1,
				Title:             "Test Post",
				Payload:           "Valid content",
				IsCommentsAllowed: true,
				CreatedAt:         now,
			},
			expectedPost: &models.Post{
				ID:                1,
				Title:             "Test Post",
				Payload:           "Valid content",
				IsCommentsAllowed: true,
				CreatedAt:         now,
				Author:            mockUser,
			},
			expectDBCalls: true,
		},
		{
			name: "user not found",
			input: models.NewPost{
				AuthorID: 999,
			},
			mockUserErr:   pgx.ErrNoRows,
			expectedError: errdefs.UserDoesNotExistError(999),
			expectDBCalls: false,
		},
		{
			name: "payload too long",
			input: models.NewPost{
				Title:    "Long Post",
				Payload:  string(make([]byte, consts.MaxPayloadSize+1)),
				AuthorID: 1,
			},
			mockUser:      mockUser,
			expectedError: errdefs.CommentTooLongError(consts.MaxPayloadSize, consts.MaxPayloadSize+1),
			expectDBCalls: false,
		},
		{
			name: "database error on create",
			input: models.NewPost{
				Title:             "Failing Post",
				Payload:           "Valid content",
				AuthorID:          1,
				IsCommentsAllowed: true,
			},
			mockUser:      mockUser,
			mockPostErr:   errors.New("database failure"),
			expectedError: errdefs.InternalServerError(),
			expectDBCalls: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			userProvider := mocks.NewMockUserProvider(ctl)
			postProvider := mocks.NewMockPostProvider(ctl)
			commentProvider := mocks.NewMockCommentProvider(ctl)
			userProvider.EXPECT().
				GetUserByID(gomock.Any(), tt.input.AuthorID).
				Return(tt.mockUser, tt.mockUserErr).
				Times(1)

			if tt.expectDBCalls {
				postProvider.EXPECT().
					CreatePost(gomock.Any(), tt.input).
					Return(tt.mockPost, tt.mockPostErr).
					Times(1)
			}

			storage := NewStorage(postProvider, commentProvider, userProvider)
			logger := zap.NewNop()
			postService := NewPostService(logger, storage)

			ctx := context.Background()
			result, err := postService.CreatePost(ctx, tt.input)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				require.NoError(t, err)
			}

			if tt.expectedPost != nil {
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedPost.ID, result.ID)
				assert.Equal(t, tt.expectedPost.Title, result.Title)
				assert.Equal(t, tt.expectedPost.Payload, result.Payload)
				assert.Equal(t, tt.expectedPost.IsCommentsAllowed, result.IsCommentsAllowed)
				assert.True(t, tt.expectedPost.CreatedAt.Equal(result.CreatedAt))
				if tt.expectedPost.Author != nil {
					assert.Equal(t, tt.expectedPost.Author.ID, result.Author.ID)
					assert.Equal(t, tt.expectedPost.Author.Username, result.Author.Username)
				}
			}
		})
	}
}

func TestPostService_GetPostByID(t *testing.T) {
	tests := []struct {
		name          string
		postID        int
		mockPost      *models.Post
		mockPostErr   error
		expectedPost  *models.Post
		expectedError error
	}{
		{
			name:   "post exists",
			postID: 1,
			mockPost: &models.Post{
				ID:      1,
				Title:   "Example Post",
				Payload: "Some content",
			},
			expectedPost: &models.Post{
				ID:      1,
				Title:   "Example Post",
				Payload: "Some content",
			},
		},
		{
			name:          "post does not exist (pgx.ErrNoRows)",
			postID:        999,
			mockPostErr:   pgx.ErrNoRows,
			expectedError: errdefs.PostDoesNotExistError(999),
		},
		{
			name:          "internal server error",
			postID:        10,
			mockPostErr:   errors.New("some db error"),
			expectedError: errdefs.InternalServerError(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			postProvider := mocks.NewMockPostProvider(ctl)
			userProvider := mocks.NewMockUserProvider(ctl)
			commentProvider := mocks.NewMockCommentProvider(ctl)
			postProvider.EXPECT().
				GetPostByID(gomock.Any(), tt.postID).
				Return(tt.mockPost, tt.mockPostErr).
				Times(1)

			storage := NewStorage(postProvider, commentProvider, userProvider)
			logger := zap.NewNop()
			postService := NewPostService(logger, storage)

			ctx := context.Background()
			result, err := postService.GetPostByID(ctx, tt.postID)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedPost.ID, result.ID)
			assert.Equal(t, tt.expectedPost.Title, result.Title)
			assert.Equal(t, tt.expectedPost.Payload, result.Payload)
		})
	}
}

func TestPostService_GetAllPosts(t *testing.T) {
	tests := []struct {
		name          string
		limit         *int
		offset        *int
		limitValue    int
		offsetValue   int
		mockPosts     []*models.Post
		mockPostsErr  error
		expectedPosts []*models.Post
		expectedError error
	}{
		{
			name:        "success without limit/offset",
			limit:       nil,
			offset:      nil,
			limitValue:  consts.DefaultLimit,
			offsetValue: consts.DefaultOffset,
			mockPosts: []*models.Post{
				{ID: 1, Title: "Post 1"},
				{ID: 2, Title: "Post 2"},
			},
			expectedPosts: []*models.Post{
				{ID: 1, Title: "Post 1"},
				{ID: 2, Title: "Post 2"},
			},
		},
		{
			name:          "internal error",
			limit:         nil,
			offset:        nil,
			limitValue:    consts.DefaultLimit,
			offsetValue:   consts.DefaultOffset,
			mockPostsErr:  errors.New("database failure"),
			expectedError: errdefs.InternalServerError(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			postProvider := mocks.NewMockPostProvider(ctl)
			userProvider := mocks.NewMockUserProvider(ctl)
			commentProvider := mocks.NewMockCommentProvider(ctl)

			postProvider.EXPECT().
				GetAllPosts(gomock.Any(), tt.limitValue, tt.offsetValue).
				Return(tt.mockPosts, tt.mockPostsErr).
				Times(1)

			storage := NewStorage(postProvider, commentProvider, userProvider)
			logger := zap.NewNop()
			postService := NewPostService(logger, storage)

			ctx := context.Background()
			result, err := postService.GetAllPosts(ctx, tt.limit, tt.offset)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expectedPosts, result)
		})
	}
}
