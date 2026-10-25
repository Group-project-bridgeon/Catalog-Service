package service

import (
	"context"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/google/uuid"
)

// StockReferenceService has no startup/HTTP wiring or implicit authority.
type StockReferenceService struct {
	repo repository.StockReferenceRepository
}

func NewStockReferenceService(repo repository.StockReferenceRepository) (*StockReferenceService, error) {
	if repo == nil {
		return nil, domain.ErrStockInvalid
	}
	return &StockReferenceService{repo}, nil
}
func (s *StockReferenceService) ClassifyPieceUnit(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return domain.ErrStockInvalid
	}
	return s.repo.ClassifyPieceUnit(ctx, id)
}
func (s *StockReferenceService) EnableStock(ctx context.Context, variant, unit uuid.UUID) error {
	if variant == uuid.Nil || unit == uuid.Nil {
		return domain.ErrStockInvalid
	}
	return s.repo.EnableStock(ctx, variant, unit)
}
func (s *StockReferenceService) CreatePack(ctx context.Context, variant uuid.UUID, code, label string, factor domain.StockCount, active bool) (domain.VariantPack, error) {
	if variant == uuid.Nil {
		return domain.VariantPack{}, domain.ErrStockInvalid
	}
	code, label, err := domain.NormalizePack(code, label, factor)
	if err != nil {
		return domain.VariantPack{}, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return domain.VariantPack{}, domain.ErrStockInternal
	}
	return s.repo.CreatePack(ctx, domain.VariantPack{ID: id, VariantID: variant, Code: code, Label: label, PiecesPerPack: factor, Active: active})
}
func (s *StockReferenceService) RevisePack(ctx context.Context, id uuid.UUID, expected, factor domain.StockCount, label string) (domain.VariantPack, error) {
	if id == uuid.Nil || expected <= 0 {
		return domain.VariantPack{}, domain.ErrStockInvalid
	}
	_, label, err := domain.NormalizePack("PACK", label, factor)
	if err != nil {
		return domain.VariantPack{}, err
	}
	return s.repo.RevisePack(ctx, id, expected, factor, label)
}
func (s *StockReferenceService) SetPackActive(ctx context.Context, id uuid.UUID, expected domain.StockCount, active bool) (domain.VariantPack, error) {
	if id == uuid.Nil || expected <= 0 {
		return domain.VariantPack{}, domain.ErrStockInvalid
	}
	return s.repo.SetPackActive(ctx, id, expected, active)
}
func (s *StockReferenceService) GetPackRevision(ctx context.Context, id uuid.UUID, revision domain.StockCount) (domain.VariantPack, error) {
	if id == uuid.Nil || revision <= 0 {
		return domain.VariantPack{}, domain.ErrStockInvalid
	}
	return s.repo.GetPackRevision(ctx, id, revision)
}
func (s *StockReferenceService) ResolveBatch(ctx context.Context, items []domain.StockReferenceRequest) (domain.StockReferenceBatch, error) {
	if err := domain.ValidateStockRequests(items); err != nil {
		return domain.StockReferenceBatch{}, err
	}
	return s.repo.ResolveBatch(ctx, items)
}
