-- name: CreateTenant :one
INSERT INTO tenants (name, status)
VALUES ($1, $2)
RETURNING *;

-- name: GetTenant :one
SELECT * FROM tenants
WHERE id = $1;

-- name: GetTenantByName :one
SELECT * FROM tenants
WHERE name = $1;

-- name: ListTenants :many
SELECT * FROM tenants
ORDER BY created_at DESC;

-- name: UpdateTenantStatus :one
UPDATE tenants
SET status = $2
WHERE id = $1
RETURNING *;

-- name: DeleteTenant :exec
DELETE FROM tenants
WHERE id = $1;