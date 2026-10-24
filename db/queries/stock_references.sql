-- name: ClassifyPieceUnit :execrows
INSERT INTO catalog_piece_units(unit_id) VALUES($1) ON CONFLICT(unit_id) DO NOTHING;

-- name: EnableVariantStock :execrows
INSERT INTO catalog_stock_eligibility(variant_id,base_unit_id) VALUES($1,$2);

-- name: LockStockVariant :one
SELECT id,unit_id FROM catalog_product_variants WHERE id=$1 FOR UPDATE;

-- name: CreateVariantPack :one
INSERT INTO catalog_variant_packs(id,variant_id,code,is_active) VALUES($1,$2,$3,$4) RETURNING *;

-- name: LockVariantPack :one
SELECT * FROM catalog_variant_packs WHERE id=$1 FOR UPDATE;

-- name: InsertPackRevision :one
INSERT INTO catalog_pack_revisions(pack_id,revision,pieces_per_pack,label) VALUES($1,$2,$3,$4) RETURNING *;

-- name: AdvancePackRevision :one
UPDATE catalog_variant_packs SET current_revision=current_revision+1 WHERE id=$1 AND current_revision=$2 RETURNING *;

-- name: SetVariantPackActive :one
UPDATE catalog_variant_packs SET is_active=$3 WHERE id=$1 AND stock_version=$2 RETURNING *;

-- name: GetPackRevision :one
SELECT p.id,p.variant_id,p.code,p.is_active,p.current_revision,p.stock_version,r.revision,r.pieces_per_pack,r.label
FROM catalog_variant_packs p JOIN catalog_pack_revisions r ON r.pack_id=p.id WHERE p.id=$1 AND r.revision=$2;

-- name: StockObservationTime :one
SELECT statement_timestamp()::timestamptz AS checked_at;

-- name: GetStockReferences :many
SELECT v.id AS variant_id,v.product_id,v.unit_id AS base_unit_id,v.sku,
    v.is_active AS variant_active,v.stock_version AS variant_version,
    p.product_code,p.name AS product_name,p.is_active AS product_active,p.stock_version AS product_version,
    u.code AS base_unit_code,u.name AS base_unit_name,u.is_active AS base_unit_active,u.stock_version AS unit_version,
    (e.variant_id IS NOT NULL)::boolean AS eligible,COALESCE(s.quantity_kind,'UNCLASSIFIED')::text AS quantity_kind
FROM catalog_product_variants v JOIN catalog_products p ON p.id=v.product_id JOIN catalog_units u ON u.id=v.unit_id
LEFT JOIN catalog_stock_eligibility e ON e.variant_id=v.id AND e.base_unit_id=v.unit_id
LEFT JOIN catalog_piece_units s ON s.unit_id=v.unit_id
WHERE v.id=ANY($1::uuid[]) ORDER BY v.id;

-- name: GetSelectedPackReferences :many
SELECT p.id,p.variant_id,p.code,p.is_active,p.current_revision,p.stock_version,
    wanted.revision::bigint AS requested_revision,r.revision,r.pieces_per_pack,r.label
FROM (SELECT unnest(sqlc.arg(pack_ids)::uuid[]) AS id,unnest(sqlc.arg(revisions)::bigint[]) AS revision) AS wanted
JOIN catalog_variant_packs p ON p.id=wanted.id
LEFT JOIN catalog_pack_revisions r ON r.pack_id=p.id AND r.revision=wanted.revision;
