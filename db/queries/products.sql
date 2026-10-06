-- name: CreateProduct :one
INSERT INTO catalog_products (
    id,
    category_id,
    brand_id,
    name,
    product_code,
    description,
    is_active
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetProductByID :one
SELECT *
FROM catalog_products
WHERE id = $1;

-- name: GetProductByCode :one
SELECT *
FROM catalog_products
WHERE product_code = $1;

-- name: ListProducts :many
SELECT *
FROM catalog_products
ORDER BY name;

-- name: UpdateProduct :one
UPDATE catalog_products
SET
    category_id = $2,
    brand_id = $3,
    name = $4,
    product_code = $5,
    description = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SetProductActive :one
UPDATE catalog_products
SET
    is_active = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
