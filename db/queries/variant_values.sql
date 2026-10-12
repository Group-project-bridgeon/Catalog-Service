-- name: CreateVariantValue :one
INSERT INTO catalog_variant_values (
    variant_id,
    attribute_id,
    attribute_value_id
)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetVariantValueByVariantAndAttribute :one
SELECT *
FROM catalog_variant_values
WHERE variant_id = $1
  AND attribute_id = $2;

-- name: ListVariantValues :many
SELECT *
FROM catalog_variant_values
WHERE
    (
        sqlc.narg(cursor_variant_id)::uuid IS NULL
        OR (
            variant_id,
            attribute_id
        ) > (
            sqlc.narg(cursor_variant_id)::uuid,
            sqlc.narg(cursor_attribute_id)::uuid
        )
    )
    AND (
        sqlc.narg(variant_id)::uuid IS NULL
        OR variant_id = sqlc.narg(variant_id)::uuid
    )
ORDER BY
    variant_id,
    attribute_id
LIMIT sqlc.arg(page_size)::int;

-- name: UpdateVariantValue :one
UPDATE catalog_variant_values
SET attribute_value_id = $3
WHERE variant_id = $1
  AND attribute_id = $2
RETURNING *;

-- name: VariantCombinationExists :one
WITH candidate_values AS (
    SELECT
        vv.attribute_id,
        vv.attribute_value_id
    FROM catalog_variant_values vv
    WHERE vv.variant_id = @variant_id
      AND vv.attribute_id <> @attribute_id

    UNION ALL

    SELECT
        @attribute_id::uuid AS attribute_id,
        @attribute_value_id::uuid AS attribute_value_id
),
target_product AS (
    SELECT product_id
    FROM catalog_product_variants
    WHERE id = @variant_id
),
other_variants AS (
    SELECT pv.id
    FROM catalog_product_variants pv
    JOIN target_product tp
      ON tp.product_id = pv.product_id
    WHERE pv.id <> @variant_id
)
SELECT EXISTS (
    SELECT 1
    FROM other_variants ov
    WHERE (
        SELECT COUNT(*)
        FROM catalog_variant_values existing
        WHERE existing.variant_id = ov.id
    ) = (
        SELECT COUNT(*)
        FROM candidate_values
    )
    AND NOT EXISTS (
        SELECT 1
        FROM candidate_values candidate
        WHERE NOT EXISTS (
            SELECT 1
            FROM catalog_variant_values existing
            WHERE existing.variant_id = ov.id
            AND existing.attribute_id = candidate.attribute_id
            AND existing.attribute_value_id = candidate.attribute_value_id
        )
    )
);
