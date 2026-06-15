package leads

import (
	"encoding/json"
	"net/http"

	db "github.com/AGX18/real_estate_crm/internal/db/sqlc"
	"github.com/AGX18/real_estate_crm/internal/httpx"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Handler struct {
	service *Service
}

type leadRequest struct {
	Phone       string        `json:"phone"`
	Description string        `json:"description"`
	Status      db.LeadStatus `json:"status"`
}

type leadDescriptionRequest struct {
	Description string `json:"description"`
}

type leadStatusRequest struct {
	Status db.LeadStatus `json:"status"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseTenantID(w, r)
	if !ok {
		return
	}

	var body leadRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Phone == "" {
		httpx.WriteError(w, http.StatusBadRequest, "phone is required")
		return
	}
	if body.Status == "" {
		body.Status = db.LeadStatusFollowUp
	}

	lead, err := h.service.Create(r.Context(), db.CreateLeadParams{
		TenantID:    tenantID,
		Phone:       body.Phone,
		Description: textParam(body.Description),
		Status:      leadStatusParam(body.Status),
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create lead")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, lead)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID, leadID, ok := parseTenantAndLeadID(w, r)
	if !ok {
		return
	}

	lead, err := h.service.Get(r.Context(), db.GetLeadByIDParams{
		ID:       leadID,
		TenantID: tenantID,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "lead not found")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, lead)
}

func (h *Handler) GetByPhone(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseTenantID(w, r)
	if !ok {
		return
	}

	phone := chi.URLParam(r, "phone")
	if phone == "" {
		httpx.WriteError(w, http.StatusBadRequest, "invalid phone")
		return
	}

	lead, err := h.service.GetByPhone(r.Context(), db.GetLeadByPhoneParams{
		Phone:    phone,
		TenantID: tenantID,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "lead not found")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, lead)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseTenantID(w, r)
	if !ok {
		return
	}

	leads, err := h.service.List(r.Context(), tenantID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch leads")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, leads)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID, leadID, ok := parseTenantAndLeadID(w, r)
	if !ok {
		return
	}

	var body leadRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Phone == "" {
		httpx.WriteError(w, http.StatusBadRequest, "phone is required")
		return
	}
	if body.Status == "" {
		body.Status = db.LeadStatusFollowUp
	}

	lead, err := h.service.Update(r.Context(), db.UpdateLeadParams{
		ID:          leadID,
		Phone:       body.Phone,
		Description: textParam(body.Description),
		Status:      leadStatusParam(body.Status),
		TenantID:    tenantID,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update lead")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, lead)
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	tenantID, leadID, ok := parseTenantAndLeadID(w, r)
	if !ok {
		return
	}

	var body leadStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Status == "" {
		httpx.WriteError(w, http.StatusBadRequest, "status is required")
		return
	}

	lead, err := h.service.UpdateStatus(r.Context(), db.UpdateLeadStatusParams{
		ID:       leadID,
		Status:   leadStatusParam(body.Status),
		TenantID: tenantID,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update lead")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, lead)
}

func (h *Handler) UpdateDescription(w http.ResponseWriter, r *http.Request) {
	tenantID, leadID, ok := parseTenantAndLeadID(w, r)
	if !ok {
		return
	}

	var body leadDescriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	lead, err := h.service.UpdateDescription(r.Context(), db.UpdateLeadDescriptionParams{
		ID:          leadID,
		Description: textParam(body.Description),
		TenantID:    tenantID,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update lead")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, lead)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID, leadID, ok := parseTenantAndLeadID(w, r)
	if !ok {
		return
	}

	err := h.service.Delete(r.Context(), db.DeleteLeadParams{
		ID:       leadID,
		TenantID: tenantID,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete lead")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "lead deleted"})
}

func parseTenantID(w http.ResponseWriter, r *http.Request) (pgtype.UUID, bool) {
	tenantID, err := httpx.ParseUUID(chi.URLParam(r, "tenant_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid tenant id")
		return pgtype.UUID{}, false
	}
	return tenantID, true
}

func parseTenantAndLeadID(w http.ResponseWriter, r *http.Request) (pgtype.UUID, int64, bool) {
	tenantID, ok := parseTenantID(w, r)
	if !ok {
		return pgtype.UUID{}, 0, false
	}

	leadID, err := httpx.ParseInt64(chi.URLParam(r, "lead_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid lead id")
		return pgtype.UUID{}, 0, false
	}

	return tenantID, leadID, true
}

func textParam(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: value != ""}
}

func leadStatusParam(value db.LeadStatus) db.NullLeadStatus {
	return db.NullLeadStatus{LeadStatus: value, Valid: value != ""}
}
