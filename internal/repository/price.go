package repository

import (
	"context"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/google/uuid"
)

type CreatePriceParams struct {
	ID          uuid.UUID
	PriceListID uuid.UUID
	VariantID   uuid.UUID
	UnitPrice   string
}

type UpdatePriceParams struct {
	ID        uuid.UUID
	UnitPrice string
}

type PriceRepository interface {
	Create(
		context.Context,
		CreatePriceParams,
	) (domain.Price, error)

	GetByID(
		context.Context,
		uuid.UUID,
	) (domain.Price, error)

	GetByPriceListAndVariant(
		context.Context,
		uuid.UUID,
		uuid.UUID,
	) (domain.Price, error)

	List(
		context.Context,
	) ([]domain.Price, error)

	ListByPriceListID(
		context.Context,
		uuid.UUID,
	) ([]domain.Price, error)

	ListByVariantID(
		context.Context,
		uuid.UUID,
	) ([]domain.Price, error)

	Update(
		context.Context,
		UpdatePriceParams,
	) (domain.Price, error)
}
