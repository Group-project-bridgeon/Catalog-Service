package postgres

import (
	"context"
	"errors"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	db "github.com/Group-project-bridgeon/Catalog-Service/internal/repository/postgres/catalogdb"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type catalogStockFixture struct {
	ctx                              context.Context
	pool                             *pgxpool.Pool
	repo                             *StockReferenceRepository
	s                                *service.StockReferenceService
	category, product, unit, variant uuid.UUID
}

func catalogStockOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) {
			t.Fatalf("fixture failure SQLSTATE=%s constraint=%s", pg.Code, pg.ConstraintName)
		}
		t.Fatalf("fixture failure type=%T; raw cause suppressed", err)
	}
}
func catalogStockDB(t *testing.T, tracers ...pgx.QueryTracer) *catalogStockFixture {
	t.Helper()
	url, run := os.Getenv("CATALOG_STOCK_TEST_DATABASE_URL"), os.Getenv("CATALOG_STOCK_TEST_RUN_ID")
	if url == "" && run == "" {
		t.Skip("use scripts/test-stock-reference-postgres.ps1 for isolated PostgreSQL")
	}
	port, e := strconv.Atoi(os.Getenv("CATALOG_STOCK_TEST_PORT"))
	if e != nil || port < 1 || port > 65535 || !strings.HasPrefix(run, "catalog-stock-test-") {
		t.Fatal("invalid Catalog fixture gate")
	}
	if _, e = uuid.Parse(strings.TrimPrefix(run, "catalog-stock-test-")); e != nil {
		t.Fatal("invalid fixture identity")
	}
	cfg, e := pgxpool.ParseConfig(url)
	if e != nil || cfg.ConnConfig.Host != "127.0.0.1" || int(cfg.ConnConfig.Port) != port || cfg.ConnConfig.Database != "catalog_stock_test" || cfg.ConnConfig.User != "catalog_stock_fixture" {
		t.Fatal("refusing non-fixture destination")
	}
	cfg.MaxConns = 8
	cfg.ConnConfig.RuntimeParams["timezone"] = "UTC"
	if len(tracers) > 0 {
		cfg.ConnConfig.Tracer = tracers[0]
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	t.Cleanup(cancel)
	pool, e := pgxpool.NewWithConfig(ctx, cfg)
	catalogStockOK(t, e)
	t.Cleanup(pool.Close)
	var marker string
	catalogStockOK(t, pool.QueryRow(ctx, "SELECT current_setting('catalog.fixture_id',true)").Scan(&marker))
	if marker != run {
		t.Fatal("fixture ownership mismatch")
	}
	f := &catalogStockFixture{ctx: ctx, pool: pool, category: uuid.New(), product: uuid.New(), unit: uuid.New(), variant: uuid.New()}
	f.repo = NewStockReferenceRepository(pool)
	f.s, e = service.NewStockReferenceService(f.repo)
	catalogStockOK(t, e)
	f.exec(t, "INSERT INTO catalog_categories(id,name,code) VALUES($1,'Synthetic',$2)", f.category, f.category.String())
	f.exec(t, "INSERT INTO catalog_units(id,code,name) VALUES($1,$2,$3)", f.unit, "U"+f.unit.String()[:20], "Unit "+f.unit.String())
	f.exec(t, "INSERT INTO catalog_products(id,category_id,name,product_code) VALUES($1,$2,'Synthetic product',$3)", f.product, f.category, f.product.String())
	f.exec(t, "INSERT INTO catalog_product_variants(id,product_id,unit_id,sku) VALUES($1,$2,$3,'SYN-SKU')", f.variant, f.product, f.unit)
	return f
}
func (f *catalogStockFixture) exec(t *testing.T, sql string, args ...any) {
	t.Helper()
	_, e := f.pool.Exec(f.ctx, sql, args...)
	catalogStockOK(t, e)
}
func (f *catalogStockFixture) eligible(t *testing.T) {
	t.Helper()
	catalogStockOK(t, f.s.ClassifyPieceUnit(f.ctx, f.unit))
	catalogStockOK(t, f.s.EnableStock(f.ctx, f.variant, f.unit))
}
func (f *catalogStockFixture) resolve(t *testing.T, request domain.StockReferenceRequest) domain.StockReferenceItem {
	t.Helper()
	b, e := f.s.ResolveBatch(f.ctx, []domain.StockReferenceRequest{request})
	catalogStockOK(t, e)
	if len(b.Items) != 1 || b.CheckedAt.IsZero() || b.Items[0].VariantID != request.VariantID {
		t.Fatal("batch correspondence")
	}
	return b.Items[0]
}
func catalogStockReject(t *testing.T, e error, code, constraint string) {
	t.Helper()
	var pg *pgconn.PgError
	if !errors.As(e, &pg) || pg.Code != code || (constraint != "" && pg.ConstraintName != constraint) {
		t.Fatal("expected named database rejection; raw cause suppressed")
	}
}

func TestCatalogStockLegacyEligibilityAndImmutableMeaning(t *testing.T) {
	f := catalogStockDB(t)
	req := domain.StockReferenceRequest{VariantID: f.variant}
	before := f.resolve(t, req)
	if before.Status != "not_stock_eligible" || before.Reference.QuantityKind != "UNCLASSIFIED" {
		t.Fatal("legacy inferred PIECE")
	}
	if e := f.s.EnableStock(f.ctx, f.variant, f.unit); !errors.Is(e, domain.ErrStockIneligible) {
		t.Fatal("unclassified unit accepted")
	}
	catalogStockOK(t, f.s.ClassifyPieceUnit(f.ctx, f.unit))
	middle := f.resolve(t, req)
	if middle.Status != "not_stock_eligible" || middle.Reference.QuantityKind != "PIECE" || middle.Reference.Revision == before.Reference.Revision {
		t.Fatal("explicit unit classification")
	}
	if e := f.s.EnableStock(f.ctx, f.variant, uuid.New()); !errors.Is(e, domain.ErrStockIneligible) {
		t.Fatal("wrong binding accepted")
	}
	catalogStockOK(t, f.s.EnableStock(f.ctx, f.variant, f.unit))
	item := f.resolve(t, req)
	if item.Status != "ok" || item.Reference.Revision == middle.Reference.Revision {
		t.Fatal("eligibility contract")
	}
	for _, sql := range []string{"UPDATE catalog_product_variants SET unit_id=gen_random_uuid() WHERE id=$1", "UPDATE catalog_product_variants SET product_id=gen_random_uuid() WHERE id=$1", "UPDATE catalog_stock_eligibility SET base_unit_id=gen_random_uuid() WHERE variant_id=$1", "DELETE FROM catalog_stock_eligibility WHERE variant_id=$1"} {
		_, e := f.pool.Exec(f.ctx, sql, f.variant)
		catalogStockReject(t, e, "23514", "catalog_stock_immutable")
	}
	_, e := f.pool.Exec(f.ctx, "UPDATE catalog_piece_units SET quantity_kind='BOX' WHERE unit_id=$1", f.unit)
	catalogStockReject(t, e, "23514", "catalog_stock_immutable")
	_, e = f.pool.Exec(f.ctx, "UPDATE catalog_units SET id=gen_random_uuid() WHERE id=$1", f.unit)
	catalogStockReject(t, e, "23503", "")
	// Existing repository update paths still work for labels, and advance references.
	_, e = NewUnitRepository(db.New(f.pool)).Update(f.ctx, repository.UpdateUnitParams{ID: f.unit, Code: "EDIT" + f.unit.String()[:20], Name: "Renamed " + f.unit.String()})
	catalogStockOK(t, e)
	if after := f.resolve(t, req); after.Status != "ok" || after.Reference.Revision == item.Reference.Revision {
		t.Fatal("unit update did not invalidate reference")
	}
	_, e = NewProductVariantRepository(db.New(f.pool)).Update(f.ctx, repository.UpdateProductVariantParams{ID: f.variant, SKU: "RENAMED"})
	catalogStockOK(t, e)
	if f.resolve(t, req).Reference.SKU != "RENAMED" {
		t.Fatal("legacy variant update incompatible")
	}
}

func TestCatalogStockPacksAndReferenceOutcomes(t *testing.T) {
	f := catalogStockDB(t)
	f.eligible(t)
	p, e := f.s.CreatePack(f.ctx, f.variant, " box-6 ", " Six pieces ", 6, true)
	catalogStockOK(t, e)
	if p.Code != "BOX-6" || p.Revision != 1 || p.PiecesPerPack != 6 || !p.Current {
		t.Fatal("pack v1")
	}
	req := domain.StockReferenceRequest{VariantID: f.variant, PackID: p.ID, PackRevision: "1"}
	old := f.resolve(t, req)
	if old.Status != "ok" {
		t.Fatal("current pack denied")
	}
	p, e = f.s.RevisePack(f.ctx, p.ID, 1, 12, "Twelve pieces")
	catalogStockOK(t, e)
	if p.Revision != 2 || p.Version != 2 {
		t.Fatal("new revision")
	}
	history, e := f.s.GetPackRevision(f.ctx, p.ID, 1)
	catalogStockOK(t, e)
	if history.PiecesPerPack != 6 || history.Label != "Six pieces" || history.Current {
		t.Fatal("history overwritten or selectable")
	}
	if f.resolve(t, req).Status != "pack_revision_obsolete" {
		t.Fatal("historical revision selectable")
	}
	req.PackRevision = "2"
	current := f.resolve(t, req)
	if current.Status != "ok" || current.Reference.Pack.PiecesPerPack != 12 || current.Reference.Revision == old.Reference.Revision {
		t.Fatal("current pack resolution")
	}
	for _, factor := range []domain.StockCount{0, -1, math.MinInt64} {
		if _, e = f.s.RevisePack(f.ctx, p.ID, 2, factor, "Bad"); !errors.Is(e, domain.ErrStockInvalid) {
			t.Fatal("invalid factor")
		}
	}
	max, e := f.s.CreatePack(f.ctx, f.variant, "MAX", "Maximum", math.MaxInt64, true)
	catalogStockOK(t, e)
	if max.PiecesPerPack != math.MaxInt64 {
		t.Fatal("factor truncated")
	}
	_, e = f.pool.Exec(f.ctx, "INSERT INTO catalog_pack_revisions(pack_id,revision,pieces_per_pack,label) VALUES($1,3,0,'Bad')", p.ID)
	catalogStockReject(t, e, "23514", "catalog_pack_factor_positive")
	_, e = f.pool.Exec(f.ctx, "UPDATE catalog_pack_revisions SET pieces_per_pack=99 WHERE pack_id=$1", p.ID)
	catalogStockReject(t, e, "23514", "catalog_stock_immutable")
	if _, e = f.s.RevisePack(f.ctx, p.ID, 1, 20, "Conflict"); !errors.Is(e, domain.ErrStockConflict) {
		t.Fatal("stale revision accepted")
	}
	p, e = f.s.SetPackActive(f.ctx, p.ID, p.Version, false)
	catalogStockOK(t, e)
	if f.resolve(t, req).Status != "reference_inactive" {
		t.Fatal("inactive pack accepted")
	}
	inactiveHistory, e := f.s.GetPackRevision(f.ctx, p.ID, 1)
	catalogStockOK(t, e)
	if inactiveHistory.PiecesPerPack != 6 || inactiveHistory.Active || inactiveHistory.Current {
		t.Fatal("inactive pack history unavailable or reinterpreted")
	}
	p, e = f.s.SetPackActive(f.ctx, p.ID, p.Version, true)
	catalogStockOK(t, e)
	if _, e = f.s.SetPackActive(f.ctx, p.ID, 1, false); !errors.Is(e, domain.ErrStockConflict) {
		t.Fatal("stale active update accepted")
	}
	missing := req
	missing.PackID = uuid.New()
	if f.resolve(t, missing).Status != "pack_not_found" {
		t.Fatal("missing pack")
	}
	missing = req
	missing.PackRevision = "99"
	if f.resolve(t, missing).Status != "pack_revision_not_found" {
		t.Fatal("missing revision")
	}
	other := uuid.New()
	f.exec(t, "INSERT INTO catalog_product_variants(id,product_id,unit_id,sku) VALUES($1,$2,$3,'OTHER')", other, f.product, f.unit)
	catalogStockOK(t, f.s.EnableStock(f.ctx, other, f.unit))
	wrong := req
	wrong.VariantID = other
	if f.resolve(t, wrong).Status != "pack_variant_mismatch" {
		t.Fatal("cross-variant pack accepted")
	}
	// No universal BOX factor: another eligible variant may own BOX-6 with factor 24.
	otherPack, e := f.s.CreatePack(f.ctx, other, "BOX-6", "Other factor", 24, true)
	catalogStockOK(t, e)
	if otherPack.PiecesPerPack != 24 {
		t.Fatal("universal conversion inferred")
	}
	if _, e = f.s.CreatePack(f.ctx, f.variant, "BOX-6", "Duplicate", 1, true); !errors.Is(e, domain.ErrStockConflict) {
		t.Fatal("duplicate code accepted")
	}
	if f.resolve(t, domain.StockReferenceRequest{VariantID: uuid.New()}).Status != "variant_not_found" {
		t.Fatal("missing variant omitted")
	}
	for _, tc := range []struct {
		table string
		id    uuid.UUID
	}{{"catalog_products", f.product}, {"catalog_product_variants", f.variant}, {"catalog_units", f.unit}} {
		f.exec(t, "UPDATE "+tc.table+" SET is_active=false WHERE id=$1", tc.id)
		if f.resolve(t, req).Status != "reference_inactive" {
			t.Fatal("inactive master accepted")
		}
		historical, e := f.s.GetPackRevision(f.ctx, p.ID, 1)
		catalogStockOK(t, e)
		if historical.PiecesPerPack != 6 || historical.Label != "Six pieces" {
			t.Fatal("inactive master erased pack history")
		}
		f.exec(t, "UPDATE "+tc.table+" SET is_active=true WHERE id=$1", tc.id)
	}
	ref := f.resolve(t, req).Reference
	req.ExpectedRevision = old.Reference.Revision
	if f.resolve(t, req).Status != "reference_changed" {
		t.Fatal("stale expected revision accepted")
	}
	req.ExpectedRevision = ref.Revision
	if f.resolve(t, req).Status != "ok" {
		t.Fatal("matching expectation rejected")
	}
}

func TestCatalogStockEveryContributingChangeAndBounds(t *testing.T) {
	f := catalogStockDB(t)
	f.eligible(t)
	req := domain.StockReferenceRequest{VariantID: f.variant}
	before := f.resolve(t, req).Reference.Revision
	for _, sql := range []string{
		"UPDATE catalog_products SET product_code=product_code||'X' WHERE id=$1",
		"UPDATE catalog_products SET name=name||'X' WHERE id=$1",
		"UPDATE catalog_products SET is_active=NOT is_active WHERE id=$1",
		"UPDATE catalog_products SET is_active=NOT is_active WHERE id=$1",
		"UPDATE catalog_product_variants SET sku=sku||'X' WHERE id=$1",
		"UPDATE catalog_product_variants SET is_active=NOT is_active WHERE id=$1",
		"UPDATE catalog_product_variants SET is_active=NOT is_active WHERE id=$1",
		"UPDATE catalog_units SET code=code||'X' WHERE id=$1",
		"UPDATE catalog_units SET name=name||'X' WHERE id=$1",
		"UPDATE catalog_units SET is_active=NOT is_active WHERE id=$1",
		"UPDATE catalog_units SET is_active=NOT is_active WHERE id=$1",
	} {
		id := f.product
		if strings.Contains(sql, "catalog_product_variants") {
			id = f.variant
		}
		if strings.Contains(sql, "catalog_units") {
			id = f.unit
		}
		f.exec(t, sql, id)
		next := f.resolve(t, req).Reference.Revision
		if next == before {
			t.Fatal("relevant mutation not versioned")
		}
		before = next
	}
	f.exec(t, "UPDATE catalog_product_variants SET default_selling_price=12.34,updated_at=now() WHERE id=$1", f.variant)
	if f.resolve(t, req).Reference.Revision != before {
		t.Fatal("unrelated price/clock changed stock revision")
	}
	_, e := f.pool.Exec(f.ctx, "UPDATE catalog_units SET stock_version=1 WHERE id=$1", f.unit)
	catalogStockReject(t, e, "23514", "catalog_stock_version_owned")
	rows, e := f.pool.Query(f.ctx, "INSERT INTO catalog_product_variants(id,product_id,unit_id,sku) SELECT gen_random_uuid(),$1,$2,'BATCH-'||n FROM generate_series(1,99) n RETURNING id", f.product, f.unit)
	catalogStockOK(t, e)
	requests := []domain.StockReferenceRequest{req}
	for rows.Next() {
		var id uuid.UUID
		catalogStockOK(t, rows.Scan(&id))
		requests = append(requests, domain.StockReferenceRequest{VariantID: id})
	}
	rows.Close()
	catalogStockOK(t, rows.Err())
	batch, e := f.s.ResolveBatch(f.ctx, requests)
	catalogStockOK(t, e)
	if len(batch.Items) != 100 {
		t.Fatal("batch incomplete")
	}
	if _, e = f.s.ResolveBatch(f.ctx, append(requests, domain.StockReferenceRequest{VariantID: uuid.New()})); !errors.Is(e, domain.ErrStockInvalid) {
		t.Fatal("oversized batch accepted")
	}
	if _, e = f.s.ResolveBatch(f.ctx, []domain.StockReferenceRequest{req, req}); !errors.Is(e, domain.ErrStockInvalid) {
		t.Fatal("duplicate batch accepted")
	}
	if f.pool.Stat().AcquiredConns() != 0 {
		t.Fatal("read connection leaked")
	}
}

func TestCatalogStockConcurrentPackRevision(t *testing.T) {
	f := catalogStockDB(t)
	f.eligible(t)
	p, e := f.s.CreatePack(f.ctx, f.variant, "BOX", "Six", 6, true)
	catalogStockOK(t, e)
	ctx, cancel := context.WithTimeout(f.ctx, 8*time.Second)
	defer cancel()
	ready := make(chan struct{}, 2)
	release := make(chan struct{})
	done := make(chan error, 2)
	for i := range 2 {
		go func(factor int) {
			select {
			case ready <- struct{}{}:
			case <-ctx.Done():
				done <- ctx.Err()
				return
			}
			select {
			case <-release:
			case <-ctx.Done():
				done <- ctx.Err()
				return
			}
			_, e := f.s.RevisePack(ctx, p.ID, 1, domain.StockCount(factor), "Concurrent")
			done <- e
		}(12 + i)
	}
	for range 2 {
		select {
		case <-ready:
		case <-ctx.Done():
			t.Fatal("worker ready timeout")
		}
	}
	close(release)
	success, conflict := 0, 0
	for range 2 {
		select {
		case e := <-done:
			if e == nil {
				success++
			} else if errors.Is(e, domain.ErrStockConflict) {
				conflict++
			} else {
				t.Fatal("unexpected worker error; raw cause suppressed")
			}
		case <-ctx.Done():
			t.Fatal("worker result timeout")
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatal("concurrent revision overwritten")
	}
	old, e := f.s.GetPackRevision(f.ctx, p.ID, 1)
	catalogStockOK(t, e)
	if old.PiecesPerPack != 6 || old.Current {
		t.Fatal("history corrupted")
	}
	latest, e := f.s.GetPackRevision(f.ctx, p.ID, 2)
	catalogStockOK(t, e)
	if !latest.Current || (latest.PiecesPerPack != 12 && latest.PiecesPerPack != 13) {
		t.Fatal("revision winner invalid")
	}
}

type catalogSnapshotBarrier struct {
	armed         atomic.Bool
	read, release chan struct{}
}
type catalogSnapshotKey struct{}

func (b *catalogSnapshotBarrier) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	if b.armed.Load() && strings.Contains(data.SQL, "GetStockReferences") {
		return context.WithValue(ctx, catalogSnapshotKey{}, true)
	}
	return ctx
}
func (b *catalogSnapshotBarrier) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryEndData) {
	if ctx.Value(catalogSnapshotKey{}) == true && b.armed.CompareAndSwap(true, false) {
		close(b.read)
		select {
		case <-b.release:
		case <-ctx.Done():
		}
	}
}
func TestCatalogStockCoherentReadDuringConcurrentChange(t *testing.T) {
	b := &catalogSnapshotBarrier{read: make(chan struct{}), release: make(chan struct{})}
	f := catalogStockDB(t, b)
	f.eligible(t)
	p, e := f.s.CreatePack(f.ctx, f.variant, "BOX", "Six", 6, true)
	catalogStockOK(t, e)
	req := domain.StockReferenceRequest{VariantID: f.variant, PackID: p.ID, PackRevision: "1"}
	old := f.resolve(t, req)
	b.armed.Store(true)
	ctx, cancel := context.WithTimeout(f.ctx, 8*time.Second)
	defer cancel()
	var once sync.Once
	defer once.Do(func() { close(b.release) })
	type result struct {
		batch domain.StockReferenceBatch
		err   error
	}
	done := make(chan result, 1)
	go func() {
		batch, e := f.s.ResolveBatch(ctx, []domain.StockReferenceRequest{req})
		done <- result{batch, e}
	}()
	collected := false
	t.Cleanup(func() {
		cancel()
		once.Do(func() { close(b.release) })
		if !collected {
			cleanup, stop := context.WithTimeout(context.Background(), 3*time.Second)
			defer stop()
			select {
			case <-done:
			case <-cleanup.Done():
				t.Error("snapshot worker failed to exit")
			}
		}
	})
	select {
	case <-b.read:
	case <-ctx.Done():
		t.Fatal("snapshot barrier timeout")
	}
	tx, e := f.pool.Begin(ctx)
	catalogStockOK(t, e)
	defer stockRollback(tx)
	q := db.New(tx)
	_, e = tx.Exec(ctx, "UPDATE catalog_products SET name='Changed atomically' WHERE id=$1", f.product)
	catalogStockOK(t, e)
	_, e = q.InsertPackRevision(ctx, db.InsertPackRevisionParams{PackID: uuidToPG(p.ID), Revision: 2, PiecesPerPack: 12, Label: "Twelve"})
	catalogStockOK(t, e)
	_, e = q.AdvancePackRevision(ctx, db.AdvancePackRevisionParams{ID: uuidToPG(p.ID), CurrentRevision: 1})
	catalogStockOK(t, e)
	catalogStockOK(t, tx.Commit(ctx))
	once.Do(func() { close(b.release) })
	var out result
	select {
	case out = <-done:
		collected = true
	case <-ctx.Done():
		t.Fatal("snapshot worker result timeout")
	}
	catalogStockOK(t, out.err)
	got := out.batch.Items[0]
	if got.Status != "ok" || got.Reference.ProductName != old.Reference.ProductName || got.Reference.Pack.PiecesPerPack != 6 || !got.Reference.Pack.Current || got.Reference.Revision != old.Reference.Revision {
		t.Fatal("mixed database snapshots")
	}
	if f.resolve(t, req).Status != "pack_revision_obsolete" {
		t.Fatal("new snapshot missed new head")
	}
	req.PackRevision = "2"
	next := f.resolve(t, req)
	if next.Status != "ok" || next.Reference.ProductName != "Changed atomically" || next.Reference.Pack.PiecesPerPack != 12 || next.Reference.Revision == old.Reference.Revision {
		t.Fatal("new coherent reference")
	}
	f.exec(t, "UPDATE catalog_products SET is_active=false WHERE id=$1", f.product)
	if !got.Reference.ProductActive || f.resolve(t, req).Status != "reference_inactive" {
		t.Fatal("R1 observation semantics")
	}
	if f.pool.Stat().AcquiredConns() != 0 {
		t.Fatal("snapshot connection leaked")
	}
}

func TestCatalogStockSQLBackstopsAndSafeClassification(t *testing.T) {
	f := catalogStockDB(t)
	f.exec(t, "UPDATE catalog_units SET is_active=false WHERE id=$1", f.unit)
	if e := f.s.ClassifyPieceUnit(f.ctx, f.unit); !errors.Is(e, domain.ErrStockIneligible) {
		t.Fatal("inactive unit classified")
	}
	f.exec(t, "UPDATE catalog_units SET is_active=true WHERE id=$1", f.unit)
	catalogStockOK(t, f.s.ClassifyPieceUnit(f.ctx, f.unit))
	f.exec(t, "UPDATE catalog_products SET is_active=false WHERE id=$1", f.product)
	if e := f.s.EnableStock(f.ctx, f.variant, f.unit); !errors.Is(e, domain.ErrStockIneligible) {
		t.Fatal("inactive product eligible")
	}
	f.exec(t, "UPDATE catalog_products SET is_active=true WHERE id=$1", f.product)
	catalogStockOK(t, f.s.EnableStock(f.ctx, f.variant, f.unit))
	id := uuid.New()
	_, e := f.pool.Exec(f.ctx, "INSERT INTO catalog_variant_packs(id,variant_id,code,is_active) VALUES($1,$2,'ORPHAN',true)", id, f.variant)
	if e == nil {
		t.Fatal("pack without revision committed")
	}
	var count int
	catalogStockOK(t, f.pool.QueryRow(f.ctx, "SELECT count(*) FROM catalog_variant_packs WHERE id=$1", id).Scan(&count))
	if count != 0 {
		t.Fatal("orphan pack survived")
	}
	p, e := f.s.CreatePack(f.ctx, f.variant, "BOX", "Six", 6, true)
	catalogStockOK(t, e)
	_, e = f.pool.Exec(f.ctx, "INSERT INTO catalog_pack_revisions(pack_id,revision,pieces_per_pack,label) VALUES($1,2,12,'Unadvanced')", p.ID)
	catalogStockReject(t, e, "23514", "catalog_pack_complete")
	_, e = f.pool.Exec(f.ctx, "INSERT INTO catalog_pack_revisions(pack_id,revision,pieces_per_pack,label) VALUES($1,3,12,'Gap')", p.ID)
	catalogStockReject(t, e, "23514", "catalog_pack_transition")
	_, e = f.pool.Exec(f.ctx, "UPDATE catalog_variant_packs SET variant_id=gen_random_uuid() WHERE id=$1", p.ID)
	catalogStockReject(t, e, "23514", "catalog_stock_immutable")
	_, e = f.pool.Exec(f.ctx, "UPDATE catalog_variant_packs SET current_revision=3 WHERE id=$1", p.ID)
	catalogStockReject(t, e, "23514", "catalog_pack_transition")
	_, e = f.pool.Exec(f.ctx, "TRUNCATE catalog_pack_revisions CASCADE")
	catalogStockReject(t, e, "23514", "catalog_stock_immutable")
	_, e = f.pool.Exec(f.ctx, "INSERT INTO catalog_pack_revisions(pack_id,revision,pieces_per_pack,label) VALUES($1,2,9223372036854775808,'Overflow')", p.ID)
	catalogStockReject(t, e, "22003", "")
	// Legacy schema permits zero IDs; do not freeze that invalid identity into stock.
	f.exec(t, "INSERT INTO catalog_products(id,category_id,product_code,name) VALUES($1,$2,'ZERO-ID','Invalid physical identity')", uuid.Nil, f.category)
	v := uuid.New()
	f.exec(t, "INSERT INTO catalog_product_variants(id,product_id,unit_id,sku) VALUES($1,$2,$3,'ZERO-PARENT')", v, uuid.Nil, f.unit)
	if e = f.s.EnableStock(f.ctx, v, f.unit); !errors.Is(e, domain.ErrStockIneligible) {
		t.Fatal("zero legacy product ID classified")
	}
	if f.resolve(t, domain.StockReferenceRequest{VariantID: v}).Status != "invalid_metadata" {
		t.Fatal("invalid legacy identity returned as eligible")
	}
}

func TestCatalogStockSafeErrorMapping(t *testing.T) {
	for _, tc := range []struct {
		code, constraint string
		want             error
	}{
		{"23503", "catalog_stock_binding_fk", domain.ErrStockIneligible},
		{"23503", "catalog_pack_current_fk", domain.ErrStockInternal},
		{"23514", "catalog_pack_complete", domain.ErrStockInternal},
		{"23505", "catalog_pack_code_unique", domain.ErrStockConflict},
		{"23505", "unexpected_unique", domain.ErrStockInternal},
	} {
		err := stockError(context.Background(), "synthetic", &pgconn.PgError{Code: tc.code, ConstraintName: tc.constraint, Message: "synthetic private detail"})
		if !errors.Is(err, tc.want) || strings.Contains(err.Error(), "private detail") {
			t.Fatal("unsafe or imprecise error mapping")
		}
		var safe *StockFailure
		if !errors.As(err, &safe) || safe.SQLState != tc.code || safe.Component != "postgres" {
			t.Fatal("diagnostic metadata missing")
		}
	}
}
