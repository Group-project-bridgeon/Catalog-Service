BEGIN;

CREATE INDEX idx_catalog_products_name_id
    ON catalog_products(name, id);

CREATE INDEX idx_catalog_products_category_name_id
    ON catalog_products(category_id, name, id);

CREATE INDEX idx_catalog_products_brand_name_id
    ON catalog_products(brand_id, name, id)
    WHERE brand_id IS NOT NULL;

CREATE INDEX idx_catalog_products_active_name_id
    ON catalog_products(name, id)
    WHERE is_active = TRUE;


CREATE INDEX idx_catalog_product_variants_product_sku_id
    ON catalog_product_variants(product_id, sku, id);

CREATE INDEX idx_catalog_product_variants_active_product_sku_id
    ON catalog_product_variants(product_id, sku, id)
    WHERE is_active = TRUE;


CREATE INDEX idx_catalog_prices_variant_price_list
    ON catalog_prices(variant_id, price_list_id);


UPDATE catalog_product_images
SET sort_order = 0
WHERE sort_order IS NULL;

ALTER TABLE catalog_product_images
ALTER COLUMN sort_order SET NOT NULL;

CREATE INDEX idx_catalog_product_images_product_sort_created_id
    ON catalog_product_images(
        product_id,
        sort_order,
        created_at,
        id
    );

CREATE INDEX idx_catalog_product_images_variant_sort_created_id
    ON catalog_product_images(
        variant_id,
        sort_order,
        created_at,
        id
    )
    WHERE variant_id IS NOT NULL;

COMMIT;