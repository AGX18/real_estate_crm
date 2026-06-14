-- +goose Up
-- +goose StatementBegin
CREATE TYPE call_sentiment AS ENUM ('positive', 'negative', 'neutral');
CREATE TYPE call_outcome AS ENUM ('follow_up', 'qualified', 'closed', 'unqualified', 'no_answer');

CREATE TABLE calls (
    id            BIGSERIAL PRIMARY KEY,
    tenant_id     UUID NOT NULL REFERENCES tenants(id),
    lead_id       BIGINT REFERENCES leads(id),
    transcript    TEXT,
    summary       TEXT,
    sentiment     call_sentiment,
    outcome       call_outcome DEFAULT 'follow_up',
    duration_secs INT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE calls;
DROP TYPE IF EXISTS call_sentiment;
DROP TYPE IF EXISTS call_outcome;
-- +goose StatementEnd