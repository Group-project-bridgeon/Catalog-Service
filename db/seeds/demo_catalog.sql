
BEGIN;

-- =========================================================
-- 1. CATEGORIES (3 categories)
-- =========================================================
INSERT INTO catalog_categories (
    id, parent_id, name, code, description, is_active
)
VALUES
    ('10000000-0000-4000-8000-000000000001', NULL, 'Men''s Wear',   'MEN_WEAR',   'Catalog category for men''s fashion products.', TRUE),
    ('10000000-0000-4000-8000-000000000002', NULL, 'Women''s Wear', 'WOMEN_WEAR', 'Catalog category for women''s fashion products.', TRUE),
    ('10000000-0000-4000-8000-000000000003', NULL, 'Kids Wear',     'KIDS_WEAR',  'Catalog category for kids fashion products.', TRUE)
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    description = EXCLUDED.description,
    is_active = TRUE,
    updated_at = NOW();

-- =========================================================
-- 2. BRANDS
-- =========================================================
INSERT INTO catalog_brands (
    id, name, code, is_active
)
VALUES
    ('11000000-0000-4000-8000-000000000001', 'Urban Loom',  'URBAN_LOOM',  TRUE),
    ('11000000-0000-4000-8000-000000000002', 'Aster Lane',  'ASTER_LANE',  TRUE),
    ('11000000-0000-4000-8000-000000000003', 'Little Trail','LITTLE_TRAIL',TRUE)
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    is_active = TRUE,
    updated_at = NOW();

-- =========================================================
-- 3. UNIT
-- =========================================================
INSERT INTO catalog_units (
    id, code, name, is_active
)
VALUES
    ('12000000-0000-4000-8000-000000000001', 'PCS', 'Piece', TRUE)
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    is_active = TRUE,
    updated_at = NOW();

-- =========================================================
-- 4. ATTRIBUTES
-- =========================================================
INSERT INTO catalog_attributes (id, name, code)
VALUES
    ('40000000-0000-4000-8000-000000000001', 'Size',  'SIZE'),
    ('40000000-0000-4000-8000-000000000002', 'Color', 'COLOR')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name;

-- =========================================================
-- 5. ATTRIBUTE VALUES
-- =========================================================
INSERT INTO catalog_attribute_values (id, attribute_id, value)
SELECT v.id::uuid, a.id, v.value
FROM (
    VALUES
        ('41000000-0000-4000-8000-000000000001', 'SIZE',  'S'),
        ('41000000-0000-4000-8000-000000000002', 'SIZE',  'M'),
        ('41000000-0000-4000-8000-000000000003', 'SIZE',  'L'),
        ('41000000-0000-4000-8000-000000000004', 'SIZE',  'XL'),
        ('41000000-0000-4000-8000-000000000005', 'COLOR', 'Blue'),
        ('41000000-0000-4000-8000-000000000006', 'COLOR', 'White'),
        ('41000000-0000-4000-8000-000000000007', 'COLOR', 'Navy'),
        ('41000000-0000-4000-8000-000000000008', 'COLOR', 'Black'),
        ('41000000-0000-4000-8000-000000000009', 'COLOR', 'Beige'),
        ('41000000-0000-4000-8000-000000000010', 'COLOR', 'Maroon'),
        ('41000000-0000-4000-8000-000000000011', 'COLOR', 'Pink'),
        ('41000000-0000-4000-8000-000000000012', 'COLOR', 'Green'),
        ('41000000-0000-4000-8000-000000000013', 'COLOR', 'Grey'),
        ('41000000-0000-4000-8000-000000000014', 'COLOR', 'Red')
) AS v(id, attribute_code, value)
JOIN catalog_attributes a
  ON a.code = v.attribute_code
ON CONFLICT (attribute_id, value) DO NOTHING;

-- =========================================================
-- 6. PRODUCTS (3 categories x 5 products = 15 products)
-- =========================================================
INSERT INTO catalog_products (
    id, category_id, brand_id, name, product_code, description, is_active
)
SELECT
    v.id::uuid,
    c.id,
    b.id,
    v.name,
    v.product_code,
    v.description,
    TRUE
FROM (
    VALUES
        -- Men's Wear
        ('20000000-0000-4000-8000-000000000001', 'MEN_WEAR',   'URBAN_LOOM',   'Cotton Casual Shirt',        'MEN-001',   'Breathable cotton casual shirt for everyday wear.'),
        ('20000000-0000-4000-8000-000000000002', 'MEN_WEAR',   'URBAN_LOOM',   'Oxford Formal Shirt',        'MEN-002',   'Classic Oxford formal shirt suitable for office and events.'),
        ('20000000-0000-4000-8000-000000000003', 'MEN_WEAR',   'URBAN_LOOM',   'Classic Polo T-Shirt',       'MEN-003',   'Soft cotton polo T-shirt with a clean classic fit.'),
        ('20000000-0000-4000-8000-000000000004', 'MEN_WEAR',   'URBAN_LOOM',   'Slim Fit Denim Jeans',       'MEN-004',   'Stretch denim jeans with a modern slim fit.'),
        ('20000000-0000-4000-8000-000000000005', 'MEN_WEAR',   'URBAN_LOOM',   'Stretch Chino Trousers',     'MEN-005',   'Comfort-stretch chino trousers for smart-casual outfits.'),

        -- Women's Wear
        ('20000000-0000-4000-8000-000000000006', 'WOMEN_WEAR', 'ASTER_LANE',   'Cotton Embroidered Kurti',   'WOMEN-001', 'Cotton kurti with lightweight embroidery and regular fit.'),
        ('20000000-0000-4000-8000-000000000007', 'WOMEN_WEAR', 'ASTER_LANE',   'Ribbed Casual Top',          'WOMEN-002', 'Soft ribbed casual top designed for everyday wear.'),
        ('20000000-0000-4000-8000-000000000008', 'WOMEN_WEAR', 'ASTER_LANE',   'Floral Maxi Dress',          'WOMEN-003', 'Flowing floral maxi dress with comfortable full-length styling.'),
        ('20000000-0000-4000-8000-000000000009', 'WOMEN_WEAR', 'ASTER_LANE',   'Wide-Leg Palazzo Pants',     'WOMEN-004', 'Relaxed wide-leg palazzo pants with an easy fit.'),
        ('20000000-0000-4000-8000-000000000010', 'WOMEN_WEAR', 'ASTER_LANE',   'Cropped Denim Jacket',       'WOMEN-005', 'Cropped denim jacket with a versatile casual silhouette.'),

        -- Kids Wear
        ('20000000-0000-4000-8000-000000000011', 'KIDS_WEAR',  'LITTLE_TRAIL', 'Printed Cotton T-Shirt',     'KIDS-001',  'Soft printed cotton T-shirt for everyday kids wear.'),
        ('20000000-0000-4000-8000-000000000012', 'KIDS_WEAR',  'LITTLE_TRAIL', 'Checked Cotton Shirt',       'KIDS-002',  'Comfortable checked cotton shirt with a regular fit.'),
        ('20000000-0000-4000-8000-000000000013', 'KIDS_WEAR',  'LITTLE_TRAIL', 'Denim Shorts',               'KIDS-003',  'Durable denim shorts designed for active daily wear.'),
        ('20000000-0000-4000-8000-000000000014', 'KIDS_WEAR',  'LITTLE_TRAIL', 'Zip-Up Hoodie',              'KIDS-004',  'Warm zip-up hoodie for cool-weather casual wear.'),
        ('20000000-0000-4000-8000-000000000015', 'KIDS_WEAR',  'LITTLE_TRAIL', 'Fleece Jogger Pants',        'KIDS-005',  'Soft fleece jogger pants with an elasticated waist.')
) AS v(id, category_code, brand_code, name, product_code, description)
JOIN catalog_categories c
  ON c.code = v.category_code
LEFT JOIN catalog_brands b
  ON b.code = v.brand_code
ON CONFLICT (product_code) DO UPDATE
SET category_id = EXCLUDED.category_id,
    brand_id = EXCLUDED.brand_id,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    is_active = TRUE,
    updated_at = NOW();

-- =========================================================
-- 7. PRODUCT VARIANTS (1 catalog SKU per product)
-- =========================================================
INSERT INTO catalog_product_variants (
    id, product_id, sku, default_selling_price, unit_id, is_active
)
SELECT
    v.id::uuid,
    p.id,
    v.sku,
    v.default_selling_price::decimal(14,2),
    u.id,
    TRUE
FROM (
    VALUES
        ('30000000-0000-4000-8000-000000000001', 'MEN-001',   'MEN-001-M-BLUE',    899.00, 'PCS'),
        ('30000000-0000-4000-8000-000000000002', 'MEN-002',   'MEN-002-L-WHITE',  1199.00, 'PCS'),
        ('30000000-0000-4000-8000-000000000003', 'MEN-003',   'MEN-003-M-NAVY',     799.00, 'PCS'),
        ('30000000-0000-4000-8000-000000000004', 'MEN-004',   'MEN-004-L-BLUE',    1499.00, 'PCS'),
        ('30000000-0000-4000-8000-000000000005', 'MEN-005',   'MEN-005-M-BEIGE',   1299.00, 'PCS'),
        ('30000000-0000-4000-8000-000000000006', 'WOMEN-001', 'WOMEN-001-M-MAROON',1099.00, 'PCS'),
        ('30000000-0000-4000-8000-000000000007', 'WOMEN-002', 'WOMEN-002-S-PINK',    699.00, 'PCS'),
        ('30000000-0000-4000-8000-000000000008', 'WOMEN-003', 'WOMEN-003-M-GREEN',  1599.00, 'PCS'),
        ('30000000-0000-4000-8000-000000000009', 'WOMEN-004', 'WOMEN-004-M-BLACK',   999.00, 'PCS'),
        ('30000000-0000-4000-8000-000000000010', 'WOMEN-005', 'WOMEN-005-S-BLUE',   1799.00, 'PCS'),
        ('30000000-0000-4000-8000-000000000011', 'KIDS-001',  'KIDS-001-S-RED',      499.00, 'PCS'),
        ('30000000-0000-4000-8000-000000000012', 'KIDS-002',  'KIDS-002-M-WHITE',    699.00, 'PCS'),
        ('30000000-0000-4000-8000-000000000013', 'KIDS-003',  'KIDS-003-M-BLUE',     799.00, 'PCS'),
        ('30000000-0000-4000-8000-000000000014', 'KIDS-004',  'KIDS-004-L-GREY',     999.00, 'PCS'),
        ('30000000-0000-4000-8000-000000000015', 'KIDS-005',  'KIDS-005-M-BLACK',    899.00, 'PCS')
) AS v(id, product_code, sku, default_selling_price, unit_code)
JOIN catalog_products p
  ON p.product_code = v.product_code
JOIN catalog_units u
  ON u.code = v.unit_code
ON CONFLICT (product_id, sku) DO UPDATE
SET default_selling_price = EXCLUDED.default_selling_price,
    unit_id = EXCLUDED.unit_id,
    is_active = TRUE,
    updated_at = NOW();

-- =========================================================
-- 8. VARIANT ATTRIBUTE VALUES (Size + Color)
-- =========================================================
INSERT INTO catalog_variant_values (
    variant_id, attribute_id, attribute_value_id
)
SELECT
    pv.id,
    a.id,
    av.id
FROM (
    VALUES
        ('MEN-001',   'MEN-001-M-BLUE',     'SIZE',  'M'),
        ('MEN-001',   'MEN-001-M-BLUE',     'COLOR', 'Blue'),
        ('MEN-002',   'MEN-002-L-WHITE',    'SIZE',  'L'),
        ('MEN-002',   'MEN-002-L-WHITE',    'COLOR', 'White'),
        ('MEN-003',   'MEN-003-M-NAVY',     'SIZE',  'M'),
        ('MEN-003',   'MEN-003-M-NAVY',     'COLOR', 'Navy'),
        ('MEN-004',   'MEN-004-L-BLUE',     'SIZE',  'L'),
        ('MEN-004',   'MEN-004-L-BLUE',     'COLOR', 'Blue'),
        ('MEN-005',   'MEN-005-M-BEIGE',    'SIZE',  'M'),
        ('MEN-005',   'MEN-005-M-BEIGE',    'COLOR', 'Beige'),
        ('WOMEN-001', 'WOMEN-001-M-MAROON', 'SIZE',  'M'),
        ('WOMEN-001', 'WOMEN-001-M-MAROON', 'COLOR', 'Maroon'),
        ('WOMEN-002', 'WOMEN-002-S-PINK',   'SIZE',  'S'),
        ('WOMEN-002', 'WOMEN-002-S-PINK',   'COLOR', 'Pink'),
        ('WOMEN-003', 'WOMEN-003-M-GREEN',  'SIZE',  'M'),
        ('WOMEN-003', 'WOMEN-003-M-GREEN',  'COLOR', 'Green'),
        ('WOMEN-004', 'WOMEN-004-M-BLACK',  'SIZE',  'M'),
        ('WOMEN-004', 'WOMEN-004-M-BLACK',  'COLOR', 'Black'),
        ('WOMEN-005', 'WOMEN-005-S-BLUE',   'SIZE',  'S'),
        ('WOMEN-005', 'WOMEN-005-S-BLUE',   'COLOR', 'Blue'),
        ('KIDS-001',  'KIDS-001-S-RED',      'SIZE',  'S'),
        ('KIDS-001',  'KIDS-001-S-RED',      'COLOR', 'Red'),
        ('KIDS-002',  'KIDS-002-M-WHITE',    'SIZE',  'M'),
        ('KIDS-002',  'KIDS-002-M-WHITE',    'COLOR', 'White'),
        ('KIDS-003',  'KIDS-003-M-BLUE',     'SIZE',  'M'),
        ('KIDS-003',  'KIDS-003-M-BLUE',     'COLOR', 'Blue'),
        ('KIDS-004',  'KIDS-004-L-GREY',     'SIZE',  'L'),
        ('KIDS-004',  'KIDS-004-L-GREY',     'COLOR', 'Grey'),
        ('KIDS-005',  'KIDS-005-M-BLACK',    'SIZE',  'M'),
        ('KIDS-005',  'KIDS-005-M-BLACK',    'COLOR', 'Black')
) AS m(product_code, sku, attribute_code, attribute_value)
JOIN catalog_products p
  ON p.product_code = m.product_code
JOIN catalog_product_variants pv
  ON pv.product_id = p.id
 AND pv.sku = m.sku
JOIN catalog_attributes a
  ON a.code = m.attribute_code
JOIN catalog_attribute_values av
  ON av.attribute_id = a.id
 AND av.value = m.attribute_value
ON CONFLICT (variant_id, attribute_id) DO UPDATE
SET attribute_value_id = EXCLUDED.attribute_value_id;

-- =========================================================
-- 9. PRICE LIST
-- =========================================================
INSERT INTO catalog_price_lists (
    id, name, code, currency_code, is_active
)
VALUES
    ('50000000-0000-4000-8000-000000000001', 'Wholesale INR', 'WHOLESALE_INR', 'INR', TRUE)
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    currency_code = EXCLUDED.currency_code,
    is_active = TRUE;

-- =========================================================
-- 10. WHOLESALE PRICES
-- =========================================================
INSERT INTO catalog_prices (
    id, price_list_id, variant_id, unit_price
)
SELECT
    m.id::uuid,
    pl.id,
    pv.id,
    m.unit_price::decimal(14,2)
FROM (
    VALUES
        ('60000000-0000-4000-8000-000000000001', 'MEN-001',   'MEN-001-M-BLUE',      760.00),
        ('60000000-0000-4000-8000-000000000002', 'MEN-002',   'MEN-002-L-WHITE',    1020.00),
        ('60000000-0000-4000-8000-000000000003', 'MEN-003',   'MEN-003-M-NAVY',      680.00),
        ('60000000-0000-4000-8000-000000000004', 'MEN-004',   'MEN-004-L-BLUE',     1275.00),
        ('60000000-0000-4000-8000-000000000005', 'MEN-005',   'MEN-005-M-BEIGE',    1100.00),
        ('60000000-0000-4000-8000-000000000006', 'WOMEN-001', 'WOMEN-001-M-MAROON',  935.00),
        ('60000000-0000-4000-8000-000000000007', 'WOMEN-002', 'WOMEN-002-S-PINK',    595.00),
        ('60000000-0000-4000-8000-000000000008', 'WOMEN-003', 'WOMEN-003-M-GREEN',  1360.00),
        ('60000000-0000-4000-8000-000000000009', 'WOMEN-004', 'WOMEN-004-M-BLACK',   850.00),
        ('60000000-0000-4000-8000-000000000010', 'WOMEN-005', 'WOMEN-005-S-BLUE',   1530.00),
        ('60000000-0000-4000-8000-000000000011', 'KIDS-001',  'KIDS-001-S-RED',      425.00),
        ('60000000-0000-4000-8000-000000000012', 'KIDS-002',  'KIDS-002-M-WHITE',    595.00),
        ('60000000-0000-4000-8000-000000000013', 'KIDS-003',  'KIDS-003-M-BLUE',     680.00),
        ('60000000-0000-4000-8000-000000000014', 'KIDS-004',  'KIDS-004-L-GREY',     850.00),
        ('60000000-0000-4000-8000-000000000015', 'KIDS-005',  'KIDS-005-M-BLACK',    765.00)
) AS m(id, product_code, sku, unit_price)
JOIN catalog_price_lists pl
  ON pl.code = 'WHOLESALE_INR'
JOIN catalog_products p
  ON p.product_code = m.product_code
JOIN catalog_product_variants pv
  ON pv.product_id = p.id
 AND pv.sku = m.sku
ON CONFLICT (price_list_id, variant_id) DO UPDATE
SET unit_price = EXCLUDED.unit_price,
    updated_at = NOW();

-- =========================================================
-- 11. PRODUCT IMAGES (one catalog image per product)
-- =========================================================
INSERT INTO catalog_product_images (
    id, product_id, variant_id, image_url, sort_order
)
SELECT
    m.id::uuid,
    p.id,
    NULL,
    m.image_url,
    0
FROM (
    VALUES
        ('70000000-0000-4000-8000-000000000001', 'MEN-001',   'https://placehold.co/800x1000?text=Cotton+Casual+Shirt'),
        ('70000000-0000-4000-8000-000000000002', 'MEN-002',   'https://placehold.co/800x1000?text=Oxford+Formal+Shirt'),
        ('70000000-0000-4000-8000-000000000003', 'MEN-003',   'https://placehold.co/800x1000?text=Classic+Polo+T-Shirt'),
        ('70000000-0000-4000-8000-000000000004', 'MEN-004',   'https://placehold.co/800x1000?text=Slim+Fit+Denim+Jeans'),
        ('70000000-0000-4000-8000-000000000005', 'MEN-005',   'https://placehold.co/800x1000?text=Stretch+Chino+Trousers'),
        ('70000000-0000-4000-8000-000000000006', 'WOMEN-001', 'https://placehold.co/800x1000?text=Cotton+Embroidered+Kurti'),
        ('70000000-0000-4000-8000-000000000007', 'WOMEN-002', 'https://placehold.co/800x1000?text=Ribbed+Casual+Top'),
        ('70000000-0000-4000-8000-000000000008', 'WOMEN-003', 'https://placehold.co/800x1000?text=Floral+Maxi+Dress'),
        ('70000000-0000-4000-8000-000000000009', 'WOMEN-004', 'https://placehold.co/800x1000?text=Wide-Leg+Palazzo+Pants'),
        ('70000000-0000-4000-8000-000000000010', 'WOMEN-005', 'https://placehold.co/800x1000?text=Cropped+Denim+Jacket'),
        ('70000000-0000-4000-8000-000000000011', 'KIDS-001',  'https://placehold.co/800x1000?text=Printed+Cotton+T-Shirt'),
        ('70000000-0000-4000-8000-000000000012', 'KIDS-002',  'https://placehold.co/800x1000?text=Checked+Cotton+Shirt'),
        ('70000000-0000-4000-8000-000000000013', 'KIDS-003',  'https://placehold.co/800x1000?text=Denim+Shorts'),
        ('70000000-0000-4000-8000-000000000014', 'KIDS-004',  'https://placehold.co/800x1000?text=Zip-Up+Hoodie'),
        ('70000000-0000-4000-8000-000000000015', 'KIDS-005',  'https://placehold.co/800x1000?text=Fleece+Jogger+Pants')
) AS m(id, product_code, image_url)
JOIN catalog_products p
  ON p.product_code = m.product_code
ON CONFLICT (id) DO UPDATE
SET product_id = EXCLUDED.product_id,
    variant_id = EXCLUDED.variant_id,
    image_url = EXCLUDED.image_url,
    sort_order = EXCLUDED.sort_order;

COMMIT;

-- =========================================================
-- VERIFICATION
-- Expected: 3 rows, each with product_count = 5
-- =========================================================
SELECT
    c.code,
    c.name,
    COUNT(p.id) AS product_count
FROM catalog_categories c
LEFT JOIN catalog_products p
       ON p.category_id = c.id
      AND (p.product_code LIKE 'MEN-%' OR p.product_code LIKE 'WOMEN-%' OR p.product_code LIKE 'KIDS-%')
WHERE c.code IN ('MEN_WEAR', 'WOMEN_WEAR', 'KIDS_WEAR')
GROUP BY c.code, c.name
ORDER BY c.code;

-- Expected total_seed_products = 15
SELECT COUNT(*) AS total_seed_products
FROM catalog_products
WHERE product_code LIKE 'MEN-%' OR product_code LIKE 'WOMEN-%' OR product_code LIKE 'KIDS-%';
