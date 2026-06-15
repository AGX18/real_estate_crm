package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	db "github.com/AGX18/real_estate_crm/internal/db/sqlc"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

const testTenantID = "550e8400-e29b-41d4-a716-446655440000"

func newBrokerTestRouter(s *Server) http.Handler {
	r := chi.NewRouter()
	s.registerBrokerRoutes(r)
	return r
}

func TestCreateBroker(t *testing.T) {
	mock := &mockQueries{broker: db.Broker{ID: 1, Username: "agent", Email: "agent@example.com", Role: db.RoleUser}}
	s := newTestServer(mock)
	r := newBrokerTestRouter(s)

	body := bytes.NewBufferString(`{"username":"agent","email":"agent@example.com","password":"secret","role":"user"}`)
	req := httptest.NewRequest(http.MethodPost, "/tenants/"+testTenantID+"/brokers", body)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected %d got %d", http.StatusCreated, w.Code)
	}
	if mock.createBrokerArg.PasswordHash == "" {
		t.Fatal("expected password hash to be stored")
	}
	if mock.createBrokerArg.PasswordHash == "secret" {
		t.Fatal("expected stored password to be hashed, not plaintext")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(mock.createBrokerArg.PasswordHash), []byte("secret")); err != nil {
		t.Fatalf("expected password hash to match submitted password: %v", err)
	}
}

func TestGetBroker(t *testing.T) {
	mock := &mockQueries{broker: db.Broker{ID: 1, Username: "agent", Email: "agent@example.com", Role: db.RoleUser}}
	s := newTestServer(mock)
	r := newBrokerTestRouter(s)

	req := httptest.NewRequest(http.MethodGet, "/tenants/"+testTenantID+"/brokers/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected %d got %d", http.StatusOK, w.Code)
	}
}

func TestListBrokers(t *testing.T) {
	mock := &mockQueries{
		brokers: []db.Broker{
			{ID: 1, Username: "agent-a", Email: "a@example.com", Role: db.RoleUser},
			{ID: 2, Username: "agent-b", Email: "b@example.com", Role: db.RoleAdmin},
		},
	}
	s := newTestServer(mock)
	r := newBrokerTestRouter(s)

	req := httptest.NewRequest(http.MethodGet, "/tenants/"+testTenantID+"/brokers", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected %d got %d", http.StatusOK, w.Code)
	}
}

func TestUpdateBrokerRole(t *testing.T) {
	mock := &mockQueries{broker: db.Broker{ID: 1, Username: "agent", Email: "agent@example.com", Role: db.RoleAdmin}}
	s := newTestServer(mock)
	r := newBrokerTestRouter(s)

	body := bytes.NewBufferString(`{"role":"admin"}`)
	req := httptest.NewRequest(http.MethodPatch, "/tenants/"+testTenantID+"/brokers/1/role", body)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected %d got %d", http.StatusOK, w.Code)
	}
}

func TestDeleteBroker(t *testing.T) {
	mock := &mockQueries{}
	s := newTestServer(mock)
	r := newBrokerTestRouter(s)

	req := httptest.NewRequest(http.MethodDelete, "/tenants/"+testTenantID+"/brokers/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected %d got %d", http.StatusOK, w.Code)
	}
}
