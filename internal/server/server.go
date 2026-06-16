package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/AGX18/real_estate_crm/internal/auth"
	"github.com/AGX18/real_estate_crm/internal/brokers"
	"github.com/AGX18/real_estate_crm/internal/config"
	"github.com/AGX18/real_estate_crm/internal/database"
	"github.com/AGX18/real_estate_crm/internal/embeddings"
	"github.com/AGX18/real_estate_crm/internal/leads"
	"github.com/AGX18/real_estate_crm/internal/logger"
	"github.com/AGX18/real_estate_crm/internal/properties"
	appstore "github.com/AGX18/real_estate_crm/internal/store"
	"github.com/AGX18/real_estate_crm/internal/tenants"
	"github.com/AGX18/real_estate_crm/internal/voice"
)

type Server struct {
	port            int
	tenantService   *tenants.Service
	brokerService   *brokers.Service
	authService     *auth.Service
	leadService     *leads.Service
	propertyService *properties.Service
	voiceService    *voice.Service
	tokenManager    *auth.TokenManager
	tenantHandler   *tenants.Handler
	brokerHandler   *brokers.Handler
	authHandler     *auth.Handler
	leadHandler     *leads.Handler
	propertyHandler *properties.Handler
	voiceHandler    *voice.Handler
	db              database.Service
	logger          *slog.Logger
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
	store := appstore.New(dbService.GetDB())
	queries := store.Queries()

	tenantService := tenants.NewService(queries)
	brokerService := brokers.NewService(queries)
	tokenManager := auth.NewTokenManager(cfg.Auth.TokenSecret)
	authService := auth.NewServiceWithStore(store, queries, tokenManager)
	leadService := leads.NewService(queries)
	embedder := embeddings.NewOpenAIEmbedder(cfg.Embeddings.APIKey, cfg.Embeddings.Model, cfg.Embeddings.Dimensions)
	propertyService := properties.NewServiceWithStore(store, embedder)
	voiceService := voice.NewServiceWithStore(store)

	NewServer := &Server{
		port:            cfg.Port,
		db:              dbService,
		tenantService:   tenantService,
		brokerService:   brokerService,
		authService:     authService,
		leadService:     leadService,
		propertyService: propertyService,
		voiceService:    voiceService,
		tokenManager:    tokenManager,
		tenantHandler:   tenants.NewHandler(tenantService),
		brokerHandler:   brokers.NewHandler(brokerService),
		authHandler:     auth.NewHandler(authService),
		leadHandler:     leads.NewHandler(leadService),
		propertyHandler: properties.NewHandler(propertyService),
		voiceHandler:    voice.NewHandler(voiceService),
		logger:          logger.Log,
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
