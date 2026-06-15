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

	s.registerTenantRoutes(r)
	s.registerBrokerRoutes(r)

	return r
}

func (s *Server) registerTenantRoutes(r chi.Router) {
	r.Post("/tenants", s.CreateTenant)
	r.Get("/tenants", s.ListTenants)
	r.Get("/tenants/{id}", s.GetTenant)
	r.Patch("/tenants/{id}/status", s.UpdateTenantStatus)
	r.Delete("/tenants/{id}", s.DeleteTenant)
}

func (s *Server) registerBrokerRoutes(r chi.Router) {
	r.Post("/tenants/{tenant_id}/brokers", s.CreateBroker)
	r.Get("/tenants/{tenant_id}/brokers", s.ListBrokers)
	r.Get("/tenants/{tenant_id}/brokers/email/{email}", s.GetBrokerByEmail)
	r.Get("/tenants/{tenant_id}/brokers/{broker_id}", s.GetBroker)
	r.Patch("/tenants/{tenant_id}/brokers/{broker_id}/role", s.UpdateBrokerRole)
	r.Delete("/tenants/{tenant_id}/brokers/{broker_id}", s.DeleteBroker)
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
