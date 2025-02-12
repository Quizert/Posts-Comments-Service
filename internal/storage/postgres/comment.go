package postgres

import (
	"context"
	"github.com/Quizert/PostCommentService/internal/models"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"time"
)

type CommentPostgresRepository struct {
	db  *pgxpool.Pool
	log *zap.Logger
}

func NewCommentPostgresRepository(db *pgxpool.Pool, log *zap.Logger) *CommentPostgresRepository {
	return &CommentPostgresRepository{
		db:  db,
		log: log,
	}
}

func (c *CommentPostgresRepository) CreateComment(ctx context.Context, input models.NewComment) (*models.Comment, error) {
	log := c.log.With(
		zap.String("Layer", "CommentPostgresRepository.CreateComment"),
		zap.Int("PostID", input.PostID),
		zap.Int("AuthorID", input.AuthorID),
	)

	query := `
		INSERT INTO comments (payload, postID, authorID, replyTo, createdAt)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, createdAt
	`

	var commentID int
	var createdAt time.Time

	err := c.db.QueryRow(ctx, query, input.Payload, input.PostID, input.AuthorID, input.ReplyTo).Scan(&commentID, &createdAt)

	if err != nil {
		log.Error("Failed to create comment", zap.Error(err))
		return nil, err
	}
	comment := &models.Comment{
		ID:        commentID,
		Payload:   input.Payload,
		PostID:    input.PostID,
		ReplyTo:   input.ReplyTo,
		CreatedAt: createdAt,
	}
	return comment, nil
}

func (c *CommentPostgresRepository) GetCommentsByPostID(ctx context.Context, limit int, offset int, postID int) ([]*models.Comment, error) {
	log := c.log.With(
		zap.String("Layer", "CommentPostgresRepository.GetCommentsByPostID"),
		zap.Int("PostID", postID),
	)

	query := `
		SELECT c.id, c.payload, c.postID, c.createdAt, u.id, u.username 
		FROM comments c join users u on c.authorID = u.id
		WHERE c.postID = $1
		AND c.replyto IS NULL 
		ORDER BY c.createdAt 
		DESC LIMIT $2 OFFSET $3
	`

	comments := make([]*models.Comment, 0, limit)
	rows, err := c.db.Query(ctx, query, postID, limit, offset)
	if err != nil {
		log.Error("Error getting comments", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var comment models.Comment
		comment.Author = &models.User{}
		err = rows.Scan(
			&comment.ID,
			&comment.Payload,
			&comment.PostID,
			&comment.CreatedAt,
			&comment.Author.ID,
			&comment.Author.Username,
		)
		if err != nil {
			log.Error("Failed to scan row", zap.Error(err))
			return nil, err
		}
		comments = append(comments, &comment)
	}
	if err = rows.Err(); err != nil {
		log.Error("Error after reading rows", zap.Error(err))
		return nil, err
	}
	return comments, nil
}

func (c *CommentPostgresRepository) Replies(ctx context.Context, commentID int, limit int, offset int) ([]*models.Comment, error) {
	log := c.log.With(
		zap.String("Layer", "CommentPostgresRepository.Replies"),
		zap.Int("CommentID", commentID),
	)

	comments := make([]*models.Comment, 0, limit)
	query := `
		SELECT c.id, c.payload, c.postID, c.createdAt, u.id, u.username
		FROM comments c join users u on c.authorID = u.id
		WHERE c.replyTo = $1 order by c.createdAt DESC LIMIT $2 OFFSET $3
	`
	rows, err := c.db.Query(ctx, query, commentID, limit, offset)
	if err != nil {
		log.Error("Error getting comments", zap.Error(err))
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var comment models.Comment
		comment.Author = &models.User{}
		err = rows.Scan(
			&comment.ID,
			&comment.Payload,
			&comment.PostID,
			&comment.CreatedAt,
			&comment.Author.ID,
			&comment.Author.Username,
		)
		if err != nil {
			log.Error("Failed to scan row", zap.Error(err))
			return nil, err
		}
		comments = append(comments, &comment)
	}
	if err = rows.Err(); err != nil {
		log.Error("Error after reading rows", zap.Error(err))
		return nil, err
	}
	return comments, nil
}
