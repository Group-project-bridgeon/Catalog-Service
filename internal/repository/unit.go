package repository

import (
	"context"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/google/uuid"
)

type CreateUnitParams struct {
	ID       uuid.UUID
	Code     string
	Name     string
	IsActive bool
}

type UpdateUnitParams struct {
	ID   uuid.UUID
	Code string
	Name string
}

type UnitRepository interface {
	Create(
		context.Context,
		CreateUnitParams,
	) (domain.Unit, error)

	GetByID(
		context.Context,
		uuid.UUID,
	) (domain.Unit, error)

	GetByCode(
		context.Context,
		string,
	) (domain.Unit, error)

	GetByName(
		context.Context,
		string,
	) (domain.Unit, error)

	List(
		context.Context,
	) ([]domain.Unit, error)

	Update(
		context.Context,
		UpdateUnitParams,
	) (domain.Unit, error)

	SetActive(
		context.Context,
		uuid.UUID,
		bool,
	) (domain.Unit, error)
}
