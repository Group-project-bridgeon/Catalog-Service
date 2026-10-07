package service

import (
	"context"
	"fmt"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/google/uuid"
)

type CreateVariantValueInput struct {
	VariantID        uuid.UUID
	AttributeID      uuid.UUID
	AttributeValueID uuid.UUID
}

type UpdateVariantValueInput struct {
	AttributeValueID uuid.UUID
}

type VariantValueService struct {
	repository repository.VariantValueRepository
}

func NewVariantValueService(
	repository repository.VariantValueRepository,
) *VariantValueService {
	return &VariantValueService{
		repository: repository,
	}
}

func (s *VariantValueService) Create(
	ctx context.Context,
	input CreateVariantValueInput,
) (domain.VariantValue, error) {
	if err := validateVariantValueIDs(
		input.VariantID,
		input.AttributeID,
		input.AttributeValueID,
	); err != nil {
		return domain.VariantValue{}, err
	}

	exists, err := s.repository.CombinationExists(
		ctx,
		input.VariantID,
		input.AttributeID,
		input.AttributeValueID,
	)
	if err != nil {
		return domain.VariantValue{},
			mapRepositoryError(
				"check variant combination",
				err,
			)
	}

	if exists {
		return domain.VariantValue{}, fmt.Errorf(
			"variant attribute combination already exists: %w",
			ErrConflict,
		)
	}

	result, err := s.repository.Create(
		ctx,
		repository.CreateVariantValueParams{
			VariantID:        input.VariantID,
			AttributeID:      input.AttributeID,
			AttributeValueID: input.AttributeValueID,
		},
	)
	if err != nil {
		return domain.VariantValue{},
			mapRepositoryError(
				"create variant value",
				err,
			)
	}

	return result, nil
}

func (s *VariantValueService) GetByVariantAndAttribute(
	ctx context.Context,
	variantID uuid.UUID,
	attributeID uuid.UUID,
) (domain.VariantValue, error) {
	if variantID == uuid.Nil {
		return domain.VariantValue{}, fmt.Errorf(
			"variant id is required: %w",
			ErrInvalidArgument,
		)
	}

	if attributeID == uuid.Nil {
		return domain.VariantValue{}, fmt.Errorf(
			"attribute id is required: %w",
			ErrInvalidArgument,
		)
	}

	result, err := s.repository.GetByVariantAndAttribute(
		ctx,
		variantID,
		attributeID,
	)
	if err != nil {
		return domain.VariantValue{},
			mapRepositoryError(
				"get variant value",
				err,
			)
	}

	return result, nil
}

func (s *VariantValueService) List(
	ctx context.Context,
) ([]domain.VariantValue, error) {
	values, err := s.repository.List(ctx)
	if err != nil {
		return nil, mapRepositoryError(
			"list variant values",
			err,
		)
	}

	return values, nil
}

func (s *VariantValueService) ListByVariantID(
	ctx context.Context,
	variantID uuid.UUID,
) ([]domain.VariantValue, error) {
	if variantID == uuid.Nil {
		return nil, fmt.Errorf(
			"variant id is required: %w",
			ErrInvalidArgument,
		)
	}

	values, err := s.repository.ListByVariantID(
		ctx,
		variantID,
	)
	if err != nil {
		return nil, mapRepositoryError(
			"list variant values by variant",
			err,
		)
	}

	return values, nil
}

func (s *VariantValueService) Update(
	ctx context.Context,
	variantID uuid.UUID,
	attributeID uuid.UUID,
	input UpdateVariantValueInput,
) (domain.VariantValue, error) {
	if err := validateVariantValueIDs(
		variantID,
		attributeID,
		input.AttributeValueID,
	); err != nil {
		return domain.VariantValue{}, err
	}

	exists, err := s.repository.CombinationExists(
		ctx,
		variantID,
		attributeID,
		input.AttributeValueID,
	)
	if err != nil {
		return domain.VariantValue{},
			mapRepositoryError(
				"check variant combination",
				err,
			)
	}

	if exists {
		return domain.VariantValue{}, fmt.Errorf(
			"variant attribute combination already exists: %w",
			ErrConflict,
		)
	}

	result, err := s.repository.Update(
		ctx,
		repository.UpdateVariantValueParams{
			VariantID:        variantID,
			AttributeID:      attributeID,
			AttributeValueID: input.AttributeValueID,
		},
	)
	if err != nil {
		return domain.VariantValue{},
			mapRepositoryError(
				"update variant value",
				err,
			)
	}

	return result, nil
}

func validateVariantValueIDs(
	variantID uuid.UUID,
	attributeID uuid.UUID,
	attributeValueID uuid.UUID,
) error {
	if variantID == uuid.Nil {
		return fmt.Errorf(
			"variant id is required: %w",
			ErrInvalidArgument,
		)
	}

	if attributeID == uuid.Nil {
		return fmt.Errorf(
			"attribute id is required: %w",
			ErrInvalidArgument,
		)
	}

	if attributeValueID == uuid.Nil {
		return fmt.Errorf(
			"attribute value id is required: %w",
			ErrInvalidArgument,
		)
	}

	return nil
}
