package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/google/uuid"
)

var unitPricePattern = regexp.MustCompile(
	`^[0-9]{1,12}(\.[0-9]{1,2})?$`,
)

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

	id, err := uuid.NewRandom()
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
) ([]domain.Price, error) {
	prices, err := s.repository.List(ctx)
	if err != nil {
		return nil, mapRepositoryError(
			"list prices",
			err,
		)
	}

	return prices, nil
}

func (s *PriceService) ListByPriceListID(
	ctx context.Context,
	priceListID uuid.UUID,
) ([]domain.Price, error) {
	if priceListID == uuid.Nil {
		return nil, fmt.Errorf(
			"price list id is required: %w",
			ErrInvalidArgument,
		)
	}

	prices, err := s.repository.ListByPriceListID(
		ctx,
		priceListID,
	)
	if err != nil {
		return nil, mapRepositoryError(
			"list prices by price list",
			err,
		)
	}

	return prices, nil
}

func (s *PriceService) ListByVariantID(
	ctx context.Context,
	variantID uuid.UUID,
) ([]domain.Price, error) {
	if variantID == uuid.Nil {
		return nil, fmt.Errorf(
			"variant id is required: %w",
			ErrInvalidArgument,
		)
	}

	prices, err := s.repository.ListByVariantID(
		ctx,
		variantID,
	)
	if err != nil {
		return nil, mapRepositoryError(
			"list prices by variant",
			err,
		)
	}

	return prices, nil
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
