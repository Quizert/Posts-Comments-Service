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

func TestCommentService_CreateComment(t *testing.T) {
	now := time.Now().UTC()
	mockUser := &models.User{
		ID:       1,
		Username: "testuser",
	}
	mockPost := &models.Post{
		ID:                1,
		Title:             "Example Post",
		Payload:           "Example payload",
		IsCommentsAllowed: true,
		CreatedAt:         now,
	}

	tests := []struct {
		name            string
		input           models.NewComment
		mockUser        *models.User
		mockUserErr     error
		mockPost        *models.Post
		mockPostErr     error
		mockComment     *models.Comment
		mockCommentErr  error
		expectedComment *models.Comment
		expectedError   error
	}{
		{
			name: "success",
			input: models.NewComment{
				PostID:   1,
				AuthorID: 1,
				Payload:  "Valid comment",
			},
			mockUser: mockUser,
			mockPost: mockPost,
			mockComment: &models.Comment{
				ID:      100,
				Payload: "Valid comment",
			},
			expectedComment: &models.Comment{
				ID:      100,
				Payload: "Valid comment",
				Author:  mockUser,
			},
		},
		{
			name: "user not found",
			input: models.NewComment{
				PostID:   1,
				AuthorID: 999,
				Payload:  "Any payload",
			},
			mockUserErr:   pgx.ErrNoRows,
			expectedError: errdefs.UserDoesNotExistError(999),
		},
		{
			name: "user provider internal error",
			input: models.NewComment{
				PostID:   1,
				AuthorID: 999,
			},
			mockUserErr:   errors.New("some db error"),
			expectedError: errdefs.InternalServerError(),
		},
		{
			name: "payload too long",
			input: models.NewComment{
				PostID:   1,
				AuthorID: 1,
				Payload:  string(make([]byte, consts.MaxPayloadSize+1)),
			},
			mockUser:      mockUser,
			expectedError: errdefs.CommentTooLongError(consts.MaxPayloadSize, consts.MaxPayloadSize+1),
		},
		{
			name: "post not found",
			input: models.NewComment{
				PostID:   2,
				AuthorID: 1,
				Payload:  "Some comment",
			},
			mockUser:      mockUser,
			mockPostErr:   pgx.ErrNoRows,
			expectedError: errdefs.PostDoesNotExistError(2),
		},
		{
			name: "post provider internal error",
			input: models.NewComment{
				PostID:   2,
				AuthorID: 1,
			},
			mockUser:      mockUser,
			mockPostErr:   errors.New("some db error"),
			expectedError: errdefs.InternalServerError(),
		},
		{
			name: "comments not allowed",
			input: models.NewComment{
				PostID:   3,
				AuthorID: 1,
				Payload:  "Comment here",
			},
			mockUser: &models.User{
				ID:       1,
				Username: "testuser",
			},
			mockPost: &models.Post{
				ID:                3,
				IsCommentsAllowed: false,
			},
			expectedError: errdefs.CommentsNotAllowed(3),
		},
		{
			name: "db error on create comment",
			input: models.NewComment{
				PostID:   1,
				AuthorID: 1,
				Payload:  "Will fail on creation",
			},
			mockUser:       mockUser,
			mockPost:       mockPost,
			mockCommentErr: errors.New("insert failed"),
			expectedError:  errdefs.InternalServerError(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			userProvider := mocks.NewMockUserProvider(ctl)
			postProvider := mocks.NewMockPostProvider(ctl)
			commentProvider := mocks.NewMockCommentProvider(ctl)

			if tt.input.AuthorID != 0 {
				userProvider.EXPECT().
					GetUserByID(gomock.Any(), tt.input.AuthorID).
					Return(tt.mockUser, tt.mockUserErr).
					Times(1)
			}

			if tt.mockUserErr == nil && len(tt.input.Payload) <= consts.MaxPayloadSize {
				postProvider.EXPECT().
					GetPostByID(gomock.Any(), tt.input.PostID).
					Return(tt.mockPost, tt.mockPostErr).
					Times(1)
			}

			canCreate := tt.mockUserErr == nil &&
				len(tt.input.Payload) <= consts.MaxPayloadSize &&
				tt.mockPostErr == nil &&
				tt.mockPost != nil &&
				tt.mockPost.IsCommentsAllowed

			if canCreate {
				commentProvider.EXPECT().
					CreateComment(gomock.Any(), tt.input).
					Return(tt.mockComment, tt.mockCommentErr).
					Times(1)
			}

			storage := NewStorage(postProvider, commentProvider, userProvider)
			logger := zap.NewNop()
			commentService := NewCommentService(logger, storage)

			ctx := context.Background()
			result, err := commentService.CreateComment(ctx, tt.input)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			require.NoError(t, err)

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedComment.ID, result.ID)
			assert.Equal(t, tt.expectedComment.Payload, result.Payload)
			if tt.expectedComment.Author != nil {
				assert.Equal(t, tt.expectedComment.Author.ID, result.Author.ID)
				assert.Equal(t, tt.expectedComment.Author.Username, result.Author.Username)
			}
		})
	}
}

func TestCommentService_GetCommentsByPostID(t *testing.T) {
	tests := []struct {
		name             string
		limit            *int
		offset           *int
		limitValue       int
		offsetValue      int
		postID           int
		mockComments     []*models.Comment
		mockCommentsErr  error
		expectedComments []*models.Comment
		expectedError    error
	}{
		{
			name:        "success",
			limit:       nil,
			offset:      nil,
			limitValue:  consts.DefaultLimit,
			offsetValue: consts.DefaultOffset,
			postID:      1,
			mockComments: []*models.Comment{
				{ID: 1, Payload: "Hello"},
				{ID: 2, Payload: "World"},
			},
			expectedComments: []*models.Comment{
				{ID: 1, Payload: "Hello"},
				{ID: 2, Payload: "World"},
			},
		},
		{
			name:            "db error",
			limit:           nil,
			offset:          nil,
			limitValue:      consts.DefaultLimit,
			offsetValue:     consts.DefaultOffset,
			postID:          1,
			mockCommentsErr: errors.New("some db error"),
			expectedError:   errdefs.InternalServerError(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			postProvider := mocks.NewMockPostProvider(ctl)
			commentProvider := mocks.NewMockCommentProvider(ctl)
			userProvider := mocks.NewMockUserProvider(ctl)

			commentProvider.EXPECT().
				GetCommentsByPostID(gomock.Any(), tt.limitValue, tt.offsetValue, tt.postID).
				Return(tt.mockComments, tt.mockCommentsErr).
				Times(1)

			storage := NewStorage(postProvider, commentProvider, userProvider)
			logger := zap.NewNop()
			commentService := NewCommentService(logger, storage)

			ctx := context.Background()
			result, err := commentService.GetCommentsByPostID(ctx, tt.limit, tt.offset, tt.postID)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expectedComments, result)
		})
	}
}

func TestCommentService_Replies(t *testing.T) {
	tests := []struct {
		name             string
		commentID        int
		limit            *int
		offset           *int
		limitValue       int
		offsetValue      int
		mockComments     []*models.Comment
		mockCommentsErr  error
		expectedComments []*models.Comment
		expectedError    error
	}{
		{
			name:        "success",
			commentID:   10,
			limit:       nil,
			offset:      nil,
			limitValue:  consts.DefaultLimit,
			offsetValue: consts.DefaultOffset,
			mockComments: []*models.Comment{
				{ID: 101, Payload: "Reply #1"},
				{ID: 102, Payload: "Reply #2"},
			},
			expectedComments: []*models.Comment{
				{ID: 101, Payload: "Reply #1"},
				{ID: 102, Payload: "Reply #2"},
			},
		},
		{
			name:            "db error",
			commentID:       123,
			limit:           nil,
			offset:          nil,
			limitValue:      consts.DefaultLimit,
			offsetValue:     consts.DefaultOffset,
			mockCommentsErr: errors.New("some db error"),
			expectedError:   errdefs.InternalServerError(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			postProvider := mocks.NewMockPostProvider(ctl)
			commentProvider := mocks.NewMockCommentProvider(ctl)
			userProvider := mocks.NewMockUserProvider(ctl)

			commentProvider.EXPECT().
				Replies(gomock.Any(), tt.commentID, tt.limitValue, tt.offsetValue).
				Return(tt.mockComments, tt.mockCommentsErr).
				Times(1)

			storage := NewStorage(postProvider, commentProvider, userProvider)
			logger := zap.NewNop()
			commentService := NewCommentService(logger, storage)

			ctx := context.Background()
			result, err := commentService.Replies(ctx, tt.commentID, tt.limit, tt.offset)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expectedComments, result)
		})
	}
}
