package repository

import (
	"context"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/google/uuid"
)

type CreateAttributeValueParams struct {
	ID          uuid.UUID
	AttributeID uuid.UUID
	Value       string
}

type UpdateAttributeValueParams struct {
	ID    uuid.UUID
	Value string
}

type AttributeValueRepository interface {
	Create(
		context.Context,
		CreateAttributeValueParams,
	) (domain.AttributeValue, error)

	GetByID(
		context.Context,
		uuid.UUID,
	) (domain.AttributeValue, error)

	GetByAttributeAndValue(
		context.Context,
		uuid.UUID,
		string,
	) (domain.AttributeValue, error)

	List(
		context.Context,
	) ([]domain.AttributeValue, error)

	ListByAttributeID(
		context.Context,
		uuid.UUID,
	) ([]domain.AttributeValue, error)

	Update(
		context.Context,
		UpdateAttributeValueParams,
	) (domain.AttributeValue, error)
}
