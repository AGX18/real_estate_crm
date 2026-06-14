-- +goose Up
-- +goose StatementBegin
CREATE TYPE property_status AS ENUM ('available', 'sold', 'rented');
CREATE TABLE properties (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    title       TEXT NOT NULL,
    description TEXT,
    price       NUMERIC(12,2),
    location    TEXT NOT NULL,
    area_sqm    NUMERIC(10,2) NOT NULL,
    status      property_status DEFAULT 'available',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd



-- goose Down
-- +goose StatementBegin
DROP TABLE units;
DROP TYPE IF EXISTS unit_status;
-- +goose StatementEnd