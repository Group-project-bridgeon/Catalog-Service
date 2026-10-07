package postgres

import (
	"context"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository/postgres/catalogdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type variantValueQueries interface {
	CreateVariantValue(
		context.Context,
		catalogdb.CreateVariantValueParams,
	) (catalogdb.CatalogVariantValue, error)

	GetVariantValueByVariantAndAttribute(
		context.Context,
		catalogdb.GetVariantValueByVariantAndAttributeParams,
	) (catalogdb.CatalogVariantValue, error)

	ListVariantValues(
		context.Context,
	) ([]catalogdb.CatalogVariantValue, error)

	ListVariantValuesByVariantID(
		context.Context,
		pgtype.UUID,
	) ([]catalogdb.CatalogVariantValue, error)

	UpdateVariantValue(
		context.Context,
		catalogdb.UpdateVariantValueParams,
	) (catalogdb.CatalogVariantValue, error)

	VariantCombinationExists(
		context.Context,
		catalogdb.VariantCombinationExistsParams,
	) (bool, error)
}

type VariantValueRepository struct {
	queries variantValueQueries
}

var _ repository.VariantValueRepository = (*VariantValueRepository)(nil)

func NewVariantValueRepository(
	queries variantValueQueries,
) *VariantValueRepository {
	return &VariantValueRepository{
		queries: queries,
	}
}

func (r *VariantValueRepository) Create(
	ctx context.Context,
	params repository.CreateVariantValueParams,
) (domain.VariantValue, error) {
	row, err := r.queries.CreateVariantValue(
		ctx,
		catalogdb.CreateVariantValueParams{
			VariantID:        uuidToPG(params.VariantID),
			AttributeID:      uuidToPG(params.AttributeID),
			AttributeValueID: uuidToPG(params.AttributeValueID),
		},
	)
	if err != nil {
		return domain.VariantValue{}, mapDBError(
			"create variant value",
			err,
		)
	}

	return variantValueToDomain(row), nil
}

func (r *VariantValueRepository) GetByVariantAndAttribute(
	ctx context.Context,
	variantID uuid.UUID,
	attributeID uuid.UUID,
) (domain.VariantValue, error) {
	row, err := r.queries.GetVariantValueByVariantAndAttribute(
		ctx,
		catalogdb.GetVariantValueByVariantAndAttributeParams{
			VariantID:   uuidToPG(variantID),
			AttributeID: uuidToPG(attributeID),
		},
	)
	if err != nil {
		return domain.VariantValue{}, mapDBError(
			"get variant value",
			err,
		)
	}

	return variantValueToDomain(row), nil
}

func (r *VariantValueRepository) List(
	ctx context.Context,
) ([]domain.VariantValue, error) {
	rows, err := r.queries.ListVariantValues(ctx)
	if err != nil {
		return nil, mapDBError(
			"list variant values",
			err,
		)
	}

	values := make(
		[]domain.VariantValue,
		0,
		len(rows),
	)

	for _, row := range rows {
		values = append(
			values,
			variantValueToDomain(row),
		)
	}

	return values, nil
}

func (r *VariantValueRepository) ListByVariantID(
	ctx context.Context,
	variantID uuid.UUID,
) ([]domain.VariantValue, error) {
	rows, err := r.queries.ListVariantValuesByVariantID(
		ctx,
		uuidToPG(variantID),
	)
	if err != nil {
		return nil, mapDBError(
			"list variant values by variant",
			err,
		)
	}

	values := make(
		[]domain.VariantValue,
		0,
		len(rows),
	)

	for _, row := range rows {
		values = append(
			values,
			variantValueToDomain(row),
		)
	}

	return values, nil
}

func (r *VariantValueRepository) Update(
	ctx context.Context,
	params repository.UpdateVariantValueParams,
) (domain.VariantValue, error) {
	row, err := r.queries.UpdateVariantValue(
		ctx,
		catalogdb.UpdateVariantValueParams{
			VariantID:        uuidToPG(params.VariantID),
			AttributeID:      uuidToPG(params.AttributeID),
			AttributeValueID: uuidToPG(params.AttributeValueID),
		},
	)
	if err != nil {
		return domain.VariantValue{}, mapDBError(
			"update variant value",
			err,
		)
	}

	return variantValueToDomain(row), nil
}

func (r *VariantValueRepository) CombinationExists(
	ctx context.Context,
	variantID uuid.UUID,
	attributeID uuid.UUID,
	attributeValueID uuid.UUID,
) (bool, error) {
	exists, err := r.queries.VariantCombinationExists(
		ctx,
		catalogdb.VariantCombinationExistsParams{
			VariantID:        uuidToPG(variantID),
			AttributeID:      uuidToPG(attributeID),
			AttributeValueID: uuidToPG(attributeValueID),
		},
	)
	if err != nil {
		return false, mapDBError(
			"check variant combination",
			err,
		)
	}

	return exists, nil
}

func variantValueToDomain(
	row catalogdb.CatalogVariantValue,
) domain.VariantValue {
	return domain.VariantValue{
		VariantID:        pgUUIDToUUID(row.VariantID),
		AttributeID:      pgUUIDToUUID(row.AttributeID),
		AttributeValueID: pgUUIDToUUID(row.AttributeValueID),
	}
}
