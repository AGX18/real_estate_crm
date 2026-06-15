package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"real_estate_crm/internal/auth"
	"real_estate_crm/internal/brokers"
	"real_estate_crm/internal/config"
	"real_estate_crm/internal/database"
	"real_estate_crm/internal/embeddings"
	"real_estate_crm/internal/leads"
	"real_estate_crm/internal/logger"
	"real_estate_crm/internal/properties"
	appstore "real_estate_crm/internal/store"
	"real_estate_crm/internal/tenants"
	"real_estate_crm/internal/voice"
)

type Server struct {
	port            int
	tenantService   *tenants.Service
	brokerService   *brokers.Service
	authService     *auth.Service
	leadService     *leads.Service
	propertyService *properties.Service
	voiceService    *voice.Service
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
	authService := auth.NewService(queries, auth.NewTokenManager(cfg.Auth.TokenSecret))
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
