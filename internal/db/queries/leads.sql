-- name: CreateLead :one
INSERT INTO leads (tenant_id, phone, description, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetLeadByID :one
SELECT * FROM leads
WHERE id = $1 AND tenant_id = $2;

-- name: GetLeadByPhone :one
SELECT * FROM leads
WHERE phone = $1 AND tenant_id = $2;

-- name: ListLeads :many
SELECT * FROM leads
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: UpdateLeadStatus :one
UPDATE leads
SET status = $2, updated_at = NOW()
WHERE id = $1 AND tenant_id = $3
RETURNING *;

-- name: UpdateLeadDescription :one
UPDATE leads
SET description = $2, updated_at = NOW()
WHERE id = $1 AND tenant_id = $3
RETURNING *;

-- name: UpdateLead :one
UPDATE leads
SET 
    phone = $2,
    description = $3,
    status = $4,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $5
RETURNING *;

-- name: DeleteLead :exec
DELETE FROM leads
WHERE id = $1 AND tenant_id = $2;