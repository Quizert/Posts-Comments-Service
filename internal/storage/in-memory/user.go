package in_memory

import (
	"context"
	"github.com/Quizert/PostCommentService/internal/models"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

type UserMemoryStorage struct {
	log     *zap.Logger
	storage *InMemoryStorage
}

func NewUserMemoryStorage(log *zap.Logger, storage *InMemoryStorage) *UserMemoryStorage {
	return &UserMemoryStorage{
		log:     log,
		storage: storage,
	}
}

func (u *UserMemoryStorage) GetUserByID(ctx context.Context, userID int) (*models.User, error) {
	log := u.log.With(
		zap.String("Layer", "PostPostgresRepository.GetUserByID"),
		zap.Int("UserID", userID),
	)

	user, ok := u.storage.users[userID]
	if !ok {
		log.Warn("User does not exist")
		return nil, pgx.ErrNoRows
	}
	return user, nil
}
