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

type productDetailQueries interface {
	GetProductByID(
		context.Context,
		pgtype.UUID,
	) (catalogdb.CatalogProduct, error)

	ListProductVariantsForDetail(
		context.Context,
		pgtype.UUID,
	) ([]catalogdb.CatalogProductVariant, error)

	ListVariantValuesForProductDetail(
		context.Context,
		pgtype.UUID,
	) ([]catalogdb.CatalogVariantValue, error)

	ListProductImagesForDetail(
		context.Context,
		pgtype.UUID,
	) ([]catalogdb.CatalogProductImage, error)

	ListPricesForProductDetail(
		context.Context,
		pgtype.UUID,
	) ([]catalogdb.CatalogPrice, error)
}

type ProductDetailRepository struct {
	queries productDetailQueries
}

var _ repository.ProductDetailRepository = (*ProductDetailRepository)(nil)

func NewProductDetailRepository(
	queries productDetailQueries,
) *ProductDetailRepository {
	return &ProductDetailRepository{
		queries: queries,
	}
}

func (r *ProductDetailRepository) Get(
	ctx context.Context,
	productID uuid.UUID,
) (domain.ProductDetail, error) {
	pgProductID := uuidToPG(productID)

	productRow, err := r.queries.GetProductByID(
		ctx,
		pgProductID,
	)
	if err != nil {
		return domain.ProductDetail{}, mapDBError(
			"get product detail: product",
			err,
		)
	}

	variantRows, err := r.queries.ListProductVariantsForDetail(
		ctx,
		pgProductID,
	)
	if err != nil {
		return domain.ProductDetail{}, mapDBError(
			"get product detail: variants",
			err,
		)
	}

	variantValueRows, err := r.queries.ListVariantValuesForProductDetail(
		ctx,
		pgProductID,
	)
	if err != nil {
		return domain.ProductDetail{}, mapDBError(
			"get product detail: variant values",
			err,
		)
	}

	imageRows, err := r.queries.ListProductImagesForDetail(
		ctx,
		pgProductID,
	)
	if err != nil {
		return domain.ProductDetail{}, mapDBError(
			"get product detail: images",
			err,
		)
	}

	priceRows, err := r.queries.ListPricesForProductDetail(
		ctx,
		pgProductID,
	)
	if err != nil {
		return domain.ProductDetail{}, mapDBError(
			"get product detail: prices",
			err,
		)
	}

	variants := make(
		[]domain.ProductVariant,
		0,
		len(variantRows),
	)
	for _, row := range variantRows {
		variant, err := productVariantToDomain(row)
		if err != nil {
			return domain.ProductDetail{}, fmt.Errorf(
				"map product variant for detail: %w",
				err,
			)
		}
		variants = append(variants, variant)
	}

	variantValues := make(
		[]domain.VariantValue,
		0,
		len(variantValueRows),
	)
	for _, row := range variantValueRows {
		variantValues = append(
			variantValues,
			variantValueToDomain(row),
		)
	}

	images := make(
		[]domain.ProductImage,
		0,
		len(imageRows),
	)
	for _, row := range imageRows {
		images = append(
			images,
			productImageToDomain(row),
		)
	}

	prices, err := pricesToDomain(priceRows)
	if err != nil {
		return domain.ProductDetail{}, fmt.Errorf(
			"map prices for detail: %w",
			err,
		)
	}

	return domain.ProductDetail{
		Product:       productToDomain(productRow),
		Variants:      variants,
		VariantValues: variantValues,
		Images:        images,
		Prices:        prices,
	}, nil
}
