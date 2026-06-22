-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS property_embeddings (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants (id),
    property_id BIGINT NOT NULL REFERENCES properties (id) ON DELETE CASCADE,
    embedding vector (1536),
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS property_embeddings_tenant_id_idx ON property_embeddings (tenant_id);

CREATE INDEX IF NOT EXISTS property_embeddings_embedding_hnsw_idx ON property_embeddings USING hnsw (embedding vector_cosine_ops);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS property_embeddings;
-- +goose StatementEnd