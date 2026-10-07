-- name: CreateProductImage :one
INSERT INTO catalog_product_images (
    id,
    product_id,
    variant_id,
    image_url,
    sort_order
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetProductImageByID :one
SELECT *
FROM catalog_product_images
WHERE id = $1;

-- name: ListProductImages :many
SELECT *
FROM catalog_product_images
ORDER BY
    product_id,
    COALESCE(sort_order, 0),
    created_at,
    id;

-- name: ListProductImagesByProductID :many
SELECT *
FROM catalog_product_images
WHERE product_id = $1
ORDER BY
    COALESCE(sort_order, 0),
    created_at,
    id;

-- name: ListProductImagesByVariantID :many
SELECT *
FROM catalog_product_images
WHERE variant_id = $1
ORDER BY
    COALESCE(sort_order, 0),
    created_at,
    id;

-- name: UpdateProductImage :one
UPDATE catalog_product_images
SET
    variant_id = $2,
    image_url = $3,
    sort_order = $4
WHERE id = $1
RETURNING *;

-- name: ProductImageVariantBelongsToProduct :one
SELECT EXISTS (
    SELECT 1
    FROM catalog_product_variants
    WHERE product_id = sqlc.arg(product_id)::uuid
      AND id = sqlc.arg(variant_id)::uuid
);
