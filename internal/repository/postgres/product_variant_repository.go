package postgres

import (
	"context"
	"fmt"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository/postgres/catalogdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type productVariantQueries interface {
	CreateProductVariant(
		context.Context,
		catalogdb.CreateProductVariantParams,
	) (catalogdb.CatalogProductVariant, error)

	GetProductVariantByID(
		context.Context,
		pgtype.UUID,
	) (catalogdb.CatalogProductVariant, error)

	GetProductVariantByProductAndSKU(
		context.Context,
		catalogdb.GetProductVariantByProductAndSKUParams,
	) (catalogdb.CatalogProductVariant, error)

	ListProductVariants(
		context.Context,
		catalogdb.ListProductVariantsParams,
	) ([]catalogdb.CatalogProductVariant, error)

	ListProductVariantsByIDs(
		context.Context,
		[]pgtype.UUID,
	) ([]catalogdb.CatalogProductVariant, error)

	UpdateProductVariant(
		context.Context,
		catalogdb.UpdateProductVariantParams,
	) (catalogdb.CatalogProductVariant, error)

	SetProductVariantActive(
		context.Context,
		catalogdb.SetProductVariantActiveParams,
	) (catalogdb.CatalogProductVariant, error)
}

type ProductVariantRepository struct {
	queries productVariantQueries
}

var _ repository.ProductVariantRepository = (*ProductVariantRepository)(nil)

func NewProductVariantRepository(
	queries productVariantQueries,
) *ProductVariantRepository {
	return &ProductVariantRepository{
		queries: queries,
	}
}

func (r *ProductVariantRepository) Create(
	ctx context.Context,
	params repository.CreateProductVariantParams,
) (domain.ProductVariant, error) {
	price, err := decimalStringToPGNumeric(
		params.DefaultSellingPrice,
	)
	if err != nil {
		return domain.ProductVariant{}, fmt.Errorf(
			"create product variant: convert price: %w",
			err,
		)
	}

	row, err := r.queries.CreateProductVariant(
		ctx,
		catalogdb.CreateProductVariantParams{
			ID:                  uuidToPG(params.ID),
			ProductID:           uuidToPG(params.ProductID),
			Sku:                 params.SKU,
			DefaultSellingPrice: price,
			UnitID:              uuidToPG(params.UnitID),
			IsActive:            params.IsActive,
		},
	)
	if err != nil {
		return domain.ProductVariant{}, mapDBError(
			"create product variant",
			err,
		)
	}

	return productVariantToDomain(row)
}

func (r *ProductVariantRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.ProductVariant, error) {
	row, err := r.queries.GetProductVariantByID(
		ctx,
		uuidToPG(id),
	)
	if err != nil {
		return domain.ProductVariant{}, mapDBError(
			"get product variant by id",
			err,
		)
	}

	return productVariantToDomain(row)
}

func (r *ProductVariantRepository) GetByProductAndSKU(
	ctx context.Context,
	productID uuid.UUID,
	sku string,
) (domain.ProductVariant, error) {
	row, err := r.queries.GetProductVariantByProductAndSKU(
		ctx,
		catalogdb.GetProductVariantByProductAndSKUParams{
			ProductID: uuidToPG(productID),
			Sku:       sku,
		},
	)
	if err != nil {
		return domain.ProductVariant{}, mapDBError(
			"get product variant by product and sku",
			err,
		)
	}

	return productVariantToDomain(row)
}

func (r *ProductVariantRepository) List(
	ctx context.Context,
	params repository.ListProductVariantsParams,
) ([]domain.ProductVariant, error) {
	rows, err := r.queries.ListProductVariants(
		ctx,
		catalogdb.ListProductVariantsParams{
			CursorProductID: nullableUUIDToPG(
				params.CursorProductID,
			),
			CursorSku: stringPtrToPGText(
				params.CursorSKU,
			),
			CursorID: nullableUUIDToPG(
				params.CursorID,
			),
			ProductID: nullableUUIDToPG(
				params.ProductID,
			),
			UnitID: nullableUUIDToPG(
				params.UnitID,
			),
			IsActive: boolPtrToPGBool(
				params.IsActive,
			),
			PageSize: params.PageSize,
		},
	)
	if err != nil {
		return nil, mapDBError(
			"list product variants",
			err,
		)
	}

	result := make(
		[]domain.ProductVariant,
		0,
		len(rows),
	)

	for _, row := range rows {
		variant, err := productVariantToDomain(row)
		if err != nil {
			return nil, fmt.Errorf(
				"list product variants: %w",
				err,
			)
		}

		result = append(result, variant)
	}

	return result, nil
}

func (r *ProductVariantRepository) ListByIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]domain.ProductVariant, error) {
	rows, err := r.queries.ListProductVariantsByIDs(
		ctx,
		uuidSliceToPG(ids),
	)
	if err != nil {
		return nil, mapDBError(
			"list product variants by ids",
			err,
		)
	}

	result := make(
		[]domain.ProductVariant,
		0,
		len(rows),
	)

	for _, row := range rows {
		variant, err := productVariantToDomain(row)
		if err != nil {
			return nil, fmt.Errorf(
				"list product variants by ids: %w",
				err,
			)
		}

		result = append(result, variant)
	}

	return result, nil
}

func (r *ProductVariantRepository) Update(
	ctx context.Context,
	params repository.UpdateProductVariantParams,
) (domain.ProductVariant, error) {
	price, err := decimalStringToPGNumeric(
		params.DefaultSellingPrice,
	)
	if err != nil {
		return domain.ProductVariant{}, fmt.Errorf(
			"update product variant: convert price: %w",
			err,
		)
	}

	row, err := r.queries.UpdateProductVariant(
		ctx,
		catalogdb.UpdateProductVariantParams{
			ID:                  uuidToPG(params.ID),
			Sku:                 params.SKU,
			DefaultSellingPrice: price,
		},
	)
	if err != nil {
		return domain.ProductVariant{}, mapDBError(
			"update product variant",
			err,
		)
	}

	return productVariantToDomain(row)
}

func (r *ProductVariantRepository) SetActive(
	ctx context.Context,
	id uuid.UUID,
	isActive bool,
) (domain.ProductVariant, error) {
	row, err := r.queries.SetProductVariantActive(
		ctx,
		catalogdb.SetProductVariantActiveParams{
			ID:       uuidToPG(id),
			IsActive: isActive,
		},
	)
	if err != nil {
		return domain.ProductVariant{}, mapDBError(
			"set product variant active",
			err,
		)
	}

	return productVariantToDomain(row)
}

func productVariantToDomain(
	row catalogdb.CatalogProductVariant,
) (domain.ProductVariant, error) {
	price, err := pgNumericToDecimalString(
		row.DefaultSellingPrice,
	)
	if err != nil {
		return domain.ProductVariant{}, fmt.Errorf(
			"convert product variant price: %w",
			err,
		)
	}

	return domain.ProductVariant{
		ID:                  pgUUIDToUUID(row.ID),
		ProductID:           pgUUIDToUUID(row.ProductID),
		SKU:                 row.Sku,
		DefaultSellingPrice: price,
		UnitID:              pgUUIDToUUID(row.UnitID),
		IsActive:            row.IsActive,
		CreatedAt:           row.CreatedAt.Time,
		UpdatedAt:           row.UpdatedAt.Time,
	}, nil
}
