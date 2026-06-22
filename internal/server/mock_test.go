package server

import (
	"context"

	db "github.com/AGX18/real_estate_crm/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

type mockQueries struct {
	tenant                       db.Tenant
	tenants                      []db.Tenant
	broker                       db.Broker
	brokers                      []db.Broker
	lead                         db.Lead
	leads                        []db.Lead
	appointment                  db.Appointment
	appointments                 []db.Appointment
	call                         db.Call
	calls                        []db.Call
	property                     db.Property
	properties                   []db.Property
	propertyEmbedding            db.PropertyEmbedding
	propertySearchResults        []db.SearchPropertyEmbeddingsRow
	createBrokerArg              db.CreateBrokerParams
	getBrokerByEmailArg          db.GetBrokerByEmailParams
	createAppointmentArg         db.CreateAppointmentParams
	createLeadArg                db.CreateLeadParams
	updateLeadArg                db.UpdateLeadParams
	createCallArg                db.CreateCallParams
	createPropertyArg            db.CreatePropertyParams
	createPropertyCalls          int
	createPropertyEmbeddingArg   db.CreatePropertyEmbeddingParams
	createPropertyEmbeddingCalls int
	searchPropertyEmbeddingsArg  db.SearchPropertyEmbeddingsParams
	leadByPhoneErr               error
	err                          error
}

// Tenants
func (m *mockQueries) CreateTenant(ctx context.Context, arg db.CreateTenantParams) (db.Tenant, error) {
	return m.tenant, m.err
}
func (m *mockQueries) GetTenant(ctx context.Context, id pgtype.UUID) (db.Tenant, error) {
	return m.tenant, m.err
}
func (m *mockQueries) GetTenantByName(ctx context.Context, name string) (db.Tenant, error) {
	return m.tenant, m.err
}
func (m *mockQueries) ListTenants(ctx context.Context) ([]db.Tenant, error) {
	return m.tenants, m.err
}
func (m *mockQueries) UpdateTenantStatus(ctx context.Context, arg db.UpdateTenantStatusParams) (db.Tenant, error) {
	return m.tenant, m.err
}
func (m *mockQueries) DeleteTenant(ctx context.Context, id pgtype.UUID) error {
	return m.err
}

// Stub everything else
func (m *mockQueries) AddLeadProperty(ctx context.Context, arg db.AddLeadPropertyParams) error {
	return nil
}
func (m *mockQueries) CreateAppointment(ctx context.Context, arg db.CreateAppointmentParams) (db.Appointment, error) {
	m.createAppointmentArg = arg
	return m.appointment, m.err
}
func (m *mockQueries) CreateBroker(ctx context.Context, arg db.CreateBrokerParams) (db.Broker, error) {
	m.createBrokerArg = arg
	return m.broker, m.err
}
func (m *mockQueries) CreateCall(ctx context.Context, arg db.CreateCallParams) (db.Call, error) {
	m.createCallArg = arg
	return m.call, m.err
}
func (m *mockQueries) CreateLead(ctx context.Context, arg db.CreateLeadParams) (db.Lead, error) {
	m.createLeadArg = arg
	return m.lead, m.err
}
func (m *mockQueries) CreateProperties(ctx context.Context, arg []db.CreatePropertiesParams) (int64, error) {
	return int64(len(arg)), m.err
}
func (m *mockQueries) CreateProperty(ctx context.Context, arg db.CreatePropertyParams) (db.Property, error) {
	m.createPropertyArg = arg
	m.createPropertyCalls++
	return m.property, m.err
}
func (m *mockQueries) CreatePropertyEmbedding(ctx context.Context, arg db.CreatePropertyEmbeddingParams) (db.PropertyEmbedding, error) {
	m.createPropertyEmbeddingArg = arg
	m.createPropertyEmbeddingCalls++
	return m.propertyEmbedding, m.err
}
func (m *mockQueries) DeleteAppointment(ctx context.Context, arg db.DeleteAppointmentParams) error {
	return m.err
}
func (m *mockQueries) DeleteBroker(ctx context.Context, arg db.DeleteBrokerParams) error {
	return m.err
}
func (m *mockQueries) DeleteCall(ctx context.Context, arg db.DeleteCallParams) error { return nil }
func (m *mockQueries) DeleteLead(ctx context.Context, arg db.DeleteLeadParams) error { return m.err }
func (m *mockQueries) DeleteProperty(ctx context.Context, arg db.DeletePropertyParams) error {
	return m.err
}
func (m *mockQueries) DeletePropertyEmbedding(ctx context.Context, arg db.DeletePropertyEmbeddingParams) error {
	return m.err
}
func (m *mockQueries) GetBrokerByEmail(ctx context.Context, arg db.GetBrokerByEmailParams) (db.Broker, error) {
	m.getBrokerByEmailArg = arg
	return m.broker, m.err
}
func (m *mockQueries) GetAppointmentByID(ctx context.Context, arg db.GetAppointmentByIDParams) (db.Appointment, error) {
	return m.appointment, m.err
}
func (m *mockQueries) GetBrokerByID(ctx context.Context, arg db.GetBrokerByIDParams) (db.Broker, error) {
	return m.broker, m.err
}
func (m *mockQueries) GetCallByID(ctx context.Context, arg db.GetCallByIDParams) (db.Call, error) {
	return db.Call{}, nil
}
func (m *mockQueries) GetLeadByID(ctx context.Context, arg db.GetLeadByIDParams) (db.Lead, error) {
	return m.lead, m.err
}
func (m *mockQueries) GetLeadByPhone(ctx context.Context, arg db.GetLeadByPhoneParams) (db.Lead, error) {
	if m.leadByPhoneErr != nil {
		return db.Lead{}, m.leadByPhoneErr
	}
	return m.lead, m.err
}
func (m *mockQueries) GetLeadProperties(ctx context.Context, arg db.GetLeadPropertiesParams) ([]db.Property, error) {
	return nil, nil
}
func (m *mockQueries) GetPropertyByID(ctx context.Context, arg db.GetPropertyByIDParams) (db.Property, error) {
	return m.property, m.err
}
func (m *mockQueries) GetPropertyLeads(ctx context.Context, arg db.GetPropertyLeadsParams) ([]db.Lead, error) {
	return nil, nil
}
func (m *mockQueries) ListAppointments(ctx context.Context, tenantID pgtype.UUID) ([]db.Appointment, error) {
	return m.appointments, m.err
}
func (m *mockQueries) ListAppointmentsByLead(ctx context.Context, arg db.ListAppointmentsByLeadParams) ([]db.Appointment, error) {
	return m.appointments, m.err
}
func (m *mockQueries) ListBrokers(ctx context.Context, tenantID pgtype.UUID) ([]db.Broker, error) {
	return m.brokers, m.err
}
func (m *mockQueries) ListCalls(ctx context.Context, tenantID pgtype.UUID) ([]db.Call, error) {
	return m.calls, m.err
}
func (m *mockQueries) ListCallsByLead(ctx context.Context, arg db.ListCallsByLeadParams) ([]db.Call, error) {
	return nil, nil
}
func (m *mockQueries) ListCallsByOutcome(ctx context.Context, arg db.ListCallsByOutcomeParams) ([]db.Call, error) {
	return nil, nil
}
func (m *mockQueries) ListLeads(ctx context.Context, tenantID pgtype.UUID) ([]db.Lead, error) {
	return m.leads, m.err
}
func (m *mockQueries) ListProperties(ctx context.Context, tenantID pgtype.UUID) ([]db.Property, error) {
	return m.properties, m.err
}
func (m *mockQueries) ListPropertiesByStatus(ctx context.Context, arg db.ListPropertiesByStatusParams) ([]db.Property, error) {
	return m.properties, m.err
}
func (m *mockQueries) ListPropertiesByType(ctx context.Context, arg db.ListPropertiesByTypeParams) ([]db.Property, error) {
	return m.properties, m.err
}
func (m *mockQueries) RemoveLeadProperty(ctx context.Context, arg db.RemoveLeadPropertyParams) error {
	return nil
}
func (m *mockQueries) SearchPropertyEmbeddings(ctx context.Context, arg db.SearchPropertyEmbeddingsParams) ([]db.SearchPropertyEmbeddingsRow, error) {
	m.searchPropertyEmbeddingsArg = arg
	return m.propertySearchResults, m.err
}
func (m *mockQueries) UpdateBrokerRole(ctx context.Context, arg db.UpdateBrokerRoleParams) (db.Broker, error) {
	return m.broker, m.err
}
func (m *mockQueries) UpdateAppointment(ctx context.Context, arg db.UpdateAppointmentParams) (db.Appointment, error) {
	return m.appointment, m.err
}
func (m *mockQueries) UpdateAppointmentStatus(ctx context.Context, arg db.UpdateAppointmentStatusParams) (db.Appointment, error) {
	return m.appointment, m.err
}
func (m *mockQueries) UpdateCall(ctx context.Context, arg db.UpdateCallParams) (db.Call, error) {
	return db.Call{}, nil
}
func (m *mockQueries) UpdateLead(ctx context.Context, arg db.UpdateLeadParams) (db.Lead, error) {
	m.updateLeadArg = arg
	return m.lead, m.err
}
func (m *mockQueries) UpdateLeadDescription(ctx context.Context, arg db.UpdateLeadDescriptionParams) (db.Lead, error) {
	return m.lead, m.err
}
func (m *mockQueries) UpdateLeadStatus(ctx context.Context, arg db.UpdateLeadStatusParams) (db.Lead, error) {
	return m.lead, m.err
}
func (m *mockQueries) UpdateProperty(ctx context.Context, arg db.UpdatePropertyParams) (db.Property, error) {
	return m.property, m.err
}
