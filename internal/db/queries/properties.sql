-- name: CreateProperty :one
INSERT INTO properties (tenant_id, description, price, location, area_sqm, type, city, governorate, bedrooms, bathrooms, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetPropertyByID :one
SELECT * FROM properties
WHERE id = $1 AND tenant_id = $2;

-- name: ListProperties :many
SELECT * FROM properties
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: ListPropertiesByStatus :many
SELECT * FROM properties
WHERE tenant_id = $1 AND status = $2
ORDER BY created_at DESC;

-- name: ListPropertiesByType :many
SELECT * FROM properties
WHERE tenant_id = $1 AND type = $2
ORDER BY created_at DESC;

-- name: UpdateProperty :one
UPDATE properties
SET
    description = $2,
    price = $3,
    location = $4,
    area_sqm = $5,
    type = $6,
    city = $7,
    governorate = $8,
    bedrooms = $9,
    bathrooms = $10,
    status = $11,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $12
RETURNING *;

-- name: CreateProperties :copyfrom
INSERT INTO properties (tenant_id, description, price, location, area_sqm, type, city, governorate, bedrooms, bathrooms, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: DeleteProperty :exec
DELETE FROM properties
WHERE id = $1 AND tenant_id = $2;
