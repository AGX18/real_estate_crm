package voice

import (
	"context"
	"errors"

	db "real_estate_crm/internal/db/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	queries db.Querier
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
}

type CreateCallResult struct {
	Lead db.Lead `json:"lead"`
	Call db.Call `json:"call"`
}

func NewService(queries db.Querier) *Service {
	return &Service{queries: queries}
}

func (s *Service) CreateCall(ctx context.Context, params CreateCallParams) (CreateCallResult, error) {
	lead, err := s.queries.GetLeadByPhone(ctx, db.GetLeadByPhoneParams{
		Phone:    params.Phone,
		TenantID: params.TenantID,
	})
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return CreateCallResult{}, err
		}

		lead, err = s.queries.CreateLead(ctx, db.CreateLeadParams{
			TenantID:    params.TenantID,
			Phone:       params.Phone,
			Description: textParam(params.Description),
			Status:      leadStatusParam(params.Status),
		})
		if err != nil {
			return CreateCallResult{}, err
		}
	} else {
		lead, err = s.queries.UpdateLead(ctx, db.UpdateLeadParams{
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

	call, err := s.queries.CreateCall(ctx, db.CreateCallParams{
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

	return CreateCallResult{Lead: lead, Call: call}, nil
}

func textParam(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: value != ""}
}

func int4Param(value int32) pgtype.Int4 {
	return pgtype.Int4{Int32: value, Valid: value > 0}
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
