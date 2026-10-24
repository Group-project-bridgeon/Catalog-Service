-- Destructive to new eligibility/pack history; preserve legacy master rows.
BEGIN;
ALTER TABLE catalog_variant_packs DROP CONSTRAINT catalog_pack_current_fk;
DROP TABLE catalog_pack_revisions;
DROP TABLE catalog_variant_packs;
DROP TABLE catalog_stock_eligibility, catalog_piece_units;
DROP TRIGGER catalog_product_stock_version ON catalog_products;
DROP TRIGGER catalog_variant_stock_version ON catalog_product_variants;
DROP TRIGGER catalog_unit_stock_version ON catalog_units;
DROP FUNCTION catalog_pack_complete(), catalog_pack_guard(), catalog_stock_classify_guard(), catalog_stock_version_guard(), catalog_stock_immutable();
ALTER TABLE catalog_product_variants DROP CONSTRAINT catalog_stock_variant_unit_unique;
ALTER TABLE catalog_product_variants DROP COLUMN stock_version;
ALTER TABLE catalog_products DROP COLUMN stock_version;
ALTER TABLE catalog_units DROP COLUMN stock_version;
COMMIT;
