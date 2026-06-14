-- name: CreatePropertyEmbedding :one
INSERT INTO property_embeddings (tenant_id, property_id, embedding, content)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: SearchPropertyEmbeddings :many
SELECT content, property_id FROM property_embeddings
WHERE tenant_id = $1
ORDER BY embedding <=> $2
LIMIT $3;

-- name: DeletePropertyEmbedding :exec
DELETE FROM property_embeddings
WHERE property_id = $1 AND tenant_id = $2;