package config

import (
	"go.uber.org/zap"
	"os"
)

func mustGetEnv(log *zap.Logger, key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatal("missing environment variable", zap.String("key", key))
	}
	return value
}

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	HTTPPort string

	StorageMode string
}

func MustLoad(log *zap.Logger) *Config {
	dbName := mustGetEnv(log, "DB_NAME")
	dbHost := mustGetEnv(log, "DB_HOST")
	dbPort := mustGetEnv(log, "DB_PORT")
	dbUser := mustGetEnv(log, "DB_USER")
	dbPassword := mustGetEnv(log, "DB_PASSWORD")
	httpPort := mustGetEnv(log, "HTTP_PORT")

	storageMode := mustGetEnv(log, "STORAGE_MODE")

	return &Config{
		DBName:      dbName,
		DBHost:      dbHost,
		DBPort:      dbPort,
		DBUser:      dbUser,
		DBPassword:  dbPassword,
		HTTPPort:    httpPort,
		StorageMode: storageMode,
	}
}
