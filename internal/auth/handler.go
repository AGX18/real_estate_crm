package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"real_estate_crm/internal/httpx"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	tenantID, err := httpx.ParseUUID(chi.URLParam(r, "tenant_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid tenant id")
		return
	}

	var body loginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Email == "" || body.Password == "" {
		httpx.WriteError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	result, err := h.service.Login(r.Context(), LoginParams{
		TenantID: tenantID,
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			httpx.WriteError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to login")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result)
}
