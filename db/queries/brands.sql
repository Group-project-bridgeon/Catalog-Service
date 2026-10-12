-- name: CreateBrand :one
INSERT INTO catalog_brands (
    id,
    name,
    code,
    is_active
)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetBrandByID :one
SELECT *
FROM catalog_brands
WHERE id = $1;

-- name: GetBrandByCode :one
SELECT *
FROM catalog_brands
WHERE code = $1;

-- name: GetBrandByName :one
SELECT *
FROM catalog_brands
WHERE name = $1;

-- name: ListBrands :many
SELECT *
FROM catalog_brands
ORDER BY name;

-- name: UpdateBrand :one
UPDATE catalog_brands
SET
    name = $2,
    code = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SetBrandActive :one
UPDATE catalog_brands
SET
    is_active = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;