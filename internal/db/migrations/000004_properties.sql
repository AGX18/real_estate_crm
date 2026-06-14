-- +goose Up
-- +goose StatementBegin
CREATE TYPE property_status AS ENUM ('available', 'sold', 'rented');
CREATE TYPE property_type AS ENUM ('شاليه', 'شقة', 'استوديو', 'دوبلكس', 'بنتهاوس', 'فيلا', 'توين هاوس', 'تاون هاوس');
CREATE TABLE properties (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    description TEXT,
    price       NUMERIC(12,2),
    location    TEXT,
    area_sqm    NUMERIC(10,2) NOT NULL,
    type        property_type DEFAULT 'شقة',
    city        VARCHAR(20),
    governorate VARCHAR(25),
    bedrooms    INTEGER NOT NULL,
    bathrooms   INTEGER NOT NULL,
    status      property_status DEFAULT 'available',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd



-- +goose Down
-- +goose StatementBegin
DROP TABLE properties;
DROP TYPE IF EXISTS property_status;
-- +goose StatementEnd