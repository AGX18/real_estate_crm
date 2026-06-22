-- +goose Up
-- +goose StatementBegin
CREATE TYPE appointment_status AS ENUM ('scheduled', 'completed', 'canceled', 'no_show');

CREATE TABLE IF NOT EXISTS appointments (
    id               BIGSERIAL PRIMARY KEY,
    tenant_id        UUID NOT NULL REFERENCES tenants(id),
    lead_id          BIGINT NOT NULL REFERENCES leads(id) ON DELETE CASCADE,
    title            TEXT NOT NULL,
    notes            TEXT,
    status           appointment_status DEFAULT 'scheduled',
    appointment_date DATE NOT NULL,
    appointment_day  TEXT NOT NULL,
    appointment_time TIME NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS appointments_tenant_date_time_idx
    ON appointments (tenant_id, appointment_date, appointment_time);

CREATE INDEX IF NOT EXISTS appointments_lead_date_time_idx
    ON appointments (lead_id, appointment_date, appointment_time);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS appointments;
DROP TYPE IF EXISTS appointment_status;
-- +goose StatementEnd
