package server

import (
	"encoding/json"
	"net/http"

	db "real_estate_crm/internal/db/sqlc"

	"github.com/go-chi/chi/v5"
)

func (s *Server) CreateBroker(w http.ResponseWriter, r *http.Request) {
	tenantID, err := parseUUID(chi.URLParam(r, "tenant_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant id")
		return
	}

	var params db.CreateBrokerParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	params.TenantID = tenantID

	broker, err := s.queries.CreateBroker(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create broker")
		return
	}

	writeJSON(w, http.StatusCreated, broker)
}

func (s *Server) GetBroker(w http.ResponseWriter, r *http.Request) {
	tenantID, err := parseUUID(chi.URLParam(r, "tenant_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant id")
		return
	}

	brokerID, err := parseInt64(chi.URLParam(r, "broker_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid broker id")
		return
	}

	broker, err := s.queries.GetBrokerByID(r.Context(), db.GetBrokerByIDParams{
		ID:       brokerID,
		TenantID: tenantID,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, "broker not found")
		return
	}

	writeJSON(w, http.StatusOK, broker)
}

func (s *Server) GetBrokerByEmail(w http.ResponseWriter, r *http.Request) {
	tenantID, err := parseUUID(chi.URLParam(r, "tenant_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant id")
		return
	}

	email := chi.URLParam(r, "email")
	if email == "" {
		writeError(w, http.StatusBadRequest, "invalid email")
		return
	}

	broker, err := s.queries.GetBrokerByEmail(r.Context(), db.GetBrokerByEmailParams{
		Email:    email,
		TenantID: tenantID,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, "broker not found")
		return
	}

	writeJSON(w, http.StatusOK, broker)
}

func (s *Server) ListBrokers(w http.ResponseWriter, r *http.Request) {
	tenantID, err := parseUUID(chi.URLParam(r, "tenant_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant id")
		return
	}

	brokers, err := s.queries.ListBrokers(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch brokers")
		return
	}

	writeJSON(w, http.StatusOK, brokers)
}

func (s *Server) UpdateBrokerRole(w http.ResponseWriter, r *http.Request) {
	tenantID, err := parseUUID(chi.URLParam(r, "tenant_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant id")
		return
	}

	brokerID, err := parseInt64(chi.URLParam(r, "broker_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid broker id")
		return
	}

	var params db.UpdateBrokerRoleParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	params.ID = brokerID
	params.TenantID = tenantID

	broker, err := s.queries.UpdateBrokerRole(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update broker")
		return
	}

	writeJSON(w, http.StatusOK, broker)
}

func (s *Server) DeleteBroker(w http.ResponseWriter, r *http.Request) {
	tenantID, err := parseUUID(chi.URLParam(r, "tenant_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant id")
		return
	}

	brokerID, err := parseInt64(chi.URLParam(r, "broker_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid broker id")
		return
	}

	err = s.queries.DeleteBroker(r.Context(), db.DeleteBrokerParams{
		ID:       brokerID,
		TenantID: tenantID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete broker")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "broker deleted"})
}
