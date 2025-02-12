package in_memory

import (
	"context"
	"testing"
	"time"

	"github.com/Quizert/PostCommentService/internal/models"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestPostMemoryStorage_CreatePost(t *testing.T) {
	logger := zap.NewNop()

	t.Run("success", func(t *testing.T) {
		storage := NewInMemoryStorage()
		postStorage := NewPostMemoryStorage(logger, storage)

		user := &models.User{
			ID:       1,
			Username: "testuser",
		}
		storage.users[1] = user

		input := models.NewPost{
			Title:             "Hello World",
			Payload:           "Some content",
			AuthorID:          1,
			IsCommentsAllowed: true,
		}

		post, err := postStorage.CreatePost(context.Background(), input)
		require.NoError(t, err)
		require.NotNil(t, post)

		assert.Equal(t, 1, post.ID, "первый пост должен получить ID = 1 (nextPostID = 1)")
		assert.Equal(t, input.Title, post.Title)
		assert.Equal(t, input.Payload, post.Payload)
		assert.Equal(t, user, post.Author)
		assert.True(t, post.IsCommentsAllowed)
		assert.WithinDuration(t, time.Now(), post.CreatedAt, time.Second, "CreatedAt должен быть близок к текущему времени")

		got, ok := storage.posts[post.ID]
		require.True(t, ok)
		assert.Equal(t, post, got)
	})

	t.Run("user does not exist", func(t *testing.T) {
		storage := NewInMemoryStorage()
		postStorage := NewPostMemoryStorage(logger, storage)

		input := models.NewPost{
			Title:             "No user",
			Payload:           "No user payload",
			AuthorID:          999,
			IsCommentsAllowed: false,
		}

		post, err := postStorage.CreatePost(context.Background(), input)
		require.Error(t, err)
		assert.Nil(t, post)

		assert.Empty(t, storage.posts)
	})
}

func TestPostMemoryStorage_GetPostByID(t *testing.T) {
	logger := zap.NewNop()

	t.Run("success", func(t *testing.T) {
		storage := NewInMemoryStorage()
		postStorage := NewPostMemoryStorage(logger, storage)

		user := &models.User{ID: 1, Username: "Alice"}
		storage.users[1] = user

		post := &models.Post{
			ID:        100,
			Title:     "Post #100",
			Payload:   "Some data",
			Author:    user,
			CreatedAt: time.Now(),
		}
		storage.posts[100] = post

		got, err := postStorage.GetPostByID(context.Background(), 100)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, 100, got.ID)
		assert.Equal(t, post, got)
	})

	t.Run("not found", func(t *testing.T) {
		storage := NewInMemoryStorage()
		postStorage := NewPostMemoryStorage(logger, storage)

		got, err := postStorage.GetPostByID(context.Background(), 999)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

func TestPostMemoryStorage_GetAllPosts(t *testing.T) {
	logger := zap.NewNop()

	t.Run("empty storage", func(t *testing.T) {
		storage := NewInMemoryStorage()
		postStorage := NewPostMemoryStorage(logger, storage)

		posts, err := postStorage.GetAllPosts(context.Background(), 10, 0)
		require.NoError(t, err)
		assert.Empty(t, posts)
	})

	t.Run("success with sorting", func(t *testing.T) {
		storage := NewInMemoryStorage()
		postStorage := NewPostMemoryStorage(logger, storage)

		now := time.Now()
		post1 := &models.Post{
			ID:        1,
			Title:     "First",
			Payload:   "First payload",
			CreatedAt: now.Add(-3 * time.Hour),
		}
		post2 := &models.Post{
			ID:        2,
			Title:     "Second",
			Payload:   "Second payload",
			CreatedAt: now.Add(-1 * time.Hour),
		}
		post3 := &models.Post{
			ID:        3,
			Title:     "Third",
			Payload:   "Third payload",
			CreatedAt: now.Add(-2 * time.Hour),
		}
		storage.posts[1] = post1
		storage.posts[2] = post2
		storage.posts[3] = post3

		posts, err := postStorage.GetAllPosts(context.Background(), 10, 0)
		require.NoError(t, err)
		require.Len(t, posts, 3)

		want := []*models.Post{post2, post3, post1}
		assert.Equal(t, want, posts)
	})

	t.Run("offset and limit", func(t *testing.T) {
		storage := NewInMemoryStorage()
		postStorage := NewPostMemoryStorage(logger, storage)

		now := time.Now()
		for i := 1; i <= 5; i++ {
			storage.posts[i] = &models.Post{
				ID:        i,
				Title:     "PostTitle",
				Payload:   "Some",
				CreatedAt: now.Add(-time.Duration(i) * time.Hour),
			}
		}

		posts, err := postStorage.GetAllPosts(context.Background(), 2, 1)
		require.NoError(t, err)

		require.Len(t, posts, 2)
		assert.Equal(t, 2, posts[0].ID)
		assert.Equal(t, 3, posts[1].ID)

		posts2, err2 := postStorage.GetAllPosts(context.Background(), 10, 10)
		require.NoError(t, err2)
		assert.Empty(t, posts2, "offset = 10 больше, чем число постов = 5 (я устал придумывать приколы на англ)")
	})
}
