package in_memory

import (
	"context"
	"github.com/Quizert/PostCommentService/internal/errdefs"
	"github.com/Quizert/PostCommentService/internal/models"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"sort"
	"time"
)

type PostMemoryStorage struct {
	log     *zap.Logger
	storage *InMemoryStorage
}

func NewPostMemoryStorage(log *zap.Logger, storage *InMemoryStorage) *PostMemoryStorage {
	return &PostMemoryStorage{
		log:     log,
		storage: storage,
	}
}

func (p *PostMemoryStorage) CreatePost(ctx context.Context, input models.NewPost) (*models.Post, error) {
	p.storage.mu.Lock()
	defer p.storage.mu.Unlock()

	log := p.log.With(
		zap.String("Layer", "PostMemoryStorage.CreatePost"),
		zap.String("Title", input.Title),
		zap.Int("AuthorID", input.AuthorID),
	)

	newID := p.storage.nextPostID
	p.storage.nextPostID++

	author, ok := p.storage.users[input.AuthorID]
	if !ok {
		log.Error("Failed to get user")
		return nil, errdefs.UserDoesNotExistError(input.AuthorID) // Не может быть так как на уровне сервиса уже проверили
	}

	post := &models.Post{
		ID:                newID,
		Title:             input.Title,
		Payload:           input.Payload,
		Author:            author,
		IsCommentsAllowed: input.IsCommentsAllowed,
		CreatedAt:         time.Now(),
	}

	p.storage.posts[newID] = post
	return post, nil
}

func (p *PostMemoryStorage) GetPostByID(ctx context.Context, id int) (*models.Post, error) {
	p.storage.mu.RLock()
	defer p.storage.mu.RUnlock()

	log := p.log.With(
		zap.String("Layer", "PostMemoryStorage.GetPostByID"),
		zap.Int("PostID", id),
	)

	post, ok := p.storage.posts[id]
	if !ok {
		log.Warn("Failed to get post: not found in memory")
		return nil, pgx.ErrNoRows
	}

	return post, nil
}

func (p *PostMemoryStorage) GetAllPosts(ctx context.Context, limit, offset int) ([]*models.Post, error) {
	p.storage.mu.RLock()
	defer p.storage.mu.RUnlock()

	postsSlice := make([]*models.Post, 0, len(p.storage.posts))
	for _, post := range p.storage.posts {
		postsSlice = append(postsSlice, post)
	}

	if offset >= len(postsSlice) {
		return []*models.Post{}, nil
	}

	// Сортируем, аналог order by
	sort.Slice(postsSlice, func(i, j int) bool {
		return postsSlice[i].CreatedAt.After(postsSlice[j].CreatedAt)
	})

	postsSlice = postsSlice[offset:]

	if limit > len(postsSlice) {
		limit = len(postsSlice)
	}
	postsSlice = postsSlice[:limit]

	return postsSlice, nil
}
