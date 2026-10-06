package repository

import (
	"context"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/google/uuid"
)

type CreateCategoryParams struct {
	ID          uuid.UUID
	ParentID    *uuid.UUID
	Name        string
	Code        string
	Description *string
	IsActive    bool
}

type UpdateCategoryParams struct {
	ID          uuid.UUID
	ParentID    *uuid.UUID
	Name        string
	Code        string
	Description *string
}

type CategoryRepository interface {
	Create(
		ctx context.Context,
		params CreateCategoryParams,
	) (domain.Category, error)

	GetByID(
		ctx context.Context,
		id uuid.UUID,
	) (domain.Category, error)

	GetByCode(
		ctx context.Context,
		code string,
	) (domain.Category, error)

	List(
		ctx context.Context,
	) ([]domain.Category, error)

	Update(
		ctx context.Context,
		params UpdateCategoryParams,
	) (domain.Category, error)

	SetActive(
		ctx context.Context,
		id uuid.UUID,
		isActive bool,
	) (domain.Category, error)
}
