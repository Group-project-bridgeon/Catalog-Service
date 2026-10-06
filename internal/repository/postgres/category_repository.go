package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository/postgres/catalogdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type categoryQueries interface {
	CreateCategory(
		context.Context,
		catalogdb.CreateCategoryParams,
	) (catalogdb.CatalogCategory, error)

	GetCategoryByID(
		context.Context,
		pgtype.UUID,
	) (catalogdb.CatalogCategory, error)

	GetCategoryByCode(
		context.Context,
		string,
	) (catalogdb.CatalogCategory, error)

	ListCategories(
		context.Context,
	) ([]catalogdb.CatalogCategory, error)

	UpdateCategory(
		context.Context,
		catalogdb.UpdateCategoryParams,
	) (catalogdb.CatalogCategory, error)

	SetCategoryActive(
		context.Context,
		catalogdb.SetCategoryActiveParams,
	) (catalogdb.CatalogCategory, error)
}

type CategoryRepository struct {
	queries categoryQueries
}

var _ repository.CategoryRepository = (*CategoryRepository)(nil)

func NewCategoryRepository(queries categoryQueries) *CategoryRepository {
	return &CategoryRepository{
		queries: queries,
	}
}

func (r *CategoryRepository) Create(
	ctx context.Context,
	params repository.CreateCategoryParams,
) (domain.Category, error) {
	row, err := r.queries.CreateCategory(
		ctx,
		catalogdb.CreateCategoryParams{
			ID:          uuidToPG(params.ID),
			ParentID:    nullableUUIDToPG(params.ParentID),
			Name:        params.Name,
			Code:        params.Code,
			Description: stringPtrToPGText(params.Description),
			IsActive:    params.IsActive,
		},
	)
	if err != nil {
		return domain.Category{}, mapCategoryDBError(
			"create category",
			err,
		)
	}

	return categoryToDomain(row), nil
}

func (r *CategoryRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Category, error) {
	row, err := r.queries.GetCategoryByID(
		ctx,
		uuidToPG(id),
	)
	if err != nil {
		return domain.Category{}, mapCategoryDBError(
			"get category by id",
			err,
		)
	}

	return categoryToDomain(row), nil
}

func (r *CategoryRepository) GetByCode(
	ctx context.Context,
	code string,
) (domain.Category, error) {
	row, err := r.queries.GetCategoryByCode(
		ctx,
		code,
	)
	if err != nil {
		return domain.Category{}, mapCategoryDBError(
			"get category by code",
			err,
		)
	}

	return categoryToDomain(row), nil
}

func (r *CategoryRepository) List(
	ctx context.Context,
) ([]domain.Category, error) {
	rows, err := r.queries.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"list categories: %w",
			err,
		)
	}

	categories := make(
		[]domain.Category,
		0,
		len(rows),
	)

	for _, row := range rows {
		categories = append(
			categories,
			categoryToDomain(row),
		)
	}

	return categories, nil
}

func (r *CategoryRepository) Update(
	ctx context.Context,
	params repository.UpdateCategoryParams,
) (domain.Category, error) {
	row, err := r.queries.UpdateCategory(
		ctx,
		catalogdb.UpdateCategoryParams{
			ID:          uuidToPG(params.ID),
			ParentID:    nullableUUIDToPG(params.ParentID),
			Name:        params.Name,
			Code:        params.Code,
			Description: stringPtrToPGText(params.Description),
		},
	)
	if err != nil {
		return domain.Category{}, mapCategoryDBError(
			"update category",
			err,
		)
	}

	return categoryToDomain(row), nil
}

func (r *CategoryRepository) SetActive(
	ctx context.Context,
	id uuid.UUID,
	isActive bool,
) (domain.Category, error) {
	row, err := r.queries.SetCategoryActive(
		ctx,
		catalogdb.SetCategoryActiveParams{
			ID:       uuidToPG(id),
			IsActive: isActive,
		},
	)
	if err != nil {
		return domain.Category{}, mapCategoryDBError(
			"set category active",
			err,
		)
	}

	return categoryToDomain(row), nil
}

func categoryToDomain(
	row catalogdb.CatalogCategory,
) domain.Category {
	return domain.Category{
		ID:          pgUUIDToUUID(row.ID),
		ParentID:    pgUUIDToNullableUUID(row.ParentID),
		Name:        row.Name,
		Code:        row.Code,
		Description: pgTextToStringPtr(row.Description),
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}

func mapCategoryDBError(
	operation string,
	err error,
) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf(
			"%s: %w",
			operation,
			repository.ErrNotFound,
		)
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return fmt.Errorf(
				"%s: %w: %w",
				operation,
				repository.ErrConflict,
				err,
			)

		case "23503":
			return fmt.Errorf(
				"%s: %w: %w",
				operation,
				repository.ErrInvalidReference,
				err,
			)
		}
	}

	return fmt.Errorf(
		"%s: %w",
		operation,
		err,
	)
}
