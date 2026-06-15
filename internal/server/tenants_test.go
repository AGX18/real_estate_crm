package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AGX18/real_estate_crm/internal/auth"
	"github.com/AGX18/real_estate_crm/internal/brokers"
	db "github.com/AGX18/real_estate_crm/internal/db/sqlc"
	"github.com/AGX18/real_estate_crm/internal/leads"
	"github.com/AGX18/real_estate_crm/internal/properties"
	"github.com/AGX18/real_estate_crm/internal/tenants"
	"github.com/AGX18/real_estate_crm/internal/voice"

	"github.com/go-chi/chi/v5"
)

func newTestServer(q db.Querier) *Server {
	tenantService := tenants.NewService(q)
	brokerService := brokers.NewService(q)
	authService := auth.NewService(q, auth.NewTokenManager("test-secret"))
	leadService := leads.NewService(q)
	propertyService := properties.NewService(q)
	voiceService := voice.NewService(q)
	s := &Server{
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
	}
	return s
}

func newTestRouter(s *Server) http.Handler {
	r := chi.NewRouter()
	s.registerTenantRoutes(r)
	return r
}
func TestGetTenant(t *testing.T) {
	mock := &mockQueries{tenant: db.Tenant{Name: "Test Agency"}}
	s := newTestServer(mock)
	r := newTestRouter(s)
	req := httptest.NewRequest(http.MethodGet, "/tenants/550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected %d got %d", http.StatusOK, w.Code)
	}
}

func TestListTenants(t *testing.T) {
	mock := &mockQueries{
		tenants: []db.Tenant{
			{Name: "Agency A"},
			{Name: "Agency B"},
		},
	}
	s := newTestServer(mock)
	req := httptest.NewRequest(http.MethodGet, "/tenants", nil)
	w := httptest.NewRecorder()
	s.tenantHandler.List(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected %d got %d", http.StatusOK, w.Code)
	}
}

func TestDeleteTenant(t *testing.T) {
	mock := &mockQueries{}
	s := newTestServer(mock)
	r := newTestRouter(s)
	req := httptest.NewRequest(http.MethodDelete, "/tenants/550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected %d got %d", http.StatusOK, w.Code)
	}
}
