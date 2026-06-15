package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"real_estate_crm/internal/database"
	db "real_estate_crm/internal/db/sqlc"
	"real_estate_crm/internal/logger"
)

type Server struct {
	port    int
	queries db.Querier
	db      database.Service
	logger  *slog.Logger
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	logger.Init("debug")
	dbService := database.New(logger.Log)

	NewServer := &Server{
		port:    port,
		db:      dbService,
		queries: db.New(dbService.GetDB()),
		logger:  logger.Log,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
