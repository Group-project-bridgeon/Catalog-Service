package postgres

import (
	"context"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository/postgres/catalogdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type unitQueries interface {
	CreateUnit(
		context.Context,
		catalogdb.CreateUnitParams,
	) (catalogdb.CatalogUnit, error)

	GetUnitByID(
		context.Context,
		pgtype.UUID,
	) (catalogdb.CatalogUnit, error)

	GetUnitByCode(
		context.Context,
		string,
	) (catalogdb.CatalogUnit, error)

	GetUnitByName(
		context.Context,
		string,
	) (catalogdb.CatalogUnit, error)

	ListUnits(
		context.Context,
	) ([]catalogdb.CatalogUnit, error)

	UpdateUnit(
		context.Context,
		catalogdb.UpdateUnitParams,
	) (catalogdb.CatalogUnit, error)

	SetUnitActive(
		context.Context,
		catalogdb.SetUnitActiveParams,
	) (catalogdb.CatalogUnit, error)
}

type UnitRepository struct {
	queries unitQueries
}

var _ repository.UnitRepository = (*UnitRepository)(nil)

func NewUnitRepository(
	queries unitQueries,
) *UnitRepository {
	return &UnitRepository{
		queries: queries,
	}
}

func (r *UnitRepository) Create(
	ctx context.Context,
	params repository.CreateUnitParams,
) (domain.Unit, error) {
	row, err := r.queries.CreateUnit(
		ctx,
		catalogdb.CreateUnitParams{
			ID:       uuidToPG(params.ID),
			Code:     params.Code,
			Name:     params.Name,
			IsActive: params.IsActive,
		},
	)
	if err != nil {
		return domain.Unit{}, mapDBError(
			"create unit",
			err,
		)
	}

	return unitToDomain(row), nil
}

func (r *UnitRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Unit, error) {
	row, err := r.queries.GetUnitByID(
		ctx,
		uuidToPG(id),
	)
	if err != nil {
		return domain.Unit{}, mapDBError(
			"get unit by id",
			err,
		)
	}

	return unitToDomain(row), nil
}

func (r *UnitRepository) GetByCode(
	ctx context.Context,
	code string,
) (domain.Unit, error) {
	row, err := r.queries.GetUnitByCode(
		ctx,
		code,
	)
	if err != nil {
		return domain.Unit{}, mapDBError(
			"get unit by code",
			err,
		)
	}

	return unitToDomain(row), nil
}

func (r *UnitRepository) GetByName(
	ctx context.Context,
	name string,
) (domain.Unit, error) {
	row, err := r.queries.GetUnitByName(
		ctx,
		name,
	)
	if err != nil {
		return domain.Unit{}, mapDBError(
			"get unit by name",
			err,
		)
	}

	return unitToDomain(row), nil
}

func (r *UnitRepository) List(
	ctx context.Context,
) ([]domain.Unit, error) {
	rows, err := r.queries.ListUnits(ctx)
	if err != nil {
		return nil, mapDBError(
			"list units",
			err,
		)
	}

	units := make(
		[]domain.Unit,
		0,
		len(rows),
	)

	for _, row := range rows {
		units = append(
			units,
			unitToDomain(row),
		)
	}

	return units, nil
}

func (r *UnitRepository) Update(
	ctx context.Context,
	params repository.UpdateUnitParams,
) (domain.Unit, error) {
	row, err := r.queries.UpdateUnit(
		ctx,
		catalogdb.UpdateUnitParams{
			ID:   uuidToPG(params.ID),
			Code: params.Code,
			Name: params.Name,
		},
	)
	if err != nil {
		return domain.Unit{}, mapDBError(
			"update unit",
			err,
		)
	}

	return unitToDomain(row), nil
}

func (r *UnitRepository) SetActive(
	ctx context.Context,
	id uuid.UUID,
	isActive bool,
) (domain.Unit, error) {
	row, err := r.queries.SetUnitActive(
		ctx,
		catalogdb.SetUnitActiveParams{
			ID:       uuidToPG(id),
			IsActive: isActive,
		},
	)
	if err != nil {
		return domain.Unit{}, mapDBError(
			"set unit active",
			err,
		)
	}

	return unitToDomain(row), nil
}

func unitToDomain(
	row catalogdb.CatalogUnit,
) domain.Unit {
	return domain.Unit{
		ID:        pgUUIDToUUID(row.ID),
		Code:      row.Code,
		Name:      row.Name,
		IsActive:  row.IsActive,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
