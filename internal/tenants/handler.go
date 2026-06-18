package tenants

import (
	"encoding/json"
	"net/http"
	"strings"

	db "github.com/AGX18/real_estate_crm/internal/db/sqlc"
	"github.com/AGX18/real_estate_crm/internal/httpx"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var params db.CreateTenantParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenant, err := h.service.Create(r.Context(), params)
	tenant.Name = strings.ToLower(tenant.Name)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create tenant")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, tenant)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	tenant, err := h.service.Get(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "tenant not found")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, tenant)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	tenants, err := h.service.List(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch tenants")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, tenants)
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var params db.UpdateTenantStatusParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	params.ID = id

	tenant, err := h.service.UpdateStatus(r.Context(), params)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update tenant")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, tenant)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	err = h.service.Delete(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete tenant")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "tenant deleted"})
}
