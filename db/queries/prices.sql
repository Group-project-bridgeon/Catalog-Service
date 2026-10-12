-- name: CreatePrice :one
INSERT INTO catalog_prices (
    id,
    price_list_id,
    variant_id,
    unit_price
)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPriceByID :one
SELECT *
FROM catalog_prices
WHERE id = $1;

-- name: GetPriceByPriceListAndVariant :one
SELECT *
FROM catalog_prices
WHERE price_list_id = $1
  AND variant_id = $2;

-- name: ListPrices :many
SELECT *
FROM catalog_prices
WHERE
    (
        sqlc.narg(cursor_price_list_id)::uuid IS NULL
        OR (
            price_list_id,
            variant_id
        ) > (
            sqlc.narg(cursor_price_list_id)::uuid,
            sqlc.narg(cursor_variant_id)::uuid
        )
    )
    AND (
        sqlc.narg(price_list_id)::uuid IS NULL
        OR price_list_id = sqlc.narg(price_list_id)::uuid
    )
    AND (
        sqlc.narg(variant_id)::uuid IS NULL
        OR variant_id = sqlc.narg(variant_id)::uuid
    )
ORDER BY
    price_list_id,
    variant_id
LIMIT sqlc.arg(page_size)::int;

-- name: UpdatePrice :one
UPDATE catalog_prices
SET
    unit_price = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
