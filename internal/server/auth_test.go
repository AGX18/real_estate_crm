package server

import (
	"bytes"
	"context"
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

type fakeTxStore struct {
	q db.Querier
}

func (s fakeTxStore) WithTx(ctx context.Context, fn func(q db.Querier) error) error {
	return fn(s.q)
}

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
		tenant: db.Tenant{ID: mustParseUUID(t, testTenantID), Name: "acme"},
		broker: db.Broker{
			ID:           1,
			TenantID:     mustParseUUID(t, testTenantID),
			Username:     "agent",
			Email:        "agent@example.com",
			PasswordHash: string(passwordHash),
			Role:         db.RoleUser,
		},
	}
	s := newTestServer(mock)
	r := newAuthTestRouter(s)

	body := bytes.NewBufferString(`{"tenant_name":"acme","email":"agent@example.com","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
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
		tenant: db.Tenant{ID: mustParseUUID(t, testTenantID), Name: "acme"},
		broker: db.Broker{
			ID:           1,
			TenantID:     mustParseUUID(t, testTenantID),
			Username:     "agent",
			Email:        "agent@example.com",
			PasswordHash: string(passwordHash),
			Role:         db.RoleUser,
		},
	}
	s := newTestServer(mock)
	r := newAuthTestRouter(s)

	body := bytes.NewBufferString(`{"tenant_name":"acme","email":"agent@example.com","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected %d got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRegisterCreatesTenantAndAdmin(t *testing.T) {
	tenantID := mustParseUUID(t, testTenantID)
	mock := &mockQueries{
		tenant: db.Tenant{ID: tenantID, Name: "acme"},
		broker: db.Broker{
			ID:       1,
			TenantID: tenantID,
			Username: "owner",
			Email:    "owner@example.com",
			Role:     db.RoleAdmin,
		},
	}
	s := newTestServer(mock)
	authService := auth.NewServiceWithStore(fakeTxStore{q: mock}, mock, s.tokenManager)
	s.authHandler = auth.NewHandler(authService)
	r := newAuthTestRouter(s)

	body := bytes.NewBufferString(`{"tenant_name":"acme","admin_username":"owner","admin_email":"owner@example.com","admin_password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/register", body)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected %d got %d body %s", http.StatusCreated, w.Code, w.Body.String())
	}
	if mock.createBrokerArg.Role != db.RoleAdmin {
		t.Fatalf("expected first broker role admin got %q", mock.createBrokerArg.Role)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(mock.createBrokerArg.PasswordHash), []byte("secret")); err != nil {
		t.Fatalf("expected admin password hash to match submitted password: %v", err)
	}

	var result auth.LoginResult
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode register response: %v", err)
	}
	if result.Token == "" {
		t.Fatal("expected token in register response")
	}
	if result.Broker.Role != db.RoleAdmin {
		t.Fatalf("expected admin broker in response got %q", result.Broker.Role)
	}
}
