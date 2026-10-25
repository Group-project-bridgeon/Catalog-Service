package repository

import (
	"context"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/google/uuid"
)

// Internal only. A future trusted/authorized adapter owns entry authorization.
type StockReferenceRepository interface {
	ClassifyPieceUnit(context.Context, uuid.UUID) error
	EnableStock(context.Context, uuid.UUID, uuid.UUID) error
	CreatePack(context.Context, domain.VariantPack) (domain.VariantPack, error)
	RevisePack(context.Context, uuid.UUID, domain.StockCount, domain.StockCount, string) (domain.VariantPack, error)
	SetPackActive(context.Context, uuid.UUID, domain.StockCount, bool) (domain.VariantPack, error)
	GetPackRevision(context.Context, uuid.UUID, domain.StockCount) (domain.VariantPack, error)
	ResolveBatch(context.Context, []domain.StockReferenceRequest) (domain.StockReferenceBatch, error)
}
