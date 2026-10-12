package service

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/google/uuid"
)

const maxAttributeValueLength = 100

type CreateAttributeValueInput struct {
	AttributeID uuid.UUID
	Value       string
}

type UpdateAttributeValueInput struct {
	Value string
}

type AttributeValueService struct {
	repository repository.AttributeValueRepository
}

func NewAttributeValueService(
	repository repository.AttributeValueRepository,
) *AttributeValueService {
	return &AttributeValueService{
		repository: repository,
	}
}

func (s *AttributeValueService) Create(
	ctx context.Context,
	input CreateAttributeValueInput,
) (domain.AttributeValue, error) {
	if input.AttributeID == uuid.Nil {
		return domain.AttributeValue{}, fmt.Errorf(
			"attribute id is required: %w",
			ErrInvalidArgument,
		)
	}

	value, err := validateAttributeValue(
		input.Value,
	)
	if err != nil {
		return domain.AttributeValue{}, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return domain.AttributeValue{}, fmt.Errorf(
			"generate attribute value id: %w",
			err,
		)
	}

	result, err := s.repository.Create(
		ctx,
		repository.CreateAttributeValueParams{
			ID:          id,
			AttributeID: input.AttributeID,
			Value:       value,
		},
	)
	if err != nil {
		return domain.AttributeValue{},
			mapRepositoryError(
				"create attribute value",
				err,
			)
	}

	return result, nil
}

func (s *AttributeValueService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.AttributeValue, error) {
	result, err := s.repository.GetByID(
		ctx,
		id,
	)
	if err != nil {
		return domain.AttributeValue{},
			mapRepositoryError(
				"get attribute value",
				err,
			)
	}

	return result, nil
}

func (s *AttributeValueService) GetByAttributeAndValue(
	ctx context.Context,
	attributeID uuid.UUID,
	value string,
) (domain.AttributeValue, error) {
	if attributeID == uuid.Nil {
		return domain.AttributeValue{}, fmt.Errorf(
			"attribute id is required: %w",
			ErrInvalidArgument,
		)
	}

	value, err := validateAttributeValue(value)
	if err != nil {
		return domain.AttributeValue{}, err
	}

	result, err := s.repository.GetByAttributeAndValue(
		ctx,
		attributeID,
		value,
	)
	if err != nil {
		return domain.AttributeValue{},
			mapRepositoryError(
				"get attribute value by attribute and value",
				err,
			)
	}

	return result, nil
}

func (s *AttributeValueService) List(
	ctx context.Context,
) ([]domain.AttributeValue, error) {
	values, err := s.repository.List(ctx)
	if err != nil {
		return nil, mapRepositoryError(
			"list attribute values",
			err,
		)
	}

	return values, nil
}

func (s *AttributeValueService) ListByAttributeID(
	ctx context.Context,
	attributeID uuid.UUID,
) ([]domain.AttributeValue, error) {
	if attributeID == uuid.Nil {
		return nil, fmt.Errorf(
			"attribute id is required: %w",
			ErrInvalidArgument,
		)
	}

	values, err := s.repository.ListByAttributeID(
		ctx,
		attributeID,
	)
	if err != nil {
		return nil, mapRepositoryError(
			"list attribute values by attribute",
			err,
		)
	}

	return values, nil
}

func (s *AttributeValueService) Update(
	ctx context.Context,
	id uuid.UUID,
	input UpdateAttributeValueInput,
) (domain.AttributeValue, error) {
	value, err := validateAttributeValue(
		input.Value,
	)
	if err != nil {
		return domain.AttributeValue{}, err
	}

	result, err := s.repository.Update(
		ctx,
		repository.UpdateAttributeValueParams{
			ID:    id,
			Value: value,
		},
	)
	if err != nil {
		return domain.AttributeValue{},
			mapRepositoryError(
				"update attribute value",
				err,
			)
	}

	return result, nil
}

func validateAttributeValue(
	value string,
) (string, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return "", fmt.Errorf(
			"attribute value is required: %w",
			ErrInvalidArgument,
		)
	}

	if utf8.RuneCountInString(value) >
		maxAttributeValueLength {
		return "", fmt.Errorf(
			"attribute value exceeds %d characters: %w",
			maxAttributeValueLength,
			ErrInvalidArgument,
		)
	}

	return value, nil
}
