-- name: CreateBroker :one
INSERT INTO brokers (tenant_id, username, email, password_hash, role)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetBrokerByID :one
SELECT * FROM brokers
WHERE id = $1 AND tenant_id = $2;

-- name: GetBrokerByEmail :one
SELECT * FROM brokers
WHERE email = $1 AND tenant_id = $2;

-- name: ListBrokers :many
SELECT * FROM brokers
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: UpdateBrokerRole :one
UPDATE brokers
SET role = $2, updated_at = NOW()
WHERE id = $1 AND tenant_id = $3
RETURNING *;

-- name: DeleteBroker :exec
DELETE FROM brokers
WHERE id = $1 AND tenant_id = $2;