-- +goose Up
-- +goose StatementBegin
DROP INDEX IF EXISTS appointments_tenant_date_time_idx;
DROP INDEX IF EXISTS appointments_lead_date_time_idx;

ALTER TABLE appointments
    DROP COLUMN IF EXISTS appointment_date,
    ALTER COLUMN appointment_time TYPE TEXT USING appointment_time::TEXT;

CREATE INDEX IF NOT EXISTS appointments_tenant_day_time_idx
    ON appointments (tenant_id, appointment_day, appointment_time);

CREATE INDEX IF NOT EXISTS appointments_lead_day_time_idx
    ON appointments (lead_id, appointment_day, appointment_time);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS appointments_tenant_day_time_idx;
DROP INDEX IF EXISTS appointments_lead_day_time_idx;

ALTER TABLE appointments
    ADD COLUMN IF NOT EXISTS appointment_date DATE NOT NULL DEFAULT CURRENT_DATE,
    ALTER COLUMN appointment_time TYPE TIME USING appointment_time::TIME;

CREATE INDEX IF NOT EXISTS appointments_tenant_date_time_idx
    ON appointments (tenant_id, appointment_date, appointment_time);

CREATE INDEX IF NOT EXISTS appointments_lead_date_time_idx
    ON appointments (lead_id, appointment_date, appointment_time);
-- +goose StatementEnd
