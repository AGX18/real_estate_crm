package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"real_estate_crm/internal/brokers"
	"real_estate_crm/internal/config"
	"real_estate_crm/internal/database"
	db "real_estate_crm/internal/db/sqlc"
	"real_estate_crm/internal/logger"
	"real_estate_crm/internal/tenants"
)

type Server struct {
	port          int
	tenantService *tenants.Service
	brokerService *brokers.Service
	tenantHandler *tenants.Handler
	brokerHandler *brokers.Handler
	db            database.Service
	logger        *slog.Logger
}

func NewServer() (*http.Server, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	logger.Init(cfg.LogLevel)
	dbService, err := database.New(logger.Log, cfg.Database)
	if err != nil {
		return nil, err
	}
	queries := db.New(dbService.GetDB())

	tenantService := tenants.NewService(queries)
	brokerService := brokers.NewService(queries)

	NewServer := &Server{
		port:          cfg.Port,
		db:            dbService,
		tenantService: tenantService,
		brokerService: brokerService,
		tenantHandler: tenants.NewHandler(tenantService),
		brokerHandler: brokers.NewHandler(brokerService),
		logger:        logger.Log,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server, nil
}
