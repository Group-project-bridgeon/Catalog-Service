package repository

import (
	"context"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/google/uuid"
)

type CreateProductParams struct {
	ID          uuid.UUID
	CategoryID  uuid.UUID
	BrandID     *uuid.UUID
	Name        string
	ProductCode string
	Description *string
	IsActive    bool
}

type UpdateProductParams struct {
	ID          uuid.UUID
	CategoryID  uuid.UUID
	BrandID     *uuid.UUID
	Name        string
	ProductCode string
	Description *string
}

type ProductRepository interface {
	Create(
		context.Context,
		CreateProductParams,
	) (domain.Product, error)

	GetByID(
		context.Context,
		uuid.UUID,
	) (domain.Product, error)

	GetByCode(
		context.Context,
		string,
	) (domain.Product, error)

	List(
		context.Context,
	) ([]domain.Product, error)

	Update(
		context.Context,
		UpdateProductParams,
	) (domain.Product, error)

	SetActive(
		context.Context,
		uuid.UUID,
		bool,
	) (domain.Product, error)
}
