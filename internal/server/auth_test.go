package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AGX18/real_estate_crm/internal/auth"
	db "github.com/AGX18/real_estate_crm/internal/db/sqlc"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
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

	var result auth.LoginResult
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	if result.Token == "" {
		t.Fatal("expected token in login response")
	}
	claims := &auth.Claims{}
	token, err := jwt.ParseWithClaims(result.Token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	if err != nil || !token.Valid {
		t.Fatalf("expected valid jwt, token valid=%v err=%v", token.Valid, err)
	}
	if claims.BrokerID != 1 {
		t.Fatalf("expected broker id claim 1 got %d", claims.BrokerID)
	}
	if claims.Email != "agent@example.com" {
		t.Fatalf("expected email claim agent@example.com got %q", claims.Email)
	}
	if claims.Role != db.RoleUser {
		t.Fatalf("expected role claim user got %q", claims.Role)
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
