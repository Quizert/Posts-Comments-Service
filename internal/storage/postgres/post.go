package postgres

import (
	"context"
	"errors"
	"github.com/Quizert/PostCommentService/internal/models"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

type PostPostgresRepository struct {
	db  *pgxpool.Pool
	log *zap.Logger
}

func NewPostPostgresRepository(db *pgxpool.Pool, log *zap.Logger) *PostPostgresRepository {
	return &PostPostgresRepository{
		db:  db,
		log: log,
	}
}

func (p *PostPostgresRepository) CreatePost(ctx context.Context, input models.NewPost) (*models.Post, error) {
	log := p.log.With(
		zap.String("Layer", "PostPostgresRepository.CreatePost"),
		zap.String("Title", input.Title),
		zap.Int("AuthorID", input.AuthorID),
	)

	query := `INSERT INTO posts (title, payload, authorID, isCommentsAllowed, createdAt)
              VALUES ($1, $2, $3, $4, NOW())
              RETURNING id, createdAt`

	var post models.Post
	err := p.db.QueryRow(ctx, query, input.Title, input.Payload, input.AuthorID, input.IsCommentsAllowed).
		Scan(&post.ID, &post.CreatedAt)

	if err != nil {
		log.Error("Failed to create post", zap.Error(err))
		return nil, err
	}

	post.Title = input.Title
	post.Payload = input.Payload
	post.IsCommentsAllowed = input.IsCommentsAllowed

	return &post, nil
}

func (p *PostPostgresRepository) GetPostByID(ctx context.Context, id int) (*models.Post, error) {
	log := p.log.With(
		zap.String("Layer", "PostPostgresRepository.GetPostByID"),
		zap.Int("PostID", id),
	)
	var post models.Post
	post.Author = &models.User{}

	query := `
		SELECT p.id, p.title, p.payload, p.isCommentsAllowed, p.createdAt, u.id as author_id, u.username as author_username
		FROM posts p JOIN users u ON p.authorID = u.id
		WHERE p.id = $1
	`

	// Выполняем запрос и сканируем результат в структуру Post
	err := p.db.QueryRow(ctx, query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Payload,
		&post.IsCommentsAllowed,
		&post.CreatedAt,
		&post.Author.ID,
		&post.Author.Username,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Warn("Failed to get post", zap.Error(err))
			return nil, err
		}
		log.Error("Failed to get post", zap.Error(err))
		return nil, err
	}
	return &post, nil
}

func (p *PostPostgresRepository) GetAllPosts(ctx context.Context, limit int, offset int) ([]*models.Post, error) {
	log := p.log.With(
		zap.String("Layer", "PostPostgresRepository.GetAllPosts"),
	)

	posts := make([]*models.Post, 0, limit)

	query := `
		SELECT p.id, p.title, p.payload, p.isCommentsAllowed, p.createdAt, u.id, u.username
		FROM posts p join users u on p.authorID = u.id 
		ORDER BY p.createdAt DESC LIMIT $1 OFFSET $2
	`

	rows, err := p.db.Query(ctx, query, limit, offset)
	if err != nil {
		log.Error("Failed to get posts", zap.Error(err))
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var post models.Post
		post.Author = &models.User{}
		err = rows.Scan(
			&post.ID,
			&post.Title,
			&post.Payload,
			&post.IsCommentsAllowed,
			&post.CreatedAt,
			&post.Author.ID,
			&post.Author.Username,
		)
		if err != nil {
			log.Error("Failed to get posts", zap.Error(err))
			return nil, err
		}
		posts = append(posts, &post)

	}

	if err = rows.Err(); err != nil {
		log.Error("Error after reading rows", zap.Error(err))
		return nil, err
	}

	return posts, nil
}
