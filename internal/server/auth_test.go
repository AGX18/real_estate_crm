package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	db "real_estate_crm/internal/db/sqlc"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

func newAuthTestRouter(s *Server) http.Handler {
	r := chi.NewRouter()
	s.registerAuthRoutes(r)
	return r
}

func TestLogin(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	mock := &mockQueries{
		broker: db.Broker{
			ID:           1,
			Username:     "agent",
			Email:        "agent@example.com",
			PasswordHash: string(passwordHash),
			Role:         db.RoleUser,
		},
	}
	s := newTestServer(mock)
	r := newAuthTestRouter(s)

	body := bytes.NewBufferString(`{"email":"agent@example.com","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/tenants/"+testTenantID+"/login", body)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected %d got %d", http.StatusOK, w.Code)
	}
}

func TestLoginInvalidPassword(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	mock := &mockQueries{
		broker: db.Broker{
			ID:           1,
			Username:     "agent",
			Email:        "agent@example.com",
			PasswordHash: string(passwordHash),
			Role:         db.RoleUser,
		},
	}
	s := newTestServer(mock)
	r := newAuthTestRouter(s)

	body := bytes.NewBufferString(`{"email":"agent@example.com","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/tenants/"+testTenantID+"/login", body)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected %d got %d", http.StatusUnauthorized, w.Code)
	}
}
