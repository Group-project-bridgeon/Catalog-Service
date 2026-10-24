package service

import (
	"context"
	"errors"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/google/uuid"
	"testing"
)

type stockRepoStub struct {
	repository.StockReferenceRepository
	pack  domain.VariantPack
	calls int
}

func (s *stockRepoStub) CreatePack(_ context.Context, p domain.VariantPack) (domain.VariantPack, error) {
	s.calls++
	s.pack = p
	return p, nil
}
func (s *stockRepoStub) ResolveBatch(_ context.Context, r []domain.StockReferenceRequest) (domain.StockReferenceBatch, error) {
	s.calls++
	return domain.StockReferenceBatch{}, nil
}
func TestStockServiceValidation(t *testing.T) {
	ctx := context.Background()
	repo := &stockRepoStub{}
	s, e := NewStockReferenceService(repo)
	if e != nil {
		t.Fatal(e)
	}
	p, e := s.CreatePack(ctx, uuid.New(), " box ", " Six ", 6, true)
	if e != nil || p.ID.Version() != 7 || repo.pack.Code != "BOX" || repo.pack.Label != "Six" {
		t.Fatal("pack creation contract")
	}
	before := repo.calls
	if _, e = s.CreatePack(ctx, uuid.Nil, "BOX", "Six", 6, true); !errors.Is(e, domain.ErrStockInvalid) {
		t.Fatal("invalid variant")
	}
	if _, e = s.CreatePack(ctx, uuid.New(), "BOX", "Six", 0, true); !errors.Is(e, domain.ErrStockInvalid) {
		t.Fatal("invalid factor")
	}
	if _, e = s.ResolveBatch(ctx, nil); !errors.Is(e, domain.ErrStockInvalid) {
		t.Fatal("invalid read")
	}
	if e = s.EnableStock(ctx, uuid.Nil, uuid.New()); !errors.Is(e, domain.ErrStockInvalid) {
		t.Fatal("invalid eligibility")
	}
	if e = s.ClassifyPieceUnit(ctx, uuid.Nil); !errors.Is(e, domain.ErrStockInvalid) {
		t.Fatal("invalid unit")
	}
	if repo.calls != before {
		t.Fatal("invalid request reached repository")
	}
	if _, e = NewStockReferenceService(nil); e == nil {
		t.Fatal("nil repository")
	}
}
