-- name: CreatePriceList :one
INSERT INTO catalog_price_lists (
    id,
    name,
    code,
    currency_code,
    is_active
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetPriceListByID :one
SELECT *
FROM catalog_price_lists
WHERE id = $1;

-- name: GetPriceListByCode :one
SELECT *
FROM catalog_price_lists
WHERE code = $1;

-- name: ListPriceLists :many
SELECT *
FROM catalog_price_lists
ORDER BY name, code;

-- name: UpdatePriceList :one
UPDATE catalog_price_lists
SET
    name = $2,
    code = $3,
    currency_code = $4
WHERE id = $1
RETURNING *;

-- name: SetPriceListActive :one
UPDATE catalog_price_lists
SET is_active = $2
WHERE id = $1
RETURNING *;
