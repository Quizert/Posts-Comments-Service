package in_memory

import (
	"context"
	"github.com/Quizert/PostCommentService/internal/models"
	"go.uber.org/zap"
	"log"
	"sort"
	"time"
)

type CommentMemoryStorage struct {
	log     *zap.Logger
	storage *InMemoryStorage
}

func NewCommentMemoryStorage(log *zap.Logger, storage *InMemoryStorage) *CommentMemoryStorage {
	return &CommentMemoryStorage{
		log:     log,
		storage: storage,
	}
}

func (c *CommentMemoryStorage) CreateComment(ctx context.Context, input models.NewComment) (*models.Comment, error) {
	c.storage.mu.Lock()
	defer c.storage.mu.Unlock()

	newID := c.storage.nextCommentID
	c.storage.nextCommentID++

	comment := &models.Comment{
		ID:        newID,
		Payload:   input.Payload,
		PostID:    input.PostID,
		ReplyTo:   input.ReplyTo,
		CreatedAt: time.Now(),
	}

	c.storage.comments[newID] = comment
	return comment, nil
}

func (c *CommentMemoryStorage) GetCommentsByPostID(ctx context.Context, limit, offset, postID int) ([]*models.Comment, error) {
	log.Println("AKOLFAKOLJFKLAWFHLJIKAWFHJIK:LAWHFIUJLAWHFLJIAWHNFJKLAWHNJF:KLWAHNFLJKAS:KHNFKSANF:ASKJFh")
	c.storage.mu.RLock()
	defer c.storage.mu.RUnlock()

	// Собираем комментарии, у которых comment.PostID == postID и comment.ReplyTo == nil
	filtered := make([]*models.Comment, 0)
	for _, comment := range c.storage.comments {
		if comment.PostID == postID && comment.ReplyTo == nil {
			filtered = append(filtered, comment)
		}
	}
	if offset >= len(filtered) {
		return []*models.Comment{}, nil
	}

	// Аналог order by
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	filtered = filtered[offset:]

	if limit > len(filtered) {
		limit = len(filtered)
	}

	result := filtered[:limit]

	return result, nil
}

func (c *CommentMemoryStorage) Replies(ctx context.Context, commentID, limit, offset int) ([]*models.Comment, error) {
	c.storage.mu.RLock()
	defer c.storage.mu.RUnlock()

	filtered := make([]*models.Comment, 0)
	for _, comment := range c.storage.comments {
		if comment.ReplyTo != nil && *comment.ReplyTo == commentID {
			filtered = append(filtered, comment)
		}
	}
	if offset >= len(filtered) {
		return []*models.Comment{}, nil
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	filtered = filtered[offset:]

	if limit > len(filtered) {
		limit = len(filtered)
	}
	result := filtered[:limit]

	return result, nil
}
