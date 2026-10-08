package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/repository"
	"github.com/google/uuid"
)

type fakePriceRepository struct {
	createFn func(
		context.Context,
		repository.CreatePriceParams,
	) (domain.Price, error)

	getByIDFn func(
		context.Context,
		uuid.UUID,
	) (domain.Price, error)

	getByPairFn func(
		context.Context,
		uuid.UUID,
		uuid.UUID,
	) (domain.Price, error)

	listFn func(
		context.Context,
		repository.ListPricesParams,
	) ([]domain.Price, error)

	updateFn func(
		context.Context,
		repository.UpdatePriceParams,
	) (domain.Price, error)
}

func (f *fakePriceRepository) Create(
	ctx context.Context,
	params repository.CreatePriceParams,
) (domain.Price, error) {
	if f.createFn == nil {
		return domain.Price{}, nil
	}

	return f.createFn(ctx, params)
}

func (f *fakePriceRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Price, error) {
	if f.getByIDFn == nil {
		return domain.Price{}, nil
	}

	return f.getByIDFn(ctx, id)
}

func (f *fakePriceRepository) GetByPriceListAndVariant(
	ctx context.Context,
	priceListID uuid.UUID,
	variantID uuid.UUID,
) (domain.Price, error) {
	if f.getByPairFn == nil {
		return domain.Price{}, nil
	}

	return f.getByPairFn(
		ctx,
		priceListID,
		variantID,
	)
}

func (f *fakePriceRepository) List(
	ctx context.Context,
	params repository.ListPricesParams,
) ([]domain.Price, error) {
	if f.listFn == nil {
		return nil, nil
	}

	return f.listFn(ctx, params)
}

func (f *fakePriceRepository) Update(
	ctx context.Context,
	params repository.UpdatePriceParams,
) (domain.Price, error) {
	if f.updateFn == nil {
		return domain.Price{}, nil
	}

	return f.updateFn(ctx, params)
}

func TestPriceServiceCreate(
	t *testing.T,
) {
	priceListID := uuid.New()
	variantID := uuid.New()

	repo := &fakePriceRepository{
		createFn: func(
			_ context.Context,
			params repository.CreatePriceParams,
		) (domain.Price, error) {
			if params.ID == uuid.Nil {
				t.Fatal("expected generated ID")
			}

			if params.PriceListID != priceListID {
				t.Fatal("unexpected price list ID")
			}

			if params.VariantID != variantID {
				t.Fatal("unexpected variant ID")
			}

			if params.UnitPrice != "850.50" {
				t.Fatalf(
					"unexpected unit price: %q",
					params.UnitPrice,
				)
			}

			return domain.Price{
				ID:          params.ID,
				PriceListID: params.PriceListID,
				VariantID:   params.VariantID,
				UnitPrice:   params.UnitPrice,
			}, nil
		},
	}

	s := NewPriceService(repo)

	price, err := s.Create(
		context.Background(),
		CreatePriceInput{
			PriceListID: priceListID,
			VariantID:   variantID,
			UnitPrice:   " 850.50 ",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if price.UnitPrice != "850.50" {
		t.Fatal("unexpected unit price")
	}
}

func TestPriceServiceCreateValidation(
	t *testing.T,
) {
	priceListID := uuid.New()
	variantID := uuid.New()

	tests := []CreatePriceInput{
		{
			VariantID: variantID,
			UnitPrice: "100",
		},
		{
			PriceListID: priceListID,
			UnitPrice:   "100",
		},
		{
			PriceListID: priceListID,
			VariantID:   variantID,
			UnitPrice:   "",
		},
		{
			PriceListID: priceListID,
			VariantID:   variantID,
			UnitPrice:   "-1.00",
		},
		{
			PriceListID: priceListID,
			VariantID:   variantID,
			UnitPrice:   "10.123",
		},
		{
			PriceListID: priceListID,
			VariantID:   variantID,
			UnitPrice:   "1234567890123.00",
		},
	}

	s := NewPriceService(
		&fakePriceRepository{},
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

func TestPriceServiceCreateConflict(
	t *testing.T,
) {
	repo := &fakePriceRepository{
		createFn: func(
			context.Context,
			repository.CreatePriceParams,
		) (domain.Price, error) {
			return domain.Price{},
				repository.ErrConflict
		},
	}

	s := NewPriceService(repo)

	_, err := s.Create(
		context.Background(),
		CreatePriceInput{
			PriceListID: uuid.New(),
			VariantID:   uuid.New(),
			UnitPrice:   "850.00",
		},
	)

	if !errors.Is(err, ErrConflict) {
		t.Fatalf(
			"expected ErrConflict, got %v",
			err,
		)
	}
}

func TestPriceServiceInvalidReference(
	t *testing.T,
) {
	repo := &fakePriceRepository{
		createFn: func(
			context.Context,
			repository.CreatePriceParams,
		) (domain.Price, error) {
			return domain.Price{},
				repository.ErrInvalidReference
		},
	}

	s := NewPriceService(repo)

	_, err := s.Create(
		context.Background(),
		CreatePriceInput{
			PriceListID: uuid.New(),
			VariantID:   uuid.New(),
			UnitPrice:   "850.00",
		},
	)

	if !errors.Is(
		err,
		ErrInvalidReference,
	) {
		t.Fatalf(
			"expected ErrInvalidReference, got %v",
			err,
		)
	}
}

func TestPriceServiceGetByIDNotFound(
	t *testing.T,
) {
	repo := &fakePriceRepository{
		getByIDFn: func(
			context.Context,
			uuid.UUID,
		) (domain.Price, error) {
			return domain.Price{},
				repository.ErrNotFound
		},
	}

	s := NewPriceService(repo)

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

func TestPriceServiceListByPriceList(
	t *testing.T,
) {
	priceListID := uuid.New()

	repo := &fakePriceRepository{
		listFn: func(
			_ context.Context,
			params repository.ListPricesParams,
		) ([]domain.Price, error) {
			if params.PriceListID == nil || *params.PriceListID != priceListID {
				t.Fatal("unexpected price list ID")
			}

			return []domain.Price{
				{
					ID:          uuid.New(),
					PriceListID: priceListID,
					VariantID:   uuid.New(),
					UnitPrice:   "100.00",
				},
			}, nil
		},
	}

	s := NewPriceService(repo)

	page, err := s.ListByPriceListID(
		context.Background(),
		priceListID,
		PageInput{Limit: 10},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(page.Items) != 1 {
		t.Fatalf(
			"expected 1 price, got %d",
			len(page.Items),
		)
	}
}

func TestPriceServiceUpdate(
	t *testing.T,
) {
	id := uuid.New()

	repo := &fakePriceRepository{
		updateFn: func(
			_ context.Context,
			params repository.UpdatePriceParams,
		) (domain.Price, error) {
			if params.ID != id {
				t.Fatal("unexpected ID")
			}

			if params.UnitPrice != "999.75" {
				t.Fatalf(
					"unexpected unit price: %q",
					params.UnitPrice,
				)
			}

			return domain.Price{
				ID:        params.ID,
				UnitPrice: params.UnitPrice,
			}, nil
		},
	}

	s := NewPriceService(repo)

	price, err := s.Update(
		context.Background(),
		id,
		UpdatePriceInput{
			UnitPrice: " 999.75 ",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if price.UnitPrice != "999.75" {
		t.Fatal("unexpected unit price")
	}
}
