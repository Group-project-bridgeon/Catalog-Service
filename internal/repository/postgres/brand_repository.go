package postgres

import (
	"context"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository/postgres/catalogdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type brandQueries interface {
	CreateBrand(
		context.Context,
		catalogdb.CreateBrandParams,
	) (catalogdb.CatalogBrand, error)

	GetBrandByID(
		context.Context,
		pgtype.UUID,
	) (catalogdb.CatalogBrand, error)

	GetBrandByCode(
		context.Context,
		string,
	) (catalogdb.CatalogBrand, error)

	GetBrandByName(
		context.Context,
		string,
	) (catalogdb.CatalogBrand, error)

	ListBrands(
		context.Context,
	) ([]catalogdb.CatalogBrand, error)

	UpdateBrand(
		context.Context,
		catalogdb.UpdateBrandParams,
	) (catalogdb.CatalogBrand, error)

	SetBrandActive(
		context.Context,
		catalogdb.SetBrandActiveParams,
	) (catalogdb.CatalogBrand, error)
}

type BrandRepository struct {
	queries brandQueries
}

var _ repository.BrandRepository = (*BrandRepository)(nil)

func NewBrandRepository(
	queries brandQueries,
) *BrandRepository {
	return &BrandRepository{
		queries: queries,
	}
}

func (r *BrandRepository) Create(
	ctx context.Context,
	params repository.CreateBrandParams,
) (domain.Brand, error) {
	row, err := r.queries.CreateBrand(
		ctx,
		catalogdb.CreateBrandParams{
			ID:       uuidToPG(params.ID),
			Name:     params.Name,
			Code:     params.Code,
			IsActive: params.IsActive,
		},
	)
	if err != nil {
		return domain.Brand{}, mapDBError(
			"create brand",
			err,
		)
	}

	return brandToDomain(row), nil
}

func (r *BrandRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Brand, error) {
	row, err := r.queries.GetBrandByID(
		ctx,
		uuidToPG(id),
	)
	if err != nil {
		return domain.Brand{}, mapDBError(
			"get brand by id",
			err,
		)
	}

	return brandToDomain(row), nil
}

func (r *BrandRepository) GetByCode(
	ctx context.Context,
	code string,
) (domain.Brand, error) {
	row, err := r.queries.GetBrandByCode(ctx, code)
	if err != nil {
		return domain.Brand{}, mapDBError(
			"get brand by code",
			err,
		)
	}

	return brandToDomain(row), nil
}

func (r *BrandRepository) GetByName(
	ctx context.Context,
	name string,
) (domain.Brand, error) {
	row, err := r.queries.GetBrandByName(ctx, name)
	if err != nil {
		return domain.Brand{}, mapDBError(
			"get brand by name",
			err,
		)
	}

	return brandToDomain(row), nil
}

func (r *BrandRepository) List(
	ctx context.Context,
) ([]domain.Brand, error) {
	rows, err := r.queries.ListBrands(ctx)
	if err != nil {
		return nil, mapDBError(
			"list brands",
			err,
		)
	}

	brands := make(
		[]domain.Brand,
		0,
		len(rows),
	)

	for _, row := range rows {
		brands = append(
			brands,
			brandToDomain(row),
		)
	}

	return brands, nil
}

func (r *BrandRepository) Update(
	ctx context.Context,
	params repository.UpdateBrandParams,
) (domain.Brand, error) {
	row, err := r.queries.UpdateBrand(
		ctx,
		catalogdb.UpdateBrandParams{
			ID:   uuidToPG(params.ID),
			Name: params.Name,
			Code: params.Code,
		},
	)
	if err != nil {
		return domain.Brand{}, mapDBError(
			"update brand",
			err,
		)
	}

	return brandToDomain(row), nil
}

func (r *BrandRepository) SetActive(
	ctx context.Context,
	id uuid.UUID,
	isActive bool,
) (domain.Brand, error) {
	row, err := r.queries.SetBrandActive(
		ctx,
		catalogdb.SetBrandActiveParams{
			ID:       uuidToPG(id),
			IsActive: isActive,
		},
	)
	if err != nil {
		return domain.Brand{}, mapDBError(
			"set brand active",
			err,
		)
	}

	return brandToDomain(row), nil
}

func brandToDomain(
	row catalogdb.CatalogBrand,
) domain.Brand {
	return domain.Brand{
		ID:        pgUUIDToUUID(row.ID),
		Name:      row.Name,
		Code:      row.Code,
		IsActive:  row.IsActive,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
