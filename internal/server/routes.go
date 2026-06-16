package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/", s.HelloWorldHandler)

	r.Get("/health", s.healthHandler)

	s.registerAuthRoutes(r)
	s.registerTenantRoutes(r)
	s.registerBrokerRoutes(r)
	s.registerLeadRoutes(r)
	s.registerPropertyRoutes(r)
	s.registerVoiceRoutes(r)

	return r
}

func (s *Server) registerAuthRoutes(r chi.Router) {
	r.Post("/login", s.authHandler.Login)
}

func (s *Server) registerTenantRoutes(r chi.Router) {
	r.Post("/tenants", s.tenantHandler.Create)
	r.Get("/tenants", s.tenantHandler.List)
	r.Get("/tenants/{id}", s.tenantHandler.Get)
	r.Patch("/tenants/{id}/status", s.tenantHandler.UpdateStatus)
	r.Delete("/tenants/{id}", s.tenantHandler.Delete)
}

func (s *Server) registerBrokerRoutes(r chi.Router) {
	r.Post("/tenants/{tenant_id}/brokers", s.brokerHandler.Create)
	r.Get("/tenants/{tenant_id}/brokers", s.brokerHandler.List)
	r.Get("/tenants/{tenant_id}/brokers/{broker_id}", s.brokerHandler.Get)
	r.Patch("/tenants/{tenant_id}/brokers/{broker_id}/role", s.brokerHandler.UpdateRole)
	r.Delete("/tenants/{tenant_id}/brokers/{broker_id}", s.brokerHandler.Delete)
}

func (s *Server) registerLeadRoutes(r chi.Router) {
	r.Post("/tenants/{tenant_id}/leads", s.leadHandler.Create)
	r.Get("/tenants/{tenant_id}/leads", s.leadHandler.List)
	r.Get("/tenants/{tenant_id}/leads/phone/{phone}", s.leadHandler.GetByPhone)
	r.Get("/tenants/{tenant_id}/leads/{lead_id}", s.leadHandler.Get)
	r.Put("/tenants/{tenant_id}/leads/{lead_id}", s.leadHandler.Update)
	r.Patch("/tenants/{tenant_id}/leads/{lead_id}/status", s.leadHandler.UpdateStatus)
	r.Patch("/tenants/{tenant_id}/leads/{lead_id}/description", s.leadHandler.UpdateDescription)
	r.Delete("/tenants/{tenant_id}/leads/{lead_id}", s.leadHandler.Delete)
}

func (s *Server) registerPropertyRoutes(r chi.Router) {
	r.Post("/tenants/{tenant_id}/properties", s.propertyHandler.Create)
	r.Post("/tenants/{tenant_id}/properties/bulk", s.propertyHandler.CreateMany)
	r.Post("/tenants/{tenant_id}/properties/import", s.propertyHandler.Import)
	r.Get("/tenants/{tenant_id}/properties", s.propertyHandler.List)
	r.Post("/tenants/{tenant_id}/properties/search", s.propertyHandler.Search)
	r.Get("/tenants/{tenant_id}/properties/{property_id}", s.propertyHandler.Get)
	r.Put("/tenants/{tenant_id}/properties/{property_id}", s.propertyHandler.Update)
	r.Delete("/tenants/{tenant_id}/properties/{property_id}", s.propertyHandler.Delete)
}

func (s *Server) registerVoiceRoutes(r chi.Router) {
	r.Post("/tenants/{tenant_id}/calls", s.voiceHandler.CreateCall)
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(s.db.Health())
	_, _ = w.Write(jsonResp)
}
