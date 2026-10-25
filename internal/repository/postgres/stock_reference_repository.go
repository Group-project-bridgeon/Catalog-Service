package postgres

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	db "github.com/Group-project-bridgeon/Catalog-Service/internal/repository/postgres/catalogdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// StockFailure preserves typed causes without rendering raw infrastructure text.
// A future trusted boundary owns one zerolog event using these safe fields.
type StockFailure struct {
	Component, Operation, Reason, SQLState string
	cause                                  error
}

func (*StockFailure) Error() string          { return "Catalog stock persistence failed" }
func (f *StockFailure) SQLStateCode() string { return f.SQLState }
func (f *StockFailure) Unwrap() error        { return f.cause }
func stockError(ctx context.Context, op string, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	for _, known := range []error{domain.ErrStockInvalid, domain.ErrStockConflict, domain.ErrStockNotFound, domain.ErrStockImmutable, domain.ErrStockIneligible, domain.ErrStockBusy, domain.ErrStockInternal} {
		if errors.Is(err, known) {
			return known
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrStockNotFound
	}
	f := &StockFailure{Component: "postgres", Operation: op, Reason: "stock_database_failure", cause: err}
	kind := domain.ErrStockInternal
	var p *pgconn.PgError
	if errors.As(err, &p) {
		f.SQLState = p.Code
		switch {
		case p.Code == "23505" && (p.ConstraintName == "catalog_pack_code_unique" || p.ConstraintName == "catalog_stock_eligibility_pkey"):
			kind = domain.ErrStockConflict
		case p.Code == "23503" && (p.ConstraintName == "catalog_stock_binding_fk" || p.ConstraintName == "catalog_stock_eligibility_base_unit_id_fkey" || p.ConstraintName == "catalog_variant_packs_variant_id_fkey" || p.ConstraintName == "catalog_piece_units_unit_id_fkey"):
			kind = domain.ErrStockIneligible
		case p.Code == "23514" && p.ConstraintName == "catalog_stock_immutable":
			kind = domain.ErrStockImmutable
		case p.Code == "23514" && p.ConstraintName == "catalog_stock_classification_invalid":
			kind = domain.ErrStockIneligible
		case p.Code == "55P03" || p.Code == "57014" || p.Code == "40001" || p.Code == "40P01":
			kind = domain.ErrStockBusy
		}
	}
	return errors.Join(kind, f)
}

type StockReferenceRepository struct{ pool *pgxpool.Pool }

func NewStockReferenceRepository(pool *pgxpool.Pool) *StockReferenceRepository {
	return &StockReferenceRepository{pool}
}
func stockRollback(tx pgx.Tx) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = tx.Rollback(ctx)
}
func (r *StockReferenceRepository) write(ctx context.Context, op string, fn func(*db.Queries) error) error {
	// Callback uses the bounded context supplied by each public operation below.
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return stockError(ctx, op, err)
	}
	defer stockRollback(tx)
	if _, err = tx.Exec(ctx, "SET LOCAL lock_timeout='500ms'; SET LOCAL statement_timeout='5s'"); err != nil {
		return stockError(ctx, op, err)
	}
	if err = fn(db.New(tx)); err != nil {
		return stockError(ctx, op, err)
	}
	if err = tx.Commit(ctx); err != nil {
		var p *pgconn.PgError
		if (errors.As(err, &p) && p.Severity != "FATAL" && p.Severity != "PANIC") || errors.Is(err, pgx.ErrTxCommitRollback) {
			return stockError(ctx, op, err)
		}
		return errors.Join(domain.ErrStockOutcomeUnknown, ctx.Err())
	}
	return nil
}
func (r *StockReferenceRepository) ClassifyPieceUnit(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return domain.ErrStockInvalid
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return r.write(ctx, "classify_piece_unit", func(q *db.Queries) error {
		n, err := q.ClassifyPieceUnit(ctx, uuidToPG(id))
		if err == nil && n != 0 && n != 1 {
			return domain.ErrStockInternal
		}
		return err
	})
}
func (r *StockReferenceRepository) EnableStock(ctx context.Context, variant, unit uuid.UUID) error {
	if variant == uuid.Nil || unit == uuid.Nil {
		return domain.ErrStockInvalid
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return r.write(ctx, "enable_stock", func(q *db.Queries) error {
		v, err := q.LockStockVariant(ctx, uuidToPG(variant))
		if err != nil {
			return err
		}
		if v.UnitID.Bytes != unit {
			return domain.ErrStockIneligible
		}
		n, err := q.EnableVariantStock(ctx, db.EnableVariantStockParams{VariantID: uuidToPG(variant), BaseUnitID: uuidToPG(unit)})
		if err == nil && n != 1 {
			return domain.ErrStockInternal
		}
		return err
	})
}
func packFromRow(p db.GetPackRevisionRow) domain.VariantPack {
	return domain.VariantPack{ID: pgUUIDToUUID(p.ID), VariantID: pgUUIDToUUID(p.VariantID), Code: p.Code, Revision: domain.StockCount(p.Revision), CurrentRevision: domain.StockCount(p.CurrentRevision), Version: domain.StockCount(p.StockVersion), PiecesPerPack: domain.StockCount(p.PiecesPerPack), Label: p.Label, Active: p.IsActive, Current: p.Revision == p.CurrentRevision}
}
func packResult(ctx context.Context, q *db.Queries, id uuid.UUID, rev int64) (domain.VariantPack, error) {
	p, err := q.GetPackRevision(ctx, db.GetPackRevisionParams{ID: uuidToPG(id), Revision: rev})
	if err != nil {
		return domain.VariantPack{}, err
	}
	return packFromRow(p), nil
}
func (r *StockReferenceRepository) CreatePack(ctx context.Context, p domain.VariantPack) (domain.VariantPack, error) {
	if p.ID == uuid.Nil || p.VariantID == uuid.Nil {
		return domain.VariantPack{}, domain.ErrStockInvalid
	}
	code, label, err := domain.NormalizePack(p.Code, p.Label, p.PiecesPerPack)
	if err != nil {
		return domain.VariantPack{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	result := domain.VariantPack{ID: p.ID}
	err = r.write(ctx, "create_pack", func(q *db.Queries) error {
		if _, err := q.CreateVariantPack(ctx, db.CreateVariantPackParams{ID: uuidToPG(p.ID), VariantID: uuidToPG(p.VariantID), Code: code, IsActive: p.Active}); err != nil {
			return err
		}
		if _, err := q.InsertPackRevision(ctx, db.InsertPackRevisionParams{PackID: uuidToPG(p.ID), Revision: 1, PiecesPerPack: int64(p.PiecesPerPack), Label: label}); err != nil {
			return err
		}
		var err error
		result, err = packResult(ctx, q, p.ID, 1)
		return err
	})
	return result, err // On uncertain create, retain generated ID for reconciliation.
}
func (r *StockReferenceRepository) RevisePack(ctx context.Context, id uuid.UUID, expected, factor domain.StockCount, label string) (domain.VariantPack, error) {
	if id == uuid.Nil || expected <= 0 || expected == math.MaxInt64 {
		return domain.VariantPack{}, domain.ErrStockInvalid
	}
	_, label, err := domain.NormalizePack("PACK", label, factor)
	if err != nil {
		return domain.VariantPack{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var result domain.VariantPack
	err = r.write(ctx, "revise_pack", func(q *db.Queries) error {
		p, err := q.LockVariantPack(ctx, uuidToPG(id))
		if err != nil {
			return err
		}
		if p.CurrentRevision != int64(expected) {
			return domain.ErrStockConflict
		}
		if _, err = q.InsertPackRevision(ctx, db.InsertPackRevisionParams{PackID: uuidToPG(id), Revision: int64(expected) + 1, PiecesPerPack: int64(factor), Label: label}); err != nil {
			return err
		}
		if _, err = q.AdvancePackRevision(ctx, db.AdvancePackRevisionParams{ID: uuidToPG(id), CurrentRevision: int64(expected)}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrStockInternal
			}
			return err
		}
		result, err = packResult(ctx, q, id, int64(expected)+1)
		return err
	})
	return result, err
}
func (r *StockReferenceRepository) SetPackActive(ctx context.Context, id uuid.UUID, expected domain.StockCount, active bool) (domain.VariantPack, error) {
	if id == uuid.Nil || expected <= 0 {
		return domain.VariantPack{}, domain.ErrStockInvalid
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var result domain.VariantPack
	err := r.write(ctx, "set_pack_active", func(q *db.Queries) error {
		p, err := q.LockVariantPack(ctx, uuidToPG(id))
		if err != nil {
			return err
		}
		if p.StockVersion != int64(expected) {
			return domain.ErrStockConflict
		}
		p, err = q.SetVariantPackActive(ctx, db.SetVariantPackActiveParams{ID: uuidToPG(id), StockVersion: int64(expected), IsActive: active})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrStockInternal
			}
			return err
		}
		result, err = packResult(ctx, q, id, p.CurrentRevision)
		return err
	})
	return result, err
}
func (r *StockReferenceRepository) GetPackRevision(ctx context.Context, id uuid.UUID, revision domain.StockCount) (domain.VariantPack, error) {
	if id == uuid.Nil || revision <= 0 {
		return domain.VariantPack{}, domain.ErrStockInvalid
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	p, err := packResult(ctx, db.New(r.pool), id, int64(revision))
	if err != nil {
		return domain.VariantPack{}, stockError(ctx, "get_pack_revision", err)
	}
	return p, nil
}

func (r *StockReferenceRepository) ResolveBatch(ctx context.Context, requests []domain.StockReferenceRequest) (domain.StockReferenceBatch, error) {
	if err := domain.ValidateStockRequests(requests); err != nil {
		return domain.StockReferenceBatch{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	fail := func(err error) (domain.StockReferenceBatch, error) {
		return domain.StockReferenceBatch{}, stockError(ctx, "resolve_stock_references", err)
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return fail(err)
	}
	defer stockRollback(tx)
	if _, err = tx.Exec(ctx, "SET LOCAL statement_timeout='5s'; SET LOCAL lock_timeout='500ms'"); err != nil {
		return fail(err)
	}
	q := db.New(tx)
	checked, err := q.StockObservationTime(ctx)
	if err != nil {
		return fail(err)
	}
	ids := make([]pgtype.UUID, 0, len(requests))
	packIDs := []pgtype.UUID{}
	revisions := []int64{}
	for _, req := range requests {
		ids = append(ids, uuidToPG(req.VariantID))
		if req.PackID != uuid.Nil {
			rev, _ := domain.ParseStockCount(req.PackRevision)
			packIDs = append(packIDs, uuidToPG(req.PackID))
			revisions = append(revisions, int64(rev))
		}
	}
	rows, err := q.GetStockReferences(ctx, ids)
	if err != nil {
		return fail(err)
	}
	byVariant := make(map[uuid.UUID]db.GetStockReferencesRow, len(rows))
	for _, row := range rows {
		byVariant[pgUUIDToUUID(row.VariantID)] = row
	}
	type packKey struct {
		id  uuid.UUID
		rev int64
	}
	byPack := map[packKey]db.GetSelectedPackReferencesRow{}
	if len(packIDs) > 0 {
		packs, err := q.GetSelectedPackReferences(ctx, db.GetSelectedPackReferencesParams{PackIds: packIDs, Revisions: revisions})
		if err != nil {
			return fail(err)
		}
		for _, p := range packs {
			byPack[packKey{pgUUIDToUUID(p.ID), p.RequestedRevision}] = p
		}
	}
	batch := domain.StockReferenceBatch{CheckedAt: checked.Time, Items: make([]domain.StockReferenceItem, 0, len(requests))}
	for _, req := range requests {
		item := domain.StockReferenceItem{VariantID: req.VariantID, Status: "variant_not_found"}
		v, ok := byVariant[req.VariantID]
		if !ok {
			batch.Items = append(batch.Items, item)
			continue
		}
		ref := domain.StockReference{ProductID: pgUUIDToUUID(v.ProductID), VariantID: req.VariantID, BaseUnitID: pgUUIDToUUID(v.BaseUnitID), QuantityKind: v.QuantityKind, Eligible: v.Eligible, ProductActive: v.ProductActive, VariantActive: v.VariantActive, BaseUnitActive: v.BaseUnitActive, ProductCode: v.ProductCode, ProductName: v.ProductName, SKU: v.Sku, BaseUnitCode: v.BaseUnitCode, BaseUnitName: v.BaseUnitName, ProductVersion: domain.StockCount(v.ProductVersion), VariantVersion: domain.StockCount(v.VariantVersion), UnitVersion: domain.StockCount(v.UnitVersion), CheckedAt: checked.Time}
		item.Status = "ok"
		if req.PackID != uuid.Nil {
			rev, _ := domain.ParseStockCount(req.PackRevision)
			p, found := byPack[packKey{req.PackID, int64(rev)}]
			switch {
			case !found:
				item.Status = "pack_not_found"
			case pgUUIDToUUID(p.VariantID) != req.VariantID:
				item.Status = "pack_variant_mismatch"
			case !p.Revision.Valid:
				item.Status = "pack_revision_not_found"
			default:
				ref.Pack = &domain.VariantPack{ID: req.PackID, VariantID: req.VariantID, Code: p.Code, Revision: domain.StockCount(p.Revision.Int64), CurrentRevision: domain.StockCount(p.CurrentRevision), Version: domain.StockCount(p.StockVersion), PiecesPerPack: domain.StockCount(p.PiecesPerPack.Int64), Label: p.Label.String, Active: p.IsActive, Current: p.CurrentRevision == int64(rev)}
				if !ref.Pack.Current {
					item.Status = "pack_revision_obsolete"
				} else if !ref.Pack.Active {
					item.Status = "reference_inactive"
				}
			}
		}
		if item.Status == "ok" {
			switch {
			case !ref.ValidMetadata():
				item.Status = "invalid_metadata"
			case !ref.Eligible || ref.QuantityKind != "PIECE":
				item.Status = "not_stock_eligible"
			case !ref.ProductActive || !ref.VariantActive || !ref.BaseUnitActive:
				item.Status = "reference_inactive"
			}
		}
		ref.Revision, err = ref.Fingerprint()
		if err != nil {
			return fail(err)
		}
		if item.Status == "ok" && req.ExpectedRevision != "" && req.ExpectedRevision != ref.Revision {
			item.Status = "reference_changed"
		}
		item.Reference = &ref
		batch.Items = append(batch.Items, item)
	}
	if err = tx.Commit(ctx); err != nil {
		return fail(err)
	}
	return batch, nil
}

var _ repository.StockReferenceRepository = (*StockReferenceRepository)(nil)
