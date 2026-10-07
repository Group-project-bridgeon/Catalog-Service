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
ORDER BY product_id, sku;

-- name: ListProductVariantsByProductID :many
SELECT *
FROM catalog_product_variants
WHERE product_id = $1
ORDER BY sku;

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
