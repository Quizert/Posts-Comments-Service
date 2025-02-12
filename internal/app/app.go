package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/Quizert/PostCommentService/graph"
	"github.com/Quizert/PostCommentService/internal/config"
	graphql "github.com/Quizert/PostCommentService/internal/resolvers"
	"github.com/Quizert/PostCommentService/internal/service"
	in_memory "github.com/Quizert/PostCommentService/internal/storage/in-memory"
	"github.com/Quizert/PostCommentService/internal/storage/postgres"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func NewDatabasePool(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*pgxpool.Pool, error) {
	log.Println(cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	logger.Info("Connecting to database", zap.String("connection_string", connString))
	return pgxpool.Connect(ctx, connString)
}

func NewInMemoryStorage() *in_memory.InMemoryStorage {
	return in_memory.NewInMemoryStorage()
}

type App struct {
	DbPool *pgxpool.Pool
	Log    *zap.Logger
	Server *http.Server
}

func InitApp(ctx context.Context) (*App, error) {
	log, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("logger init error: %w", err)
	}
	log.Info("Initializing app")
	cfg := config.MustLoad(log)

	var (
		postProvider    service.PostProvider
		commentProvider service.CommentProvider
		userProvider    service.UserProvider
	)
	var dbPool *pgxpool.Pool
	switch cfg.StorageMode {
	case "memory":
		memoryStorage := NewInMemoryStorage()

		postProvider = in_memory.NewPostMemoryStorage(log, memoryStorage)
		commentProvider = in_memory.NewCommentMemoryStorage(log, memoryStorage)
		userProvider = in_memory.NewUserMemoryStorage(log, memoryStorage)

		log.Info("Using in-memory storage")
	case "postgres":
		dbPool, err = NewDatabasePool(ctx, cfg, log)
		if err != nil {
			log.Fatal("Error connecting to database", zap.Error(err))
		}

		postProvider = postgres.NewPostPostgresRepository(dbPool, log)
		commentProvider = postgres.NewCommentPostgresRepository(dbPool, log)
		userProvider = postgres.NewUserPostgresRepository(dbPool, log)

		log.Info("Using postgres storage")
	}

	storage := service.NewStorage(postProvider, commentProvider, userProvider)

	postService := service.NewPostService(log, storage)
	commentService := service.NewCommentService(log, storage)
	subManager := service.NewSubscriptionService()
	resolver := graphql.NewResolver(log, postService, commentService, subManager)

	mux := http.NewServeMux()
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	mux.Handle("/", playground.Handler("GraphQL Playground", "/query"))
	mux.Handle("/query", srv)

	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: mux,
	}

	app := &App{
		DbPool: dbPool,
		Log:    log,
		Server: server,
	}

	return app, nil
}

func (a *App) Start() error {
	a.Log.Info("Starting GraphQL server...")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.Log.Fatal("GraphQL server error", zap.Error(err))
		}
	}()

	a.Log.Info("Server is running", zap.String("address", a.Server.Addr))

	sig := <-signalChan
	a.Log.Info("Received shutdown signal", zap.String("signal", sig.String()))

	return a.Stop()
}

func (a *App) Stop() error {
	a.Log.Info("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.Server.Shutdown(ctx); err != nil {
		a.Log.Error("Server forced to shutdown", zap.Error(err))
		return err
	}
	a.Log.Info("HTTP server stopped gracefully")

	if a.DbPool != nil {
		a.DbPool.Close()
		a.Log.Info("Database connection closed")
	}

	a.Log.Info("Application stopped successfully")
	return nil
}
