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
	maxPriceListNameLength = 100
	maxPriceListCodeLength = 50
	currencyCodeLength     = 3
)

type CreatePriceListInput struct {
	Name         string
	Code         string
	CurrencyCode string
}

type UpdatePriceListInput struct {
	Name         string
	Code         string
	CurrencyCode string
}

type PriceListService struct {
	repository repository.PriceListRepository
}

func NewPriceListService(
	repository repository.PriceListRepository,
) *PriceListService {
	return &PriceListService{
		repository: repository,
	}
}

func (s *PriceListService) Create(
	ctx context.Context,
	input CreatePriceListInput,
) (domain.PriceList, error) {
	name, code, currencyCode, err :=
		validatePriceListFields(
			input.Name,
			input.Code,
			input.CurrencyCode,
		)
	if err != nil {
		return domain.PriceList{}, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return domain.PriceList{}, fmt.Errorf(
			"generate price list id: %w",
			err,
		)
	}

	priceList, err := s.repository.Create(
		ctx,
		repository.CreatePriceListParams{
			ID:           id,
			Name:         name,
			Code:         code,
			CurrencyCode: currencyCode,
			IsActive:     true,
		},
	)
	if err != nil {
		return domain.PriceList{},
			mapRepositoryError(
				"create price list",
				err,
			)
	}

	return priceList, nil
}

func (s *PriceListService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.PriceList, error) {
	priceList, err := s.repository.GetByID(
		ctx,
		id,
	)
	if err != nil {
		return domain.PriceList{},
			mapRepositoryError(
				"get price list",
				err,
			)
	}

	return priceList, nil
}

func (s *PriceListService) GetByCode(
	ctx context.Context,
	code string,
) (domain.PriceList, error) {
	code = strings.TrimSpace(code)

	if code == "" {
		return domain.PriceList{}, fmt.Errorf(
			"price list code is required: %w",
			ErrInvalidArgument,
		)
	}

	if utf8.RuneCountInString(code) >
		maxPriceListCodeLength {
		return domain.PriceList{}, fmt.Errorf(
			"price list code exceeds %d characters: %w",
			maxPriceListCodeLength,
			ErrInvalidArgument,
		)
	}

	priceList, err := s.repository.GetByCode(
		ctx,
		code,
	)
	if err != nil {
		return domain.PriceList{},
			mapRepositoryError(
				"get price list by code",
				err,
			)
	}

	return priceList, nil
}

func (s *PriceListService) List(
	ctx context.Context,
) ([]domain.PriceList, error) {
	priceLists, err := s.repository.List(ctx)
	if err != nil {
		return nil, mapRepositoryError(
			"list price lists",
			err,
		)
	}

	return priceLists, nil
}

func (s *PriceListService) Update(
	ctx context.Context,
	id uuid.UUID,
	input UpdatePriceListInput,
) (domain.PriceList, error) {
	name, code, currencyCode, err :=
		validatePriceListFields(
			input.Name,
			input.Code,
			input.CurrencyCode,
		)
	if err != nil {
		return domain.PriceList{}, err
	}

	priceList, err := s.repository.Update(
		ctx,
		repository.UpdatePriceListParams{
			ID:           id,
			Name:         name,
			Code:         code,
			CurrencyCode: currencyCode,
		},
	)
	if err != nil {
		return domain.PriceList{},
			mapRepositoryError(
				"update price list",
				err,
			)
	}

	return priceList, nil
}

func (s *PriceListService) SetActive(
	ctx context.Context,
	id uuid.UUID,
	isActive bool,
) (domain.PriceList, error) {
	priceList, err := s.repository.SetActive(
		ctx,
		id,
		isActive,
	)
	if err != nil {
		return domain.PriceList{},
			mapRepositoryError(
				"set price list active",
				err,
			)
	}

	return priceList, nil
}

func validatePriceListFields(
	name string,
	code string,
	currencyCode string,
) (string, string, string, error) {
	name = strings.TrimSpace(name)
	code = strings.TrimSpace(code)
	currencyCode = strings.TrimSpace(currencyCode)

	if name == "" {
		return "", "", "", fmt.Errorf(
			"price list name is required: %w",
			ErrInvalidArgument,
		)
	}

	if utf8.RuneCountInString(name) >
		maxPriceListNameLength {
		return "", "", "", fmt.Errorf(
			"price list name exceeds %d characters: %w",
			maxPriceListNameLength,
			ErrInvalidArgument,
		)
	}

	if code == "" {
		return "", "", "", fmt.Errorf(
			"price list code is required: %w",
			ErrInvalidArgument,
		)
	}

	if utf8.RuneCountInString(code) >
		maxPriceListCodeLength {
		return "", "", "", fmt.Errorf(
			"price list code exceeds %d characters: %w",
			maxPriceListCodeLength,
			ErrInvalidArgument,
		)
	}

	if utf8.RuneCountInString(currencyCode) !=
		currencyCodeLength {
		return "", "", "", fmt.Errorf(
			"currency code must be exactly %d characters: %w",
			currencyCodeLength,
			ErrInvalidArgument,
		)
	}

	return name, code, currencyCode, nil
}
