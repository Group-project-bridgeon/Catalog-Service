BEGIN;

CREATE TABLE catalog_categories (
    id UUID PRIMARY KEY,
    parent_id UUID,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_catalog_categories_parent
        FOREIGN KEY (parent_id)
        REFERENCES catalog_categories(id)
);

CREATE INDEX idx_catalog_categories_parent_id
    ON catalog_categories(parent_id);


CREATE TABLE catalog_brands (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    code VARCHAR(50) NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TABLE catalog_units (
    id UUID PRIMARY KEY,
    code VARCHAR(30) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TABLE catalog_attributes (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    code VARCHAR(50) NOT NULL UNIQUE
);


CREATE TABLE catalog_products (
    id UUID PRIMARY KEY,
    category_id UUID NOT NULL,
    brand_id UUID,
    name VARCHAR(200) NOT NULL,
    product_code VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_catalog_products_category
        FOREIGN KEY (category_id)
        REFERENCES catalog_categories(id),

    CONSTRAINT fk_catalog_products_brand
        FOREIGN KEY (brand_id)
        REFERENCES catalog_brands(id)
);

CREATE INDEX idx_catalog_products_category_id
    ON catalog_products(category_id);

CREATE INDEX idx_catalog_products_brand_id
    ON catalog_products(brand_id);


CREATE TABLE catalog_product_variants (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL,
    sku VARCHAR(100) NOT NULL,
    default_selling_price DECIMAL(14, 2),
    unit_id UUID NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_catalog_product_variants_product
        FOREIGN KEY (product_id)
        REFERENCES catalog_products(id),

    CONSTRAINT fk_catalog_product_variants_unit
        FOREIGN KEY (unit_id)
        REFERENCES catalog_units(id),

    CONSTRAINT uq_catalog_product_variants_product_sku
        UNIQUE (product_id, sku),

    CONSTRAINT chk_catalog_product_variants_price
        CHECK (
            default_selling_price IS NULL
            OR default_selling_price >= 0
        )
);

CREATE INDEX idx_catalog_product_variants_sku
    ON catalog_product_variants(sku);

CREATE INDEX idx_catalog_product_variants_unit_id
    ON catalog_product_variants(unit_id);


CREATE TABLE catalog_attribute_values (
    id UUID PRIMARY KEY,
    attribute_id UUID NOT NULL,
    value VARCHAR(100) NOT NULL,

    CONSTRAINT fk_catalog_attribute_values_attribute
        FOREIGN KEY (attribute_id)
        REFERENCES catalog_attributes(id),

    CONSTRAINT uq_catalog_attribute_values_attribute_value
        UNIQUE (attribute_id, value),

    CONSTRAINT uq_catalog_attribute_values_attribute_id
        UNIQUE (attribute_id, id)
);


CREATE TABLE catalog_variant_values (
    variant_id UUID NOT NULL,
    attribute_id UUID NOT NULL,
    attribute_value_id UUID NOT NULL,

    CONSTRAINT pk_catalog_variant_values
        PRIMARY KEY (variant_id, attribute_id),

    CONSTRAINT fk_catalog_variant_values_variant
        FOREIGN KEY (variant_id)
        REFERENCES catalog_product_variants(id),

    CONSTRAINT fk_catalog_variant_values_attribute_value
        FOREIGN KEY (attribute_id, attribute_value_id)
        REFERENCES catalog_attribute_values(attribute_id, id)
);

CREATE INDEX idx_catalog_variant_values_attribute_value_id
    ON catalog_variant_values(attribute_value_id);


CREATE TABLE catalog_product_images (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL,
    variant_id UUID,
    image_url TEXT NOT NULL,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_catalog_product_images_product
        FOREIGN KEY (product_id)
        REFERENCES catalog_products(id),

    CONSTRAINT fk_catalog_product_images_variant
        FOREIGN KEY (variant_id)
        REFERENCES catalog_product_variants(id)
);

CREATE INDEX idx_catalog_product_images_product_id
    ON catalog_product_images(product_id);

CREATE INDEX idx_catalog_product_images_variant_id
    ON catalog_product_images(variant_id);


CREATE TABLE catalog_price_lists (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    currency_code VARCHAR(3) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);


CREATE TABLE catalog_prices (
    id UUID PRIMARY KEY,
    price_list_id UUID NOT NULL,
    variant_id UUID NOT NULL,
    unit_price DECIMAL(14, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_catalog_prices_price_list
        FOREIGN KEY (price_list_id)
        REFERENCES catalog_price_lists(id),

    CONSTRAINT fk_catalog_prices_variant
        FOREIGN KEY (variant_id)
        REFERENCES catalog_product_variants(id),

    CONSTRAINT uq_catalog_prices_price_list_variant
        UNIQUE (price_list_id, variant_id),

    CONSTRAINT chk_catalog_prices_unit_price
        CHECK (unit_price >= 0)
);

CREATE INDEX idx_catalog_prices_variant_id
    ON catalog_prices(variant_id);

COMMIT;