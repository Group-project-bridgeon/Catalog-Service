-- name: CreateAttributeValue :one
INSERT INTO catalog_attribute_values (
    id,
    attribute_id,
    value
)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetAttributeValueByID :one
SELECT *
FROM catalog_attribute_values
WHERE id = $1;

-- name: GetAttributeValueByAttributeAndValue :one
SELECT *
FROM catalog_attribute_values
WHERE attribute_id = $1
  AND value = $2;

-- name: ListAttributeValues :many
SELECT *
FROM catalog_attribute_values
ORDER BY attribute_id, value;

-- name: ListAttributeValuesByAttributeID :many
SELECT *
FROM catalog_attribute_values
WHERE attribute_id = $1
ORDER BY value;

-- name: UpdateAttributeValue :one
UPDATE catalog_attribute_values
SET
    value = $2
WHERE id = $1
RETURNING *;
