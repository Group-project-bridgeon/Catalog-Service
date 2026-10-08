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
	maxBrandNameLength = 100
	maxBrandCodeLength = 50
)

type CreateBrandInput struct {
	Name string
	Code string
}

type UpdateBrandInput struct {
	Name string
	Code string
}

type BrandService struct {
	repository repository.BrandRepository
}

func NewBrandService(
	repository repository.BrandRepository,
) *BrandService {
	return &BrandService{
		repository: repository,
	}
}

func (s *BrandService) Create(
	ctx context.Context,
	input CreateBrandInput,
) (domain.Brand, error) {
	name, code, err := validateBrandFields(
		input.Name,
		input.Code,
	)
	if err != nil {
		return domain.Brand{}, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return domain.Brand{}, fmt.Errorf(
			"generate brand id: %w",
			err,
		)
	}

	brand, err := s.repository.Create(
		ctx,
		repository.CreateBrandParams{
			ID:       id,
			Name:     name,
			Code:     code,
			IsActive: true,
		},
	)
	if err != nil {
		return domain.Brand{}, mapRepositoryError(
			"create brand",
			err,
		)
	}

	return brand, nil
}

func (s *BrandService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Brand, error) {
	brand, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return domain.Brand{}, mapRepositoryError(
			"get brand",
			err,
		)
	}

	return brand, nil
}

func (s *BrandService) GetByCode(
	ctx context.Context,
	code string,
) (domain.Brand, error) {
	code = strings.TrimSpace(code)

	if code == "" {
		return domain.Brand{}, fmt.Errorf(
			"brand code is required: %w",
			ErrInvalidArgument,
		)
	}

	brand, err := s.repository.GetByCode(ctx, code)
	if err != nil {
		return domain.Brand{}, mapRepositoryError(
			"get brand by code",
			err,
		)
	}

	return brand, nil
}

func (s *BrandService) GetByName(
	ctx context.Context,
	name string,
) (domain.Brand, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return domain.Brand{}, fmt.Errorf(
			"brand name is required: %w",
			ErrInvalidArgument,
		)
	}

	brand, err := s.repository.GetByName(ctx, name)
	if err != nil {
		return domain.Brand{}, mapRepositoryError(
			"get brand by name",
			err,
		)
	}

	return brand, nil
}

func (s *BrandService) List(
	ctx context.Context,
) ([]domain.Brand, error) {
	brands, err := s.repository.List(ctx)
	if err != nil {
		return nil, mapRepositoryError(
			"list brands",
			err,
		)
	}

	return brands, nil
}

func (s *BrandService) Update(
	ctx context.Context,
	id uuid.UUID,
	input UpdateBrandInput,
) (domain.Brand, error) {
	name, code, err := validateBrandFields(
		input.Name,
		input.Code,
	)
	if err != nil {
		return domain.Brand{}, err
	}

	brand, err := s.repository.Update(
		ctx,
		repository.UpdateBrandParams{
			ID:   id,
			Name: name,
			Code: code,
		},
	)
	if err != nil {
		return domain.Brand{}, mapRepositoryError(
			"update brand",
			err,
		)
	}

	return brand, nil
}

func (s *BrandService) SetActive(
	ctx context.Context,
	id uuid.UUID,
	isActive bool,
) (domain.Brand, error) {
	brand, err := s.repository.SetActive(
		ctx,
		id,
		isActive,
	)
	if err != nil {
		return domain.Brand{}, mapRepositoryError(
			"set brand active",
			err,
		)
	}

	return brand, nil
}

func validateBrandFields(
	name string,
	code string,
) (string, string, error) {
	name = strings.TrimSpace(name)
	code = strings.TrimSpace(code)

	if name == "" {
		return "", "", fmt.Errorf(
			"brand name is required: %w",
			ErrInvalidArgument,
		)
	}

	if utf8.RuneCountInString(name) > maxBrandNameLength {
		return "", "", fmt.Errorf(
			"brand name exceeds %d characters: %w",
			maxBrandNameLength,
			ErrInvalidArgument,
		)
	}

	if code == "" {
		return "", "", fmt.Errorf(
			"brand code is required: %w",
			ErrInvalidArgument,
		)
	}

	if utf8.RuneCountInString(code) > maxBrandCodeLength {
		return "", "", fmt.Errorf(
			"brand code exceeds %d characters: %w",
			maxBrandCodeLength,
			ErrInvalidArgument,
		)
	}

	return name, code, nil
}
