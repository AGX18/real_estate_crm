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

type registerRequest struct {
	TenantName    string `json:"tenant_name"`
	AdminUsername string `json:"admin_username"`
	AdminEmail    string `json:"admin_email"`
	AdminPassword string `json:"admin_password"`
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

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var body registerRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.TenantName == "" || body.AdminUsername == "" || body.AdminEmail == "" || body.AdminPassword == "" {
		httpx.WriteError(w, http.StatusBadRequest, "tenant_name, admin_username, admin_email and admin_password are required")
		return
	}

	result, err := h.service.Register(r.Context(), RegisterParams{
		TenantName:    body.TenantName,
		AdminUsername: body.AdminUsername,
		AdminEmail:    body.AdminEmail,
		AdminPassword: body.AdminPassword,
	})
	if err != nil {
		if errors.Is(err, ErrRegistrationUnavailable) {
			httpx.WriteError(w, http.StatusInternalServerError, "registration is unavailable")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to register tenant")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, result)
}
