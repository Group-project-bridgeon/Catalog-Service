package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/pagination"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/google/uuid"
)

const maxProductVariantSKULength = 100

var productVariantPricePattern = regexp.MustCompile(
	`^[0-9]{1,12}(\.[0-9]{1,2})?$`,
)

type ListProductVariantsInput struct {
	Page      PageInput
	ProductID *uuid.UUID
	UnitID    *uuid.UUID
	IsActive  *bool
}

type productVariantCursor struct {
	ProductID uuid.UUID `json:"product_id"`
	SKU       string    `json:"sku"`
	ID        uuid.UUID `json:"id"`
}

type CreateProductVariantInput struct {
	ProductID           uuid.UUID
	SKU                 string
	DefaultSellingPrice *string
	UnitID              uuid.UUID
}

type UpdateProductVariantInput struct {
	SKU                 string
	DefaultSellingPrice *string
}

type ProductVariantService struct {
	repository repository.ProductVariantRepository
}

func NewProductVariantService(
	repository repository.ProductVariantRepository,
) *ProductVariantService {
	return &ProductVariantService{
		repository: repository,
	}
}

func (s *ProductVariantService) Create(
	ctx context.Context,
	input CreateProductVariantInput,
) (domain.ProductVariant, error) {
	if input.ProductID == uuid.Nil {
		return domain.ProductVariant{}, fmt.Errorf(
			"product id is required: %w",
			ErrInvalidArgument,
		)
	}

	if input.UnitID == uuid.Nil {
		return domain.ProductVariant{}, fmt.Errorf(
			"unit id is required: %w",
			ErrInvalidArgument,
		)
	}

	sku, err := validateProductVariantSKU(input.SKU)
	if err != nil {
		return domain.ProductVariant{}, err
	}

	price, err := validateProductVariantPrice(
		input.DefaultSellingPrice,
	)
	if err != nil {
		return domain.ProductVariant{}, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return domain.ProductVariant{}, fmt.Errorf(
			"generate product variant id: %w",
			err,
		)
	}

	variant, err := s.repository.Create(
		ctx,
		repository.CreateProductVariantParams{
			ID:                  id,
			ProductID:           input.ProductID,
			SKU:                 sku,
			DefaultSellingPrice: price,
			UnitID:              input.UnitID,
			IsActive:            true,
		},
	)
	if err != nil {
		return domain.ProductVariant{},
			mapRepositoryError(
				"create product variant",
				err,
			)
	}

	return variant, nil
}

func (s *ProductVariantService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.ProductVariant, error) {
	variant, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return domain.ProductVariant{},
			mapRepositoryError(
				"get product variant",
				err,
			)
	}

	return variant, nil
}

func (s *ProductVariantService) GetByProductAndSKU(
	ctx context.Context,
	productID uuid.UUID,
	sku string,
) (domain.ProductVariant, error) {
	if productID == uuid.Nil {
		return domain.ProductVariant{}, fmt.Errorf(
			"product id is required: %w",
			ErrInvalidArgument,
		)
	}

	sku, err := validateProductVariantSKU(sku)
	if err != nil {
		return domain.ProductVariant{}, err
	}

	variant, err := s.repository.GetByProductAndSKU(
		ctx,
		productID,
		sku,
	)
	if err != nil {
		return domain.ProductVariant{},
			mapRepositoryError(
				"get product variant by product and sku",
				err,
			)
	}

	return variant, nil
}

func (s *ProductVariantService) List(
	ctx context.Context,
	input ListProductVariantsInput,
) (Page[domain.ProductVariant], error) {
	page, err := normalizePageInput(input.Page)
	if err != nil {
		return Page[domain.ProductVariant]{}, fmt.Errorf(
			"invalid pagination: %w",
			ErrInvalidArgument,
		)
	}

	var cursor productVariantCursor

	if page.Cursor != "" {
		if err := pagination.Decode(
			page.Cursor,
			&cursor,
		); err != nil {
			return Page[domain.ProductVariant]{}, fmt.Errorf(
				"invalid cursor: %w",
				ErrInvalidArgument,
			)
		}
	}

	var cursorProductID *uuid.UUID
	var cursorSKU *string
	var cursorID *uuid.UUID

	if page.Cursor != "" {
		cursorProductID = &cursor.ProductID
		cursorSKU = &cursor.SKU
		cursorID = &cursor.ID
	}

	variants, err := s.repository.List(
		ctx,
		repository.ListProductVariantsParams{
			CursorProductID: cursorProductID,
			CursorSKU:       cursorSKU,
			CursorID:        cursorID,
			ProductID:       input.ProductID,
			UnitID:          input.UnitID,
			IsActive:        input.IsActive,
			PageSize:        int32(page.Limit + 1),
		},
	)
	if err != nil {
		return Page[domain.ProductVariant]{},
			mapRepositoryError(
				"list product variants",
				err,
			)
	}

	variants, hasMore := trimPage(
		variants,
		page.Limit,
	)

	var nextCursor *string

	if hasMore && len(variants) > 0 {
		last := variants[len(variants)-1]

		encoded, err := pagination.Encode(
			productVariantCursor{
				ProductID: last.ProductID,
				SKU:       last.SKU,
				ID:        last.ID,
			},
		)
		if err != nil {
			return Page[domain.ProductVariant]{},
				fmt.Errorf(
					"encode product variant cursor: %w",
					err,
				)
		}

		nextCursor = &encoded
	}

	return Page[domain.ProductVariant]{
		Items:      variants,
		NextCursor: nextCursor,
	}, nil
}

func (s *ProductVariantService) ListByProductID(
	ctx context.Context,
	productID uuid.UUID,
	page PageInput,
) (Page[domain.ProductVariant], error) {
	if productID == uuid.Nil {
		return Page[domain.ProductVariant]{},
			fmt.Errorf(
				"product id is required: %w",
				ErrInvalidArgument,
			)
	}

	return s.List(
		ctx,
		ListProductVariantsInput{
			Page:      page,
			ProductID: &productID,
		},
	)
}

func (s *ProductVariantService) ListByIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]domain.ProductVariant, error) {
	if len(ids) == 0 {
		return []domain.ProductVariant{}, nil
	}

	if len(ids) > 100 {
		return nil, fmt.Errorf(
			"maximum 100 variant ids are allowed: %w",
			ErrInvalidArgument,
		)
	}

	for _, id := range ids {
		if id == uuid.Nil {
			return nil, fmt.Errorf(
				"invalid variant id: %w",
				ErrInvalidArgument,
			)
		}
	}

	variants, err := s.repository.ListByIDs(
		ctx,
		ids,
	)
	if err != nil {
		return nil, mapRepositoryError(
			"list product variants by ids",
			err,
		)
	}

	return variants, nil
}

func (s *ProductVariantService) Update(
	ctx context.Context,
	id uuid.UUID,
	input UpdateProductVariantInput,
) (domain.ProductVariant, error) {
	sku, err := validateProductVariantSKU(input.SKU)
	if err != nil {
		return domain.ProductVariant{}, err
	}

	price, err := validateProductVariantPrice(
		input.DefaultSellingPrice,
	)
	if err != nil {
		return domain.ProductVariant{}, err
	}

	variant, err := s.repository.Update(
		ctx,
		repository.UpdateProductVariantParams{
			ID:                  id,
			SKU:                 sku,
			DefaultSellingPrice: price,
		},
	)
	if err != nil {
		return domain.ProductVariant{},
			mapRepositoryError(
				"update product variant",
				err,
			)
	}

	return variant, nil
}

func (s *ProductVariantService) SetActive(
	ctx context.Context,
	id uuid.UUID,
	isActive bool,
) (domain.ProductVariant, error) {
	variant, err := s.repository.SetActive(
		ctx,
		id,
		isActive,
	)
	if err != nil {
		return domain.ProductVariant{},
			mapRepositoryError(
				"set product variant active",
				err,
			)
	}

	return variant, nil
}

func validateProductVariantSKU(
	sku string,
) (string, error) {
	sku = strings.TrimSpace(sku)

	if sku == "" {
		return "", fmt.Errorf(
			"sku is required: %w",
			ErrInvalidArgument,
		)
	}

	if utf8.RuneCountInString(sku) >
		maxProductVariantSKULength {
		return "", fmt.Errorf(
			"sku exceeds %d characters: %w",
			maxProductVariantSKULength,
			ErrInvalidArgument,
		)
	}

	return sku, nil
}

func validateProductVariantPrice(
	price *string,
) (*string, error) {
	if price == nil {
		return nil, nil
	}

	value := strings.TrimSpace(*price)

	if value == "" {
		return nil, fmt.Errorf(
			"default selling price is invalid: %w",
			ErrInvalidArgument,
		)
	}

	if !productVariantPricePattern.MatchString(value) {
		return nil, fmt.Errorf(
			"default selling price must be a non-negative decimal with at most 12 integer digits and 2 decimal places: %w",
			ErrInvalidArgument,
		)
	}

	return &value, nil
}
