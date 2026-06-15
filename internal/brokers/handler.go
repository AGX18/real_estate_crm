package brokers

import (
	"encoding/json"
	"net/http"

	db "real_estate_crm/internal/db/sqlc"
	"real_estate_crm/internal/httpx"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	service *Service
}

type createRequest struct {
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Role     db.Role `json:"role"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID, err := httpx.ParseUUID(chi.URLParam(r, "tenant_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid tenant id")
		return
	}

	var body createRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Username == "" || body.Email == "" || body.Password == "" {
		httpx.WriteError(w, http.StatusBadRequest, "username, email and password are required")
		return
	}
	if body.Role == "" {
		body.Role = db.RoleUser
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create broker")
		return
	}

	broker, err := h.service.Create(r.Context(), db.CreateBrokerParams{
		TenantID:     tenantID,
		Username:     body.Username,
		Email:        body.Email,
		PasswordHash: string(passwordHash),
		Role:         body.Role,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create broker")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, broker)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID, err := httpx.ParseUUID(chi.URLParam(r, "tenant_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid tenant id")
		return
	}

	brokerID, err := httpx.ParseInt64(chi.URLParam(r, "broker_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid broker id")
		return
	}

	broker, err := h.service.Get(r.Context(), db.GetBrokerByIDParams{
		ID:       brokerID,
		TenantID: tenantID,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "broker not found")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, broker)
}

func (h *Handler) GetByEmail(w http.ResponseWriter, r *http.Request) {
	tenantID, err := httpx.ParseUUID(chi.URLParam(r, "tenant_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid tenant id")
		return
	}

	email := chi.URLParam(r, "email")
	if email == "" {
		httpx.WriteError(w, http.StatusBadRequest, "invalid email")
		return
	}

	broker, err := h.service.GetByEmail(r.Context(), db.GetBrokerByEmailParams{
		Email:    email,
		TenantID: tenantID,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "broker not found")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, broker)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, err := httpx.ParseUUID(chi.URLParam(r, "tenant_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid tenant id")
		return
	}

	brokers, err := h.service.List(r.Context(), tenantID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch brokers")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, brokers)
}

func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	tenantID, err := httpx.ParseUUID(chi.URLParam(r, "tenant_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid tenant id")
		return
	}

	brokerID, err := httpx.ParseInt64(chi.URLParam(r, "broker_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid broker id")
		return
	}

	var params db.UpdateBrokerRoleParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	params.ID = brokerID
	params.TenantID = tenantID

	broker, err := h.service.UpdateRole(r.Context(), params)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update broker")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, broker)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID, err := httpx.ParseUUID(chi.URLParam(r, "tenant_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid tenant id")
		return
	}

	brokerID, err := httpx.ParseInt64(chi.URLParam(r, "broker_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid broker id")
		return
	}

	err = h.service.Delete(r.Context(), db.DeleteBrokerParams{
		ID:       brokerID,
		TenantID: tenantID,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete broker")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "broker deleted"})
}
