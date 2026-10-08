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

type priceQueries interface {
	CreatePrice(
		context.Context,
		catalogdb.CreatePriceParams,
	) (catalogdb.CatalogPrice, error)

	GetPriceByID(
		context.Context,
		pgtype.UUID,
	) (catalogdb.CatalogPrice, error)

	GetPriceByPriceListAndVariant(
		context.Context,
		catalogdb.GetPriceByPriceListAndVariantParams,
	) (catalogdb.CatalogPrice, error)

	ListPrices(
		context.Context,
		catalogdb.ListPricesParams,
	) ([]catalogdb.CatalogPrice, error)

	UpdatePrice(
		context.Context,
		catalogdb.UpdatePriceParams,
	) (catalogdb.CatalogPrice, error)
}

type PriceRepository struct {
	queries priceQueries
}

var _ repository.PriceRepository = (*PriceRepository)(nil)

func NewPriceRepository(
	queries priceQueries,
) *PriceRepository {
	return &PriceRepository{
		queries: queries,
	}
}

func (r *PriceRepository) Create(
	ctx context.Context,
	params repository.CreatePriceParams,
) (domain.Price, error) {
	unitPrice, err := decimalStringToPGNumeric(
		&params.UnitPrice,
	)
	if err != nil {
		return domain.Price{}, fmt.Errorf(
			"create price: convert unit price: %w",
			err,
		)
	}

	row, err := r.queries.CreatePrice(
		ctx,
		catalogdb.CreatePriceParams{
			ID:          uuidToPG(params.ID),
			PriceListID: uuidToPG(params.PriceListID),
			VariantID:   uuidToPG(params.VariantID),
			UnitPrice:   unitPrice,
		},
	)
	if err != nil {
		return domain.Price{}, mapDBError(
			"create price",
			err,
		)
	}

	return priceToDomain(row)
}

func (r *PriceRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Price, error) {
	row, err := r.queries.GetPriceByID(
		ctx,
		uuidToPG(id),
	)
	if err != nil {
		return domain.Price{}, mapDBError(
			"get price by id",
			err,
		)
	}

	return priceToDomain(row)
}

func (r *PriceRepository) GetByPriceListAndVariant(
	ctx context.Context,
	priceListID uuid.UUID,
	variantID uuid.UUID,
) (domain.Price, error) {
	row, err :=
		r.queries.GetPriceByPriceListAndVariant(
			ctx,
			catalogdb.GetPriceByPriceListAndVariantParams{
				PriceListID: uuidToPG(priceListID),
				VariantID:   uuidToPG(variantID),
			},
		)
	if err != nil {
		return domain.Price{}, mapDBError(
			"get price by price list and variant",
			err,
		)
	}

	return priceToDomain(row)
}

func (r *PriceRepository) List(
	ctx context.Context,
	params repository.ListPricesParams,
) ([]domain.Price, error) {
	rows, err := r.queries.ListPrices(
		ctx,
		catalogdb.ListPricesParams{
			CursorPriceListID: nullableUUIDToPG(
				params.CursorPriceListID,
			),
			CursorVariantID: nullableUUIDToPG(
				params.CursorVariantID,
			),
			PriceListID: nullableUUIDToPG(
				params.PriceListID,
			),
			VariantID: nullableUUIDToPG(
				params.VariantID,
			),
			PageSize: params.PageSize,
		},
	)
	if err != nil {
		return nil, mapDBError(
			"list prices",
			err,
		)
	}

	return pricesToDomain(rows)
}

func (r *PriceRepository) Update(
	ctx context.Context,
	params repository.UpdatePriceParams,
) (domain.Price, error) {
	unitPrice, err := decimalStringToPGNumeric(
		&params.UnitPrice,
	)
	if err != nil {
		return domain.Price{}, fmt.Errorf(
			"update price: convert unit price: %w",
			err,
		)
	}

	row, err := r.queries.UpdatePrice(
		ctx,
		catalogdb.UpdatePriceParams{
			ID:        uuidToPG(params.ID),
			UnitPrice: unitPrice,
		},
	)
	if err != nil {
		return domain.Price{}, mapDBError(
			"update price",
			err,
		)
	}

	return priceToDomain(row)
}

func priceToDomain(
	row catalogdb.CatalogPrice,
) (domain.Price, error) {
	unitPrice, err := pgNumericToDecimalString(
		row.UnitPrice,
	)
	if err != nil {
		return domain.Price{}, fmt.Errorf(
			"convert price unit price: %w",
			err,
		)
	}

	if unitPrice == nil {
		return domain.Price{}, fmt.Errorf(
			"convert price unit price: value is null",
		)
	}

	return domain.Price{
		ID:          pgUUIDToUUID(row.ID),
		PriceListID: pgUUIDToUUID(row.PriceListID),
		VariantID:   pgUUIDToUUID(row.VariantID),
		UnitPrice:   *unitPrice,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}, nil
}

func pricesToDomain(
	rows []catalogdb.CatalogPrice,
) ([]domain.Price, error) {
	prices := make(
		[]domain.Price,
		0,
		len(rows),
	)

	for _, row := range rows {
		price, err := priceToDomain(row)
		if err != nil {
			return nil, err
		}

		prices = append(
			prices,
			price,
		)
	}

	return prices, nil
}
