-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tenants (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL UNIQUE,
    status      VARCHAR(50) DEFAULT 'active',
    created_at  TIMESTAMP DEFAULT NOW()
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE tenants;
-- +goose StatementEnd