-- name: CreateCall :one
INSERT INTO calls (tenant_id, lead_id, transcript, details, summary, sentiment, outcome, duration_secs)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetCallByID :one
SELECT * FROM calls
WHERE id = $1 AND tenant_id = $2;

-- name: ListCalls :many
SELECT * FROM calls
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: ListCallsByLead :many
SELECT * FROM calls
WHERE lead_id = $1 AND tenant_id = $2
ORDER BY created_at DESC;

-- name: ListCallsByOutcome :many
SELECT * FROM calls
WHERE tenant_id = $1 AND outcome = $2
ORDER BY created_at DESC;

-- name: UpdateCall :one
UPDATE calls
SET
    transcript = $2,
    details = $3,
    summary = $4,
    sentiment = $5,
    outcome = $6,
    duration_secs = $7
WHERE id = $1 AND tenant_id = $8
RETURNING *;

-- name: DeleteCall :exec
DELETE FROM calls
WHERE id = $1 AND tenant_id = $2;
