package tenants

import (
	"context"

	db "github.com/AGX18/real_estate_crm/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	queries db.Querier
}

func NewService(queries db.Querier) *Service {
	return &Service{queries: queries}
}

func (s *Service) Create(ctx context.Context, params db.CreateTenantParams) (db.Tenant, error) {
	return s.queries.CreateTenant(ctx, params)
}

func (s *Service) Get(ctx context.Context, id pgtype.UUID) (db.Tenant, error) {
	return s.queries.GetTenant(ctx, id)
}

func (s *Service) GetByName(ctx context.Context, name string) (db.Tenant, error) {
	return s.queries.GetTenantByName(ctx, name)
}

func (s *Service) List(ctx context.Context) ([]db.Tenant, error) {
	return s.queries.ListTenants(ctx)
}

func (s *Service) UpdateStatus(ctx context.Context, params db.UpdateTenantStatusParams) (db.Tenant, error) {
	return s.queries.UpdateTenantStatus(ctx, params)
}

func (s *Service) Delete(ctx context.Context, id pgtype.UUID) error {
	return s.queries.DeleteTenant(ctx, id)
}
