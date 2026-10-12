package postgres

import (
	"context"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository/postgres/catalogdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type priceListQueries interface {
	CreatePriceList(
		context.Context,
		catalogdb.CreatePriceListParams,
	) (catalogdb.CatalogPriceList, error)

	GetPriceListByID(
		context.Context,
		pgtype.UUID,
	) (catalogdb.CatalogPriceList, error)

	GetPriceListByCode(
		context.Context,
		string,
	) (catalogdb.CatalogPriceList, error)

	ListPriceLists(
		context.Context,
	) ([]catalogdb.CatalogPriceList, error)

	UpdatePriceList(
		context.Context,
		catalogdb.UpdatePriceListParams,
	) (catalogdb.CatalogPriceList, error)

	SetPriceListActive(
		context.Context,
		catalogdb.SetPriceListActiveParams,
	) (catalogdb.CatalogPriceList, error)
}

type PriceListRepository struct {
	queries priceListQueries
}

var _ repository.PriceListRepository = (*PriceListRepository)(nil)

func NewPriceListRepository(
	queries priceListQueries,
) *PriceListRepository {
	return &PriceListRepository{
		queries: queries,
	}
}

func (r *PriceListRepository) Create(
	ctx context.Context,
	params repository.CreatePriceListParams,
) (domain.PriceList, error) {
	row, err := r.queries.CreatePriceList(
		ctx,
		catalogdb.CreatePriceListParams{
			ID:           uuidToPG(params.ID),
			Name:         params.Name,
			Code:         params.Code,
			CurrencyCode: params.CurrencyCode,
			IsActive:     params.IsActive,
		},
	)
	if err != nil {
		return domain.PriceList{}, mapDBError(
			"create price list",
			err,
		)
	}

	return priceListToDomain(row), nil
}

func (r *PriceListRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.PriceList, error) {
	row, err := r.queries.GetPriceListByID(
		ctx,
		uuidToPG(id),
	)
	if err != nil {
		return domain.PriceList{}, mapDBError(
			"get price list by id",
			err,
		)
	}

	return priceListToDomain(row), nil
}

func (r *PriceListRepository) GetByCode(
	ctx context.Context,
	code string,
) (domain.PriceList, error) {
	row, err := r.queries.GetPriceListByCode(
		ctx,
		code,
	)
	if err != nil {
		return domain.PriceList{}, mapDBError(
			"get price list by code",
			err,
		)
	}

	return priceListToDomain(row), nil
}

func (r *PriceListRepository) List(
	ctx context.Context,
) ([]domain.PriceList, error) {
	rows, err := r.queries.ListPriceLists(ctx)
	if err != nil {
		return nil, mapDBError(
			"list price lists",
			err,
		)
	}

	priceLists := make(
		[]domain.PriceList,
		0,
		len(rows),
	)

	for _, row := range rows {
		priceLists = append(
			priceLists,
			priceListToDomain(row),
		)
	}

	return priceLists, nil
}

func (r *PriceListRepository) Update(
	ctx context.Context,
	params repository.UpdatePriceListParams,
) (domain.PriceList, error) {
	row, err := r.queries.UpdatePriceList(
		ctx,
		catalogdb.UpdatePriceListParams{
			ID:           uuidToPG(params.ID),
			Name:         params.Name,
			Code:         params.Code,
			CurrencyCode: params.CurrencyCode,
		},
	)
	if err != nil {
		return domain.PriceList{}, mapDBError(
			"update price list",
			err,
		)
	}

	return priceListToDomain(row), nil
}

func (r *PriceListRepository) SetActive(
	ctx context.Context,
	id uuid.UUID,
	isActive bool,
) (domain.PriceList, error) {
	row, err := r.queries.SetPriceListActive(
		ctx,
		catalogdb.SetPriceListActiveParams{
			ID:       uuidToPG(id),
			IsActive: isActive,
		},
	)
	if err != nil {
		return domain.PriceList{}, mapDBError(
			"set price list active",
			err,
		)
	}

	return priceListToDomain(row), nil
}

func priceListToDomain(
	row catalogdb.CatalogPriceList,
) domain.PriceList {
	return domain.PriceList{
		ID:           pgUUIDToUUID(row.ID),
		Name:         row.Name,
		Code:         row.Code,
		CurrencyCode: row.CurrencyCode,
		IsActive:     row.IsActive,
	}
}
