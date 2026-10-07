package repository

import (
	"context"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/google/uuid"
)

type CreateVariantValueParams struct {
	VariantID        uuid.UUID
	AttributeID      uuid.UUID
	AttributeValueID uuid.UUID
}

type UpdateVariantValueParams struct {
	VariantID        uuid.UUID
	AttributeID      uuid.UUID
	AttributeValueID uuid.UUID
}

type VariantValueRepository interface {
	Create(
		context.Context,
		CreateVariantValueParams,
	) (domain.VariantValue, error)

	GetByVariantAndAttribute(
		context.Context,
		uuid.UUID,
		uuid.UUID,
	) (domain.VariantValue, error)

	List(
		context.Context,
	) ([]domain.VariantValue, error)

	ListByVariantID(
		context.Context,
		uuid.UUID,
	) ([]domain.VariantValue, error)

	Update(
		context.Context,
		UpdateVariantValueParams,
	) (domain.VariantValue, error)

	CombinationExists(
		context.Context,
		uuid.UUID,
		uuid.UUID,
		uuid.UUID,
	) (bool, error)
}
