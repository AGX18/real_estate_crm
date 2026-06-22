package appointments

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

func (s *Service) Create(ctx context.Context, params db.CreateAppointmentParams) (db.Appointment, error) {
	return s.queries.CreateAppointment(ctx, params)
}

func (s *Service) Get(ctx context.Context, params db.GetAppointmentByIDParams) (db.Appointment, error) {
	return s.queries.GetAppointmentByID(ctx, params)
}

func (s *Service) List(ctx context.Context, tenantID pgtype.UUID) ([]db.Appointment, error) {
	return s.queries.ListAppointments(ctx, tenantID)
}

func (s *Service) ListByLead(ctx context.Context, params db.ListAppointmentsByLeadParams) ([]db.Appointment, error) {
	return s.queries.ListAppointmentsByLead(ctx, params)
}

func (s *Service) Update(ctx context.Context, params db.UpdateAppointmentParams) (db.Appointment, error) {
	return s.queries.UpdateAppointment(ctx, params)
}

func (s *Service) UpdateStatus(ctx context.Context, params db.UpdateAppointmentStatusParams) (db.Appointment, error) {
	return s.queries.UpdateAppointmentStatus(ctx, params)
}

func (s *Service) Delete(ctx context.Context, params db.DeleteAppointmentParams) error {
	return s.queries.DeleteAppointment(ctx, params)
}
