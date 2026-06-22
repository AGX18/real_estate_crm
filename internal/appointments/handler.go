package appointments

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	db "github.com/AGX18/real_estate_crm/internal/db/sqlc"
	"github.com/AGX18/real_estate_crm/internal/httpx"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Handler struct {
	service *Service
}

type appointmentRequest struct {
	LeadID int64                `json:"lead_id"`
	Title  string               `json:"title"`
	Notes  string               `json:"notes"`
	Status db.AppointmentStatus `json:"status"`
	Date   string               `json:"date"`
	Day    string               `json:"day"`
	Time   string               `json:"time"`
}

type appointmentStatusRequest struct {
	Status db.AppointmentStatus `json:"status"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseTenantID(w, r)
	if !ok {
		return
	}

	var body appointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	params, err := createParams(tenantID, body)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	appointment, err := h.service.Create(r.Context(), params)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create appointment")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, appointment)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID, appointmentID, ok := parseTenantAndAppointmentID(w, r)
	if !ok {
		return
	}

	appointment, err := h.service.Get(r.Context(), db.GetAppointmentByIDParams{
		ID:       appointmentID,
		TenantID: tenantID,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "appointment not found")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, appointment)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseTenantID(w, r)
	if !ok {
		return
	}

	leadIDValue := r.URL.Query().Get("lead_id")
	if leadIDValue != "" {
		leadID, err := httpx.ParseInt64(leadIDValue)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid lead id")
			return
		}
		appointments, err := h.service.ListByLead(r.Context(), db.ListAppointmentsByLeadParams{
			LeadID:   leadID,
			TenantID: tenantID,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch appointments")
			return
		}
		httpx.WriteJSON(w, http.StatusOK, appointments)
		return
	}

	appointments, err := h.service.List(r.Context(), tenantID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch appointments")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, appointments)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID, appointmentID, ok := parseTenantAndAppointmentID(w, r)
	if !ok {
		return
	}

	var body appointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	params, err := updateParams(tenantID, appointmentID, body)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	appointment, err := h.service.Update(r.Context(), params)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update appointment")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, appointment)
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	tenantID, appointmentID, ok := parseTenantAndAppointmentID(w, r)
	if !ok {
		return
	}

	var body appointmentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if !validAppointmentStatus(body.Status) {
		httpx.WriteError(w, http.StatusBadRequest, "invalid appointment status")
		return
	}

	appointment, err := h.service.UpdateStatus(r.Context(), db.UpdateAppointmentStatusParams{
		ID:       appointmentID,
		Status:   appointmentStatusParam(body.Status),
		TenantID: tenantID,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update appointment")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, appointment)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID, appointmentID, ok := parseTenantAndAppointmentID(w, r)
	if !ok {
		return
	}

	if err := h.service.Delete(r.Context(), db.DeleteAppointmentParams{ID: appointmentID, TenantID: tenantID}); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete appointment")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "appointment deleted"})
}

func createParams(tenantID pgtype.UUID, body appointmentRequest) (db.CreateAppointmentParams, error) {
	if body.LeadID == 0 {
		return db.CreateAppointmentParams{}, fmt.Errorf("lead_id is required")
	}
	if strings.TrimSpace(body.Title) == "" {
		return db.CreateAppointmentParams{}, fmt.Errorf("title is required")
	}
	appointmentDate, appointmentTime, appointmentDay, err := parseAppointmentSchedule(body.Date, body.Day, body.Time)
	if err != nil {
		return db.CreateAppointmentParams{}, err
	}
	if body.Status == "" {
		body.Status = db.AppointmentStatusScheduled
	}
	if !validAppointmentStatus(body.Status) {
		return db.CreateAppointmentParams{}, fmt.Errorf("invalid appointment status")
	}

	return db.CreateAppointmentParams{
		TenantID:        tenantID,
		LeadID:          body.LeadID,
		Title:           strings.TrimSpace(body.Title),
		Notes:           textParam(body.Notes),
		Status:          appointmentStatusParam(body.Status),
		AppointmentDate: dateParam(appointmentDate),
		AppointmentDay:  appointmentDay,
		AppointmentTime: timeParam(appointmentTime),
	}, nil
}

func updateParams(tenantID pgtype.UUID, appointmentID int64, body appointmentRequest) (db.UpdateAppointmentParams, error) {
	if strings.TrimSpace(body.Title) == "" {
		return db.UpdateAppointmentParams{}, fmt.Errorf("title is required")
	}
	appointmentDate, appointmentTime, appointmentDay, err := parseAppointmentSchedule(body.Date, body.Day, body.Time)
	if err != nil {
		return db.UpdateAppointmentParams{}, err
	}
	if body.Status == "" {
		body.Status = db.AppointmentStatusScheduled
	}
	if !validAppointmentStatus(body.Status) {
		return db.UpdateAppointmentParams{}, fmt.Errorf("invalid appointment status")
	}

	return db.UpdateAppointmentParams{
		ID:              appointmentID,
		Title:           strings.TrimSpace(body.Title),
		Notes:           textParam(body.Notes),
		Status:          appointmentStatusParam(body.Status),
		AppointmentDate: dateParam(appointmentDate),
		AppointmentDay:  appointmentDay,
		AppointmentTime: timeParam(appointmentTime),
		TenantID:        tenantID,
	}, nil
}

func parseTenantID(w http.ResponseWriter, r *http.Request) (pgtype.UUID, bool) {
	tenantID, err := httpx.ParseUUID(chi.URLParam(r, "tenant_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid tenant id")
		return pgtype.UUID{}, false
	}
	return tenantID, true
}

func parseTenantAndAppointmentID(w http.ResponseWriter, r *http.Request) (pgtype.UUID, int64, bool) {
	tenantID, ok := parseTenantID(w, r)
	if !ok {
		return pgtype.UUID{}, 0, false
	}

	appointmentID, err := httpx.ParseInt64(chi.URLParam(r, "appointment_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid appointment id")
		return pgtype.UUID{}, 0, false
	}

	return tenantID, appointmentID, true
}

func parseAppointmentSchedule(dateValue string, dayValue string, timeValue string) (time.Time, time.Time, string, error) {
	dateValue = strings.TrimSpace(dateValue)
	dayValue = strings.TrimSpace(dayValue)
	timeValue = strings.TrimSpace(timeValue)
	if dateValue == "" {
		return time.Time{}, time.Time{}, "", fmt.Errorf("date is required")
	}
	if dayValue == "" {
		return time.Time{}, time.Time{}, "", fmt.Errorf("day is required")
	}
	if timeValue == "" {
		return time.Time{}, time.Time{}, "", fmt.Errorf("time is required")
	}

	appointmentDate, err := time.ParseInLocation("2006-01-02", dateValue, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, "", fmt.Errorf("date must be YYYY-MM-DD")
	}

	appointmentTime, err := parseAppointmentTime(timeValue)
	if err != nil {
		return time.Time{}, time.Time{}, "", err
	}
	if !weekdayMatches(dayValue, appointmentDate.Weekday()) {
		return time.Time{}, time.Time{}, "", fmt.Errorf("day must match date")
	}

	return appointmentDate, appointmentTime, dayValue, nil
}

func parseAppointmentTime(value string) (time.Time, error) {
	for _, layout := range []string{"15:04", "15:04:05", "3:04 PM", "3:04PM"} {
		parsed, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("time must be HH:MM")
}

func weekdayMatches(value string, weekday time.Weekday) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	full := strings.ToLower(weekday.String())
	return normalized == full || normalized == full[:3]
}

func validAppointmentStatus(value db.AppointmentStatus) bool {
	switch value {
	case db.AppointmentStatusScheduled, db.AppointmentStatusCompleted, db.AppointmentStatusCanceled, db.AppointmentStatusNoShow:
		return true
	default:
		return false
	}
}

func textParam(value string) pgtype.Text {
	value = strings.TrimSpace(value)
	return pgtype.Text{String: value, Valid: value != ""}
}

func int4Param(value int32) pgtype.Int4 {
	return pgtype.Int4{Int32: value, Valid: value > 0}
}

func dateParam(value time.Time) pgtype.Date {
	return pgtype.Date{Time: value, Valid: true}
}

func timeParam(value time.Time) pgtype.Time {
	microseconds := int64(value.Hour()) * int64(time.Hour/time.Microsecond)
	microseconds += int64(value.Minute()) * int64(time.Minute/time.Microsecond)
	microseconds += int64(value.Second()) * int64(time.Second/time.Microsecond)
	return pgtype.Time{Microseconds: microseconds, Valid: true}
}

func appointmentStatusParam(value db.AppointmentStatus) db.NullAppointmentStatus {
	return db.NullAppointmentStatus{AppointmentStatus: value, Valid: value != ""}
}
