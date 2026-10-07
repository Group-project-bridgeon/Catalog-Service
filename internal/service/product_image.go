package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/google/uuid"
)

type CreateProductImageInput struct {
	ProductID uuid.UUID
	VariantID *uuid.UUID
	ImageURL  string
	SortOrder int32
}

type UpdateProductImageInput struct {
	VariantID *uuid.UUID
	ImageURL  string
	SortOrder int32
}

type ProductImageService struct {
	repository repository.ProductImageRepository
}

func NewProductImageService(
	repository repository.ProductImageRepository,
) *ProductImageService {
	return &ProductImageService{
		repository: repository,
	}
}

func (s *ProductImageService) Create(
	ctx context.Context,
	input CreateProductImageInput,
) (domain.ProductImage, error) {
	if input.ProductID == uuid.Nil {
		return domain.ProductImage{}, fmt.Errorf(
			"product id is required: %w",
			ErrInvalidArgument,
		)
	}

	imageURL, err := validateProductImageURL(
		input.ImageURL,
	)
	if err != nil {
		return domain.ProductImage{}, err
	}

	if err := s.validateVariant(
		ctx,
		input.ProductID,
		input.VariantID,
	); err != nil {
		return domain.ProductImage{}, err
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return domain.ProductImage{}, fmt.Errorf(
			"generate product image id: %w",
			err,
		)
	}

	image, err := s.repository.Create(
		ctx,
		repository.CreateProductImageParams{
			ID:        id,
			ProductID: input.ProductID,
			VariantID: input.VariantID,
			ImageURL:  imageURL,
			SortOrder: input.SortOrder,
		},
	)
	if err != nil {
		return domain.ProductImage{},
			mapRepositoryError(
				"create product image",
				err,
			)
	}

	return image, nil
}

func (s *ProductImageService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.ProductImage, error) {
	image, err := s.repository.GetByID(
		ctx,
		id,
	)
	if err != nil {
		return domain.ProductImage{},
			mapRepositoryError(
				"get product image",
				err,
			)
	}

	return image, nil
}

func (s *ProductImageService) List(
	ctx context.Context,
) ([]domain.ProductImage, error) {
	images, err := s.repository.List(ctx)
	if err != nil {
		return nil, mapRepositoryError(
			"list product images",
			err,
		)
	}

	return images, nil
}

func (s *ProductImageService) ListByProductID(
	ctx context.Context,
	productID uuid.UUID,
) ([]domain.ProductImage, error) {
	if productID == uuid.Nil {
		return nil, fmt.Errorf(
			"product id is required: %w",
			ErrInvalidArgument,
		)
	}

	images, err := s.repository.ListByProductID(
		ctx,
		productID,
	)
	if err != nil {
		return nil, mapRepositoryError(
			"list product images by product",
			err,
		)
	}

	return images, nil
}

func (s *ProductImageService) ListByVariantID(
	ctx context.Context,
	variantID uuid.UUID,
) ([]domain.ProductImage, error) {
	if variantID == uuid.Nil {
		return nil, fmt.Errorf(
			"variant id is required: %w",
			ErrInvalidArgument,
		)
	}

	images, err := s.repository.ListByVariantID(
		ctx,
		variantID,
	)
	if err != nil {
		return nil, mapRepositoryError(
			"list product images by variant",
			err,
		)
	}

	return images, nil
}

func (s *ProductImageService) Update(
	ctx context.Context,
	id uuid.UUID,
	input UpdateProductImageInput,
) (domain.ProductImage, error) {
	existing, err := s.repository.GetByID(
		ctx,
		id,
	)
	if err != nil {
		return domain.ProductImage{},
			mapRepositoryError(
				"get product image for update",
				err,
			)
	}

	imageURL, err := validateProductImageURL(
		input.ImageURL,
	)
	if err != nil {
		return domain.ProductImage{}, err
	}

	if err := s.validateVariant(
		ctx,
		existing.ProductID,
		input.VariantID,
	); err != nil {
		return domain.ProductImage{}, err
	}

	image, err := s.repository.Update(
		ctx,
		repository.UpdateProductImageParams{
			ID:        id,
			VariantID: input.VariantID,
			ImageURL:  imageURL,
			SortOrder: input.SortOrder,
		},
	)
	if err != nil {
		return domain.ProductImage{},
			mapRepositoryError(
				"update product image",
				err,
			)
	}

	return image, nil
}

func (s *ProductImageService) validateVariant(
	ctx context.Context,
	productID uuid.UUID,
	variantID *uuid.UUID,
) error {
	if variantID == nil {
		return nil
	}

	if *variantID == uuid.Nil {
		return fmt.Errorf(
			"variant id is invalid: %w",
			ErrInvalidArgument,
		)
	}

	belongs, err :=
		s.repository.VariantBelongsToProduct(
			ctx,
			productID,
			*variantID,
		)
	if err != nil {
		return mapRepositoryError(
			"check product image variant",
			err,
		)
	}

	if !belongs {
		return fmt.Errorf(
			"variant does not belong to product: %w",
			ErrInvalidReference,
		)
	}

	return nil
}

func validateProductImageURL(
	value string,
) (string, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return "", fmt.Errorf(
			"image url is required: %w",
			ErrInvalidArgument,
		)
	}

	parsed, err := url.ParseRequestURI(value)
	if err != nil {
		return "", fmt.Errorf(
			"image url is invalid: %w",
			ErrInvalidArgument,
		)
	}

	if parsed.Scheme != "http" &&
		parsed.Scheme != "https" {
		return "", fmt.Errorf(
			"image url must use http or https: %w",
			ErrInvalidArgument,
		)
	}

	if parsed.Host == "" {
		return "", fmt.Errorf(
			"image url host is required: %w",
			ErrInvalidArgument,
		)
	}

	return value, nil
}
