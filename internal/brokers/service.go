package brokers

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

func (s *Service) Create(ctx context.Context, params db.CreateBrokerParams) (db.Broker, error) {
	return s.queries.CreateBroker(ctx, params)
}

func (s *Service) Get(ctx context.Context, params db.GetBrokerByIDParams) (db.Broker, error) {
	return s.queries.GetBrokerByID(ctx, params)
}

func (s *Service) GetByEmail(ctx context.Context, params db.GetBrokerByEmailParams) (db.Broker, error) {
	return s.queries.GetBrokerByEmail(ctx, params)
}

func (s *Service) List(ctx context.Context, tenantID pgtype.UUID) ([]db.Broker, error) {
	return s.queries.ListBrokers(ctx, tenantID)
}

func (s *Service) UpdateRole(ctx context.Context, params db.UpdateBrokerRoleParams) (db.Broker, error) {
	return s.queries.UpdateBrokerRole(ctx, params)
}

func (s *Service) Delete(ctx context.Context, params db.DeleteBrokerParams) error {
	return s.queries.DeleteBroker(ctx, params)
}
