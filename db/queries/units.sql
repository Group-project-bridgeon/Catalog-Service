-- name: CreateUnit :one
INSERT INTO catalog_units (
    id,
    code,
    name,
    is_active
)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUnitByID :one
SELECT *
FROM catalog_units
WHERE id = $1;

-- name: GetUnitByCode :one
SELECT *
FROM catalog_units
WHERE code = $1;

-- name: GetUnitByName :one
SELECT *
FROM catalog_units
WHERE name = $1;

-- name: ListUnits :many
SELECT *
FROM catalog_units
ORDER BY name;

-- name: UpdateUnit :one
UPDATE catalog_units
SET
    code = $2,
    name = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SetUnitActive :one
UPDATE catalog_units
SET
    is_active = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
