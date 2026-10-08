BEGIN;

DROP INDEX IF EXISTS
    idx_catalog_product_images_variant_sort_created_id;

DROP INDEX IF EXISTS
    idx_catalog_product_images_product_sort_created_id;

ALTER TABLE catalog_product_images
ALTER COLUMN sort_order DROP NOT NULL;

DROP INDEX IF EXISTS
    idx_catalog_prices_variant_price_list;

DROP INDEX IF EXISTS
    idx_catalog_product_variants_active_product_sku_id;

DROP INDEX IF EXISTS
    idx_catalog_product_variants_product_sku_id;

DROP INDEX IF EXISTS
    idx_catalog_products_active_name_id;

DROP INDEX IF EXISTS
    idx_catalog_products_brand_name_id;

DROP INDEX IF EXISTS
    idx_catalog_products_category_name_id;

DROP INDEX IF EXISTS
    idx_catalog_products_name_id;

COMMIT;