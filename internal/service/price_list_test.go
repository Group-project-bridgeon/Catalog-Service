package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/google/uuid"
)

type fakePriceListRepository struct {
	createFn func(
		context.Context,
		repository.CreatePriceListParams,
	) (domain.PriceList, error)

	getByIDFn func(
		context.Context,
		uuid.UUID,
	) (domain.PriceList, error)

	getByCodeFn func(
		context.Context,
		string,
	) (domain.PriceList, error)

	listFn func(
		context.Context,
	) ([]domain.PriceList, error)

	updateFn func(
		context.Context,
		repository.UpdatePriceListParams,
	) (domain.PriceList, error)

	setActiveFn func(
		context.Context,
		uuid.UUID,
		bool,
	) (domain.PriceList, error)
}

func (f *fakePriceListRepository) Create(
	ctx context.Context,
	params repository.CreatePriceListParams,
) (domain.PriceList, error) {
	if f.createFn == nil {
		return domain.PriceList{}, nil
	}

	return f.createFn(ctx, params)
}

func (f *fakePriceListRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.PriceList, error) {
	if f.getByIDFn == nil {
		return domain.PriceList{}, nil
	}

	return f.getByIDFn(ctx, id)
}

func (f *fakePriceListRepository) GetByCode(
	ctx context.Context,
	code string,
) (domain.PriceList, error) {
	if f.getByCodeFn == nil {
		return domain.PriceList{}, nil
	}

	return f.getByCodeFn(ctx, code)
}

func (f *fakePriceListRepository) List(
	ctx context.Context,
) ([]domain.PriceList, error) {
	if f.listFn == nil {
		return nil, nil
	}

	return f.listFn(ctx)
}

func (f *fakePriceListRepository) Update(
	ctx context.Context,
	params repository.UpdatePriceListParams,
) (domain.PriceList, error) {
	if f.updateFn == nil {
		return domain.PriceList{}, nil
	}

	return f.updateFn(ctx, params)
}

func (f *fakePriceListRepository) SetActive(
	ctx context.Context,
	id uuid.UUID,
	active bool,
) (domain.PriceList, error) {
	if f.setActiveFn == nil {
		return domain.PriceList{}, nil
	}

	return f.setActiveFn(
		ctx,
		id,
		active,
	)
}

func TestPriceListServiceCreate(
	t *testing.T,
) {
	repo := &fakePriceListRepository{
		createFn: func(
			_ context.Context,
			params repository.CreatePriceListParams,
		) (domain.PriceList, error) {
			if params.ID == uuid.Nil {
				t.Fatal("expected generated ID")
			}

			if params.Name != "Wholesale India" {
				t.Fatalf(
					"unexpected name: %q",
					params.Name,
				)
			}

			if params.Code != "WHOLESALE-IN" {
				t.Fatalf(
					"unexpected code: %q",
					params.Code,
				)
			}

			if params.CurrencyCode != "INR" {
				t.Fatalf(
					"unexpected currency code: %q",
					params.CurrencyCode,
				)
			}

			if !params.IsActive {
				t.Fatal("new price list should be active")
			}

			return domain.PriceList{
				ID:           params.ID,
				Name:         params.Name,
				Code:         params.Code,
				CurrencyCode: params.CurrencyCode,
				IsActive:     params.IsActive,
			}, nil
		},
	}

	s := NewPriceListService(repo)

	result, err := s.Create(
		context.Background(),
		CreatePriceListInput{
			Name:         " Wholesale India ",
			Code:         " WHOLESALE-IN ",
			CurrencyCode: " INR ",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.CurrencyCode != "INR" {
		t.Fatal("unexpected currency code")
	}
}

func TestPriceListServiceCreateValidation(
	t *testing.T,
) {
	tests := []CreatePriceListInput{
		{
			Code:         "WHOLESALE",
			CurrencyCode: "INR",
		},
		{
			Name:         strings.Repeat("a", 101),
			Code:         "WHOLESALE",
			CurrencyCode: "INR",
		},
		{
			Name:         "Wholesale",
			CurrencyCode: "INR",
		},
		{
			Name: "Wholesale",
			Code: strings.Repeat(
				"a",
				51,
			),
			CurrencyCode: "INR",
		},
		{
			Name:         "Wholesale",
			Code:         "WHOLESALE",
			CurrencyCode: "IN",
		},
		{
			Name:         "Wholesale",
			Code:         "WHOLESALE",
			CurrencyCode: "INRR",
		},
	}

	s := NewPriceListService(
		&fakePriceListRepository{},
	)

	for index, input := range tests {
		_, err := s.Create(
			context.Background(),
			input,
		)

		if !errors.Is(
			err,
			ErrInvalidArgument,
		) {
			t.Fatalf(
				"case %d: expected ErrInvalidArgument, got %v",
				index,
				err,
			)
		}
	}
}

func TestPriceListServiceConflict(
	t *testing.T,
) {
	repo := &fakePriceListRepository{
		createFn: func(
			context.Context,
			repository.CreatePriceListParams,
		) (domain.PriceList, error) {
			return domain.PriceList{},
				repository.ErrConflict
		},
	}

	s := NewPriceListService(repo)

	_, err := s.Create(
		context.Background(),
		CreatePriceListInput{
			Name:         "Wholesale",
			Code:         "WHOLESALE",
			CurrencyCode: "INR",
		},
	)

	if !errors.Is(err, ErrConflict) {
		t.Fatalf(
			"expected ErrConflict, got %v",
			err,
		)
	}
}

func TestPriceListServiceGetByIDNotFound(
	t *testing.T,
) {
	repo := &fakePriceListRepository{
		getByIDFn: func(
			context.Context,
			uuid.UUID,
		) (domain.PriceList, error) {
			return domain.PriceList{},
				repository.ErrNotFound
		},
	}

	s := NewPriceListService(repo)

	_, err := s.GetByID(
		context.Background(),
		uuid.New(),
	)

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf(
			"expected ErrNotFound, got %v",
			err,
		)
	}
}

func TestPriceListServiceGetByCode(
	t *testing.T,
) {
	repo := &fakePriceListRepository{
		getByCodeFn: func(
			_ context.Context,
			code string,
		) (domain.PriceList, error) {
			if code != "WHOLESALE" {
				t.Fatalf(
					"unexpected code: %q",
					code,
				)
			}

			return domain.PriceList{
				Code: code,
			}, nil
		},
	}

	s := NewPriceListService(repo)

	_, err := s.GetByCode(
		context.Background(),
		" WHOLESALE ",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPriceListServiceUpdate(
	t *testing.T,
) {
	id := uuid.New()

	repo := &fakePriceListRepository{
		updateFn: func(
			_ context.Context,
			params repository.UpdatePriceListParams,
		) (domain.PriceList, error) {
			if params.ID != id {
				t.Fatal("unexpected ID")
			}

			return domain.PriceList{
				ID:           params.ID,
				Name:         params.Name,
				Code:         params.Code,
				CurrencyCode: params.CurrencyCode,
				IsActive:     true,
			}, nil
		},
	}

	s := NewPriceListService(repo)

	result, err := s.Update(
		context.Background(),
		id,
		UpdatePriceListInput{
			Name:         " Updated Wholesale ",
			Code:         " WHOLESALE-UPDATED ",
			CurrencyCode: " USD ",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.CurrencyCode != "USD" {
		t.Fatal("unexpected currency code")
	}
}

func TestPriceListServiceSetActive(
	t *testing.T,
) {
	id := uuid.New()

	repo := &fakePriceListRepository{
		setActiveFn: func(
			_ context.Context,
			gotID uuid.UUID,
			active bool,
		) (domain.PriceList, error) {
			return domain.PriceList{
				ID:       gotID,
				IsActive: active,
			}, nil
		},
	}

	s := NewPriceListService(repo)

	result, err := s.SetActive(
		context.Background(),
		id,
		false,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsActive {
		t.Fatal("expected inactive price list")
	}
}
