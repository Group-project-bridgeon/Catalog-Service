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
WHERE
    (
        sqlc.narg(cursor_product_id)::uuid IS NULL
        OR (
            product_id,
            sort_order,
            created_at,
            id
        ) > (
            sqlc.narg(cursor_product_id)::uuid,
            sqlc.narg(cursor_sort_order)::int,
            sqlc.narg(cursor_created_at)::timestamptz,
            sqlc.narg(cursor_id)::uuid
        )
    )
    AND (
        sqlc.narg(product_id)::uuid IS NULL
        OR product_id = sqlc.narg(product_id)::uuid
    )
    AND (
        sqlc.narg(variant_id)::uuid IS NULL
        OR variant_id = sqlc.narg(variant_id)::uuid
    )
ORDER BY
    product_id,
    sort_order,
    created_at,
    id
LIMIT sqlc.arg(page_size)::int;

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
