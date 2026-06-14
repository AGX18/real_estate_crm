-- name: CreateCall :one
INSERT INTO calls (tenant_id, lead_id, transcript, summary, sentiment, outcome, duration_secs)
VALUES ($1, $2, $3, $4, $5, $6, $7)
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
    summary = $3,
    sentiment = $4,
    outcome = $5,
    duration_secs = $6
WHERE id = $1 AND tenant_id = $7
RETURNING *;

-- name: DeleteCall :exec
DELETE FROM calls
WHERE id = $1 AND tenant_id = $2;