package repository

import (
	"context"
	"time"

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

type ListProductImagesParams struct {
	CursorProductID *uuid.UUID
	CursorSortOrder *int32
	CursorCreatedAt *time.Time
	CursorID        *uuid.UUID
	ProductID       *uuid.UUID
	VariantID       *uuid.UUID
	PageSize        int32
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
		ListProductImagesParams,
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
