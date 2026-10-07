package postgres

import (
	"context"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository/postgres/catalogdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type attributeQueries interface {
	CreateAttribute(
		context.Context,
		catalogdb.CreateAttributeParams,
	) (catalogdb.CatalogAttribute, error)

	GetAttributeByID(
		context.Context,
		pgtype.UUID,
	) (catalogdb.CatalogAttribute, error)

	GetAttributeByCode(
		context.Context,
		string,
	) (catalogdb.CatalogAttribute, error)

	GetAttributeByName(
		context.Context,
		string,
	) (catalogdb.CatalogAttribute, error)

	ListAttributes(
		context.Context,
	) ([]catalogdb.CatalogAttribute, error)

	UpdateAttribute(
		context.Context,
		catalogdb.UpdateAttributeParams,
	) (catalogdb.CatalogAttribute, error)
}

type AttributeRepository struct {
	queries attributeQueries
}

var _ repository.AttributeRepository = (*AttributeRepository)(nil)

func NewAttributeRepository(
	queries attributeQueries,
) *AttributeRepository {
	return &AttributeRepository{
		queries: queries,
	}
}

func (r *AttributeRepository) Create(
	ctx context.Context,
	params repository.CreateAttributeParams,
) (domain.Attribute, error) {
	row, err := r.queries.CreateAttribute(
		ctx,
		catalogdb.CreateAttributeParams{
			ID:   uuidToPG(params.ID),
			Name: params.Name,
			Code: params.Code,
		},
	)
	if err != nil {
		return domain.Attribute{}, mapDBError(
			"create attribute",
			err,
		)
	}

	return attributeToDomain(row), nil
}

func (r *AttributeRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Attribute, error) {
	row, err := r.queries.GetAttributeByID(
		ctx,
		uuidToPG(id),
	)
	if err != nil {
		return domain.Attribute{}, mapDBError(
			"get attribute by id",
			err,
		)
	}

	return attributeToDomain(row), nil
}

func (r *AttributeRepository) GetByCode(
	ctx context.Context,
	code string,
) (domain.Attribute, error) {
	row, err := r.queries.GetAttributeByCode(
		ctx,
		code,
	)
	if err != nil {
		return domain.Attribute{}, mapDBError(
			"get attribute by code",
			err,
		)
	}

	return attributeToDomain(row), nil
}

func (r *AttributeRepository) GetByName(
	ctx context.Context,
	name string,
) (domain.Attribute, error) {
	row, err := r.queries.GetAttributeByName(
		ctx,
		name,
	)
	if err != nil {
		return domain.Attribute{}, mapDBError(
			"get attribute by name",
			err,
		)
	}

	return attributeToDomain(row), nil
}

func (r *AttributeRepository) List(
	ctx context.Context,
) ([]domain.Attribute, error) {
	rows, err := r.queries.ListAttributes(ctx)
	if err != nil {
		return nil, mapDBError(
			"list attributes",
			err,
		)
	}

	attributes := make(
		[]domain.Attribute,
		0,
		len(rows),
	)

	for _, row := range rows {
		attributes = append(
			attributes,
			attributeToDomain(row),
		)
	}

	return attributes, nil
}

func (r *AttributeRepository) Update(
	ctx context.Context,
	params repository.UpdateAttributeParams,
) (domain.Attribute, error) {
	row, err := r.queries.UpdateAttribute(
		ctx,
		catalogdb.UpdateAttributeParams{
			ID:   uuidToPG(params.ID),
			Name: params.Name,
			Code: params.Code,
		},
	)
	if err != nil {
		return domain.Attribute{}, mapDBError(
			"update attribute",
			err,
		)
	}

	return attributeToDomain(row), nil
}

func attributeToDomain(
	row catalogdb.CatalogAttribute,
) domain.Attribute {
	return domain.Attribute{
		ID:   pgUUIDToUUID(row.ID),
		Name: row.Name,
		Code: row.Code,
	}
}
