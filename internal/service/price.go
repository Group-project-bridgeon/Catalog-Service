package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/pagination"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/google/uuid"
)

var unitPricePattern = regexp.MustCompile(
	`^[0-9]{1,12}(\.[0-9]{1,2})?$`,
)

type ListPricesInput struct {
	Page        PageInput
	PriceListID *uuid.UUID
	VariantID   *uuid.UUID
}

type priceCursor struct {
	PriceListID uuid.UUID `json:"price_list_id"`
	VariantID   uuid.UUID `json:"variant_id"`
}

type CreatePriceInput struct {
	PriceListID uuid.UUID
	VariantID   uuid.UUID
	UnitPrice   string
}

type UpdatePriceInput struct {
	UnitPrice string
}

type PriceService struct {
	repository repository.PriceRepository
}

func NewPriceService(
	repository repository.PriceRepository,
) *PriceService {
	return &PriceService{
		repository: repository,
	}
}

func (s *PriceService) Create(
	ctx context.Context,
	input CreatePriceInput,
) (domain.Price, error) {
	if input.PriceListID == uuid.Nil {
		return domain.Price{}, fmt.Errorf(
			"price list id is required: %w",
			ErrInvalidArgument,
		)
	}

	if input.VariantID == uuid.Nil {
		return domain.Price{}, fmt.Errorf(
			"variant id is required: %w",
			ErrInvalidArgument,
		)
	}

	unitPrice, err := validateUnitPrice(
		input.UnitPrice,
	)
	if err != nil {
		return domain.Price{}, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return domain.Price{}, fmt.Errorf(
			"generate price id: %w",
			err,
		)
	}

	price, err := s.repository.Create(
		ctx,
		repository.CreatePriceParams{
			ID:          id,
			PriceListID: input.PriceListID,
			VariantID:   input.VariantID,
			UnitPrice:   unitPrice,
		},
	)
	if err != nil {
		return domain.Price{},
			mapRepositoryError(
				"create price",
				err,
			)
	}

	return price, nil
}

func (s *PriceService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Price, error) {
	price, err := s.repository.GetByID(
		ctx,
		id,
	)
	if err != nil {
		return domain.Price{},
			mapRepositoryError(
				"get price",
				err,
			)
	}

	return price, nil
}

func (s *PriceService) GetByPriceListAndVariant(
	ctx context.Context,
	priceListID uuid.UUID,
	variantID uuid.UUID,
) (domain.Price, error) {
	if priceListID == uuid.Nil {
		return domain.Price{}, fmt.Errorf(
			"price list id is required: %w",
			ErrInvalidArgument,
		)
	}

	if variantID == uuid.Nil {
		return domain.Price{}, fmt.Errorf(
			"variant id is required: %w",
			ErrInvalidArgument,
		)
	}

	price, err :=
		s.repository.GetByPriceListAndVariant(
			ctx,
			priceListID,
			variantID,
		)
	if err != nil {
		return domain.Price{},
			mapRepositoryError(
				"get price by price list and variant",
				err,
			)
	}

	return price, nil
}

func (s *PriceService) List(
	ctx context.Context,
	input ListPricesInput,
) (Page[domain.Price], error) {
	page, err := normalizePageInput(input.Page)
	if err != nil {
		return Page[domain.Price]{}, fmt.Errorf(
			"invalid pagination: %w",
			ErrInvalidArgument,
		)
	}

	var cursor priceCursor

	if page.Cursor != "" {
		if err := pagination.Decode(
			page.Cursor,
			&cursor,
		); err != nil {
			return Page[domain.Price]{}, fmt.Errorf(
				"invalid cursor: %w",
				ErrInvalidArgument,
			)
		}
	}

	var cursorPriceListID *uuid.UUID
	var cursorVariantID *uuid.UUID

	if page.Cursor != "" {
		cursorPriceListID = &cursor.PriceListID
		cursorVariantID = &cursor.VariantID
	}

	prices, err := s.repository.List(
		ctx,
		repository.ListPricesParams{
			CursorPriceListID: cursorPriceListID,
			CursorVariantID:   cursorVariantID,
			PriceListID:       input.PriceListID,
			VariantID:         input.VariantID,
			PageSize:          int32(page.Limit + 1),
		},
	)
	if err != nil {
		return Page[domain.Price]{},
			mapRepositoryError(
				"list prices",
				err,
			)
	}

	prices, hasMore := trimPage(
		prices,
		page.Limit,
	)

	var nextCursor *string

	if hasMore && len(prices) > 0 {
		last := prices[len(prices)-1]

		encoded, err := pagination.Encode(
			priceCursor{
				PriceListID: last.PriceListID,
				VariantID:   last.VariantID,
			},
		)
		if err != nil {
			return Page[domain.Price]{},
				fmt.Errorf(
					"encode price cursor: %w",
					err,
				)
		}

		nextCursor = &encoded
	}

	return Page[domain.Price]{
		Items:      prices,
		NextCursor: nextCursor,
	}, nil
}

func (s *PriceService) ListByPriceListID(
	ctx context.Context,
	priceListID uuid.UUID,
	page PageInput,
) (Page[domain.Price], error) {
	if priceListID == uuid.Nil {
		return Page[domain.Price]{}, fmt.Errorf(
			"price list id is required: %w",
			ErrInvalidArgument,
		)
	}

	return s.List(
		ctx,
		ListPricesInput{
			Page:        page,
			PriceListID: &priceListID,
		},
	)
}

func (s *PriceService) ListByVariantID(
	ctx context.Context,
	variantID uuid.UUID,
	page PageInput,
) (Page[domain.Price], error) {
	if variantID == uuid.Nil {
		return Page[domain.Price]{}, fmt.Errorf(
			"variant id is required: %w",
			ErrInvalidArgument,
		)
	}

	return s.List(
		ctx,
		ListPricesInput{
			Page:      page,
			VariantID: &variantID,
		},
	)
}

func (s *PriceService) Update(
	ctx context.Context,
	id uuid.UUID,
	input UpdatePriceInput,
) (domain.Price, error) {
	unitPrice, err := validateUnitPrice(
		input.UnitPrice,
	)
	if err != nil {
		return domain.Price{}, err
	}

	price, err := s.repository.Update(
		ctx,
		repository.UpdatePriceParams{
			ID:        id,
			UnitPrice: unitPrice,
		},
	)
	if err != nil {
		return domain.Price{},
			mapRepositoryError(
				"update price",
				err,
			)
	}

	return price, nil
}

func validateUnitPrice(
	value string,
) (string, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return "", fmt.Errorf(
			"unit price is required: %w",
			ErrInvalidArgument,
		)
	}

	if !unitPricePattern.MatchString(value) {
		return "", fmt.Errorf(
			"unit price must be a non-negative decimal with at most 12 integer digits and 2 decimal places: %w",
			ErrInvalidArgument,
		)
	}

	return value, nil
}
