package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/AGX18/real_estate_crm/internal/config"
	"github.com/AGX18/real_estate_crm/internal/db/migrations"

	"github.com/jackc/pgx/v5/stdlib"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"github.com/pressly/goose/v3"
)

type Service interface {
	Health() map[string]string
	Close()
	GetDB() *pgxpool.Pool
}

type service struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

var (
	dbInstance *service
)

func New(logger *slog.Logger, cfg config.Database) (Service, error) {
	if dbInstance != nil {
		return dbInstance, nil
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
	var pool *pgxpool.Pool
	var err error
	if cfg.URL != "" {
		pool, err = pgxpool.New(context.Background(), cfg.URL)
		if err != nil {
			return nil, err
		}
	} else {
		pool, err = pgxpool.New(context.Background(), connStr)
		if err != nil {
			return nil, err
		}
	}

	logger.Debug("created pgxpool", "connection string:", pool.Config().ConnString())

	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}
	sqlDB := stdlib.OpenDBFromPool(pool)

	if err := goose.Up(sqlDB, "."); err != nil {
		logger.Error("Failed to migrate", "error", err.Error())
		return nil, err
	}

	dbInstance = &service{db: pool, logger: logger}
	return dbInstance, nil
}

func (s *service) GetDB() *pgxpool.Pool {
	return s.db
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stats := make(map[string]string)
	err := s.db.Ping(ctx)
	if err != nil {
		s.logger.Error("pinging the database", "error", err)
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		return stats
	}

	stats["status"] = "up"
	stats["message"] = "It's healthy"
	return stats
}

func (s *service) Close() {
	s.db.Close()
}
