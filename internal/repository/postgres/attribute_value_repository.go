package postgres

import (
	"context"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository/postgres/catalogdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type attributeValueQueries interface {
	CreateAttributeValue(
		context.Context,
		catalogdb.CreateAttributeValueParams,
	) (catalogdb.CatalogAttributeValue, error)

	GetAttributeValueByID(
		context.Context,
		pgtype.UUID,
	) (catalogdb.CatalogAttributeValue, error)

	GetAttributeValueByAttributeAndValue(
		context.Context,
		catalogdb.GetAttributeValueByAttributeAndValueParams,
	) (catalogdb.CatalogAttributeValue, error)

	ListAttributeValues(
		context.Context,
	) ([]catalogdb.CatalogAttributeValue, error)

	ListAttributeValuesByAttributeID(
		context.Context,
		pgtype.UUID,
	) ([]catalogdb.CatalogAttributeValue, error)

	UpdateAttributeValue(
		context.Context,
		catalogdb.UpdateAttributeValueParams,
	) (catalogdb.CatalogAttributeValue, error)
}

type AttributeValueRepository struct {
	queries attributeValueQueries
}

var _ repository.AttributeValueRepository = (*AttributeValueRepository)(nil)

func NewAttributeValueRepository(
	queries attributeValueQueries,
) *AttributeValueRepository {
	return &AttributeValueRepository{
		queries: queries,
	}
}

func (r *AttributeValueRepository) Create(
	ctx context.Context,
	params repository.CreateAttributeValueParams,
) (domain.AttributeValue, error) {
	row, err := r.queries.CreateAttributeValue(
		ctx,
		catalogdb.CreateAttributeValueParams{
			ID:          uuidToPG(params.ID),
			AttributeID: uuidToPG(params.AttributeID),
			Value:       params.Value,
		},
	)
	if err != nil {
		return domain.AttributeValue{}, mapDBError(
			"create attribute value",
			err,
		)
	}

	return attributeValueToDomain(row), nil
}

func (r *AttributeValueRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.AttributeValue, error) {
	row, err := r.queries.GetAttributeValueByID(
		ctx,
		uuidToPG(id),
	)
	if err != nil {
		return domain.AttributeValue{}, mapDBError(
			"get attribute value by id",
			err,
		)
	}

	return attributeValueToDomain(row), nil
}

func (r *AttributeValueRepository) GetByAttributeAndValue(
	ctx context.Context,
	attributeID uuid.UUID,
	value string,
) (domain.AttributeValue, error) {
	row, err := r.queries.GetAttributeValueByAttributeAndValue(
		ctx,
		catalogdb.GetAttributeValueByAttributeAndValueParams{
			AttributeID: uuidToPG(attributeID),
			Value:       value,
		},
	)
	if err != nil {
		return domain.AttributeValue{}, mapDBError(
			"get attribute value by attribute and value",
			err,
		)
	}

	return attributeValueToDomain(row), nil
}

func (r *AttributeValueRepository) List(
	ctx context.Context,
) ([]domain.AttributeValue, error) {
	rows, err := r.queries.ListAttributeValues(ctx)
	if err != nil {
		return nil, mapDBError(
			"list attribute values",
			err,
		)
	}

	values := make(
		[]domain.AttributeValue,
		0,
		len(rows),
	)

	for _, row := range rows {
		values = append(
			values,
			attributeValueToDomain(row),
		)
	}

	return values, nil
}

func (r *AttributeValueRepository) ListByAttributeID(
	ctx context.Context,
	attributeID uuid.UUID,
) ([]domain.AttributeValue, error) {
	rows, err := r.queries.ListAttributeValuesByAttributeID(
		ctx,
		uuidToPG(attributeID),
	)
	if err != nil {
		return nil, mapDBError(
			"list attribute values by attribute",
			err,
		)
	}

	values := make(
		[]domain.AttributeValue,
		0,
		len(rows),
	)

	for _, row := range rows {
		values = append(
			values,
			attributeValueToDomain(row),
		)
	}

	return values, nil
}

func (r *AttributeValueRepository) Update(
	ctx context.Context,
	params repository.UpdateAttributeValueParams,
) (domain.AttributeValue, error) {
	row, err := r.queries.UpdateAttributeValue(
		ctx,
		catalogdb.UpdateAttributeValueParams{
			ID:    uuidToPG(params.ID),
			Value: params.Value,
		},
	)
	if err != nil {
		return domain.AttributeValue{}, mapDBError(
			"update attribute value",
			err,
		)
	}

	return attributeValueToDomain(row), nil
}

func attributeValueToDomain(
	row catalogdb.CatalogAttributeValue,
) domain.AttributeValue {
	return domain.AttributeValue{
		ID:          pgUUIDToUUID(row.ID),
		AttributeID: pgUUIDToUUID(row.AttributeID),
		Value:       row.Value,
	}
}
