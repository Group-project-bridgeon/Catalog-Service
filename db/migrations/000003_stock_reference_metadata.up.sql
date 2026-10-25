BEGIN;

ALTER TABLE catalog_products ADD COLUMN stock_version BIGINT NOT NULL DEFAULT 1 CHECK (stock_version>0);
ALTER TABLE catalog_product_variants ADD COLUMN stock_version BIGINT NOT NULL DEFAULT 1 CHECK (stock_version>0);
ALTER TABLE catalog_units ADD COLUMN stock_version BIGINT NOT NULL DEFAULT 1 CHECK (stock_version>0);
ALTER TABLE catalog_product_variants ADD CONSTRAINT catalog_stock_variant_unit_unique UNIQUE(id,unit_id);

CREATE TABLE catalog_piece_units (
    unit_id UUID PRIMARY KEY REFERENCES catalog_units(id) ON DELETE RESTRICT,
    quantity_kind TEXT NOT NULL DEFAULT 'PIECE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT catalog_piece_kind CHECK (quantity_kind='PIECE'),
    CONSTRAINT catalog_piece_id CHECK (unit_id<>'00000000-0000-0000-0000-000000000000')
);
CREATE TABLE catalog_stock_eligibility (
    variant_id UUID PRIMARY KEY,
    base_unit_id UUID NOT NULL REFERENCES catalog_piece_units(unit_id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT catalog_stock_binding_fk FOREIGN KEY(variant_id,base_unit_id)
        REFERENCES catalog_product_variants(id,unit_id) ON DELETE RESTRICT,
    CONSTRAINT catalog_stock_variant_nonzero CHECK (variant_id<>'00000000-0000-0000-0000-000000000000')
);
CREATE INDEX catalog_stock_eligibility_unit ON catalog_stock_eligibility(base_unit_id);

CREATE TABLE catalog_variant_packs (
    id UUID PRIMARY KEY,
    variant_id UUID NOT NULL REFERENCES catalog_stock_eligibility(variant_id) ON DELETE RESTRICT,
    code TEXT NOT NULL,
    current_revision BIGINT NOT NULL DEFAULT 1,
    stock_version BIGINT NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT catalog_pack_code_unique UNIQUE(variant_id,code),
    CONSTRAINT catalog_pack_code_valid CHECK (code ~ '^[A-Z0-9][A-Z0-9_-]{0,49}$'),
    CONSTRAINT catalog_pack_versions_positive CHECK (current_revision>0 AND stock_version>0),
    CONSTRAINT catalog_pack_id_valid CHECK (id<>'00000000-0000-0000-0000-000000000000')
);
CREATE TABLE catalog_pack_revisions (
    pack_id UUID NOT NULL REFERENCES catalog_variant_packs(id) ON DELETE RESTRICT,
    revision BIGINT NOT NULL,
    pieces_per_pack BIGINT NOT NULL,
    label TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY(pack_id,revision),
    CONSTRAINT catalog_pack_revision_positive CHECK (revision>0),
    CONSTRAINT catalog_pack_factor_positive CHECK (pieces_per_pack>0),
    CONSTRAINT catalog_pack_label_valid CHECK (char_length(btrim(label)) BETWEEN 1 AND 200 AND label !~ '[[:cntrl:]]')
);
ALTER TABLE catalog_variant_packs ADD CONSTRAINT catalog_pack_current_fk
    FOREIGN KEY(id,current_revision) REFERENCES catalog_pack_revisions(pack_id,revision)
    DEFERRABLE INITIALLY DEFERRED;

CREATE FUNCTION catalog_stock_immutable() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'immutable Catalog stock definition' USING ERRCODE='23514',CONSTRAINT='catalog_stock_immutable';
END $$;
CREATE TRIGGER catalog_piece_immutable BEFORE UPDATE OR DELETE OR TRUNCATE ON catalog_piece_units
    FOR EACH STATEMENT EXECUTE FUNCTION catalog_stock_immutable();
CREATE TRIGGER catalog_eligibility_immutable BEFORE UPDATE OR DELETE OR TRUNCATE ON catalog_stock_eligibility
    FOR EACH STATEMENT EXECUTE FUNCTION catalog_stock_immutable();
CREATE TRIGGER catalog_pack_revision_immutable BEFORE UPDATE OR DELETE OR TRUNCATE ON catalog_pack_revisions
    FOR EACH STATEMENT EXECUTE FUNCTION catalog_stock_immutable();
CREATE TRIGGER catalog_pack_no_remove BEFORE DELETE OR TRUNCATE ON catalog_variant_packs
    FOR EACH STATEMENT EXECUTE FUNCTION catalog_stock_immutable();

CREATE FUNCTION catalog_stock_version_guard() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE changed BOOLEAN;
BEGIN
    IF TG_OP='INSERT' THEN
        IF NEW.stock_version<>1 THEN
            RAISE EXCEPTION 'stock versions are server owned' USING ERRCODE='23514',CONSTRAINT='catalog_stock_version_owned';
        END IF;
        RETURN NEW;
    END IF;
    IF TG_TABLE_NAME='catalog_products' THEN
        changed := (NEW.id,NEW.product_code,NEW.name,NEW.is_active) IS DISTINCT FROM (OLD.id,OLD.product_code,OLD.name,OLD.is_active);
    ELSIF TG_TABLE_NAME='catalog_units' THEN
        changed := (NEW.id,NEW.code,NEW.name,NEW.is_active) IS DISTINCT FROM (OLD.id,OLD.code,OLD.name,OLD.is_active);
    ELSE
        IF (NEW.id,NEW.product_id,NEW.unit_id) IS DISTINCT FROM (OLD.id,OLD.product_id,OLD.unit_id)
           AND EXISTS(SELECT 1 FROM catalog_stock_eligibility WHERE variant_id=OLD.id) THEN
            RAISE EXCEPTION 'eligible variant binding is immutable' USING ERRCODE='23514',CONSTRAINT='catalog_stock_immutable';
        END IF;
        changed := (NEW.id,NEW.product_id,NEW.unit_id,NEW.sku,NEW.is_active) IS DISTINCT FROM (OLD.id,OLD.product_id,OLD.unit_id,OLD.sku,OLD.is_active);
    END IF;
    IF NEW.stock_version<>OLD.stock_version THEN
        RAISE EXCEPTION 'stock versions are server owned' USING ERRCODE='23514',CONSTRAINT='catalog_stock_version_owned';
    END IF;
    IF changed THEN
        IF OLD.stock_version=9223372036854775807 THEN
            RAISE EXCEPTION 'stock version exhausted' USING ERRCODE='23514',CONSTRAINT='catalog_stock_version_exhausted';
        END IF;
        NEW.stock_version:=OLD.stock_version+1;
    END IF;
    RETURN NEW;
END $$;
CREATE TRIGGER catalog_product_stock_version BEFORE INSERT OR UPDATE ON catalog_products FOR EACH ROW EXECUTE FUNCTION catalog_stock_version_guard();
CREATE TRIGGER catalog_variant_stock_version BEFORE INSERT OR UPDATE ON catalog_product_variants FOR EACH ROW EXECUTE FUNCTION catalog_stock_version_guard();
CREATE TRIGGER catalog_unit_stock_version BEFORE INSERT OR UPDATE ON catalog_units FOR EACH ROW EXECUTE FUNCTION catalog_stock_version_guard();

CREATE FUNCTION catalog_stock_classify_guard() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE v catalog_product_variants%ROWTYPE; u catalog_units%ROWTYPE; active_product BOOLEAN;
BEGIN
    IF TG_TABLE_NAME='catalog_piece_units' THEN
        SELECT * INTO u FROM catalog_units WHERE id=NEW.unit_id FOR UPDATE;
        IF NOT FOUND OR NOT u.is_active THEN
            RAISE EXCEPTION 'active unit required' USING ERRCODE='23514',CONSTRAINT='catalog_stock_classification_invalid';
        END IF;
    ELSE
        -- Serializes eligibility establishment with direct binding updates too.
        SELECT * INTO v FROM catalog_product_variants WHERE id=NEW.variant_id FOR UPDATE;
        IF NOT FOUND THEN
            RAISE EXCEPTION 'variant required' USING ERRCODE='23514',CONSTRAINT='catalog_stock_classification_invalid';
        END IF;
        SELECT * INTO u FROM catalog_units WHERE id=v.unit_id;
        SELECT is_active INTO active_product FROM catalog_products WHERE id=v.product_id;
        IF NOT v.is_active OR NOT u.is_active OR NOT active_product OR v.unit_id<>NEW.base_unit_id
           OR v.product_id='00000000-0000-0000-0000-000000000000' THEN
            RAISE EXCEPTION 'active matching piece binding required' USING ERRCODE='23514',CONSTRAINT='catalog_stock_classification_invalid';
        END IF;
    END IF;
    RETURN NEW;
END $$;
CREATE TRIGGER catalog_piece_insert BEFORE INSERT ON catalog_piece_units FOR EACH ROW EXECUTE FUNCTION catalog_stock_classify_guard();
CREATE TRIGGER catalog_eligibility_insert BEFORE INSERT ON catalog_stock_eligibility FOR EACH ROW EXECUTE FUNCTION catalog_stock_classify_guard();

CREATE FUNCTION catalog_pack_guard() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE p catalog_variant_packs%ROWTYPE;
BEGIN
    IF TG_TABLE_NAME='catalog_pack_revisions' THEN
        SELECT * INTO p FROM catalog_variant_packs WHERE id=NEW.pack_id FOR UPDATE;
        IF NOT FOUND OR NOT (
            (NEW.revision=1 AND p.current_revision=1 AND NOT EXISTS(SELECT 1 FROM catalog_pack_revisions WHERE pack_id=p.id)) OR
            (p.current_revision<9223372036854775807 AND NEW.revision::numeric=p.current_revision::numeric+1
                AND EXISTS(SELECT 1 FROM catalog_pack_revisions WHERE pack_id=p.id AND revision=p.current_revision))) THEN
            RAISE EXCEPTION 'invalid pack revision sequence' USING ERRCODE='23514',CONSTRAINT='catalog_pack_transition';
        END IF;
        RETURN NEW;
    END IF;
    IF TG_OP='INSERT' THEN
        IF NEW.current_revision<>1 OR NEW.stock_version<>1 THEN
            RAISE EXCEPTION 'invalid initial pack version' USING ERRCODE='23514',CONSTRAINT='catalog_pack_transition';
        END IF;
    ELSE
        IF (NEW.id,NEW.variant_id,NEW.code,NEW.created_at) IS DISTINCT FROM (OLD.id,OLD.variant_id,OLD.code,OLD.created_at)
            OR NEW.stock_version<>OLD.stock_version THEN
            RAISE EXCEPTION 'immutable pack identity' USING ERRCODE='23514',CONSTRAINT='catalog_stock_immutable';
        END IF;
        IF NEW.current_revision<>OLD.current_revision AND (
            NEW.current_revision::numeric<>OLD.current_revision::numeric+1 OR
            NOT EXISTS(SELECT 1 FROM catalog_pack_revisions WHERE pack_id=OLD.id AND revision=NEW.current_revision)) THEN
            RAISE EXCEPTION 'invalid pack head advance' USING ERRCODE='23514',CONSTRAINT='catalog_pack_transition';
        END IF;
        IF (NEW.current_revision,NEW.is_active) IS DISTINCT FROM (OLD.current_revision,OLD.is_active) THEN
            IF OLD.stock_version=9223372036854775807 THEN
                RAISE EXCEPTION 'pack version exhausted' USING ERRCODE='23514',CONSTRAINT='catalog_stock_version_exhausted';
            END IF;
            NEW.stock_version:=OLD.stock_version+1;
        END IF;
    END IF;
    RETURN NEW;
END $$;
CREATE TRIGGER catalog_pack_head_guard BEFORE INSERT OR UPDATE ON catalog_variant_packs FOR EACH ROW EXECUTE FUNCTION catalog_pack_guard();
CREATE TRIGGER catalog_pack_revision_guard BEFORE INSERT ON catalog_pack_revisions FOR EACH ROW EXECUTE FUNCTION catalog_pack_guard();

CREATE FUNCTION catalog_pack_complete() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE pack UUID; head BIGINT; latest BIGINT;
BEGIN
    IF TG_TABLE_NAME='catalog_variant_packs' THEN pack:=NEW.id; ELSE pack:=NEW.pack_id; END IF;
    SELECT current_revision INTO head FROM catalog_variant_packs WHERE id=pack;
    SELECT revision INTO latest FROM catalog_pack_revisions WHERE pack_id=pack ORDER BY revision DESC LIMIT 1;
    IF latest IS NULL OR head<>latest THEN
        RAISE EXCEPTION 'pack head and revisions disagree' USING ERRCODE='23514',CONSTRAINT='catalog_pack_complete';
    END IF;
    RETURN NULL;
END $$;
CREATE CONSTRAINT TRIGGER catalog_pack_head_complete AFTER INSERT OR UPDATE ON catalog_variant_packs
    DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION catalog_pack_complete();
CREATE CONSTRAINT TRIGGER catalog_pack_revision_complete AFTER INSERT ON catalog_pack_revisions
    DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION catalog_pack_complete();
COMMIT;
