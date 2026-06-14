
-- +goose Up
-- +goose StatementBegin
CREATE TYPE user_role AS ENUM ('user', 'admin');
CREATE TABLE IF NOT EXISTS users (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    username    VARCHAR(60) NOT NULL,
    email       VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role        user_role NOT NULL DEFAULT 'user',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(username, tenant_id),
    UNIQUE(email, tenant_id)
);
-- +goose StatementEnd


-- goose Down
-- +goose StatementBegin
DROP TABLE users;
DROP TYPE IF EXISTS user_role;
-- +goose StatementEnd