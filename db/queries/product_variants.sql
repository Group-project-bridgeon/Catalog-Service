-- name: CreateProductVariant :one
INSERT INTO catalog_product_variants (
    id,
    product_id,
    sku,
    default_selling_price,
    unit_id,
    is_active
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetProductVariantByID :one
SELECT *
FROM catalog_product_variants
WHERE id = $1;

-- name: GetProductVariantByProductAndSKU :one
SELECT *
FROM catalog_product_variants
WHERE product_id = $1
  AND sku = $2;

-- name: ListProductVariants :many
SELECT *
FROM catalog_product_variants
WHERE
    (
        sqlc.narg(cursor_product_id)::uuid IS NULL
        OR (
            product_id,
            sku,
            id
        ) > (
            sqlc.narg(cursor_product_id)::uuid,
            sqlc.narg(cursor_sku)::text,
            sqlc.narg(cursor_id)::uuid
        )
    )
    AND (
        sqlc.narg(product_id)::uuid IS NULL
        OR product_id = sqlc.narg(product_id)::uuid
    )
    AND (
        sqlc.narg(unit_id)::uuid IS NULL
        OR unit_id = sqlc.narg(unit_id)::uuid
    )
    AND (
        sqlc.narg(is_active)::boolean IS NULL
        OR is_active = sqlc.narg(is_active)::boolean
    )
ORDER BY
    product_id,
    sku,
    id
LIMIT sqlc.arg(page_size)::int;

-- name: ListProductVariantsByIDs :many
SELECT *
FROM catalog_product_variants
WHERE id = ANY(sqlc.arg(ids)::uuid[])
ORDER BY
    product_id,
    sku,
    id;

-- name: UpdateProductVariant :one
UPDATE catalog_product_variants
SET
    sku = $2,
    default_selling_price = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SetProductVariantActive :one
UPDATE catalog_product_variants
SET
    is_active = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
