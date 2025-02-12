package postgres

import (
	"context"
	"errors"
	"github.com/Quizert/PostCommentService/internal/models"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

type UserPostgresRepository struct {
	db  *pgxpool.Pool
	log *zap.Logger
}

func NewUserPostgresRepository(db *pgxpool.Pool, log *zap.Logger) *UserPostgresRepository {
	return &UserPostgresRepository{
		db:  db,
		log: log,
	}
}

func (p *UserPostgresRepository) GetUserByID(ctx context.Context, userID int) (*models.User, error) {
	log := p.log.With(
		zap.String("Layer", "PostPostgresRepository.GetUserByID"),
		zap.Int("UserID", userID),
	)

	var user models.User
	query := `SELECT id, username FROM users WHERE id = $1`
	err := p.db.QueryRow(ctx, query, userID).Scan(&user.ID, &user.Username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Warn("User does not exist")
			return nil, err
		}
		log.Error("Error getting user", zap.Error(err))
		return nil, err
	}

	return &user, nil
}
