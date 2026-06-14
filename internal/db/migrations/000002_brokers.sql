
-- +goose Up
-- +goose StatementBegin
CREATE TYPE role AS ENUM ('user', 'admin');
CREATE TABLE IF NOT EXISTS brokers (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    username    VARCHAR(60) NOT NULL,
    email       VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role        role NOT NULL DEFAULT 'user',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(username, tenant_id),
    UNIQUE(email, tenant_id)
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE brokers;
DROP TYPE IF EXISTS role;
-- +goose StatementEnd