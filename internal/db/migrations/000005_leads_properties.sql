-- +goose Up
-- +goose StatementBegin
CREATE TABLE lead_properties (
    lead_id       BIGINT NOT NULL REFERENCES leads(id) ON DELETE CASCADE,
    property_id   BIGINT NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (lead_id, property_id)
);
-- +goose StatementEnd


-- goose Down
-- +goose StatementBegin
DROP TABLE lead_properties;
-- +goose StatementEnd