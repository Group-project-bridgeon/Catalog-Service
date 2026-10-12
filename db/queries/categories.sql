-- name: CreateCategory :one
INSERT INTO catalog_categories (
    id,
    parent_id,
    name,
    code,
    description,
    is_active
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetCategoryByID :one
SELECT *
FROM catalog_categories
WHERE id = $1;

-- name: GetCategoryByCode :one
SELECT *
FROM catalog_categories
WHERE code = $1;

-- name: ListCategories :many
SELECT *
FROM catalog_categories
ORDER BY name;

-- name: UpdateCategory :one
UPDATE catalog_categories
SET
    parent_id = $2,
    name = $3,
    code = $4,
    description = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SetCategoryActive :one
UPDATE catalog_categories
SET
    is_active = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;