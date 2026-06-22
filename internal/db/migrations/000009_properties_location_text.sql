-- +goose Up
-- +goose StatementBegin
ALTER TABLE properties
    ALTER COLUMN city TYPE TEXT,
    ALTER COLUMN governorate TYPE TEXT;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
ALTER TABLE properties
    ALTER COLUMN city TYPE VARCHAR(20),
    ALTER COLUMN governorate TYPE VARCHAR(25);
-- +goose StatementEnd
