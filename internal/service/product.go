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
	maxProductNameLength = 200
	maxProductCodeLength = 50
)

type CreateProductInput struct {
	CategoryID  uuid.UUID
	BrandID     *uuid.UUID
	Name        string
	ProductCode string
	Description *string
}

type UpdateProductInput struct {
	CategoryID  uuid.UUID
	BrandID     *uuid.UUID
	Name        string
	ProductCode string
	Description *string
}

type ProductService struct {
	repository repository.ProductRepository
}

func NewProductService(
	repository repository.ProductRepository,
) *ProductService {
	return &ProductService{
		repository: repository,
	}
}

func (s *ProductService) Create(
	ctx context.Context,
	input CreateProductInput,
) (domain.Product, error) {
	input, err := validateCreateProductInput(input)
	if err != nil {
		return domain.Product{}, err
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return domain.Product{}, fmt.Errorf(
			"generate product id: %w",
			err,
		)
	}

	product, err := s.repository.Create(
		ctx,
		repository.CreateProductParams{
			ID:          id,
			CategoryID:  input.CategoryID,
			BrandID:     input.BrandID,
			Name:        input.Name,
			ProductCode: input.ProductCode,
			Description: input.Description,
			IsActive:    true,
		},
	)
	if err != nil {
		return domain.Product{}, mapRepositoryError(
			"create product",
			err,
		)
	}

	return product, nil
}

func (s *ProductService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Product, error) {
	product, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return domain.Product{}, mapRepositoryError(
			"get product",
			err,
		)
	}

	return product, nil
}

func (s *ProductService) GetByCode(
	ctx context.Context,
	code string,
) (domain.Product, error) {
	code = strings.TrimSpace(code)

	if code == "" {
		return domain.Product{}, fmt.Errorf(
			"product code is required: %w",
			ErrInvalidArgument,
		)
	}

	product, err := s.repository.GetByCode(ctx, code)
	if err != nil {
		return domain.Product{}, mapRepositoryError(
			"get product by code",
			err,
		)
	}

	return product, nil
}

func (s *ProductService) List(
	ctx context.Context,
) ([]domain.Product, error) {
	products, err := s.repository.List(ctx)
	if err != nil {
		return nil, mapRepositoryError(
			"list products",
			err,
		)
	}

	return products, nil
}

func (s *ProductService) Update(
	ctx context.Context,
	id uuid.UUID,
	input UpdateProductInput,
) (domain.Product, error) {
	input, err := validateUpdateProductInput(input)
	if err != nil {
		return domain.Product{}, err
	}

	product, err := s.repository.Update(
		ctx,
		repository.UpdateProductParams{
			ID:          id,
			CategoryID:  input.CategoryID,
			BrandID:     input.BrandID,
			Name:        input.Name,
			ProductCode: input.ProductCode,
			Description: input.Description,
		},
	)
	if err != nil {
		return domain.Product{}, mapRepositoryError(
			"update product",
			err,
		)
	}

	return product, nil
}

func (s *ProductService) SetActive(
	ctx context.Context,
	id uuid.UUID,
	isActive bool,
) (domain.Product, error) {
	product, err := s.repository.SetActive(
		ctx,
		id,
		isActive,
	)
	if err != nil {
		return domain.Product{}, mapRepositoryError(
			"set product active",
			err,
		)
	}

	return product, nil
}

func validateCreateProductInput(
	input CreateProductInput,
) (CreateProductInput, error) {
	name, code, description, err := validateProductFields(
		input.CategoryID,
		input.BrandID,
		input.Name,
		input.ProductCode,
		input.Description,
	)
	if err != nil {
		return CreateProductInput{}, err
	}

	input.Name = name
	input.ProductCode = code
	input.Description = description

	return input, nil
}

func validateUpdateProductInput(
	input UpdateProductInput,
) (UpdateProductInput, error) {
	name, code, description, err := validateProductFields(
		input.CategoryID,
		input.BrandID,
		input.Name,
		input.ProductCode,
		input.Description,
	)
	if err != nil {
		return UpdateProductInput{}, err
	}

	input.Name = name
	input.ProductCode = code
	input.Description = description

	return input, nil
}

func validateProductFields(
	categoryID uuid.UUID,
	brandID *uuid.UUID,
	name string,
	productCode string,
	description *string,
) (string, string, *string, error) {
	if categoryID == uuid.Nil {
		return "", "", nil, fmt.Errorf(
			"category id is required: %w",
			ErrInvalidArgument,
		)
	}

	if brandID != nil && *brandID == uuid.Nil {
		return "", "", nil, fmt.Errorf(
			"brand id is invalid: %w",
			ErrInvalidArgument,
		)
	}

	name = strings.TrimSpace(name)

	if name == "" {
		return "", "", nil, fmt.Errorf(
			"product name is required: %w",
			ErrInvalidArgument,
		)
	}

	if utf8.RuneCountInString(name) > maxProductNameLength {
		return "", "", nil, fmt.Errorf(
			"product name exceeds %d characters: %w",
			maxProductNameLength,
			ErrInvalidArgument,
		)
	}

	productCode = strings.TrimSpace(productCode)

	if productCode == "" {
		return "", "", nil, fmt.Errorf(
			"product code is required: %w",
			ErrInvalidArgument,
		)
	}

	if utf8.RuneCountInString(productCode) > maxProductCodeLength {
		return "", "", nil, fmt.Errorf(
			"product code exceeds %d characters: %w",
			maxProductCodeLength,
			ErrInvalidArgument,
		)
	}

	description = normalizeOptionalString(description)

	return name, productCode, description, nil
}

func normalizeOptionalString(
	value *string,
) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)

	if trimmed == "" {
		return nil
	}

	return &trimmed
}
