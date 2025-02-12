package in_memory

import (
	"github.com/Quizert/PostCommentService/internal/models"
	"sync"
)

type InMemoryStorage struct {
	posts    map[int]*models.Post
	comments map[int]*models.Comment
	users    map[int]*models.User

	nextPostID    int
	nextCommentID int
	nextUserID    int

	mu sync.RWMutex
}

func NewInMemoryStorage() *InMemoryStorage {
	storage := &InMemoryStorage{
		posts:         make(map[int]*models.Post),
		comments:      make(map[int]*models.Comment),
		users:         make(map[int]*models.User),
		nextPostID:    1,
		nextCommentID: 1,
		nextUserID:    4,
	}

	user1 := &models.User{
		ID:       1,
		Username: "Alice",
	}
	user2 := &models.User{
		ID:       2,
		Username: "Quizert",
	}
	user3 := &models.User{
		ID:       3,
		Username: "Alen",
	}
	storage.users[user1.ID] = user1
	storage.users[user2.ID] = user2
	storage.users[user3.ID] = user3
	return storage
}
