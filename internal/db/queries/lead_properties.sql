-- name: AddLeadProperty :exec
INSERT INTO lead_properties (lead_id, property_id)
VALUES ($1, $2);

-- name: GetLeadProperties :many
SELECT p.* FROM properties p
INNER JOIN lead_properties lp ON lp.property_id = p.id
WHERE lp.lead_id = $1 AND p.tenant_id = $2
ORDER BY lp.created_at DESC;

-- name: GetPropertyLeads :many
SELECT l.* FROM leads l
INNER JOIN lead_properties lp ON lp.lead_id = l.id
WHERE lp.property_id = $1 AND l.tenant_id = $2
ORDER BY lp.created_at DESC;

-- name: RemoveLeadProperty :exec
DELETE FROM lead_properties
WHERE lead_id = $1 AND property_id = $2;