package postgres

import (
	"context"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository/postgres/catalogdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type productQueries interface {
	CreateProduct(
		context.Context,
		catalogdb.CreateProductParams,
	) (catalogdb.CatalogProduct, error)

	GetProductByID(
		context.Context,
		pgtype.UUID,
	) (catalogdb.CatalogProduct, error)

	GetProductByCode(
		context.Context,
		string,
	) (catalogdb.CatalogProduct, error)

	ListProducts(
		context.Context,
		catalogdb.ListProductsParams,
	) ([]catalogdb.CatalogProduct, error)

	UpdateProduct(
		context.Context,
		catalogdb.UpdateProductParams,
	) (catalogdb.CatalogProduct, error)

	SetProductActive(
		context.Context,
		catalogdb.SetProductActiveParams,
	) (catalogdb.CatalogProduct, error)
}

type ProductRepository struct {
	queries productQueries
}

var _ repository.ProductRepository = (*ProductRepository)(nil)

func NewProductRepository(
	queries productQueries,
) *ProductRepository {
	return &ProductRepository{
		queries: queries,
	}
}

func (r *ProductRepository) Create(
	ctx context.Context,
	params repository.CreateProductParams,
) (domain.Product, error) {
	row, err := r.queries.CreateProduct(
		ctx,
		catalogdb.CreateProductParams{
			ID:          uuidToPG(params.ID),
			CategoryID:  uuidToPG(params.CategoryID),
			BrandID:     nullableUUIDToPG(params.BrandID),
			Name:        params.Name,
			ProductCode: params.ProductCode,
			Description: stringPtrToPGText(params.Description),
			IsActive:    params.IsActive,
		},
	)
	if err != nil {
		return domain.Product{}, mapDBError(
			"create product",
			err,
		)
	}

	return productToDomain(row), nil
}

func (r *ProductRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Product, error) {
	row, err := r.queries.GetProductByID(
		ctx,
		uuidToPG(id),
	)
	if err != nil {
		return domain.Product{}, mapDBError(
			"get product by id",
			err,
		)
	}

	return productToDomain(row), nil
}

func (r *ProductRepository) GetByCode(
	ctx context.Context,
	code string,
) (domain.Product, error) {
	row, err := r.queries.GetProductByCode(
		ctx,
		code,
	)
	if err != nil {
		return domain.Product{}, mapDBError(
			"get product by code",
			err,
		)
	}

	return productToDomain(row), nil
}

func (r *ProductRepository) List(
	ctx context.Context,
	params repository.ListProductsParams,
) ([]domain.Product, error) {
	rows, err := r.queries.ListProducts(
		ctx,
		catalogdb.ListProductsParams{
			CursorName: stringPtrToPGText(
				params.CursorName,
			),
			CursorID: nullableUUIDToPG(
				params.CursorID,
			),
			CategoryID: nullableUUIDToPG(
				params.CategoryID,
			),
			BrandID: nullableUUIDToPG(
				params.BrandID,
			),
			IsActive: boolPtrToPGBool(
				params.IsActive,
			),
			PageSize: params.PageSize,
		},
	)
	if err != nil {
		return nil, mapDBError(
			"list products",
			err,
		)
	}

	products := make(
		[]domain.Product,
		0,
		len(rows),
	)

	for _, row := range rows {
		products = append(
			products,
			productToDomain(row),
		)
	}

	return products, nil
}

func (r *ProductRepository) Update(
	ctx context.Context,
	params repository.UpdateProductParams,
) (domain.Product, error) {
	row, err := r.queries.UpdateProduct(
		ctx,
		catalogdb.UpdateProductParams{
			ID:          uuidToPG(params.ID),
			CategoryID:  uuidToPG(params.CategoryID),
			BrandID:     nullableUUIDToPG(params.BrandID),
			Name:        params.Name,
			ProductCode: params.ProductCode,
			Description: stringPtrToPGText(params.Description),
		},
	)
	if err != nil {
		return domain.Product{}, mapDBError(
			"update product",
			err,
		)
	}

	return productToDomain(row), nil
}

func (r *ProductRepository) SetActive(
	ctx context.Context,
	id uuid.UUID,
	isActive bool,
) (domain.Product, error) {
	row, err := r.queries.SetProductActive(
		ctx,
		catalogdb.SetProductActiveParams{
			ID:       uuidToPG(id),
			IsActive: isActive,
		},
	)
	if err != nil {
		return domain.Product{}, mapDBError(
			"set product active",
			err,
		)
	}

	return productToDomain(row), nil
}

func productToDomain(
	row catalogdb.CatalogProduct,
) domain.Product {
	return domain.Product{
		ID:          pgUUIDToUUID(row.ID),
		CategoryID:  pgUUIDToUUID(row.CategoryID),
		BrandID:     pgUUIDToNullableUUID(row.BrandID),
		Name:        row.Name,
		ProductCode: row.ProductCode,
		Description: pgTextToStringPtr(row.Description),
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}
