package repository

import (
	"context"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/google/uuid"
)

type CreatePriceListParams struct {
	ID           uuid.UUID
	Name         string
	Code         string
	CurrencyCode string
	IsActive     bool
}

type UpdatePriceListParams struct {
	ID           uuid.UUID
	Name         string
	Code         string
	CurrencyCode string
}

type PriceListRepository interface {
	Create(
		context.Context,
		CreatePriceListParams,
	) (domain.PriceList, error)

	GetByID(
		context.Context,
		uuid.UUID,
	) (domain.PriceList, error)

	GetByCode(
		context.Context,
		string,
	) (domain.PriceList, error)

	List(
		context.Context,
	) ([]domain.PriceList, error)

	Update(
		context.Context,
		UpdatePriceListParams,
	) (domain.PriceList, error)

	SetActive(
		context.Context,
		uuid.UUID,
		bool,
	) (domain.PriceList, error)
}
