-- name: ListProductVariantsForDetail :many
SELECT *
FROM catalog_product_variants
WHERE product_id = $1
ORDER BY
    sku,
    id;

-- name: ListVariantValuesForProductDetail :many
SELECT vv.*
FROM catalog_variant_values vv
JOIN catalog_product_variants pv
  ON pv.id = vv.variant_id
WHERE pv.product_id = $1
ORDER BY
    vv.variant_id,
    vv.attribute_id;

-- name: ListProductImagesForDetail :many
SELECT *
FROM catalog_product_images
WHERE product_id = $1
ORDER BY
    sort_order,
    created_at,
    id;

-- name: ListPricesForProductDetail :many
SELECT p.*
FROM catalog_prices p
JOIN catalog_product_variants pv
  ON pv.id = p.variant_id
WHERE pv.product_id = $1
ORDER BY
    p.price_list_id,
    p.variant_id;
