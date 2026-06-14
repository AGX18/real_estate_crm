-- +goose Up
-- +goose StatementBegin
CREATE TYPE lead_status AS ENUM ('Follow_Up', 'qualified', 'closed', 'unqualified');
CREATE TABLE leads (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    phone       TEXT NOT NULL CHECK (phone ~ '^\+?[0-9]{7,15}$'),
    description TEXT,
    status      lead_status DEFAULT 'Follow_Up',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(phone, tenant_id)
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE leads;
DROP TYPE IF EXISTS lead_status;
-- +goose StatementEnd