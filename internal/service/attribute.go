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

const (
	maxAttributeNameLength = 100
	maxAttributeCodeLength = 50
)

type CreateAttributeInput struct {
	Name string
	Code string
}

type UpdateAttributeInput struct {
	Name string
	Code string
}

type AttributeService struct {
	repository repository.AttributeRepository
}

func NewAttributeService(
	repository repository.AttributeRepository,
) *AttributeService {
	return &AttributeService{
		repository: repository,
	}
}

func (s *AttributeService) Create(
	ctx context.Context,
	input CreateAttributeInput,
) (domain.Attribute, error) {
	name, code, err := validateAttributeFields(
		input.Name,
		input.Code,
	)
	if err != nil {
		return domain.Attribute{}, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return domain.Attribute{}, fmt.Errorf(
			"generate attribute id: %w",
			err,
		)
	}

	attribute, err := s.repository.Create(
		ctx,
		repository.CreateAttributeParams{
			ID:   id,
			Name: name,
			Code: code,
		},
	)
	if err != nil {
		return domain.Attribute{},
			mapRepositoryError(
				"create attribute",
				err,
			)
	}

	return attribute, nil
}

func (s *AttributeService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Attribute, error) {
	attribute, err := s.repository.GetByID(
		ctx,
		id,
	)
	if err != nil {
		return domain.Attribute{},
			mapRepositoryError(
				"get attribute",
				err,
			)
	}

	return attribute, nil
}

func (s *AttributeService) GetByCode(
	ctx context.Context,
	code string,
) (domain.Attribute, error) {
	code = strings.TrimSpace(code)

	if code == "" {
		return domain.Attribute{}, fmt.Errorf(
			"attribute code is required: %w",
			ErrInvalidArgument,
		)
	}

	attribute, err := s.repository.GetByCode(
		ctx,
		code,
	)
	if err != nil {
		return domain.Attribute{},
			mapRepositoryError(
				"get attribute by code",
				err,
			)
	}

	return attribute, nil
}

func (s *AttributeService) GetByName(
	ctx context.Context,
	name string,
) (domain.Attribute, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return domain.Attribute{}, fmt.Errorf(
			"attribute name is required: %w",
			ErrInvalidArgument,
		)
	}

	attribute, err := s.repository.GetByName(
		ctx,
		name,
	)
	if err != nil {
		return domain.Attribute{},
			mapRepositoryError(
				"get attribute by name",
				err,
			)
	}

	return attribute, nil
}

func (s *AttributeService) List(
	ctx context.Context,
) ([]domain.Attribute, error) {
	attributes, err := s.repository.List(ctx)
	if err != nil {
		return nil, mapRepositoryError(
			"list attributes",
			err,
		)
	}

	return attributes, nil
}

func (s *AttributeService) Update(
	ctx context.Context,
	id uuid.UUID,
	input UpdateAttributeInput,
) (domain.Attribute, error) {
	name, code, err := validateAttributeFields(
		input.Name,
		input.Code,
	)
	if err != nil {
		return domain.Attribute{}, err
	}

	attribute, err := s.repository.Update(
		ctx,
		repository.UpdateAttributeParams{
			ID:   id,
			Name: name,
			Code: code,
		},
	)
	if err != nil {
		return domain.Attribute{},
			mapRepositoryError(
				"update attribute",
				err,
			)
	}

	return attribute, nil
}

func validateAttributeFields(
	name string,
	code string,
) (string, string, error) {
	name = strings.TrimSpace(name)
	code = strings.TrimSpace(code)

	if name == "" {
		return "", "", fmt.Errorf(
			"attribute name is required: %w",
			ErrInvalidArgument,
		)
	}

	if utf8.RuneCountInString(name) >
		maxAttributeNameLength {
		return "", "", fmt.Errorf(
			"attribute name exceeds %d characters: %w",
			maxAttributeNameLength,
			ErrInvalidArgument,
		)
	}

	if code == "" {
		return "", "", fmt.Errorf(
			"attribute code is required: %w",
			ErrInvalidArgument,
		)
	}

	if utf8.RuneCountInString(code) >
		maxAttributeCodeLength {
		return "", "", fmt.Errorf(
			"attribute code exceeds %d characters: %w",
			maxAttributeCodeLength,
			ErrInvalidArgument,
		)
	}

	return name, code, nil
}
