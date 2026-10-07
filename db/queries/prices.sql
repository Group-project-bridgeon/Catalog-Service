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
ORDER BY price_list_id, variant_id;

-- name: ListPricesByPriceListID :many
SELECT *
FROM catalog_prices
WHERE price_list_id = $1
ORDER BY variant_id;

-- name: ListPricesByVariantID :many
SELECT *
FROM catalog_prices
WHERE variant_id = $1
ORDER BY price_list_id;

-- name: UpdatePrice :one
UPDATE catalog_prices
SET
    unit_price = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
