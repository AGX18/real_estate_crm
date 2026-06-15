package server

import (
	"context"
	db "real_estate_crm/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

type mockQueries struct {
	tenant  db.Tenant
	tenants []db.Tenant
	broker  db.Broker
	brokers []db.Broker
	err     error
}

// Tenants
func (m *mockQueries) CreateTenant(ctx context.Context, arg db.CreateTenantParams) (db.Tenant, error) {
	return m.tenant, m.err
}
func (m *mockQueries) GetTenant(ctx context.Context, id pgtype.UUID) (db.Tenant, error) {
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
func (m *mockQueries) CreateBroker(ctx context.Context, arg db.CreateBrokerParams) (db.Broker, error) {
	return m.broker, m.err
}
func (m *mockQueries) CreateCall(ctx context.Context, arg db.CreateCallParams) (db.Call, error) {
	return db.Call{}, nil
}
func (m *mockQueries) CreateLead(ctx context.Context, arg db.CreateLeadParams) (db.Lead, error) {
	return db.Lead{}, nil
}
func (m *mockQueries) CreateProperties(ctx context.Context, arg []db.CreatePropertiesParams) (int64, error) {
	return 0, nil
}
func (m *mockQueries) CreateProperty(ctx context.Context, arg db.CreatePropertyParams) (db.Property, error) {
	return db.Property{}, nil
}
func (m *mockQueries) CreatePropertyEmbedding(ctx context.Context, arg db.CreatePropertyEmbeddingParams) (db.PropertyEmbedding, error) {
	return db.PropertyEmbedding{}, nil
}
func (m *mockQueries) DeleteBroker(ctx context.Context, arg db.DeleteBrokerParams) error {
	return m.err
}
func (m *mockQueries) DeleteCall(ctx context.Context, arg db.DeleteCallParams) error { return nil }
func (m *mockQueries) DeleteLead(ctx context.Context, arg db.DeleteLeadParams) error { return nil }
func (m *mockQueries) DeleteProperty(ctx context.Context, arg db.DeletePropertyParams) error {
	return nil
}
func (m *mockQueries) DeletePropertyEmbedding(ctx context.Context, arg db.DeletePropertyEmbeddingParams) error {
	return nil
}
func (m *mockQueries) GetBrokerByEmail(ctx context.Context, arg db.GetBrokerByEmailParams) (db.Broker, error) {
	return m.broker, m.err
}
func (m *mockQueries) GetBrokerByID(ctx context.Context, arg db.GetBrokerByIDParams) (db.Broker, error) {
	return m.broker, m.err
}
func (m *mockQueries) GetCallByID(ctx context.Context, arg db.GetCallByIDParams) (db.Call, error) {
	return db.Call{}, nil
}
func (m *mockQueries) GetLeadByID(ctx context.Context, arg db.GetLeadByIDParams) (db.Lead, error) {
	return db.Lead{}, nil
}
func (m *mockQueries) GetLeadByPhone(ctx context.Context, arg db.GetLeadByPhoneParams) (db.Lead, error) {
	return db.Lead{}, nil
}
func (m *mockQueries) GetLeadProperties(ctx context.Context, arg db.GetLeadPropertiesParams) ([]db.Property, error) {
	return nil, nil
}
func (m *mockQueries) GetPropertyByID(ctx context.Context, arg db.GetPropertyByIDParams) (db.Property, error) {
	return db.Property{}, nil
}
func (m *mockQueries) GetPropertyLeads(ctx context.Context, arg db.GetPropertyLeadsParams) ([]db.Lead, error) {
	return nil, nil
}
func (m *mockQueries) ListBrokers(ctx context.Context, tenantID pgtype.UUID) ([]db.Broker, error) {
	return m.brokers, m.err
}
func (m *mockQueries) ListCalls(ctx context.Context, tenantID pgtype.UUID) ([]db.Call, error) {
	return nil, nil
}
func (m *mockQueries) ListCallsByLead(ctx context.Context, arg db.ListCallsByLeadParams) ([]db.Call, error) {
	return nil, nil
}
func (m *mockQueries) ListCallsByOutcome(ctx context.Context, arg db.ListCallsByOutcomeParams) ([]db.Call, error) {
	return nil, nil
}
func (m *mockQueries) ListLeads(ctx context.Context, tenantID pgtype.UUID) ([]db.Lead, error) {
	return nil, nil
}
func (m *mockQueries) ListProperties(ctx context.Context, tenantID pgtype.UUID) ([]db.Property, error) {
	return nil, nil
}
func (m *mockQueries) ListPropertiesByStatus(ctx context.Context, arg db.ListPropertiesByStatusParams) ([]db.Property, error) {
	return nil, nil
}
func (m *mockQueries) ListPropertiesByType(ctx context.Context, arg db.ListPropertiesByTypeParams) ([]db.Property, error) {
	return nil, nil
}
func (m *mockQueries) RemoveLeadProperty(ctx context.Context, arg db.RemoveLeadPropertyParams) error {
	return nil
}
func (m *mockQueries) SearchPropertyEmbeddings(ctx context.Context, arg db.SearchPropertyEmbeddingsParams) ([]db.SearchPropertyEmbeddingsRow, error) {
	return nil, nil
}
func (m *mockQueries) UpdateBrokerRole(ctx context.Context, arg db.UpdateBrokerRoleParams) (db.Broker, error) {
	return m.broker, m.err
}
func (m *mockQueries) UpdateCall(ctx context.Context, arg db.UpdateCallParams) (db.Call, error) {
	return db.Call{}, nil
}
func (m *mockQueries) UpdateLead(ctx context.Context, arg db.UpdateLeadParams) (db.Lead, error) {
	return db.Lead{}, nil
}
func (m *mockQueries) UpdateLeadDescription(ctx context.Context, arg db.UpdateLeadDescriptionParams) (db.Lead, error) {
	return db.Lead{}, nil
}
func (m *mockQueries) UpdateLeadStatus(ctx context.Context, arg db.UpdateLeadStatusParams) (db.Lead, error) {
	return db.Lead{}, nil
}
func (m *mockQueries) UpdateProperty(ctx context.Context, arg db.UpdatePropertyParams) (db.Property, error) {
	return db.Property{}, nil
}
