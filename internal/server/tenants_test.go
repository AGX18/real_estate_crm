package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	db "real_estate_crm/internal/db/sqlc"

	"github.com/go-chi/chi/v5"
)

func newTestServer(q db.Querier) *Server {
	s := &Server{queries: q}
	return s
}

func newTestRouter(s *Server) http.Handler {
	r := chi.NewRouter()
	r.Post("/tenants", s.CreateTenant)
	r.Get("/tenants", s.ListTenants)
	r.Get("/tenants/{id}", s.GetTenant)
	r.Patch("/tenants/{id}/status", s.UpdateTenantStatus)
	r.Delete("/tenants/{id}", s.DeleteTenant)
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
	s.ListTenants(w, req)
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
