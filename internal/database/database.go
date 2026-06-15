package database

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"real_estate_crm/internal/db/migrations"

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
	database   = os.Getenv("BLUEPRINT_DB_DATABASE")
	password   = os.Getenv("BLUEPRINT_DB_PASSWORD")
	username   = os.Getenv("BLUEPRINT_DB_USERNAME")
	port       = os.Getenv("BLUEPRINT_DB_PORT")
	host       = os.Getenv("BLUEPRINT_DB_HOST")
	dbInstance *service
)

func New(logger *slog.Logger) Service {
	if dbInstance != nil {
		return dbInstance
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		username, password, host, port, database)
	var pool *pgxpool.Pool
	var err error
	db_url := os.Getenv("DATABASE_URL")
	if db_url != "" {
		pool, err = pgxpool.New(context.Background(), db_url)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		pool, err = pgxpool.New(context.Background(), connStr)
		if err != nil {
			log.Fatal(err)
		}
	}

	logger.Debug("created pgxpool", "connection string:", pool.Config().ConnString())

	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}
	sqlDB := stdlib.OpenDBFromPool(pool)

	if err := goose.Up(sqlDB, "."); err != nil {
		logger.Error("Failed to migrate", "error", err.Error())
		panic(err)
	}

	dbInstance = &service{db: pool, logger: logger}
	return dbInstance
}

func migrate() {

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
	log.Printf("Disconnected from database: %s", database)
	s.db.Close()
}
