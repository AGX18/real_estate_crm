package leads

import (
	"context"

	db "real_estate_crm/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	queries db.Querier
}

func NewService(queries db.Querier) *Service {
	return &Service{queries: queries}
}

func (s *Service) Create(ctx context.Context, params db.CreateLeadParams) (db.Lead, error) {
	return s.queries.CreateLead(ctx, params)
}

func (s *Service) Get(ctx context.Context, params db.GetLeadByIDParams) (db.Lead, error) {
	return s.queries.GetLeadByID(ctx, params)
}

func (s *Service) GetByPhone(ctx context.Context, params db.GetLeadByPhoneParams) (db.Lead, error) {
	return s.queries.GetLeadByPhone(ctx, params)
}

func (s *Service) List(ctx context.Context, tenantID pgtype.UUID) ([]db.Lead, error) {
	return s.queries.ListLeads(ctx, tenantID)
}

func (s *Service) Update(ctx context.Context, params db.UpdateLeadParams) (db.Lead, error) {
	return s.queries.UpdateLead(ctx, params)
}

func (s *Service) UpdateStatus(ctx context.Context, params db.UpdateLeadStatusParams) (db.Lead, error) {
	return s.queries.UpdateLeadStatus(ctx, params)
}

func (s *Service) UpdateDescription(ctx context.Context, params db.UpdateLeadDescriptionParams) (db.Lead, error) {
	return s.queries.UpdateLeadDescription(ctx, params)
}

func (s *Service) Delete(ctx context.Context, params db.DeleteLeadParams) error {
	return s.queries.DeleteLead(ctx, params)
}
