package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/AGX18/real_estate_crm/internal/httpx"
)

type Handler struct {
	service *Service
}

type loginRequest struct {
	TenantName string `json:"tenant_name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body loginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.TenantName == "" || body.Email == "" || body.Password == "" {
		httpx.WriteError(w, http.StatusBadRequest, "tenant_name, email and password are required")
		return
	}

	result, err := h.service.Login(r.Context(), LoginParams{
		TenantName: body.TenantName,
		Email:      body.Email,
		Password:   body.Password,
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
