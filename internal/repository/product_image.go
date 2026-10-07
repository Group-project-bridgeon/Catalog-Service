package repository

import (
	"context"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/google/uuid"
)

type CreateProductImageParams struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	VariantID *uuid.UUID
	ImageURL  string
	SortOrder int32
}

type UpdateProductImageParams struct {
	ID        uuid.UUID
	VariantID *uuid.UUID
	ImageURL  string
	SortOrder int32
}

type ProductImageRepository interface {
	Create(
		context.Context,
		CreateProductImageParams,
	) (domain.ProductImage, error)

	GetByID(
		context.Context,
		uuid.UUID,
	) (domain.ProductImage, error)

	List(
		context.Context,
	) ([]domain.ProductImage, error)

	ListByProductID(
		context.Context,
		uuid.UUID,
	) ([]domain.ProductImage, error)

	ListByVariantID(
		context.Context,
		uuid.UUID,
	) ([]domain.ProductImage, error)

	Update(
		context.Context,
		UpdateProductImageParams,
	) (domain.ProductImage, error)

	VariantBelongsToProduct(
		context.Context,
		uuid.UUID,
		uuid.UUID,
	) (bool, error)
}
