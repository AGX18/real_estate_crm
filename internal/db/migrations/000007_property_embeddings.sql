-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION vector;
CREATE TABLE IF NOT EXISTS property_embeddings (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    property_id BIGINT NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    embedding   vector(1536),
    content     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON property_embeddings
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS property_embeddings;
-- +goose StatementEnd