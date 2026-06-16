package server

import (
	"net/http"
	"strings"

	"github.com/AGX18/real_estate_crm/internal/db/sqlc"
	"github.com/AGX18/real_estate_crm/internal/httpx"

	"github.com/go-chi/chi/v5"
)

func (s *Server) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.tokenManager == nil {
			httpx.WriteError(w, http.StatusInternalServerError, "auth unavailable")
			return
		}

		header := r.Header.Get("Authorization")
		tokenString, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || strings.TrimSpace(tokenString) == "" {
			httpx.WriteError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}

		claims, err := s.tokenManager.Verify(strings.TrimSpace(tokenString))
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "invalid bearer token")
			return
		}
		if claims.Role != db.RoleAdmin {
			httpx.WriteError(w, http.StatusForbidden, "admin role required")
			return
		}

		tenantID := chi.URLParam(r, "tenant_id")
		if tenantID != "" && tenantID != claims.TenantID {
			httpx.WriteError(w, http.StatusForbidden, "tenant access denied")
			return
		}

		next.ServeHTTP(w, r)
	})
}
