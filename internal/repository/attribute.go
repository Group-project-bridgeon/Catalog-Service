package repository

import (
	"context"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/google/uuid"
)

type CreateAttributeParams struct {
	ID   uuid.UUID
	Name string
	Code string
}

type UpdateAttributeParams struct {
	ID   uuid.UUID
	Name string
	Code string
}

type AttributeRepository interface {
	Create(
		context.Context,
		CreateAttributeParams,
	) (domain.Attribute, error)

	GetByID(
		context.Context,
		uuid.UUID,
	) (domain.Attribute, error)

	GetByCode(
		context.Context,
		string,
	) (domain.Attribute, error)

	GetByName(
		context.Context,
		string,
	) (domain.Attribute, error)

	List(
		context.Context,
	) ([]domain.Attribute, error)

	Update(
		context.Context,
		UpdateAttributeParams,
	) (domain.Attribute, error)
}
