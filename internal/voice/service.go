package voice

import (
	"context"
	"errors"
	"time"

	db "github.com/AGX18/real_estate_crm/internal/db/sqlc"
	"github.com/AGX18/real_estate_crm/internal/store"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	queries db.Querier
	store   txRunner
}

type txRunner interface {
	WithTx(ctx context.Context, fn func(q db.Querier) error) error
}

type CreateCallParams struct {
	TenantID     pgtype.UUID
	Phone        string
	Description  string
	Status       db.LeadStatus
	Transcript   string
	Details      string
	Summary      string
	Sentiment    db.CallSentiment
	Outcome      db.CallOutcome
	DurationSecs int32
	Appointment  *CreateAppointmentParams
}

type CreateCallResult struct {
	Lead        db.Lead         `json:"lead"`
	Call        db.Call         `json:"call"`
	Appointment *db.Appointment `json:"appointment,omitempty"`
}

type CreateAppointmentParams struct {
	Date time.Time
	Day  string
	Time time.Time
}

func NewService(queries db.Querier) *Service {
	return &Service{queries: queries}
}

func NewServiceWithStore(store *store.Store) *Service {
	return &Service{queries: store.Queries(), store: store}
}

func (s *Service) CreateCall(ctx context.Context, params CreateCallParams) (CreateCallResult, error) {
	if s.store == nil {
		return s.createCall(ctx, s.queries, params)
	}

	var result CreateCallResult
	err := s.store.WithTx(ctx, func(q db.Querier) error {
		var err error
		result, err = s.createCall(ctx, q, params)
		return err
	})
	return result, err
}

func (s *Service) ListCalls(ctx context.Context, tenantID pgtype.UUID) ([]db.Call, error) {
	return s.queries.ListCalls(ctx, tenantID)
}

func (s *Service) createCall(ctx context.Context, q db.Querier, params CreateCallParams) (CreateCallResult, error) {
	lead, err := q.GetLeadByPhone(ctx, db.GetLeadByPhoneParams{
		Phone:    params.Phone,
		TenantID: params.TenantID,
	})
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return CreateCallResult{}, err
		}

		lead, err = q.CreateLead(ctx, db.CreateLeadParams{
			TenantID:    params.TenantID,
			Phone:       params.Phone,
			Description: textParam(params.Description),
			Status:      leadStatusParam(params.Status),
		})
		if err != nil {
			return CreateCallResult{}, err
		}
	} else {
		lead, err = q.UpdateLead(ctx, db.UpdateLeadParams{
			ID:          lead.ID,
			Phone:       params.Phone,
			Description: textParam(params.Description),
			Status:      leadStatusParam(params.Status),
			TenantID:    params.TenantID,
		})
		if err != nil {
			return CreateCallResult{}, err
		}
	}

	call, err := q.CreateCall(ctx, db.CreateCallParams{
		TenantID:     params.TenantID,
		LeadID:       pgtype.Int8{Int64: lead.ID, Valid: true},
		Transcript:   textParam(params.Transcript),
		Details:      textParam(params.Details),
		Summary:      textParam(params.Summary),
		Sentiment:    callSentimentParam(params.Sentiment),
		Outcome:      callOutcomeParam(params.Outcome),
		DurationSecs: int4Param(params.DurationSecs),
	})
	if err != nil {
		return CreateCallResult{}, err
	}

	result := CreateCallResult{Lead: lead, Call: call}
	if params.Appointment != nil {
		appointment, err := q.CreateAppointment(ctx, db.CreateAppointmentParams{
			TenantID:        params.TenantID,
			LeadID:          lead.ID,
			Title:           "Property viewing",
			Notes:           textParam("Scheduled from V2 call summary. Day: " + params.Appointment.Day),
			Status:          appointmentStatusParam(db.AppointmentStatusScheduled),
			AppointmentDate: dateParam(params.Appointment.Date),
			AppointmentDay:  params.Appointment.Day,
			AppointmentTime: timeParam(params.Appointment.Time),
		})
		if err != nil {
			return CreateCallResult{}, err
		}
		result.Appointment = &appointment
	}

	return result, nil
}

func textParam(value string) pgtype.Text {
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

func leadStatusParam(value db.LeadStatus) db.NullLeadStatus {
	return db.NullLeadStatus{LeadStatus: value, Valid: value != ""}
}

func callSentimentParam(value db.CallSentiment) db.NullCallSentiment {
	return db.NullCallSentiment{CallSentiment: value, Valid: value != ""}
}

func callOutcomeParam(value db.CallOutcome) db.NullCallOutcome {
	return db.NullCallOutcome{CallOutcome: value, Valid: value != ""}
}

func appointmentStatusParam(value db.AppointmentStatus) db.NullAppointmentStatus {
	return db.NullAppointmentStatus{AppointmentStatus: value, Valid: value != ""}
}
