package postgres

import (
	"context"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository/postgres/catalogdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type productImageQueries interface {
	CreateProductImage(
		context.Context,
		catalogdb.CreateProductImageParams,
	) (catalogdb.CatalogProductImage, error)

	GetProductImageByID(
		context.Context,
		pgtype.UUID,
	) (catalogdb.CatalogProductImage, error)

	ListProductImages(
		context.Context,
		catalogdb.ListProductImagesParams,
	) ([]catalogdb.CatalogProductImage, error)

	UpdateProductImage(
		context.Context,
		catalogdb.UpdateProductImageParams,
	) (catalogdb.CatalogProductImage, error)

	ProductImageVariantBelongsToProduct(
		context.Context,
		catalogdb.ProductImageVariantBelongsToProductParams,
	) (bool, error)
}

type ProductImageRepository struct {
	queries productImageQueries
}

var _ repository.ProductImageRepository = (*ProductImageRepository)(nil)

func NewProductImageRepository(
	queries productImageQueries,
) *ProductImageRepository {
	return &ProductImageRepository{
		queries: queries,
	}
}

func (r *ProductImageRepository) Create(
	ctx context.Context,
	params repository.CreateProductImageParams,
) (domain.ProductImage, error) {
	row, err := r.queries.CreateProductImage(
		ctx,
		catalogdb.CreateProductImageParams{
			ID:        uuidToPG(params.ID),
			ProductID: uuidToPG(params.ProductID),
			VariantID: nullableUUIDToPG(params.VariantID),
			ImageUrl:  params.ImageURL,
			SortOrder: params.SortOrder,
		},
	)
	if err != nil {
		return domain.ProductImage{}, mapDBError(
			"create product image",
			err,
		)
	}

	return productImageToDomain(row), nil
}

func (r *ProductImageRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.ProductImage, error) {
	row, err := r.queries.GetProductImageByID(
		ctx,
		uuidToPG(id),
	)
	if err != nil {
		return domain.ProductImage{}, mapDBError(
			"get product image by id",
			err,
		)
	}

	return productImageToDomain(row), nil
}

func (r *ProductImageRepository) List(
	ctx context.Context,
	params repository.ListProductImagesParams,
) ([]domain.ProductImage, error) {
	rows, err := r.queries.ListProductImages(
		ctx,
		catalogdb.ListProductImagesParams{
			CursorProductID: nullableUUIDToPG(
				params.CursorProductID,
			),
			CursorSortOrder: int32PtrToPGInt4(
				params.CursorSortOrder,
			),
			CursorCreatedAt: timePtrToPGTimestamptz(
				params.CursorCreatedAt,
			),
			CursorID: nullableUUIDToPG(
				params.CursorID,
			),
			ProductID: nullableUUIDToPG(
				params.ProductID,
			),
			VariantID: nullableUUIDToPG(
				params.VariantID,
			),
			PageSize: params.PageSize,
		},
	)
	if err != nil {
		return nil, mapDBError(
			"list product images",
			err,
		)
	}

	return productImagesToDomain(rows), nil
}

func (r *ProductImageRepository) Update(
	ctx context.Context,
	params repository.UpdateProductImageParams,
) (domain.ProductImage, error) {
	row, err := r.queries.UpdateProductImage(
		ctx,
		catalogdb.UpdateProductImageParams{
			ID:        uuidToPG(params.ID),
			VariantID: nullableUUIDToPG(params.VariantID),
			ImageUrl:  params.ImageURL,
			SortOrder: params.SortOrder,
		},
	)
	if err != nil {
		return domain.ProductImage{}, mapDBError(
			"update product image",
			err,
		)
	}

	return productImageToDomain(row), nil
}

func (r *ProductImageRepository) VariantBelongsToProduct(
	ctx context.Context,
	productID uuid.UUID,
	variantID uuid.UUID,
) (bool, error) {
	belongs, err :=
		r.queries.ProductImageVariantBelongsToProduct(
			ctx,
			catalogdb.ProductImageVariantBelongsToProductParams{
				ProductID: uuidToPG(productID),
				VariantID: uuidToPG(variantID),
			},
		)
	if err != nil {
		return false, mapDBError(
			"check product image variant",
			err,
		)
	}

	return belongs, nil
}

func productImageToDomain(
	row catalogdb.CatalogProductImage,
) domain.ProductImage {
	return domain.ProductImage{
		ID:        pgUUIDToUUID(row.ID),
		ProductID: pgUUIDToUUID(row.ProductID),
		VariantID: pgUUIDToNullableUUID(row.VariantID),
		ImageURL:  row.ImageUrl,
		SortOrder: row.SortOrder,
		CreatedAt: row.CreatedAt.Time,
	}
}

func productImagesToDomain(
	rows []catalogdb.CatalogProductImage,
) []domain.ProductImage {
	images := make(
		[]domain.ProductImage,
		0,
		len(rows),
	)

	for _, row := range rows {
		images = append(
			images,
			productImageToDomain(row),
		)
	}

	return images
}
