package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	db "real_estate_crm/internal/db/sqlc"

	"github.com/go-chi/chi/v5"
)

func newLeadTestRouter(s *Server) http.Handler {
	r := chi.NewRouter()
	s.registerLeadRoutes(r)
	return r
}

func TestCreateLead(t *testing.T) {
	mock := &mockQueries{lead: db.Lead{ID: 1, Phone: "+201001234567"}}
	s := newTestServer(mock)
	r := newLeadTestRouter(s)

	body := bytes.NewBufferString(`{"phone":"+201001234567","description":"Looking for apartment","status":"Follow_Up"}`)
	req := httptest.NewRequest(http.MethodPost, "/tenants/"+testTenantID+"/leads", body)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected %d got %d", http.StatusCreated, w.Code)
	}
}

func TestGetLead(t *testing.T) {
	mock := &mockQueries{lead: db.Lead{ID: 1, Phone: "+201001234567"}}
	s := newTestServer(mock)
	r := newLeadTestRouter(s)

	req := httptest.NewRequest(http.MethodGet, "/tenants/"+testTenantID+"/leads/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected %d got %d", http.StatusOK, w.Code)
	}
}

func TestListLeads(t *testing.T) {
	mock := &mockQueries{
		leads: []db.Lead{
			{ID: 1, Phone: "+201001234567"},
			{ID: 2, Phone: "+201009876543"},
		},
	}
	s := newTestServer(mock)
	r := newLeadTestRouter(s)

	req := httptest.NewRequest(http.MethodGet, "/tenants/"+testTenantID+"/leads", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected %d got %d", http.StatusOK, w.Code)
	}
}

func TestUpdateLeadStatus(t *testing.T) {
	mock := &mockQueries{
		lead: db.Lead{
			ID:     1,
			Phone:  "+201001234567",
			Status: db.NullLeadStatus{LeadStatus: db.LeadStatusQualified, Valid: true},
		},
	}
	s := newTestServer(mock)
	r := newLeadTestRouter(s)

	body := bytes.NewBufferString(`{"status":"qualified"}`)
	req := httptest.NewRequest(http.MethodPatch, "/tenants/"+testTenantID+"/leads/1/status", body)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected %d got %d", http.StatusOK, w.Code)
	}
}

func TestDeleteLead(t *testing.T) {
	mock := &mockQueries{}
	s := newTestServer(mock)
	r := newLeadTestRouter(s)

	req := httptest.NewRequest(http.MethodDelete, "/tenants/"+testTenantID+"/leads/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected %d got %d", http.StatusOK, w.Code)
	}
}
