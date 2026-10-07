package repository

import (
	"context"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/google/uuid"
)

type CreateProductVariantParams struct {
	ID                  uuid.UUID
	ProductID           uuid.UUID
	SKU                 string
	DefaultSellingPrice *string
	UnitID              uuid.UUID
	IsActive            bool
}

type UpdateProductVariantParams struct {
	ID                  uuid.UUID
	SKU                 string
	DefaultSellingPrice *string
}

type ProductVariantRepository interface {
	Create(
		context.Context,
		CreateProductVariantParams,
	) (domain.ProductVariant, error)

	GetByID(
		context.Context,
		uuid.UUID,
	) (domain.ProductVariant, error)

	GetByProductAndSKU(
		context.Context,
		uuid.UUID,
		string,
	) (domain.ProductVariant, error)

	List(
		context.Context,
	) ([]domain.ProductVariant, error)

	ListByProductID(
		context.Context,
		uuid.UUID,
	) ([]domain.ProductVariant, error)

	Update(
		context.Context,
		UpdateProductVariantParams,
	) (domain.ProductVariant, error)

	SetActive(
		context.Context,
		uuid.UUID,
		bool,
	) (domain.ProductVariant, error)
}
