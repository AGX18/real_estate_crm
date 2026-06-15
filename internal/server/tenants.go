package server

import (
	"encoding/json"
	"net/http"

	db "real_estate_crm/internal/db/sqlc"

	"github.com/go-chi/chi/v5"
)

func (s *Server) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var params db.CreateTenantParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenant, err := s.queries.CreateTenant(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create tenant")
		return
	}

	writeJSON(w, http.StatusCreated, tenant)
}

func (s *Server) GetTenant(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	tenant, err := s.queries.GetTenant(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}

	writeJSON(w, http.StatusOK, tenant)
}

func (s *Server) ListTenants(w http.ResponseWriter, r *http.Request) {
	tenants, err := s.queries.ListTenants(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch tenants")
		return
	}

	writeJSON(w, http.StatusOK, tenants)
}

func (s *Server) UpdateTenantStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var params db.UpdateTenantStatusParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	params.ID = id

	tenant, err := s.queries.UpdateTenantStatus(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update tenant")
		return
	}

	writeJSON(w, http.StatusOK, tenant)
}

func (s *Server) DeleteTenant(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	err = s.queries.DeleteTenant(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete tenant")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "tenant deleted"})
}
