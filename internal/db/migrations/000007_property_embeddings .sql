-- +goose Up
-- +goose StatementBegin
CREATE TABLE property_embeddings (
    id UUID,
    tenant_id UUID,
    embedding vector(1536),
    content TEXT,
    metadata JSONB
);
-- +goose StatementEnd

-- goose Down
-- +goose StatementBegin
DROP TABLE property_embeddings;
-- +goose StatementEnd