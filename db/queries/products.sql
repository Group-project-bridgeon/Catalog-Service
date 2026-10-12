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
WHERE
    (
        sqlc.narg(cursor_name)::text IS NULL
        OR (
            name,
            id
        ) > (
            sqlc.narg(cursor_name)::text,
            sqlc.narg(cursor_id)::uuid
        )
    )
    AND (
        sqlc.narg(category_id)::uuid IS NULL
        OR category_id = sqlc.narg(category_id)::uuid
    )
    AND (
        sqlc.narg(brand_id)::uuid IS NULL
        OR brand_id = sqlc.narg(brand_id)::uuid
    )
    AND (
        sqlc.narg(is_active)::boolean IS NULL
        OR is_active = sqlc.narg(is_active)::boolean
    )
ORDER BY
    name,
    id
LIMIT sqlc.arg(page_size)::int;

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
