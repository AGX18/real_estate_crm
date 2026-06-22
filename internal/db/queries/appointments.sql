-- name: CreateAppointment :one
INSERT INTO appointments (tenant_id, lead_id, title, notes, status, appointment_day, appointment_time)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetAppointmentByID :one
SELECT * FROM appointments
WHERE id = $1 AND tenant_id = $2;

-- name: ListAppointments :many
SELECT * FROM appointments
WHERE tenant_id = $1
ORDER BY appointment_day ASC, appointment_time ASC;

-- name: ListAppointmentsByLead :many
SELECT * FROM appointments
WHERE lead_id = $1 AND tenant_id = $2
ORDER BY appointment_day ASC, appointment_time ASC;

-- name: UpdateAppointment :one
UPDATE appointments
SET
    title = $2,
    notes = $3,
    status = $4,
    appointment_day = $5,
    appointment_time = $6,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $7
RETURNING *;

-- name: UpdateAppointmentStatus :one
UPDATE appointments
SET status = $2, updated_at = NOW()
WHERE id = $1 AND tenant_id = $3
RETURNING *;

-- name: DeleteAppointment :exec
DELETE FROM appointments
WHERE id = $1 AND tenant_id = $2;
