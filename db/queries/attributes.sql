-- name: CreateAttribute :one
INSERT INTO catalog_attributes (
    id,
    name,
    code
)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetAttributeByID :one
SELECT *
FROM catalog_attributes
WHERE id = $1;

-- name: GetAttributeByCode :one
SELECT *
FROM catalog_attributes
WHERE code = $1;

-- name: GetAttributeByName :one
SELECT *
FROM catalog_attributes
WHERE name = $1;

-- name: ListAttributes :many
SELECT *
FROM catalog_attributes
ORDER BY name;

-- name: UpdateAttribute :one
UPDATE catalog_attributes
SET
    name = $2,
    code = $3
WHERE id = $1
RETURNING *;
