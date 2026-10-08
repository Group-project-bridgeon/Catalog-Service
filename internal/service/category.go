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
	maxCategoryNameLength = 100
	maxCategoryCodeLength = 50
)

type CreateCategoryInput struct {
	ParentID    *uuid.UUID
	Name        string
	Code        string
	Description *string
}

type UpdateCategoryInput struct {
	ParentID    *uuid.UUID
	Name        string
	Code        string
	Description *string
}

type CategoryService struct {
	repository repository.CategoryRepository
}

func NewCategoryService(
	repository repository.CategoryRepository,
) *CategoryService {
	return &CategoryService{
		repository: repository,
	}
}

func (s *CategoryService) Create(
	ctx context.Context,
	input CreateCategoryInput,
) (domain.Category, error) {
	name, code, err := validateCategoryFields(
		input.Name,
		input.Code,
	)
	if err != nil {
		return domain.Category{}, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return domain.Category{}, fmt.Errorf(
			"generate category id: %w",
			err,
		)
	}

	category, err := s.repository.Create(
		ctx,
		repository.CreateCategoryParams{
			ID:          id,
			ParentID:    input.ParentID,
			Name:        name,
			Code:        code,
			Description: input.Description,
			IsActive:    true,
		},
	)
	if err != nil {
		return domain.Category{}, mapRepositoryError(
			"create category",
			err,
		)
	}

	return category, nil
}

func (s *CategoryService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Category, error) {
	category, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return domain.Category{}, mapRepositoryError(
			"get category",
			err,
		)
	}

	return category, nil
}

func (s *CategoryService) GetByCode(
	ctx context.Context,
	code string,
) (domain.Category, error) {
	code = strings.TrimSpace(code)

	if code == "" {
		return domain.Category{}, fmt.Errorf(
			"category code is required: %w",
			ErrInvalidArgument,
		)
	}

	category, err := s.repository.GetByCode(ctx, code)
	if err != nil {
		return domain.Category{}, mapRepositoryError(
			"get category by code",
			err,
		)
	}

	return category, nil
}

func (s *CategoryService) List(
	ctx context.Context,
) ([]domain.Category, error) {
	categories, err := s.repository.List(ctx)
	if err != nil {
		return nil, mapRepositoryError(
			"list categories",
			err,
		)
	}

	return categories, nil
}

func (s *CategoryService) Update(
	ctx context.Context,
	id uuid.UUID,
	input UpdateCategoryInput,
) (domain.Category, error) {
	name, code, err := validateCategoryFields(
		input.Name,
		input.Code,
	)
	if err != nil {
		return domain.Category{}, err
	}

	category, err := s.repository.Update(
		ctx,
		repository.UpdateCategoryParams{
			ID:          id,
			ParentID:    input.ParentID,
			Name:        name,
			Code:        code,
			Description: input.Description,
		},
	)
	if err != nil {
		return domain.Category{}, mapRepositoryError(
			"update category",
			err,
		)
	}

	return category, nil
}

func (s *CategoryService) SetActive(
	ctx context.Context,
	id uuid.UUID,
	isActive bool,
) (domain.Category, error) {
	category, err := s.repository.SetActive(
		ctx,
		id,
		isActive,
	)
	if err != nil {
		return domain.Category{}, mapRepositoryError(
			"set category active",
			err,
		)
	}

	return category, nil
}

func validateCategoryFields(
	name string,
	code string,
) (string, string, error) {
	name = strings.TrimSpace(name)
	code = strings.TrimSpace(code)

	if name == "" {
		return "", "", fmt.Errorf(
			"category name is required: %w",
			ErrInvalidArgument,
		)
	}

	if utf8.RuneCountInString(name) > maxCategoryNameLength {
		return "", "", fmt.Errorf(
			"category name exceeds %d characters: %w",
			maxCategoryNameLength,
			ErrInvalidArgument,
		)
	}

	if code == "" {
		return "", "", fmt.Errorf(
			"category code is required: %w",
			ErrInvalidArgument,
		)
	}

	if utf8.RuneCountInString(code) > maxCategoryCodeLength {
		return "", "", fmt.Errorf(
			"category code exceeds %d characters: %w",
			maxCategoryCodeLength,
			ErrInvalidArgument,
		)
	}

	return name, code, nil
}
