package repository

//Defines what Brand persistence operations the service layer can use.

import (
	"context"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/google/uuid"
)

// data containers used to pass the required information from the service layer to the repository.
type CreateBrandParams struct {
	ID       uuid.UUID
	Name     string
	Code     string
	IsActive bool
}

// data containers used to pass the required information from the service layer to the repository.
type UpdateBrandParams struct {
	ID   uuid.UUID
	Name string
	Code string
}

type BrandRepository interface {
	Create(
		context.Context,
		CreateBrandParams,
	) (domain.Brand, error)

	GetByID(
		context.Context,
		uuid.UUID,
	) (domain.Brand, error)

	GetByCode(
		context.Context,
		string,
	) (domain.Brand, error)

	GetByName(
		context.Context,
		string,
	) (domain.Brand, error)

	List(
		context.Context,
	) ([]domain.Brand, error)

	Update(
		context.Context,
		UpdateBrandParams,
	) (domain.Brand, error)

	SetActive(
		context.Context,
		uuid.UUID,
		bool,
	) (domain.Brand, error)
}
