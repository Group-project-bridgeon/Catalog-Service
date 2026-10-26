package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository/postgres"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestLocalCatalogProvisioning(t *testing.T) {
	raw := os.Getenv("LOCAL_PROVISION_CATALOG_URL")
	if raw == "" {
		t.Skip("requires owned provisioning fixture")
	}
	c, e := pgxpool.ParseConfig(raw)
	if e != nil || c.ConnConfig.Host != "127.0.0.1" || c.ConnConfig.Database != "local_provision_catalog_test" || c.ConnConfig.User != "local_provision_test" {
		t.Fatal("unsafe fixture")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	p, e := pgxpool.NewWithConfig(ctx, c)
	if e != nil {
		t.Fatal("pool")
	}
	defer p.Close()
	var marker string
	if p.QueryRow(ctx, "SELECT current_setting('local.provision_fixture')").Scan(&marker) != nil || marker != os.Getenv("LOCAL_PROVISION_RUN_ID") || marker == "" {
		t.Fatal("marker")
	}
	category, product, unit, variant := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	for _, q := range []struct {
		sql  string
		args []any
	}{
		{"INSERT INTO catalog_categories(id,name,code) VALUES($1,'Local fixture','LOCAL')", []any{category}},
		{"INSERT INTO catalog_units(id,name,code) VALUES($1,'Local physical piece','LOCAL-PC')", []any{unit}},
		{"INSERT INTO catalog_products(id,category_id,name,product_code) VALUES($1,$2,'Local fixture','LOCAL')", []any{product, category}},
		{"INSERT INTO catalog_product_variants(id,product_id,sku,unit_id) VALUES($1,$2,'LOCAL',$3)", []any{variant, product, unit}},
	} {
		if _, e = p.Exec(ctx, q.sql, q.args...); e != nil {
			t.Fatal("fixture seed")
		}
	}
	s, e := service.NewStockReferenceService(postgres.NewStockReferenceRepository(p))
	if e != nil {
		t.Fatal("service")
	}
	plan := selection{Unit: unit, Variants: []uuid.UUID{variant}, Confirmation: "one base unit is one physical piece for every variant using this unit"}
	bad := plan
	bad.Confirmation = ""
	if apply(ctx, s, bad) == nil {
		t.Fatal("unconfirmed classification allowed")
	}
	if apply(ctx, s, plan) != nil {
		t.Fatal("classification failed")
	}
	var units, eligible, packs int
	if p.QueryRow(ctx, "SELECT (SELECT count(*) FROM catalog_piece_units),(SELECT count(*) FROM catalog_stock_eligibility),(SELECT count(*) FROM catalog_variant_packs)").Scan(&units, &eligible, &packs) != nil || units != 1 || eligible != 1 || packs != 0 {
		t.Fatal("classification scope")
	}
	t.Log("confirmed service classification, one variant eligibility, no packs passed")
}
